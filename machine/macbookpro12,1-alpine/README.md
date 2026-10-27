# MacBookPro12,1 machine base

Measured idle/app-running CPU, RAM, and GPU baseline: [RESOURCE-PROFILE.md](RESOURCE-PROFILE.md).

This folder contains the machine-specific Alpine Linux base and app target for
the Early 2015 13-inch MacBook Pro. It prepares the hardware and runtime that
the repository's Audio Utility app runs on; product UI and app source remain
under `src/`.

## Contents

- [`roadmap.md`](roadmap.md): the original boot and hardware roadmap, preserved
  as the machine/base roadmap.
- [`docs/first-boot.md`](docs/first-boot.md): SD card layout, verified Alpine
  source image, and first-boot instructions.
- [`base/scripts/`](base/scripts/): Alpine startup, storage, SSH, and hardware
  reporting scripts.
- [`base/build-base.sh`](base/build-base.sh): renders the local machine/network
  settings from this folder's `.env` and assembles a base configuration payload.
- [`app/`](app/): machine-specific Go build, GitHub Actions target, and KMSDRM
  test/deployment instructions for the workstation app.

## Build the configured base payload

1. Copy `.env.example` in this folder to `.env` here and fill in this machine's
   values. The machine-local `.env` is ignored by Git.
2. Run `machine/macbookpro12,1-alpine/base/build-base.sh` from the repository root.
3. Review the generated files under `build/macbookpro12,1/`. The archive contains
   machine startup scripts, a generated Wi-Fi/network configuration overlay,
   the manual `/usr/local/bin/audio-utility` launcher, and the SSH public key.
   It can contain local network credentials, so the entire `build/` directory
   is ignored by Git.
4. Apply the generated files to the existing Alpine base and persist the
   overlay with Alpine's `lbu` utility. Preserve the current EFI layout, kernel,
   initramfs, offline APK repository, and known-good recovery path.

The builder creates a configured payload, not a complete SD image. The verified
Alpine boot files and signed offline package repository are maintained as base
inputs and are not copied from this source tree. This keeps machine firmware,
network settings, and app source in distinct places while allowing the payload
to be rebuilt from the machine-specific `.env`.

The SD card stores the compiled machine app, not a source checkout. The binary
is on the writable `audio-utility` partition; SDL3/Mesa APKs are stored there
too and loaded into the RAM root at each boot. A fresh-card setup must restore
these runtime packages and install a compatible binary from a verified GitHub
release before it is considered equivalent to this card.

## Base behavior and status

The intended base uses Alpine diskless mode: `/` is RAM-backed, and the
read-only `ALPINE` boot volume remains mounted for inspection and access to the
module image. `audio-utility` remains mounted read-write for reports and logs.
The setup commits configuration before remounting `ALPINE` read-only, avoiding
routine writes to the boot volume while keeping both card volumes available.
Wi-Fi and SSH start at boot. The hardware report checks association, addressing,
gateway, DNS, and HTTPS. The Apple keyboard backlight is enabled when the kernel
exposes it.

The current live system has working SSH and Wi-Fi association. Both card data
partitions are mounted now: `ALPINE` is read-only and `audio-utility` is
read-write. The startup change is deployed and its resulting cold-boot behavior
still needs confirmation. See the machine roadmap for the remaining base
milestone.

The live kernel exposes the Intel Iris 6100 through `i915`, with `/dev/dri/card1`,
`renderD128`, and the internal `eDP-1` connector. A live SDL3 KMSDRM fullscreen
smoke test created an OpenGL 4.6 Core context (GLSL 4.60), compiled and linked
shaders, and rendered a triangle at 2560×1600 through Mesa's Intel Iris 6100
driver. The signed SDL3/Mesa runtime package closure belongs on the writable
`audio-utility/base-packages/sdl3-kmsdrm/` directory; startup loads it into RAM.
The compiled test belongs at `audio-utility/graphics-test/sdl3-kmsdrm-smoke` and
is launched with `graphics-test` after boot. It runs on a separate Linux virtual
terminal, keeps diagnostics off the display, then returns to the starting VT
when it ends. The triangle stays up for 30 seconds. SDL3's atomic
pageflip failed on this Mac, so the test explicitly selects SDL's legacy KMS
pageflip path and verifies a yellow center pixel before reporting success. The
visible test is manual and does not start during normal startup.

For a fresh base/data volume, run `stage-graphics-runtime.sh` once after Wi-Fi
is connected; it stores the signed Alpine 3.24 runtime packages on the writable
data volume. On every boot, startup installs those packages into RAM and makes
`graphics-test` available. The checked-in C source and staging helper are in
`base/scripts/` for rebuilding or restaging the test.

The internal macOS disk must remain untouched. Use the recovery procedure in
[`docs/first-boot.md`](docs/first-boot.md) if Alpine fails to start.

After boot, the global `audio-utility` command opens the app on a spare virtual
terminal and returns to the current shell when the app exits. This is a manual
launch; app-on-boot remains disabled.
