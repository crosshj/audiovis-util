#!/bin/sh
# Compile only the C ABI bridge locally; rendering is implemented in Go.
set -eu
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo=$(CDPATH= cd -- "$here/../.." && pwd)
exp=$repo/src
out=$repo/build/ultralight-eval
if [ -n "${ZIG:-}" ]; then
 zig=$ZIG
elif command -v zig >/dev/null 2>&1; then
 zig=$(command -v zig)
elif [ "$(uname -sm)" = 'Darwin arm64' ]; then
 archive=$out/zig-0.15.2.tar.xz
 [ -f "$archive" ] || curl -fL https://ziglang.org/download/0.15.2/zig-aarch64-macos-0.15.2.tar.xz -o "$archive"
 printf '%s  %s\n' 3cc2bab367e185cdfb27501c4b30b1b0653c28d9f73df8dc91488e66ece5fa6b "$archive" | shasum -a 256 -c -
 zig=$out/zig-aarch64-macos-0.15.2/zig
 [ -x "$zig" ] || tar -xf "$archive" -C "$out"
else
 echo 'Zig 0.15.2 is required; add zig to PATH or set ZIG=/path/to/zig.' >&2
 exit 1
fi
zig_version=$("$zig" version)
[ "$zig_version" = '0.15.2' ] || { echo "Zig 0.15.2 is required; found $zig_version." >&2; exit 1; }
"$zig" cc -target x86_64-linux-gnu -shared -fPIC -O2 \
 -I "$out/sdk/include" "$exp/internal/renderer/native/gpu_bridge.c" \
 -o "$out/libgpu_bridge.so"
python3 "$here/prepare-shaders.py" "$out/sdk/shaders/glsl" "$out/gpu-shaders"
