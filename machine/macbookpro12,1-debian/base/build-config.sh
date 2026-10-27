#!/bin/sh
# Render the Debian Live startup payload from machine-local settings.
set -eu
umask 077

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
machine_dir=$repo_root/machine/macbookpro12,1-debian
env_file=$machine_dir/.env
output_root=${1:-$repo_root/build/macbookpro12,1-debian}

[ -r "$env_file" ] || {
	echo "Copy $machine_dir/.env.example to .env and fill in machine settings." >&2
	exit 1
}
set -a
. "$env_file"
set +a
: "${WIFI_SSID:?Set WIFI_SSID in .env}"
: "${WIFI_PASSWORD:?Set WIFI_PASSWORD in .env}"
: "${MACBOOK_IPV4_CIDR:?Set MACBOOK_IPV4_CIDR in .env}"
: "${LAN_GATEWAY:?Set LAN_GATEWAY in .env}"
: "${SSH_PRIVATE_KEY_FILE:?Set SSH_PRIVATE_KEY_FILE in .env}"
[ -r "$SSH_PRIVATE_KEY_FILE" ] || {
	echo "SSH private key is not readable: $SSH_PRIVATE_KEY_FILE" >&2
	exit 1
}
[ ${#WIFI_PASSWORD} -ge 8 ] && [ ${#WIFI_PASSWORD} -le 63 ] || {
	echo 'WIFI_PASSWORD must contain 8 to 63 characters.' >&2
	exit 1
}

cache=$output_root/packages
[ -r "$cache/SHA256SUMS" ] || {
	echo "Offline packages missing: $cache. Run stage-packages.sh in the local Debian preparation VM first." >&2
	exit 1
}
host_keys=$output_root/ssh-host-keys
mkdir -p "$host_keys"
if [ ! -f "$host_keys/ssh_host_ed25519_key" ]; then
	ssh-keygen -q -t ed25519 -N '' -C audiovis-util-debian -f "$host_keys/ssh_host_ed25519_key"
fi
ssh-keygen -y -f "$host_keys/ssh_host_ed25519_key" |
	awk '{print "macbookpro12-1-debian " $0}' > "$output_root/known_hosts"
stage=$output_root/card-payload
rm -rf "$stage"
mkdir -p "$stage/live/config-hooks" "$stage/audiovis-util"
mkdir -p "$stage/efi"
cp "$machine_dir/base/grub.cfg" "$stage/efi/grub.cfg"
cp -R "$cache" "$stage/audiovis-util/packages"
if [ -r "$output_root/runtime-packages/SHA256SUMS" ]; then
    cp -R "$output_root/runtime-packages" "$stage/audiovis-util/runtime-packages"
fi
if [ -r "$output_root/sdl-runtime/SHA256SUMS" ]; then
    cp -R "$output_root/sdl-runtime" "$stage/audiovis-util/sdl-runtime"
fi
if [ -x "$output_root/audio-utility" ]; then
    mkdir -p "$stage/audiovis-util/apps/macbookpro12,1"
    cp "$output_root/audio-utility" "$stage/audiovis-util/apps/macbookpro12,1/"
    cp "$machine_dir/app/run-test.sh" "$stage/audiovis-util/apps/macbookpro12,1/"
    if [ -x "$output_root/check-sdl-runtime" ]; then
        cp "$output_root/check-sdl-runtime" "$stage/audiovis-util/apps/macbookpro12,1/"
    fi
fi
cp -R "$host_keys" "$stage/audiovis-util/ssh-host-keys"
cp "$machine_dir/base/console-font.sh" "$stage/audiovis-util/console-font.sh"
cp "$machine_dir/base/install-root-update.sh" "$stage/audiovis-util/install-root-update.sh"
cp "$machine_dir/base/install-offline.sh" "$stage/audiovis-util/install-offline.sh"
cp "$machine_dir/base/audio-utility" "$stage/audiovis-util/audio-utility"
cp "$machine_dir/base/live-startup.sh" "$stage/audiovis-util/live-startup.sh"
cp "$machine_dir/base/hardware-test.sh" "$stage/audiovis-util/hardware-test.sh"
cp "$machine_dir/base/live-hook.sh" "$stage/audiovis-util/live-hook.sh"
cp "$machine_dir/base/live-hook-launcher" "$stage/live/config-hooks/0100-audiovis-util"
chmod 755 "$stage/audiovis-util/live-startup.sh" \
	"$stage/audiovis-util/hardware-test.sh" "$stage/audiovis-util/live-hook.sh"
ssh-keygen -y -f "$SSH_PRIVATE_KEY_FILE" > "$stage/audiovis-util/ssh-authorized-key"
chmod 644 "$stage/audiovis-util/ssh-authorized-key"

python3 - "$stage/audiovis-util" "$WIFI_SSID" "$WIFI_PASSWORD" \
	"$MACBOOK_IPV4_CIDR" "$LAN_GATEWAY" <<'PY'
import pathlib, sys
root = pathlib.Path(sys.argv[1])
ssid, password, address, gateway = sys.argv[2:]
def quote(value):
    return '"' + value.replace('\\', '\\\\').replace('"', '\\"') + '"'
conf = f'''ctrl_interface=/run/wpa_supplicant
update_config=0
country=US

network={{
    ssid={quote(ssid)}
    key_mgmt=WPA-PSK
    psk={quote(password)}
}}
'''
(root / 'wpa_supplicant.conf').write_text(conf)
(root / 'network.conf').write_text(f'''ADDRESS={address}
GATEWAY={gateway}
''')
PY
chmod 600 "$stage/audiovis-util/wpa_supplicant.conf" "$stage/audiovis-util/network.conf"

chmod 755 "$stage/live/config-hooks/0100-audiovis-util"

cat > "$stage/README.txt" <<'EOF'
Debian card payload for MacBookPro12,1.

Copy live/ to DEBIAN and the contents of audiovis-util/ to the
audiovis-util data partition. DEBIAN's only machine-specific file is the small
Live hook that locates and runs live-hook.sh from audiovis-util. Wi-Fi
credentials, SSH key, and all full setup/report scripts belong on audiovis-util.
Keep this payload private and excluded from Git.
EOF
chmod -R go-rwx "$stage"
printf 'Built Debian card payload: %s\n' "$stage"
printf 'Contains local Wi-Fi settings and SSH public key; keep it private.\n'
