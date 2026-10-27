#!/bin/bash
# Build only in a disposable local QEMU VM; never in the physical target root.
set -euo pipefail
source_image=${1:?Usage: build-trimmed-root.sh source.squashfs scratch-directory output-directory [squashfs|tar]}
scratch=${2:?}
output=${3:?}
mode=${4:-squashfs}
case "$mode" in squashfs|tar) ;; *) echo 'Output mode must be squashfs or tar' >&2; exit 1 ;; esac
case "$(cat /sys/class/dmi/id/product_name)" in
    'QEMU '*|'Standard PC ('*) ;;
    *) echo 'This builder requires a disposable local QEMU VM.' >&2; exit 1 ;;
esac
scripts=$(cd -- "$(dirname -- "$0")" && pwd)
printf '%s  %s\n' 6af8bcb58ec0122620c0ccbe59c2e1acf9f07f519950931662642c9916e53b15 "$source_image" | sha256sum -c -
root=$scratch/root
[ ! -e "$root" ] || { echo "Refusing to overwrite existing work: $root" >&2; exit 1; }
mkdir -p "$output"
unsquashfs -s "$source_image" > "$output/source-filesystem.txt"
unsquashfs -no-progress -processors 2 -d "$root" "$source_image"
chroot "$root" dpkg-query -W -f='${db:Status-Abbrev} ${Package}\t${Version}\n' > "$output/packages-before.tsv"
mapfile -t remove < "$scripts/remove-packages.txt"
sort -u "$scripts/remove-packages.txt" > "$output/expected-removals.txt"
chroot "$root" apt-get --simulate purge "${remove[@]}" > "$output/removal-plan.txt"
awk '$1 == "Remv" || $1 == "Purg" {print $2}' "$output/removal-plan.txt" | sort -u > "$output/planned-removals.txt"
cmp "$output/expected-removals.txt" "$output/planned-removals.txt"
# Block services and avoid regenerating the separately supplied Live initramfs.
[ ! -e "$root/usr/sbin/policy-rc.d" ]
cp -p "$root/etc/initramfs-tools/update-initramfs.conf" "$scratch/update-initramfs.conf"
sed -i 's/^update_initramfs=.*/update_initramfs=no/' "$root/etc/initramfs-tools/update-initramfs.conf"
printf '#!/bin/sh\nexit 101\n' > "$root/usr/sbin/policy-rc.d"
chmod 755 "$root/usr/sbin/policy-rc.d"
cleanup() {
    for mount in sys proc dev; do
        if mountpoint -q "$root/$mount"; then umount "$root/$mount"; fi
    done
    cp -p "$scratch/update-initramfs.conf" "$root/etc/initramfs-tools/update-initramfs.conf"
    rm -f "$root/usr/sbin/policy-rc.d"
}
trap cleanup EXIT
(cd "$root" && find boot -type f -exec sha256sum {} \; | sort) > "$output/boot-before.sha256"
for mount in dev proc sys; do mount --bind "/$mount" "$root/$mount"; done
DEBIAN_FRONTEND=noninteractive chroot "$root" apt-get -y purge "${remove[@]}"
chroot "$root" dpkg --audit > "$output/dpkg-audit.txt"
[ ! -s "$output/dpkg-audit.txt" ]
chroot "$root" dpkg-query -W -f='${db:Status-Abbrev} ${Package}\t${Version}\n' > "$output/packages-after.tsv"
awk '$1 == "ii" {print $2}' "$output/packages-before.tsv" | sort > "$output/installed-before.txt"
awk '$1 == "ii" {print $2}' "$output/packages-after.tsv" | sort > "$output/installed-after.txt"
comm -23 "$output/installed-before.txt" "$output/installed-after.txt" > "$output/actual-removals.txt"
cmp "$output/expected-removals.txt" "$output/actual-removals.txt"
comm -13 "$output/installed-before.txt" "$output/installed-after.txt" > "$output/unexpected-additions.txt"
[ ! -s "$output/unexpected-additions.txt" ]
while IFS= read -r package; do
    grep -Fx "$package" "$output/installed-after.txt" >/dev/null || {
        echo "Required package missing: $package" >&2; exit 1;
    }
done < "$scripts/keep-packages.txt"
(cd "$root" && find boot -type f -exec sha256sum {} \; | sort) > "$output/boot-after.sha256"
cmp "$output/boot-before.sha256" "$output/boot-after.sha256"
cleanup
trap - EXIT
# Native sqfstar can compress this archive on a different architecture without
# extracting Linux ownership/device metadata onto the host filesystem.
if [ "$mode" = tar ]; then
    tar --numeric-owner --xattrs --acls --sort=name --one-file-system \
        -C "$root" -cf "$output/root.tar" .
    printf 'Verified root exported; compress with compress-root-tar.sh.\n'
    exit 0
fi
# Preserve file ownership, modes, links and xattrs on the Linux scratch filesystem.
mksquashfs "$root" "$scratch/filesystem.squashfs" -noappend -comp xz -b 1M -processors 2 -no-progress
unsquashfs -s "$scratch/filesystem.squashfs" > "$output/trimmed-filesystem.txt"
cp "$scratch/filesystem.squashfs" "$output/filesystem.squashfs"
(cd "$output" && sha256sum filesystem.squashfs > SHA256SUMS)
stat -c '%s %n' "$source_image" "$output/filesystem.squashfs" > "$output/image-sizes.txt"
printf 'Trimmed filesystem built and package/boot invariants verified.\n'
