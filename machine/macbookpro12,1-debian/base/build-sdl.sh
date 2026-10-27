#!/bin/sh
# Run only in the disposable local Debian amd64 preparation VM.
set -eu
output=${1:?Usage: build-sdl.sh /mounted-output}
archive=$output/sources/SDL3-3.4.0.tar.gz
printf '%s  %s\n' 082cbf5f429e0d80820f68dc2b507a94d4cc1b4e70817b119bbb8ec6a69584b8 "$archive" | sha256sum -c -
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT HUP INT TERM
tar -xzf "$archive" -C "$work"
cmake -S "$work/SDL3-3.4.0" -B "$work/build" \
    -DCMAKE_BUILD_TYPE=Release -DSDL_SHARED=ON -DSDL_STATIC=OFF \
    -DSDL_TESTS=OFF -DSDL_EXAMPLES=OFF -DSDL_X11=OFF -DSDL_WAYLAND=OFF \
    -DSDL_UNIX_CONSOLE_BUILD=ON -DSDL_KMSDRM=ON -DSDL_OPENGLES=ON -DSDL_ALSA=ON \
    -DSDL_PULSEAUDIO=OFF -DSDL_PIPEWIRE=OFF -DSDL_SNDIO=OFF
cmake --build "$work/build" --parallel 2
mkdir -p "$output/sdl-runtime"
cp -L "$work/build/libSDL3.so.0" "$output/sdl-runtime/libSDL3.so.0"
(cd "$output/sdl-runtime" && sha256sum libSDL3.so.0 > SHA256SUMS)
ldd "$output/sdl-runtime/libSDL3.so.0"
