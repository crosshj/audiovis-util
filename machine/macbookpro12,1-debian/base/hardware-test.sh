#!/bin/sh
# Debian counterpart to the Alpine hardware report.
set -u
stamp=$(date -u +%Y%m%dT%H%M%SZ)
report=/var/tmp/hardware-report-$stamp-$$.txt
run() {
  printf '\n$'
  printf ' %s' "$@"
  printf '\n'
  if command -v "$1" >/dev/null 2>&1; then "$@" 2>&1 || printf '(exit %s)\n' "$?"
  else printf '(unavailable: %s)\n' "$1"; fi
}
{
  echo '===== Debian, model, CPU, RAM ====='
  run cat /etc/debian_version
  run uname -a
  run cat /sys/class/dmi/id/product_name
  run grep 'model name' /proc/cpuinfo
  run free -h
  echo '\n===== Services and storage ====='
  run systemctl --no-pager --failed
  run systemctl --no-pager --type=service --state=running
  run lsblk -f
  run findmnt
  run df -h
  echo '\n===== PCI/USB and kernel drivers ====='
  run lspci -nnk
  run lsusb -t
  echo '\n===== Wi-Fi, firmware, and network ====='
  run ip link
  run ip address
  run ip route
  run iw dev
  wifi_interface=$(iw dev 2>/dev/null | awk '$1 == "Interface" { print $2; exit }')
  if [ -n "$wifi_interface" ]; then run iw dev "$wifi_interface" link
  else echo '(no wireless interface detected)'; fi
  run journalctl -k -b --no-pager -g 'brcm|firmware|wlan|wifi|apple|smc|error'
  echo '\n===== Apple keyboard and input ====='
  run lsmod
  run ls -l /sys/class/leds
  run cat /sys/class/leds/smc::kbd_backlight/brightness
  run cat /proc/bus/input/devices
  echo '\n===== Audio ====='
  run cat /proc/asound/cards
  run aplay -l
  run arecord -l
  echo '\n===== Graphics and camera ====='
  run stty -F /dev/tty1 size
  run cat /sys/class/graphics/fb0/virtual_size
  run ls -l /dev/dri
  run ls -l /dev/video*
  run glxinfo -B
  echo '\n===== Connectivity ====='
  run ping -c 1 -W 3 1.1.1.1
  run getent hosts debian.org
  run curl -I --max-time 10 https://www.debian.org/
} > "$report" 2>&1
mkdir -p /media/audiovis-util/hardware-reports 2>/dev/null || true
if [ -d /media/audiovis-util/hardware-reports ] &&
  cp "$report" /media/audiovis-util/hardware-reports/ 2>/dev/null; then
  echo "Hardware report: /media/audiovis-util/hardware-reports/${report##*/}"
else
  echo "Hardware report: $report"
fi
