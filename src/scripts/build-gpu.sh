#!/bin/sh
# Compile only the C ABI bridge locally; rendering is implemented in Go.
set -eu
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo=$(CDPATH= cd -- "$here/../.." && pwd)
exp=$repo/src
out=$repo/build/ultralight-eval
if [ -n "${ZIG:-}" ]; then
 zig=$ZIG
else
 [ "$(uname -sm)" = 'Darwin arm64' ] || { echo 'Set ZIG to a Zig 0.15.2 executable for this build host.' >&2; exit 1; }
 archive=$out/zig-0.15.2.tar.xz
 [ -f "$archive" ] || curl -fL https://ziglang.org/download/0.15.2/zig-aarch64-macos-0.15.2.tar.xz -o "$archive"
 printf '%s  %s\n' 3cc2bab367e185cdfb27501c4b30b1b0653c28d9f73df8dc91488e66ece5fa6b "$archive" | shasum -a 256 -c -
 zig=$out/zig-aarch64-macos-0.15.2/zig
 [ -x "$zig" ] || tar -xf "$archive" -C "$out"
fi
"$zig" cc -target x86_64-linux-gnu -shared -fPIC -O2 \
 -I "$out/sdk/include" "$exp/internal/renderer/native/gpu_bridge.c" \
 -o "$out/libgpu_bridge.so"
python3 "$here/prepare-shaders.py" "$out/sdk/shaders/glsl" "$out/gpu-shaders"
