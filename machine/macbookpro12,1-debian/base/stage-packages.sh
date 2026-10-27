#!/bin/sh
# Run in a disposable local Debian amd64 VM booted from the same Live ISO.
# Never run this on the reference Alpine machine. APT verifies signed indexes
# and archive hashes; checksum manifests verify subsequent copies.
set -eu
output=${1:?Usage: stage-packages.sh /mounted-output}
[ "$(dpkg --print-architecture)" = amd64 ]
. /etc/os-release
[ "$ID" = debian ] && [ "$VERSION_CODENAME" = trixie ]
sources=$(mktemp)
trap 'rm -f "$sources"' EXIT HUP INT TERM
printf 'deb http://deb.debian.org/debian trixie main non-free-firmware\n' > "$sources"
# Avoid file: ISO sources: APT may reuse those archives without copying them.
apt_local() {
    apt-get -o Dir::Etc::sourcelist="$sources" -o Dir::Etc::sourceparts=- "$@"
}
apt_local update
for group in packages runtime-packages; do
    cache=$output/$group
    mkdir -p "$cache/partial"
    case "$group" in
        packages) set -- openssh-server openssl pciutils usbutils alsa-utils curl ca-certificates dnsutils ;;
        runtime-packages) set -- libsdl3-0 libgl1-mesa-dri libegl1 libgbm1 libgles2 kbd ;;
    esac
    apt_local -y --download-only --no-install-recommends \
        -o Dir::Cache::archives="$cache" install "$@"
    (cd "$cache" && sha256sum ./*.deb > SHA256SUMS)
    dpkg-query -W -f='${Package}\t${Version}\n' > "$cache/base-packages.tsv"
done
printf 'Offline package closures staged at %s\n' "$output"
