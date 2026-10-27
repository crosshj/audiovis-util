# First boot: MacBookPro12,1 Debian headless RAM root

This follows the Alpine removable-media workflow: prepare labeled partitions
from verified stock files, run setup automatically at boot, and use SSH for
follow-up work. It does not install Debian onto the internal Mac disk or start
a graphical desktop.

## Card preparation

The Debian card uses three partitions:

- `EFI`: FAT EFI boot files, GRUB, and its configuration only.
- `DEBIAN`: FAT32 Debian Live kernel, initramfs, squashfs, and one small hook
  that mounts the data volume. No full machine scripts or settings belong here.
- `audiovis-util`: unjournaled HFS+ with the full machine scripts, generated
  Wi-Fi settings, SSH public key, persistent reports/logs, and later app files.
  HFS+ preserves the requested long volume name and keeps this volume readable
  on both macOS and Debian.

The source is the checksum-verified Debian 13.7 amd64 Standard Live ISO. The
card preparation script extracts its stock EFI/GRUB/Live files, copies them to
the labeled partitions, and adds the generated machine payload to
`audiovis-util`; it does not write the ISO as a raw disk image. The small hook
on `DEBIAN` only mounts `audiovis-util` and runs the full hook stored there.

Build the network/key payload and prepare the card as described in the machine
[`README.md`](../README.md). Re-check that the selected whole disk is the
external removable card of about 59 GiB before confirming erase. The device
number may change after reinsertion. Do not select the internal macOS disk.

## Boot behavior

1. Insert the prepared SD card and power on while holding **Option / Alt**.
2. Select the card's external EFI entry.
3. GRUB starts Debian Live in text mode with `toram`; the Debian root runs from
   RAM and changes to it are discarded at shutdown.
4. The Live medium hook installs the machine startup service into the RAM root.
   It configures root console auto-login on tty1, connects the built-in
   Broadcom Wi-Fi with the local settings, installs/enables OpenSSH with the
   staged Ed25519 public key, and runs the hardware report. No XFCE login is
   part of the flow.
5. Reports and startup logs are saved to
   `/media/audiovis-util/hardware-reports`. If the data volume is unavailable,
   output remains in RAM for that boot.

Package installation uses verified archives on the card, without network access.
SSH host keys are retained on the data partition across boots. A successful
first boot should print startup progress and the SSH address on the console;
the key is required and SSH password login is disabled.

## Remote follow-up and verification

From the development Mac, run:

```sh
machine/macbookpro12,1-debian/base/macbook-ssh.sh
```

The helper pins the Debian card's generated host key under a separate host alias,
so switching between Alpine and Debian at the same IP does not overwrite or
bypass either identity. Keep `build/macbookpro12,1-debian/ssh-host-keys` when
rebuilding the payload.

Run the acceptance checks without modifying the reference Alpine system:

```sh
machine/macbookpro12,1-debian/base/macbook-ssh.sh 'sh -s' \
  < machine/macbookpro12,1-debian/base/verify-boot.sh
```

Then inspect the latest report under `/media/audiovis-util/hardware-reports`.
Confirm the actual root and boot/data mounts, Wi-Fi association, IPv4 address,
gateway, DNS, HTTPS, Broadcom driver/firmware, Apple keyboard backlight, ALSA,
and Intel Iris 6100 device nodes. Device enumeration alone does not prove that
audio, camera, or graphics features work end-to-end.

After checking these results, shut down cleanly and cold boot again. Confirm
that Wi-Fi, SSH, text-console auto-login, RAM-backed root, and persistent data
volume reporting all work without a desktop session. Do not treat a service
restart as a substitute for the cold-boot check.

The internal macOS disk remains untouched. For recovery or routine setup
changes, return the card to the development Mac, mount the three volumes, and
run `copy-card-payload.sh` to update scripts on `audiovis-util` and the small
bridge hook on `DEBIAN`; repartitioning is unnecessary.

Use the mounted-volume paths and generated payload:

```sh
machine/macbookpro12,1-debian/copy-card-payload.sh \
  /Volumes/EFI /Volumes/DEBIAN /Volumes/audiovis-util \
  build/macbookpro12,1-debian/card-payload
```

## Console text size

The card applies Debian's stock `Lat15-Terminus32x16.psf.gz` to tty1–tty6 after
`console-setup.service`. This is a 16-pixel-wide, 32-pixel-high font, matching
Alpine's observed 160-column × 50-row console at 2560×1600. The font dimensions
were verified from its PSF header. No alternate display mode is forced.

Startup logs and hardware reports record `stty -F /dev/tty1 size` and the
framebuffer resolution. The physical Debian boot confirmed `50 160` and `2560,1600`, matching
Alpine exactly. The
settings are restored from the data partition into the RAM root every boot.
