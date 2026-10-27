#!/bin/sh
# Install the locally built SDL runtime on the Debian card after it boots.
set -eu
repo=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
runtime=$repo/build/macbookpro12,1-debian/sdl-runtime
ssh=$repo/machine/macbookpro12,1-debian/base/macbook-ssh.sh
[ -r "$runtime/SHA256SUMS" ] && [ -r "$runtime/libSDL3.so.0" ] || {
    echo 'Build SDL 3.4 in the local preparation VM first.' >&2; exit 1;
}
# The SSH helper pins this card's host identity. Also check the OS before writes.
"$ssh" '. /etc/os-release; [ "$ID" = debian ] || exit 1'
tar -C "$runtime" -cf - libSDL3.so.0 SHA256SUMS | "$ssh" '
set -eu
. /etc/os-release
[ "$ID" = debian ] || exit 1
mountpoint -q /media/audiovis-util
stage=$(mktemp -d /media/audiovis-util/.sdl-update.XXXXXX)
trap '\''rm -rf "$stage"'\'' EXIT HUP INT TERM
tar -xf - -C "$stage"
(cd "$stage" && sha256sum -c SHA256SUMS)
mkdir -p /media/audiovis-util/sdl-runtime
cp "$stage/libSDL3.so.0" /media/audiovis-util/sdl-runtime/libSDL3.so.0.new
mv /media/audiovis-util/sdl-runtime/libSDL3.so.0.new /media/audiovis-util/sdl-runtime/libSDL3.so.0
cp "$stage/SHA256SUMS" /media/audiovis-util/sdl-runtime/
install -m 755 "$stage/libSDL3.so.0" /usr/local/lib/libSDL3.so.0.new
mv /usr/local/lib/libSDL3.so.0.new /usr/local/lib/libSDL3.so.0
ldconfig
sync
echo "SDL runtime installed in RAM and retained on the Debian data volume."
'
