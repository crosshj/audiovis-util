#!/bin/sh
# Run over SSH on Debian after startup. --vm skips physical hardware/Wi-Fi checks.
set -eu
. /etc/os-release
[ "$ID" = debian ] || { echo 'Refusing checks: this is not Debian.' >&2; exit 1; }
[ "$(findmnt -n -o FSTYPE /)" = overlay ]
[ "$(findmnt -n -o FSTYPE /run/live/overlay)" = tmpfs ]
[ "$(findmnt -n -o FSTYPE /run/live/medium)" = tmpfs ]
findmnt -n -o OPTIONS /run/live/medium | grep -Eq '(^|,)ro(,|$)'
[ -z "$(swapon --noheadings --show)" ]
echo 'PASS: operating root and OS medium are in RAM; OS medium is read-only; no swap.'
[ "$(findmnt -n -o FSTYPE /media/audiovis-util)" = hfsplus ]
probe=$(mktemp /media/audiovis-util/.verify-write.XXXXXX)
rm -f "$probe"
echo 'PASS: persistent data is mounted and writable.'
systemctl is-active --quiet ssh.service
settings=$(/usr/sbin/sshd -T)
printf '%s\n' "$settings" | grep -qx 'passwordauthentication no'
printf '%s\n' "$settings" | grep -qx 'kbdinteractiveauthentication no'
printf '%s\n' "$settings" | grep -qx 'allowusers root'
cmp /etc/ssh/ssh_host_ed25519_key /media/audiovis-util/ssh-host-keys/ssh_host_ed25519_key
echo 'PASS: SSH is active, key-only, and uses the retained host identity.'
systemctl show getty@tty1.service -p ExecStart | grep -q -- '--autologin root'
echo 'PASS: tty1 automatically logs in as root.'
if [ "${1:-}" != --vm ]; then
    [ "$(cat /sys/class/dmi/id/product_name)" = MacBookPro12,1 ]
    interface=$(iw dev | awk '$1 == "Interface" {print $2; exit}')
    [ -n "$interface" ]
    iw dev "$interface" link | grep -q 'Connected to'
    . /etc/audiovis-util-network.conf
    ip -4 -o address show dev "$interface" | grep -Fq " $ADDRESS "
    echo 'PASS: expected MacBook hardware, associated Wi-Fi, and configured address.'
else
    echo 'SKIP: physical MacBook hardware and Broadcom Wi-Fi require a card boot.'
fi
systemctl is-active --quiet audiovis-util-live-startup.service
[ "$(systemctl show audiovis-util-live-startup.service -p Result --value)" = success ]
echo 'PASS: automatic startup completed successfully.'
