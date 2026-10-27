# MacBookPro12,1 Debian headless base

Measured CPU/RAM/GPU use and comparison with Alpine: [RESOURCE-PROFILE.md](RESOURCE-PROFILE.md).

Debian equivalent of the Alpine headless, RAM-root machine. The Alpine
implementation remains preserved at
[`../macbookpro12,1-alpine/`](../macbookpro12,1-alpine/).

The Debian card is prepared as partitions and files, following Alpine's
file-based setup. The verified Debian Live ISO is only a source for stock EFI,
GRUB, kernel, initramfs, and root filesystem files; it is not written raw to
the card. The target has no desktop environment and boots to a text console.

## Prepare the card

The official Debian 13.7 amd64 Standard Live ISO is staged at
`/private/tmp/debian-live-13.7.0-amd64-standard.iso`; its official checksums
are in `/private/tmp/debian-live-13.7.0-standard-SHA512SUMS`. The checksum has
been verified. The standard image includes the Broadcom firmware, `iw`,
`wpa_supplicant`, and Debian Live boot/config tools needed for headless setup.

1. Copy `.env.example` to `.env` in this folder and fill in the local Wi-Fi,
   network, and SSH key path. Keep `.env` private; it is Git-ignored.
2. Prepare offline packages in a disposable **local** Debian VM using
   [local-preparation.md](docs/local-preparation.md), then render the payload:

   ```sh
   machine/macbookpro12,1-debian/base/build-config.sh
   ```

3. Identify the removable card by its external/removable status and capacity,
   then prepare it with the displayed device number:

   ```sh
   machine/macbookpro12,1-debian/prepare-card.sh \
     /private/tmp/debian-live-13.7.0-amd64-standard.iso \
     /private/tmp/debian-live-13.7.0-standard-SHA512SUMS \
     /dev/diskN build/macbookpro12,1-debian/card-payload
   ```

For an already partitioned card, mount its three volumes and apply or update
the generated files without reformatting:

```sh
machine/macbookpro12,1-debian/copy-card-payload.sh \
  /Volumes/EFI /Volumes/DEBIAN /Volumes/audiovis-util \
  build/macbookpro12,1-debian/card-payload
```

The script verifies the ISO checksum, requires an external removable disk,
and asks for an explicit erase confirmation. It creates the `EFI`, `DEBIAN`,
and `audiovis-util` partitions. EFI/GRUB files go on `EFI`; Debian's kernel,
initramfs, and Live root go on `DEBIAN`; full machine scripts, Wi-Fi settings,
SSH public key, reports, and app files go on `audiovis-util`. `DEBIAN` contains
only Debian's files plus a small Live hook that mounts `audiovis-util` and
hands off setup. The data partition uses unjournaled HFS+ to preserve the
requested long volume name and remain readable on macOS. A 64 GB card appears
as about 59 GiB.
Never select the internal macOS drive.

## Boot and connect

1. Insert the card into the Early 2015 MacBook Pro, power on, and hold
   **Option / Alt** to choose its external EFI boot entry.
2. Debian boots the standard headless Live system with `toram`; its operating
   root and runtime writes are in RAM. The main text console automatically
   logs in as root. No XFCE or graphical login is installed or started.
3. The startup hook brings up Broadcom Wi-Fi, installs SSH from the card
   into the temporary RAM root without Internet access, enables key-only SSH, and
   records hardware/network results. The persistent `audiovis-util` partition
   holds reports and startup logs.
4. Connect from the development Mac using the configured host and key:

   ```sh
   machine/macbookpro12,1-debian/base/macbook-ssh.sh
   ```

The startup path is automatic and does not require a local desktop login or a
password prompt on the MacBook. Finish additional machine setup over SSH.
Changes to the Live root disappear at shutdown; keep durable reports, app
files, and other data under `/media/audiovis-util`.

## Preparation boundary

The Alpine target is strictly read-only reference material. All downloads,
builds, package staging, and emulator tests run on the development Mac. The
Debian card is the only physical deployment destination.

## Current progress

The base payload has been applied to the physical card, compared byte for byte,
and safely ejected. A local EFI cold boot with external networking disabled
passed automatic key-only SSH, retained host identity, root tty1 auto-login,
RAM-backed OS/overlay, no swap, read-only OS medium, writable HFS+ data, and
startup-service completion. Physical MacBook boot has now passed the same acceptance checks. Broadcom
Wi-Fi, DNS/HTTPS, and key-only SSH work; console resolution and text grid
match Alpine exactly (2560×1600, 160 columns × 50 rows). The Go app rendered
through OpenGL ES at 2560×1600, accepted trackpad input, and exited cleanly
back to tty1. See [physical validation](docs/physical-validation.md).

The Go binary is staged. SDL 3.4 has been built locally because Debian's SDL 3.2
cannot load the existing Go bindings. The exact Go binding loaded successfully,
KMSDRM/OpenGL ES drivers were enumerated, and the app completed a three-second
offscreen rendering run in the local Debian VM. The updated local payload now
includes this runtime. SDL 3.4.0 and `check-sdl-runtime` have also been copied
to the physical card, verified byte for byte, and safely ejected. Startup
installs the retained library into RAM on every boot. Future runtime updates
can use `base/deploy-runtime.sh` over pinned-key SSH **only on Debian**. The app is a manual launch, as on Alpine.

See [`TODO.md`](TODO.md) for the ordered Alpine-equivalent work plan and
[`docs/first-boot.md`](docs/first-boot.md) for the card boot and validation
sequence. Hardware reports cover CPU/RAM, PCI/USB, firmware, Wi-Fi, input,
ALSA audio, graphics, and storage. The app build/deploy path is under [`app/`](app/).

The conservative OS trim is scripted under [`base/trim/`](base/trim/), with
the exact removal/retention manifests, package checks and metadata validation.
See [`docs/trimmed-root.md`](docs/trimmed-root.md) for rebuilding, updating an
already-running card and rollback. Resource measurements are recorded in
[`RESOURCE-PROFILE.md`](RESOURCE-PROFILE.md).
