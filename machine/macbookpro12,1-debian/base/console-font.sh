#!/bin/sh
# Match Alpine's observed 16x32-pixel text cells at native 2560x1600.
# Run after Debian console-setup so its default small font cannot override this.
set -eu
font=/usr/share/consolefonts/Lat15-Terminus32x16.psf.gz
[ -r "$font" ] || { echo "Console font missing: $font" >&2; exit 1; }
for tty in /dev/tty1 /dev/tty2 /dev/tty3 /dev/tty4 /dev/tty5 /dev/tty6; do
    [ -c "$tty" ] || continue
    setfont -C "$tty" "$font"
done
printf 'Console font: Terminus 16x32 pixels; tty1 rows/columns: '
stty -F /dev/tty1 size
for size in /sys/class/graphics/fb*/virtual_size; do
    [ -r "$size" ] || continue
    printf 'Framebuffer %s: ' "$size"
    cat "$size"
done
