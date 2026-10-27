#!/bin/sh
# Locate the audio-utility partition as the third partition beside ALPINE.
# Source this file, then call find_audio_utility with the boot-volume mountpoint.
find_audio_utility() {
	AUDIO_UTILITY=
	AUDIO_UTILITY_MOUNTED_BY_HELPER=0
	_au_media=$1
	_au_boot_device=
	for _au_marker in "$_au_media"/.alpine-release /media/*/.alpine-release; do
		[ -f "$_au_marker" ] || continue
		_au_candidate=${_au_marker%/.alpine-release}
		_au_boot_device=$(awk -v path="$_au_candidate" '$2 == path { print $1; exit }' /proc/mounts)
		[ -z "$_au_boot_device" ] || break
	done
	if [ -z "$_au_boot_device" ] && [ -r /run/hardware-setup/boot-device ]; then
		_au_boot_device=$(cat /run/hardware-setup/boot-device)
	fi
	[ -n "$_au_boot_device" ] || return 1

	case "$_au_boot_device" in
		/dev/sd*2|/dev/mmcblk*p2) _au_data_device=${_au_boot_device%2}3 ;;
		*) return 1 ;;
	esac
	[ -b "$_au_data_device" ] || return 1

	modprobe hfsplus 2>/dev/null || :
	_au_mountpoint=$(awk -v device="$_au_data_device" '$1 == device && $3 == "hfsplus" { print $2; exit }' /proc/mounts)
	if [ -n "$_au_mountpoint" ]; then
		mount -o remount,rw "$_au_mountpoint" 2>/dev/null || :
	else
		_au_mountpoint=/media/audio-utility
		mkdir -p "$_au_mountpoint" || return 1
		mount -t hfsplus -o rw "$_au_data_device" "$_au_mountpoint" 2>/dev/null || return 1
		AUDIO_UTILITY_MOUNTED_BY_HELPER=1
	fi

	awk -v device="$_au_data_device" -v path="$_au_mountpoint" \
		'$1 == device && $2 == path && $3 == "hfsplus" && $4 ~ /(^|,)rw(,|$)/ { found=1 } END { exit !found }' \
		/proc/mounts || return 1

	_au_probe=$_au_mountpoint/.audio-utility-write-test-$$
	if ! (umask 077; : > "$_au_probe") 2>/dev/null; then
		return 1
	fi
	rm -f "$_au_probe" || return 1
	AUDIO_UTILITY=$_au_mountpoint
	return 0
}

