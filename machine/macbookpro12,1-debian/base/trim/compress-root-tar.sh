#!/bin/bash
# Run with native squashfs-tools; the tar must come from the verified VM builder.
set -euo pipefail
output=${1:?Usage: compress-root-tar.sh trimmed-root-directory}
sqfstar=${SQFSTAR:-sqfstar}
unsquashfs=${UNSQUASHFS:-unsquashfs}
cmp "$output/expected-removals.txt" "$output/actual-removals.txt"
cmp "$output/boot-before.sha256" "$output/boot-after.sha256"
[ ! -s "$output/dpkg-audit.txt" ]
[ ! -s "$output/unexpected-additions.txt" ]
[ ! -e "$output/filesystem.squashfs" ]
"$sqfstar" -exit-on-error -numeric-owner -root-uid 0 -root-gid 0 -root-mode 0755 \
    -comp xz -b 1M -processors 4 -mem 1G -no-progress \
    "$output/filesystem.squashfs" < "$output/root.tar"
"$unsquashfs" -s "$output/filesystem.squashfs" > "$output/trimmed-filesystem.txt"
scripts=$(cd -- "$(dirname -- "$0")" && pwd)
UNSQUASHFS="$unsquashfs" python3 "$scripts/verify-tar-metadata.py" \
    "$output/root.tar" "$output/filesystem.squashfs" > "$output/metadata-check.txt"
(cd "$output" && shasum -a 256 filesystem.squashfs > SHA256SUMS)
wc -c < "$output/filesystem.squashfs" > "$output/trimmed-size-bytes.txt"
