#!/bin/sh
set -eu
umask 077

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
machine_dir=$repo_root/machine/macbookpro12,1-alpine
env_file=$machine_dir/.env
output_root=${1:-$repo_root/build/macbookpro12,1}
scripts_dir=$repo_root/machine/macbookpro12,1-alpine/base/scripts

if [ ! -r "$env_file" ]; then
	printf 'Missing %s. Copy %s/.env.example to .env and fill in machine settings.\n' \
		"$env_file" "$machine_dir" >&2
	exit 1
fi

# The machine-local .env is a trusted, shell-compatible configuration file.
set -a
. "$env_file"
set +a

: "${WIFI_SSID:?Set WIFI_SSID in .env}"
: "${WIFI_PASSWORD:?Set WIFI_PASSWORD in .env}"
: "${MACBOOK_IPV4_CIDR:?Set MACBOOK_IPV4_CIDR in .env}"
: "${LAN_GATEWAY:?Set LAN_GATEWAY in .env}"
: "${MACBOOK_SSH_HOST:?Set MACBOOK_SSH_HOST in .env}"
: "${SSH_PRIVATE_KEY_FILE:?Set SSH_PRIVATE_KEY_FILE in .env}"
[ -r "$SSH_PRIVATE_KEY_FILE" ] || {
	printf 'SSH private key is not readable: %s\n' "$SSH_PRIVATE_KEY_FILE" >&2
	exit 1
}

stage=$output_root/base-payload
rm -rf "$stage"
mkdir -p "$stage/ALPINE" "$stage/overlay/etc/network" \
	"$stage/overlay/etc/wpa_supplicant" "$stage/overlay/etc/local.d" \
	"$stage/overlay/usr/local/bin"

for file in hardware-setup-launcher hardware-setup.start hardware-storage.sh hardware-test.sh \
	graphics-test stage-graphics-runtime.sh graphics-test.c; do
	[ -r "$scripts_dir/$file" ] || {
		printf 'Missing machine base script: %s\n' "$scripts_dir/$file" >&2
		exit 1
	}
	cp "$scripts_dir/$file" "$stage/ALPINE/$file"
done
chmod 755 "$stage/ALPINE/"*
ssh-keygen -y -f "$SSH_PRIVATE_KEY_FILE" > "$stage/ALPINE/ssh-authorized-key"
chmod 644 "$stage/ALPINE/ssh-authorized-key"
cp "$scripts_dir/hardware-setup-launcher" \
	"$stage/overlay/etc/local.d/hardware-setup.start"
chmod 755 "$stage/overlay/etc/local.d/hardware-setup.start"
cp "$scripts_dir/audio-utility" "$stage/overlay/usr/local/bin/audio-utility"
chmod 755 "$stage/overlay/usr/local/bin/audio-utility"

export STAGE_DIR=$stage
python3 - <<'PY'
import os
from pathlib import Path

stage = Path(os.environ["STAGE_DIR"])

def required(name):
    value = os.environ.get(name, "")
    if not value:
        raise SystemExit(f"Missing {name} in .env")
    return value

def quote(value):
    return '"' + value.replace("\\", "\\\\").replace('"', '\\"') + '"'

ssid = required("WIFI_SSID")
password = required("WIFI_PASSWORD")
address = required("MACBOOK_IPV4_CIDR")
gateway = required("LAN_GATEWAY")

if not 8 <= len(password) <= 63:
    raise SystemExit("WIFI_PASSWORD must contain 8 to 63 characters")

wpa = """ctrl_interface=/var/run/wpa_supplicant
update_config=0
country=US

network={
    ssid=%s
    key_mgmt=WPA-PSK
    psk=%s
}
""" % (quote(ssid), quote(password))

interfaces = """auto lo
iface lo
    use loopback

auto wlan0
iface wlan0
    address %s
    gateway %s
""" % (address, gateway)

(stage / "overlay/etc/wpa_supplicant/wpa_supplicant.conf").write_text(wpa)
(stage / "overlay/etc/network/interfaces").write_text(interfaces)
(stage / "overlay/etc/wpa_supplicant/wpa_supplicant.conf").chmod(0o600)
(stage / "overlay/etc/network/interfaces").chmod(0o644)
PY

cat > "$stage/README.txt" <<EOF
Machine base payload for MacBookPro12,1

ALPINE/ contains machine startup, report, graphics-test and package-staging
scripts plus the SSH public key. SDL3/Mesa runtime packages and the compiled
smoke-test binary are stored on the writable audio-utility data volume.
overlay/ contains generated network configuration. It was rendered from this
machine folder's .env and may include sensitive local network settings. It also
contains the manual /usr/local/bin/audio-utility launcher; it does not enable
the app to start automatically during boot.

Apply this payload to the existing Alpine base. Merge overlay/ into the Alpine
root and use lbu include/commit to persist it; do not replace the complete
localhost.apkovl.tar.gz with this partial configuration overlay. Run
stage-graphics-runtime.sh once on the MacBook after applying this payload to
place signed SDL3/Mesa packages on audio-utility. Keep the compiled smoke-test
at audio-utility/graphics-test/sdl3-kmsdrm-smoke. Startup loads the graphics
runtime into RAM and exposes the manual graphics-test command; it does not
start the visible triangle automatically.
EOF

mkdir -p "$output_root"
LC_ALL=C tar -czf "$output_root/macbookpro12,1-base-payload.tar.gz" -C "$stage" .
chmod 600 "$output_root/macbookpro12,1-base-payload.tar.gz"
printf 'Built machine base payload: %s\n' "$output_root/macbookpro12,1-base-payload.tar.gz"
printf 'Generated network configuration is under: %s/overlay\n' "$stage"
printf 'Treat the output as sensitive; it is excluded from Git.\n'
