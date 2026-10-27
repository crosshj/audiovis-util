#!/bin/sh
# Run the staged app on the MacBook's display for 30 seconds, then restore VT.
set -eu

app=/media/audiovis-util/apps/macbookpro12,1/audio-utility
log=/tmp/audio-utility-app-test.log
status_file=/tmp/audio-utility-app-test.status
[ -x "$app" ] || {
	echo "App binary is missing or not executable: $app" >&2
	exit 1
}

active_vt=$(cat /sys/class/tty/tty0/active)
original_vt=${active_vt#tty}
export SDL_VIDEODRIVER=kmsdrm
export SDL_KMSDRM_REQUIRE_DRM_MASTER=1
export SDL_KMSDRM_ATOMIC=0
test_seconds=${1:-30}
case "$test_seconds" in
	*[!0-9]*|'') echo 'Duration must be a whole number of seconds.' >&2; exit 2 ;;
esac
export AUDIO_UTILITY_TEST_SECONDS=$test_seconds

rm -f "$status_file"
if openvt -s -w -- sh -c "'$app' >'$log' 2>&1; rc=\$?; echo \"\$rc\" >'$status_file'"; then
	openvt_status=0
else
	openvt_status=$?
fi
if [ -n "$original_vt" ]; then
	chvt "$original_vt" || echo "Could not switch back to $active_vt." >&2
fi
if [ -r "$status_file" ]; then
	status=$(cat "$status_file")
else
	status=$openvt_status
	echo 'Could not read the app exit status from its virtual terminal.' >&2
fi
cat "$log"
exit "$status"
