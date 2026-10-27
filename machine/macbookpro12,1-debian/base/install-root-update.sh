#!/bin/bash
# Run on Debian after uploading a locally verified image to the data partition.
# Backups live on data, never alongside /live/*.squashfs (which Live may load).
set -euo pipefail
image=${1:?Usage: install-root-update.sh /media/audiovis-util/os-updates/filesystem.squashfs SHA256}
expected=${2:?}
[[ "$expected" =~ ^[0-9a-f]{64}$ ]]
. /etc/os-release
[ "$ID" = debian ]
[ "$(findmnt -n -o FSTYPE /run/live/medium)" = tmpfs ]
data=/media/audiovis-util
mountpoint -q "$data"
case "$image" in "$data"/os-updates/*) ;; *) echo 'Image must be staged on the data partition.' >&2; exit 1;; esac
printf '%s  %s\n' "$expected" "$image" | sha256sum -c -
device=$(blkid -L DEBIAN)
data_device=$(findmnt -n -o SOURCE "$data")
[ "$(lsblk -ndo PKNAME "$device")" = "$(lsblk -ndo PKNAME "$data_device")" ]
[ "$(blkid -s TYPE -o value "$device")" = vfat ]
[ -z "$(findmnt -rn -S "$device" -o TARGET)" ] || {
    echo 'DEBIAN is already mounted; inspect before updating.' >&2; exit 1;
}
mount_dir=$(mktemp -d /mnt/debian-update.XXXXXX)
cleanup() {
    if mountpoint -q "$mount_dir"; then umount "$mount_dir"; fi
    rmdir "$mount_dir"
}
trap cleanup EXIT
mount -t vfat -o rw "$device" "$mount_dir"
old=$mount_dir/live/filesystem.squashfs
[ -f "$old" ] && [ -f "$mount_dir/live/vmlinuz" ] && [ -f "$mount_dir/live/initrd.img" ]
old_hash=$(sha256sum "$old" | awk '{print $1}')
if [ "$old_hash" = "$expected" ]; then
    echo 'Requested filesystem is already installed.'
    exit 0
fi
backup=$data/os-backups/$old_hash
mkdir -p "$backup"
if [ ! -f "$backup/filesystem.squashfs" ]; then
    cp "$old" "$backup/filesystem.squashfs.new"
    printf '%s  %s\n' "$old_hash" "$backup/filesystem.squashfs.new" | sha256sum -c -
    mv "$backup/filesystem.squashfs.new" "$backup/filesystem.squashfs"
fi
printf '%s  %s\n' "$old_hash" "$backup/filesystem.squashfs" | sha256sum -c -
printf '%s  filesystem.squashfs\n' "$old_hash" > "$backup/SHA256SUMS"
cp "$image" "$old.new"
printf '%s  %s\n' "$expected" "$old.new" | sha256sum -c -
sync
mv "$old.new" "$old"
sync
printf '%s  %s\n' "$expected" "$old" | sha256sum -c -
printf 'Installed %s\nRollback image: %s/filesystem.squashfs\n' "$expected" "$backup"
printf 'Running RAM root is unchanged; reboot to use the new image.\n'
