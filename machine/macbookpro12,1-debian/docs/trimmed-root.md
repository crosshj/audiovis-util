# Smaller Debian Live filesystem

The approved removal set is tracked in
[`base/trim/remove-packages.txt`](../base/trim/remove-packages.txt): printing,
enterprise networking/storage firmware, ASpeed/AMD/NVIDIA graphics firmware,
and their two umbrella firmware metapackages. No blanket `autoremove` is used.
Bluetooth, ordinary peripheral firmware, Intel graphics/audio, development and
recovery tools, network discovery, and audio/video functionality are retained.

## Background services

The original physical boot ran 14 services. Removing printing eliminates
`cups.service` and `cups-browsed.service`, including their automatic activation
units. `ipp-usb` is also removed; it was installed but not running in the profile.
Keep Bluetooth, Avahi discovery, SMART disk monitoring, cron, D-Bus, the console
login, SSH, journald, logind, udev, the root user manager and wpa_supplicant.
No additional service disabling is included in this conservative removal set.
Firmware and other package removal reduce the compressed OS copied into RAM;
removing running services additionally removes their process memory. Measure
the combined result after reboot rather than adding package sizes and RSS.

## Applied and verified — 2026-10-27

The physical card now boots the trimmed filesystem. The stock file was
1,284,431,872 bytes; the replacement is 1,081,774,080 bytes, **193.27 MiB smaller**.
Its SHA-256 is
`c26eb3afe6de2db98f786bfff782c34acf433116a2f0c97d81cc8aeaea4cde50`.
The live RAM copy matched that checksum after reboot. The verified original
remains on data under `os-backups/6af8bcb58ec0122620c0ccbe59c2e1acf9f07f519950931662642c9916e53b15/`.

The offline local boot test passed before deployment. Physical verification
then passed Wi-Fi, pinned key-only SSH, RAM mounts/no swap, autologin, retained
Bluetooth/Avahi/SMART services, exact package removals and absence of printing
activation units. Running services fell from 14 to 12. Console dimensions
remain 2560×1600 with a 160×50 text grid. The SDL app rendered through OpenGL ES
at 2560×1600, initialized trackpad input, and exited cleanly to tty1.
This does not establish actual audio output or Bluetooth peripheral pairing.

Evidence is retained in [`trim-validation/`](trim-validation/) and
[`resource-profiles/2026-10-27-trimmed-context.txt`](resource-profiles/2026-10-27-trimmed-context.txt).

## Rebuild locally

Use the disposable Debian VM described in `local-preparation.md`, with a blank
Linux scratch disk large enough for the extracted root and compressed output
(8 GiB was allocated). Do not build on the Alpine reference or inside the
physical target's running root. Install `squashfs-tools` in the preparation VM,
format/mount **its virtual scratch disk**, then run:

```sh
bash /work/machine/macbookpro12,1-debian/base/trim/build-trimmed-root.sh \
  /run/live/medium/live/filesystem.squashfs /scratch \
  /work/build/macbookpro12,1-debian/trimmed-root
```

The builder requires a QEMU guest and pins the stock filesystem SHA-256. It
extracts onto a Linux filesystem to preserve ownership, modes, links and
xattrs; checks APT's simulated removals against the explicit manifest; blocks
service starts during removal; temporarily disables initramfs regeneration;
and checks the resulting package set and boot files. Temporary build controls
are restored before creating the image. The firmware packages removed from
the root do not require changing this hardware's known-working kernel/initrd.

The output uses XZ compression with 1 MiB blocks (stock used 128 KiB blocks).
For an emulated x86 guest on an ARM Mac, append `tar` to the builder command
to export `root.tar` after the same checks. On the host, run
`bash base/trim/compress-root-tar.sh build-output-directory` with native
`sqfstar` and `unsquashfs` on PATH (or set `SQFSTAR` and `UNSQUASHFS`). This
compresses directly from the Linux tar metadata; never extract the archive
onto the Mac filesystem. The default builder still compresses inside Linux.
Native compression explicitly preserves numeric owners and sets the root to
`root:root`, mode `0755`; a full archive-to-image ownership/mode comparison
must pass before generating the deployment checksum.

Consequently the measured size change includes both package removal and the
compression-block change. Build logs, package inventories, actual removal list,
protected-package checks, original/trimmed superblock details, file sizes and
SHA-256 are retained under ignored `build/macbookpro12,1-debian/`.

For a fresh card, the optional fifth argument to `prepare-card.sh` selects this
verified `trimmed-root` directory. Without that argument the script uses the
stock ISO filesystem. Routine payload updates do not replace the OS image.

## Update an already-running Debian card

First boot-test the generated image locally with the existing machine payload.
Upload the image to `/media/audiovis-util/os-updates/` through the Debian helper
(which pins the card's SSH host key), then run the repository's
`base/install-root-update.sh` on Debian with the uploaded image path and its
expected SHA-256.

The updater verifies that the running Live medium is already in RAM, identifies
the DEBIAN partition on the same disk as audiovis-util, and temporarily mounts
that OS partition. It backs up the original filesystem under
`audiovis-util/os-backups/<old-sha256>/`, verifies it, writes/verifies a new file,
then replaces the boot filesystem by rename and unmounts the OS partition.
Neither the internal disk nor the EFI partition is modified.

Backups must not be placed alongside `/live/*.squashfs`, where Debian Live may
load them as additional layers, or elsewhere on DEBIAN, where `toram` would
copy them into RAM. Keeping them on data avoids both problems.

The running RAM root remains unchanged until reboot. After reboot, verify the
package manifest, protected features, absent printing services, Wi-Fi/SSH,
console, RAM mounts and app. Repeat the resource profile to measure actual RAM
savings rather than equating uncompressed package sizes with memory saved.

## Rollback

If boot fails, return the card to the development Mac. Copy
`audiovis-util/os-backups/<old-sha256>/filesystem.squashfs` back to
`DEBIAN/live/filesystem.squashfs`, verify it against the backup's `SHA256SUMS`,
sync and eject. The kernel, initramfs, bootloader and machine payload are not
changed by the root-filesystem update.
