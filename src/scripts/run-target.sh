#!/bin/sh
# Run only on the Debian target; stage this file beside the product binary.
set -eu
cd "$(dirname "$0")"
. /etc/os-release
[ "$ID" = debian ] || { echo 'Debian required'; exit 1; }
if pgrep -x audio-utility >/dev/null; then
 echo 'An app is already running; refusing to compete for DRM.' >&2
 exit 1
fi
vt=$(cat /sys/class/tty/tty0/active)
trap 'chvt "${vt#tty}"' EXIT HUP INT TERM
export SDL_VIDEODRIVER=kmsdrm SDL_KMSDRM_REQUIRE_DRM_MASTER=1 SDL_KMSDRM_ATOMIC=0
export LD_LIBRARY_PATH="$PWD/bin${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
unset DISPLAY WAYLAND_DISPLAY
rm -f result.status
openvt -s -w -- sh -c './audio-utility "$@" >run.log 2>&1; code=$?; echo "$code" >result.status; exit "$code"' sh "$@" || true
cat run.log
[ -r result.status ] && [ "$(cat result.status)" = 0 ]
