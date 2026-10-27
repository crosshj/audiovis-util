#!/bin/sh
# Build the Debian amd64 dashboard bundle on macOS or Linux.
set -eu
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo=$(CDPATH= cd -- "$here/../.." && pwd)
exp=$repo/src
out=${1:-$repo/build/macbookpro12,1-debian}
cache=$repo/build/ultralight-eval
mkdir -p "$out" "$cache"
archive=$cache/sdk-1.4.0.7z
sha=75e3079a9bcae834e98607af0d8ec6ef8f8144abc674ce961cf878fb2d816ddd
[ -f "$archive" ] || curl -fL 'https://ultralig.ht/api/v1/sdk/download?platform=linux-x64&version=1.4.0' -o "$archive"
printf '%s  %s\n' "$sha" "$archive" | shasum -a 256 -c -
# The legacy release API publishes no checksum. This pins the first retrieved
# official archive; a changed archive must be reviewed before changing this pin.
mkdir -p "$cache/sdk"
tar -xf "$archive" -C "$cache/sdk"
"$here/build-gpu.sh"
cd "$repo/src"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -buildid=' -o "$out/audio-utility" ./app
rm -rf "$out/ui"
cp -R "$exp/ui" "$out/ui"
cp "$here/run-target.sh" "$out/"
python3 "$here/prepare-assets.py" "$out"
cp -R "$cache/sdk/bin" "$cache/sdk/resources" "$cache/sdk/license" "$out/"
cp "$cache/libgpu_bridge.so" "$out/"
cp -R "$cache/gpu-shaders" "$out/"
tar -czf "$out/runtime.tar.gz" -C "$out" bin resources license
tar -czf "$out/audio-utility-debian-amd64.tar.gz" -C "$out" audio-utility ui run-target.sh bin resources license libgpu_bridge.so gpu-shaders
printf 'Built %s\n' "$out/audio-utility"
