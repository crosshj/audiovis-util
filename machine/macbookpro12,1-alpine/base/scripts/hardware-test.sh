#!/bin/sh
# Run from the Alpine SD card as root: sh /media/*/hardware-test.sh
set -u

if [ "$(uname -s)" != Linux ] || [ "$(id -u)" != 0 ]; then
    echo 'Run this as root after booting Alpine on the MacBook.' >&2
    exit 1
fi

media=${1:-}
ram=${2:-/run/hardware-setup}
if [ -z "$media" ]; then
	for marker in /media/*/.alpine-release; do
		[ -f "$marker" ] || continue
		media=${marker%/.alpine-release}
		break
	done
fi
if [ -z "$media" ] && [ -r /run/hardware-setup/boot-device ]; then
	media=/media/sdb2
fi
if [ -z "$media" ] || [ ! -d "$ram/hardware-test/apks" ]; then
	echo 'Usage: hardware-test.sh [ALPINE mountpoint] [RAM working directory]' >&2
	exit 1
fi

stamp=$(date -u +%Y%m%dT%H%M%SZ)
report=/tmp/hardware-report-$stamp-$$.txt

section() {
    printf '\n===== %s =====\n' "$*"
}

run() {
    printf '\n$'
    printf ' %s' "$@"
    printf '\n'
    if command -v "$1" >/dev/null 2>&1; then
        "$@" 2>&1 || printf '(command exited %s)\n' "$?"
    else
        printf '(command unavailable: %s)\n' "$1"
    fi
}

check_cmd() {
    check_name=$1
    shift
    printf '\n$'
    printf ' %s' "$@"
    printf '\n'
    check_output=/tmp/hardware-check-$$.txt
    if "$@" > "$check_output" 2>&1; then
        check_result=PASS
    else
        check_result=FAIL
    fi
    cat "$check_output"
    rm -f "$check_output"
    printf 'CHECK_RESULT|%s|%s\n' "$check_name" "$check_result"
}

{
    section 'Mac model, stock Alpine boot, and RAM'
    run cat /etc/alpine-release
    run uname -a
    run cat /sys/class/dmi/id/product_name
    run grep 'model name' /proc/cpuinfo
    run free -m
    run cat /proc/cmdline

    section 'Runtime use and active services'
    run top -b -n 1
    run rc-status -a

    section 'PCI and USB devices and attached drivers'
    run lspci -nnk
    run lsusb
    run lsusb -t

    section 'Wi-Fi radio, firmware, and network state'
    run ip link
    run ip address
    run ip route
    run iw dev
    run iw phy
    run wpa_cli -i wlan0 status
    run ls -l /lib/firmware/brcm/brcmfmac43602*
    printf '\nSelected kernel messages:\n'
    dmesg | grep -Ei 'brcm|firmware|wlan|wifi|apple|smc|kbd|error' || true

    section 'Apple keyboard and backlight'
    run lsmod
    run ls -l /sys/class/leds
    run cat /sys/class/leds/smc::kbd_backlight/brightness
    run cat /sys/class/leds/smc::kbd_backlight/max_brightness

    section 'Audio devices'
    run cat /proc/asound/cards
    run aplay -l
    run arecord -l

    section 'Touchpad and input devices'
    run cat /proc/bus/input/devices

    section 'Graphics, camera, and storage'
    run ls -l /dev/dri
    run ls -l /dev/video*
    run cat /proc/partitions
    run blkid
    run mount
    run df -h

    section 'Install signed offline diagnostic tools into RAM'
    # Include the SDL/Mesa packages in the same transaction so APK does not
    # prune the runtime-only packages loaded earlier during startup.
    graphics_packages=
    if [ -r "$ram/hardware-storage.sh" ]; then
        . "$ram/hardware-storage.sh"
        if find_audio_utility "$media"; then
            graphics_packages=$AUDIO_UTILITY/base-packages/sdl3-kmsdrm
        fi
    fi
    if [ -n "$graphics_packages" ]; then
        set -- "$graphics_packages"/*.apk
        if [ -f "$1" ]; then
            set -- "$ram"/hardware-test/apks/*.apk "$graphics_packages"/*.apk
        else
            set -- "$ram"/hardware-test/apks/*.apk
        fi
    else
        set -- "$ram"/hardware-test/apks/*.apk
    fi

    world_backup=/tmp/apk-world-before-hardware-test-$$
    if cp /etc/apk/world "$world_backup"; then
        if apk --no-network --force-non-repository --repositories-file /dev/null add "$@"; then
            echo 'Offline diagnostic tools installed.'
        else
            echo 'Offline package install failed; the tools remain on the ALPINE volume.'
        fi
        if cp "$world_backup" /etc/apk/world; then
            rm -f "$world_backup"
        else
            echo 'WARNING: Could not restore /etc/apk/world after diagnostic package install.'
        fi
    else
        echo 'Could not protect Alpine package selections; diagnostic packages were not installed.'
    fi

    section 'Audio device enumeration after installing ALSA tools'
    run cat /proc/asound/cards
    run aplay -l
    run arecord -l

    section 'Detailed Wi-Fi and PCI results after installing the tools'
    run lspci -nnk
    run iw dev
    run iw phy
    run wpa_cli -i wlan0 status
    run ip address
    run ip route

    section 'Automatic Wi-Fi association and Internet checks'
    echo 'Wait up to 30 seconds for wlan0 to associate.'
    attempt=0
    while [ "$attempt" -lt 30 ]; do
        if iw dev wlan0 link 2>/dev/null | grep -q 'Connected to'; then
            break
        fi
        sleep 1
        attempt=$((attempt + 1))
    done
    check_cmd 'Wi-Fi association' sh -c 'iw dev wlan0 link; iw dev wlan0 link | grep -q "Connected to"'
    wifi_association=$check_result

    echo 'Wait up to 30 seconds for the configured IPv4 address.'
    attempt=0
    while [ "$attempt" -lt 30 ]; do
        if ip -4 address show dev wlan0 2>/dev/null | grep -q 'inet '; then
            break
        fi
        sleep 1
        attempt=$((attempt + 1))
    done
    check_cmd 'IPv4 address' sh -c 'ip -4 address show dev wlan0; ip -4 address show dev wlan0 | grep -q "inet "'
    ipv4_result=$check_result

    gateway=$(ip route show default 2>/dev/null | awk 'NR == 1 { print $3; exit }')
    if [ -n "$gateway" ]; then
        check_cmd 'LAN gateway ping' ping -c 1 -W 3 "$gateway"
        gateway_result=$check_result
    else
        echo 'No default gateway was assigned.'
        gateway_result=FAIL
    fi
    check_cmd 'Public IP ping' ping -c 1 -W 3 1.1.1.1
    public_ip_result=$check_result
    check_cmd 'DNS lookup' nslookup alpinelinux.org
    dns_result=$check_result
    check_cmd 'HTTPS Internet fetch' wget -q -T 10 -O /dev/null https://alpinelinux.org/
    https_result=$check_result

    section 'Result summary'
    printf 'Wi-Fi association: %s\n' "$wifi_association"
    printf 'IPv4 address configured: %s\n' "$ipv4_result"
    printf 'LAN gateway ping: %s' "$gateway_result"
    [ -z "$gateway" ] || printf ' (%s)' "$gateway"
    printf '\nPublic IP ping: %s\nDNS lookup: %s\nHTTPS Internet fetch: %s\n' \
        "$public_ip_result" "$dns_result" "$https_result"
    echo 'A failed public ping can mean ICMP is blocked; HTTPS is the stronger Internet check.'
} > "$report" 2>&1

# Save reports only to audio-utility, never to the Alpine boot volume.
destination=
dmesg_report=/tmp/dmesg-$stamp-$$.txt
dmesg > "$dmesg_report"
if [ -r "$ram/hardware-storage.sh" ]; then
	. "$ram/hardware-storage.sh"
    if find_audio_utility "$media"; then
        destination=$AUDIO_UTILITY
    fi
fi

if [ -n "$destination" ] && mkdir -p "$destination/hardware-reports" && cp "$report" "$destination/hardware-reports/"; then
    saved=$destination/hardware-reports/$(basename "$report")
    cp "$dmesg_report" "$destination/hardware-reports/"
    printf '%s\n' "$saved" > "$destination/hardware-reports/latest-report"
    sync
    echo "Saved report: $saved"
    echo 'Wi-Fi and Internet check results:'
    sed -n '/^===== Result summary =====$/,$p' "$report"
else
    echo "Could not save to audio-utility. Report is in RAM: $report"
fi
echo "Read the report with: less $report"
echo 'Use poweroff before removing the card after testing.'
