#!/bin/sh
# Apply the generated startup payload to already-prepared, mounted card volumes.
set -eu

if [ "$#" -ne 4 ]; then
	echo "Usage: $0 /Volumes/EFI /Volumes/DEBIAN /Volumes/audiovis-util /path/to/card-payload" >&2
	exit 2
fi
efi=$1
debian=$2
data=$3
payload=$4

check_volume() {
	path=$1
	expected=$2
	[ -d "$path" ] || { echo "Volume is not mounted: $path" >&2; exit 1; }
	info=$(diskutil info "$path")
	name=$(printf '%s\n' "$info" | sed -n 's/^[[:space:]]*Volume Name:[[:space:]]*//p' | head -n 1)
	[ "$name" = "$expected" ] || {
		echo "Expected volume $expected at $path; found $name." >&2; exit 1;
	}
	printf '%s\n' "$info" | grep -q 'Volume Read-Only:.*No' || {
		echo "Volume is read-only: $path" >&2; exit 1;
	}
}

check_volume "$efi" EFI
check_volume "$debian" DEBIAN
check_volume "$data" audiovis-util

# Labels alone cannot establish that these are partitions of the same card.
whole=
for volume in "$efi" "$debian" "$data"; do
    parent=$(diskutil info "$volume" | sed -n 's/^[[:space:]]*Part of Whole:[[:space:]]*//p')
    [ -n "$parent" ] || { echo "Cannot identify parent of $volume" >&2; exit 1; }
    [ -z "$whole" ] && whole=$parent
    [ "$parent" = "$whole" ] || { echo 'Volumes belong to different disks.' >&2; exit 1; }
done
info=$(diskutil info "$whole")
printf '%s\n' "$info" | grep -q 'Device Location:.*External'
printf '%s\n' "$info" | grep -q 'Removable Media:.*Removable'
[ -r "$payload/audiovis-util/packages/SHA256SUMS" ] || {
    echo 'Offline SSH packages have not been staged.' >&2; exit 1;
}
[ -r "$payload/live/config-hooks/0100-audiovis-util" ] || {
	echo 'Generated live startup hook is missing.' >&2; exit 1;
}
[ -r "$payload/audiovis-util/live-hook.sh" ] &&
	[ -r "$payload/audiovis-util/live-startup.sh" ] || {
	echo 'Generated startup scripts are missing.' >&2; exit 1;
}
[ -r "$debian/live/filesystem.squashfs" ] && [ -r "$debian/live/vmlinuz" ] &&
	[ -r "$debian/live/initrd.img" ] || {
		echo 'Debian Live root, kernel, or initramfs is missing from DEBIAN.' >&2; exit 1;
}

# Earlier attempts put the generated scripts in DEBIAN. Remove those exact
# generated files so DEBIAN contains only Debian and its small bridge hook.
for file in live-hook.sh live-startup.sh hardware-test.sh ssh-authorized-key \
		wpa_supplicant.conf network.conf; do
	rm -f "$debian/audiovis-util/$file"
done
rmdir "$debian/audiovis-util" 2>/dev/null || :
rmdir "$debian/EFI/boot" "$debian/EFI" 2>/dev/null || :
if [ -e "$debian/audiovis-util" ] || [ -e "$debian/EFI" ]; then
	echo 'Unexpected generated directories remain on DEBIAN; refusing to finish.' >&2
	exit 1
fi

# GRUB belongs on EFI. Move its source files off DEBIAN if this is an earlier
# card layout, then leave only the Debian Live root and its bridge hook there.
if [ -d "$debian/boot/grub" ]; then
	mkdir -p "$efi/boot/grub"
	ditto "$debian/boot/grub/" "$efi/boot/grub/"
	for item in "$debian/boot"/*; do
		[ -e "$item" ] || continue
		[ "${item##*/}" = grub ] || {
			echo "Unexpected content under $debian/boot; refusing to remove it." >&2
			exit 1
		}
	done
	chmod -R u+w "$debian/boot" 2>/dev/null || :
	rm -rf "$debian/boot"
fi
if [ -e "$debian/boot" ]; then
	echo 'Bootloader files remain under DEBIAN; refusing to finish.' >&2
	exit 1
fi

echo 'Copying the bridge hook to DEBIAN and full scripts/settings to audiovis-util.'
ditto "$payload/live/" "$debian/live/"
ditto "$payload/audiovis-util/" "$data/"
rm -f "$debian/live/filesystem.packages"


mkdir -p "$efi/EFI/BOOT" "$efi/boot/grub"
cp "$payload/efi/grub.cfg" "$efi/boot/grub/grub.cfg"
cp "$payload/efi/grub.cfg" "$efi/EFI/BOOT/grub.cfg"
mkdir -p "$efi/boot/grub/x86_64-efi"
cp "$payload/efi/grub.cfg" "$efi/boot/grub/x86_64-efi/grub.cfg"
chmod 755 "$data/live-hook.sh" "$data/live-startup.sh" "$data/hardware-test.sh"
[ -x "$debian/live/config-hooks/0100-audiovis-util" ] &&
	[ -x "$data/live-hook.sh" ] && [ -x "$data/live-startup.sh" ] || {
		echo 'Card payload placement check failed.' >&2; exit 1;
}
# Compare every generated file after copying, including private configuration.
python3 - "$payload" "$efi" "$debian" "$data" <<'PYVERIFY'
import pathlib, sys
payload, efi, debian, data = map(pathlib.Path, sys.argv[1:])
for source, dest in [(payload/'live', debian/'live'), (payload/'audiovis-util', data)]:
    for path in source.rglob('*'):
        if path.is_file():
            target = dest / path.relative_to(source)
            if not target.is_file() or path.read_bytes() != target.read_bytes():
                raise SystemExit(f'Copy verification failed: {target}')
for rel in ['boot/grub/grub.cfg', 'boot/grub/x86_64-efi/grub.cfg', 'EFI/BOOT/grub.cfg']:
    if (efi/rel).read_bytes() != (payload/'efi/grub.cfg').read_bytes():
        raise SystemExit(f'GRUB verification failed: {rel}')
print('Verified all generated card files byte for byte.')
PYVERIFY
sync
echo 'Card updated: EFI holds all boot files, DEBIAN holds the Live OS plus its small bridge hook, and audiovis-util holds the full machine scripts/settings.'
partition_id=$(diskutil info "$efi" | sed -n 's/^[[:space:]]*Device Identifier:[[:space:]]*//p' | head -n 1)
whole_disk=${partition_id%s*}
case "$whole_disk" in
	disk[0-9]*) diskutil eject "/dev/$whole_disk" ;;
	*) echo "Could not determine the SD card from $efi; eject it manually." >&2; exit 1 ;;
esac
echo "Safely ejected /dev/$whole_disk."
