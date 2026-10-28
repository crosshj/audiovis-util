#!/bin/bash
# First-boot and every-boot setup for the Debian Live RAM root.
set -u

stamp=$(date -u +%Y%m%dT%H%M%SZ)
log=/tmp/audiovis-util-startup-$stamp-$$.log
exec > >(tee -a "$log") 2>&1
say() { printf '[%s] %s\n' "$(date -u +%H:%M:%S)" "$*"; }
say 'Debian headless startup setup beginning.'

data_device=$(blkid -L audiovis-util 2>/dev/null || true)
data=/media/audiovis-util
if [ -n "$data_device" ]; then
	modprobe hfsplus 2>/dev/null || :
	mkdir -p "$data"
	if ! mountpoint -q "$data"; then
		mount "$data_device" "$data" 2>/dev/null || :
	fi
fi
data_writable=0
if mountpoint -q "$data"; then
	probe=$data/.write-test-$$
	if (umask 077; : > "$probe") 2>/dev/null; then
		rm -f "$probe" || :
		data_writable=1
	fi
fi
if [ "$data_writable" -eq 1 ]; then
	mkdir -p "$data/hardware-reports"
	log=$data/hardware-reports/startup-$stamp.log
	# Continue recording on the persistent data volume once it is mounted.
	exec > >(tee -a "$log") 2>&1
	say "Persistent reports: $data/hardware-reports"
else
	say "WARNING: audiovis-util is unavailable or read-only; logs remain in $log."
fi

/usr/local/sbin/audiovis-util-console-font || say 'WARNING: console font setup failed.'

modprobe applesmc 2>/dev/null || :
led=/sys/class/leds/smc::kbd_backlight
if [ -r "$led/max_brightness" ] && [ -w "$led/brightness" ]; then
	max=$(cat "$led/max_brightness")
	echo none > "$led/trigger" 2>/dev/null || :
	echo $((max / 2)) > "$led/brightness" 2>/dev/null || :
	say 'Apple keyboard backlight enabled.'
fi

if [ ! -r /etc/audiovis-util-network.conf ]; then
	say 'ERROR: generated network settings are missing.'
	exit 1
fi
. /etc/audiovis-util-network.conf
if [ ! -r /etc/wpa_supplicant/audiovis-util.conf ]; then
	say 'ERROR: generated Wi-Fi settings are missing.'
	exit 1
fi

modprobe brcmfmac 2>/dev/null || :
if command -v rfkill >/dev/null 2>&1; then rfkill unblock wlan || :; fi
wifi_interface=
attempt=0
while [ "$attempt" -lt 20 ]; do
    wifi_interface=$(iw dev 2>/dev/null | awk '$1 == "Interface" { print $2; exit }')
    [ -n "$wifi_interface" ] && break
    sleep 1
    attempt=$((attempt + 1))
done

if [ -z "$wifi_interface" ]; then
	say 'ERROR: no wireless interface was detected.'
else
	say "Using wireless interface $wifi_interface."
	ip link set "$wifi_interface" up || :
	if ! pgrep -af "wpa_supplicant.*-i $wifi_interface" >/dev/null 2>&1; then
		wpa_supplicant -B -i "$wifi_interface" -c /etc/wpa_supplicant/audiovis-util.conf || :
	fi
	ip address replace "$ADDRESS" dev "$wifi_interface" || :
	ip route replace default via "$GATEWAY" dev "$wifi_interface" || :
	rm -f /etc/resolv.conf
	printf 'nameserver 1.1.1.1\nnameserver 8.8.8.8\n' > /etc/resolv.conf
fi

attempt=0
while [ "$attempt" -lt 40 ] && [ -n "$wifi_interface" ]; do
	if iw dev "$wifi_interface" link 2>/dev/null | grep -q 'Connected to'; then break; fi
	sleep 1
	attempt=$((attempt + 1))
done
if [ -n "$wifi_interface" ] && iw dev "$wifi_interface" link 2>/dev/null | grep -q 'Connected to'; then
	say "Wi-Fi connected; configured address ${ADDRESS%%/*}."
else
	say 'WARNING: Wi-Fi did not associate; no route or no interface is available.'
fi

# SSH must be available even when the Internet or Debian mirrors are down.
export DEBIAN_FRONTEND=noninteractive
ssh_ready=0
if /usr/local/sbin/audiovis-util-install-offline "$data/packages"; then
    install -d -m 700 /root/.ssh
    install -m 600 "$data/ssh-authorized-key" /root/.ssh/authorized_keys
    mkdir -p /etc/ssh/sshd_config.d /run/sshd
    # Host keys are generated once on the preparation Mac and retained on data.
    for key in "$data"/ssh-host-keys/ssh_host_*; do
        [ -f "$key" ] || continue
        install -m 600 "$key" /etc/ssh/
    done
    if passwd -S root | awk '$2 == "L" {locked=1} END {exit !locked}'; then
        password=$(openssl rand -base64 48)
        printf 'root:%s\n' "$password" | chpasswd
        unset password
    fi
    cat > /etc/ssh/sshd_config.d/00-audiovis-util.conf <<'EOF'
AddressFamily inet
PubkeyAuthentication yes
PermitRootLogin prohibit-password
PasswordAuthentication no
KbdInteractiveAuthentication no
AllowUsers root
AllowAgentForwarding no
AllowTcpForwarding no
X11Forwarding no
PermitTunnel no
UseDNS no
HostKey /etc/ssh/ssh_host_ed25519_key
EOF
    if /usr/sbin/sshd -t && systemctl enable --now ssh.service; then
        ssh_ready=1
        say "SSH listening at root@${ADDRESS%%/*} (public key required)."
    else
        say 'ERROR: SSH configuration or service failed.'
    fi
else
    say 'ERROR: Offline SSH package cache is missing, corrupt, or cannot be installed.'
fi

# Optional graphics packages must never prevent SSH from starting.
if [ -r "$data/runtime-packages/SHA256SUMS" ]; then
    /usr/local/sbin/audiovis-util-install-offline "$data/runtime-packages" ||
        say 'WARNING: graphics runtime installation failed; SSH remains available.'
fi

if [ -r "$data/sdl-runtime/SHA256SUMS" ] &&
    (cd "$data/sdl-runtime" && sha256sum -c SHA256SUMS); then
    install -m 755 "$data/sdl-runtime/libSDL3.so.0" /usr/local/lib/libSDL3.so.0
    ldconfig
fi

if [ -x /usr/local/sbin/audiovis-util-hardware-test ]; then
	/usr/local/sbin/audiovis-util-hardware-test || :
fi

cat > /etc/motd <<EOF
Debian headless RAM-root machine: MacBookPro12,1
Wi-Fi interface: ${wifi_interface:-unavailable}
Wi-Fi: $([ -n "$wifi_interface" ] && iw dev "$wifi_interface" link 2>/dev/null | sed -n 's/^[[:space:]]*SSID: /connected to /p' | head -n 1)
SSH service ready: $ssh_ready; address root@${ADDRESS%%/*} (configured key required)
Reports: /media/audiovis-util/hardware-reports (when the data volume is mounted)
Launch the fullscreen UI from this root console with: audio-utility
Press Escape to close the UI and return to the console.
Root changes are temporary and will be lost at reboot.
EOF
say 'Debian headless startup setup complete.'
