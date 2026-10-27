# MacBookPro12,1 app target

This folder contains the machine-specific build and test launcher for the Go
workstation app. App source stays in the repository's `src/`; the SD card
stores the compatible compiled binary on its writable data partition.

## Build

From the repository root, run:

```sh
machine/macbookpro12,1-alpine/app/build.sh
```

It produces `build/macbookpro12,1/audio-utility`, a Linux x86_64 musl executable
with the Inter font embedded. The build uses Alpine's Go and C toolchain so
Go's dynamic loader binds to musl instead of glibc. On Linux it loads the
system `libSDL3.so.0`, using the target base's SDL3 build with KMSDRM support.
Text is rasterized in Go; SDL_ttf and a separate installed font package are
not required. On macOS, use Docker or set `CC` to an x86_64-linux-musl compiler.

The same script is called by `.github/workflows/build-app.yml`. The workflow
uploads the target binary as an artifact and publishes it as a release asset
for `v*` tags.

## Deploy and test

With the target MacBook booted and reachable over the configured SSH link:

```sh
machine/macbookpro12,1-alpine/app/deploy-test.sh
machine/macbookpro12,1-alpine/base/scripts/macbook-ssh.sh "sh /media/audio-utility/apps/macbookpro12,1/run-test.sh"
```

The deploy script writes to the writable `audio-utility` partition, not the
read-only Alpine boot partition. The test launcher selects the KMSDRM video
driver and SDL `opengles2` renderer, starts fullscreen on a spare virtual
terminal, displays the app for 30 seconds, records diagnostics in `/tmp`, and
returns to the original console. It reads the app's exit code from the VT
because BusyBox `openvt` itself returns 1 even when its child exits
successfully. The top-bar cursor slider adjusts the pointer during the run and
prints each size change in pixels in the SSH output. It defaults to 60 px, the
center mark on the 32–60–88 px scale. On this target, the app reads the
`bcm5974` trackpad's multitouch events and accumulates finger movement into
cursor deltas. The first finger to touch establishes the motion origin and
controls the pointer until a click is held; while held, the app follows the
finger that is moving so either finger can move the cursor during a drag.
Touching or tapping to click does not teleport the cursor. The app draws its
own visible cursor. The local layout preview remains windowed and keeps normal
macOS cursor behavior.

After boot, run `audio-utility` from the root console to open the app on a
fullscreen virtual terminal. Press Ctrl+C there to exit and return to the
original console. The command runs the app indefinitely and does not enable
app-on-boot.
