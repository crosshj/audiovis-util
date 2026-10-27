#!/bin/sh
# Build the Debian Live card as labeled partitions and copied files, not a raw ISO.
set -eu

if [ "$#" -lt 4 ] || [ "$#" -gt 5 ]; then
	echo "Usage: $0 /path/to/debian-live-amd64-standard.iso /path/to/SHA512SUMS /dev/diskN /path/to/card-payload [trimmed-root-directory]" >&2
	exit 2
fi
image=$1
sums=$2
target=$3
payload=$4
trimmed=${5:-}
if [ -n "$trimmed" ]; then
    [ -r "$trimmed/filesystem.squashfs" ] && [ -r "$trimmed/SHA256SUMS" ] || {
        echo 'Trimmed filesystem or checksum is missing.' >&2; exit 1;
    }
    (cd "$trimmed" && shasum -a 256 -c SHA256SUMS)
fi
[ -r "$image" ] && [ -r "$sums" ] || { echo 'ISO or checksum file is not readable.' >&2; exit 1; }
[ -d "$payload/live/config-hooks" ] && [ -d "$payload/audiovis-util" ] || {
	echo 'Card payload is incomplete; run base/build-config.sh first.' >&2
	exit 1
}
case "$target" in
	/dev/disk[0-9]*) ;;
	*) echo 'Target must be a whole macOS disk such as /dev/disk5.' >&2; exit 1 ;;
esac
case "$target" in
	/dev/disk[0-9]*s[0-9]*) echo 'Target must be a whole disk, not a partition.' >&2; exit 1 ;;
esac

disk_info=$(diskutil info "$target")
printf '%s\n' "$disk_info" | grep -q 'Device Location:.*External' || {
	echo 'Refusing to write: target is not external.' >&2; exit 1;
}
printf '%s\n' "$disk_info" | grep -q 'Removable Media:.*Removable' || {
	echo 'Refusing to write: target is not marked removable.' >&2; exit 1;
}
image_name=$(basename "$image")
sum_line=$(awk -v name="$image_name" '$2 == name { print $1 "  " $2 }' "$sums")
[ -n "$sum_line" ] || { echo "No published checksum entry for $image_name." >&2; exit 1; }
printf '%s\n' "$sum_line" | (cd "$(dirname "$image")" && shasum -a 512 -c -)

bsdtar -tf "$image" | grep -q '^live/vmlinuz$' || { echo 'Debian Live kernel is missing.' >&2; exit 1; }
bsdtar -tf "$image" | grep -q '^live/initrd.img$' || { echo 'Debian Live initramfs is missing.' >&2; exit 1; }

printf '\nTarget: %s\n' "$target"
printf '%s\n' "$disk_info" | grep -E 'Device / Media Name:|Disk Size:|Device Location:|Removable Media:'
printf 'This erases the entire target disk. Type ERASE %s to continue: ' "$target" >&2
IFS= read -r confirmation
[ "$confirmation" = "ERASE $target" ] || { echo 'Cancelled.' >&2; exit 1; }

work=$(mktemp -d /private/tmp/debian-card.XXXXXX)
cleanup() {
	chmod -R u+w "$work" 2>/dev/null || :
	rm -rf "$work"
}
trap cleanup EXIT HUP INT TERM
mkdir -p "$work/source"
echo 'Extracting the selected EFI, GRUB, and Live system files from the verified ISO.'
bsdtar -xf "$image" -C "$work/source" boot/grub EFI/boot live

sudo -v
echo "Unmounting $target before partitioning."
sudo diskutil unmountDisk "$target"
echo 'Creating the built-in EFI system partition plus DEBIAN and audiovis-util.'
sudo diskutil partitionDisk "$target" GPTFormat MS-DOS DEBIAN 4G HFS+ audiovis-util R

volume_mount() {
	partition=$1
	mount_point=$(diskutil info "$partition" | sed -n 's/^[[:space:]]*Mount Point: *//p')
	if [ -z "$mount_point" ] || [ "$mount_point" = 'Not applicable (no file system)' ] || [ "$mount_point" = 'Not mounted' ]; then
		sudo diskutil mount "$partition" >/dev/null
		mount_point=$(diskutil info "$partition" | sed -n 's/^[[:space:]]*Mount Point: *//p')
	fi
	[ -n "$mount_point" ] && [ "$mount_point" != 'Not applicable (no file system)' ] || {
		echo "Could not mount $partition." >&2; exit 1;
	}
	printf '%s\n' "$mount_point"
}

efi=$(volume_mount "${target}s1")
debian=$(volume_mount "${target}s2")
data=$(volume_mount "${target}s3")
for volume in "$efi" "$debian" "$data"; do
	[ -d "$volume" ] || { echo "Expected mounted volume $volume was not created." >&2; exit 1; }
done

echo 'Copying Debian boot files and the generated headless startup payload.'
mkdir -p "$efi/EFI/BOOT" "$efi/boot/grub" "$debian/live"
ditto "$work/source/EFI/boot/" "$efi/EFI/BOOT/"
ditto "$work/source/boot/grub/" "$efi/boot/grub/"
# Copy only the three runtime files; dereference the ISO's hardlink aliases.
for file in filesystem.squashfs vmlinuz initrd.img; do
    source_file=$work/source/live/$file
    if [ "$file" = filesystem.squashfs ] && [ -n "$trimmed" ]; then
        source_file=$trimmed/filesystem.squashfs
    fi
    cp -L "$source_file" "$debian/live/$file"
    cmp "$source_file" "$debian/live/$file"
done
ditto "$payload/live/" "$debian/live/"
ditto "$payload/audiovis-util/" "$data/"
rm -f "$debian/live/filesystem.packages"


mkdir -p "$efi/boot/grub" "$efi/EFI/BOOT"
cp "$payload/efi/grub.cfg" "$efi/boot/grub/grub.cfg"
cp "$payload/efi/grub.cfg" "$efi/EFI/BOOT/grub.cfg"
mkdir -p "$efi/boot/grub/x86_64-efi"
cp "$payload/efi/grub.cfg" "$efi/boot/grub/x86_64-efi/grub.cfg"
rmdir "$debian/EFI/boot" "$debian/EFI" 2>/dev/null || :

sync
sudo diskutil eject "$target"
echo "Prepared $target with EFI, DEBIAN, and audiovis-util partitions."
echo 'The Debian Live OS boots in text mode with toram; its root changes are temporary.'
