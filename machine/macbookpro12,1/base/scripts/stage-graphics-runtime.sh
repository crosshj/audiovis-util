#!/bin/sh
# Download the signed SDL3/Mesa runtime closure into the writable data volume.
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
media=
for marker in /media/*/.alpine-release; do
	[ -f "$marker" ] || continue
	media=${marker%/.alpine-release}
	break
done
if [ -z "$media" ] && [ -r /run/hardware-setup/boot-device ]; then
	boot_device=$(cat /run/hardware-setup/boot-device)
	media=/media/${boot_device##*/}
	mkdir -p "$media"
	if ! awk -v path="$media" '$2 == path { found=1 } END { exit !found }' /proc/mounts; then
	mount -o ro "$boot_device" "$media" || {
		echo 'Could not remount the ALPINE boot volume.' >&2
		exit 1
	}
	fi
fi
[ -n "$media" ] || {
	echo 'Could not find the mounted ALPINE volume.' >&2
	exit 1
}

. "$script_dir/hardware-storage.sh"
find_audio_utility "$media" || {
	echo 'Could not mount audio-utility read-write.' >&2
	exit 1
}

if [ "$(cat /etc/alpine-release)" != 3.24.3 ]; then
	echo 'This package set is pinned to the Alpine 3.24 base; verify the target release before staging.' >&2
	exit 1
fi

destination=$AUDIO_UTILITY/base-packages/sdl3-kmsdrm
mkdir -p "$destination"
apk fetch --recursive --repositories-file /dev/null \
	--repository https://dl-cdn.alpinelinux.org/alpine/v3.24/main \
	--repository https://dl-cdn.alpinelinux.org/alpine/v3.24/community \
	--output "$destination" sdl3 mesa-dri-gallium mesa-egl mesa-gl
rm -f "$destination"/*.tmp.*
[ -n "$(ls "$destination"/sdl3-*.apk 2>/dev/null)" ] || {
	echo 'SDL3 package was not staged successfully.' >&2
	exit 1
}
[ -n "$(ls "$destination"/mesa-egl-*.apk 2>/dev/null)" ] || {
	echo 'Mesa EGL package was not staged successfully.' >&2
	exit 1
}
sync
echo "SDL3/Mesa runtime packages saved to $destination"
