# Debian equivalent of the Alpine RAM-only machine

## Boundary

The running Alpine target is strictly **read-only reference material**. Never
stage files, install packages, or run preparation jobs on it. All preparation
runs on the development Mac or a disposable local Debian VM. The inserted
Debian SD card is the only physical deployment destination. Do not modify the
MacBook's internal disk.

## Alpine behavior to preserve

| Behavior | Debian implementation |
| --- | --- |
| File-based removable boot | Stock signed EFI/GRUB files on `EFI` |
| Distribution files on OS partition | Kernel, initramfs, squashfs on `DEBIAN` |
| RAM-only operating root | Live `toram`, tmpfs writable overlay, `noswap` |
| Automatic startup setup | Explicit local Live hook, then systemd service |
| Offline bootstrap packages | APT-verified `.deb` closures on data; `dpkg` at boot |
| Broadcom Wi-Fi | Stock firmware/brcmfmac, discovered interface, wpa_supplicant |
| Fixed address and key-only root SSH | Generated private settings, retained SSH host key |
| Root text-console auto-login | Final tty1 getty override, multi-user target |
| Apple keyboard backlight | applesmc control in startup |
| Persistent scripts, apps, reports | Unjournaled HFS+ `audiovis-util`, mounted read-write |
| Manual fullscreen Go/SDL app | `audio-utility`, KMSDRM, EGL/GLES, legacy pageflip |

## 1. Audit and source verification

- [x] Inspect Alpine scripts and read-only SSH runtime: root is tmpfs; Alpine OS
  and its data volume currently mount read-only. Do not assume data is writable.
- [x] Identify physical Debian card: 63,864,569,856 bytes; EFI / DEBIAN /
  audiovis-util. Reidentify before every write; device numbers can change.
- [x] Verify Debian 13.7 amd64 Standard Live source checksum.
- [x] Compare card kernel, initramfs, squashfs against extracted source hashes.
- [x] Inspect signed GRUB's embedded configuration lookup; deploy the same menu
  to its main, architecture-specific, and EFI fallback paths.

## 2. Script preparation locally

- [x] Stage signed offline SSH dependency closure against the exact Live root.
- [x] Retain SSH host keys across payload rebuilds and card boots.
- [x] Keep credentials and generated artifacts excluded from Git.
- [x] Discover wireless interface and wait for driver/firmware initialization.
- [x] Queue startup asynchronously from early live-config; install packages
  without starting services before the key-only SSH configuration is in place.
- [x] Fix hook path: this image's `hooks=medium` searches a nonexistent
  `/lib/live/mount/medium`; use explicit `file:///run/live/medium/...` instead.
- [x] Fix console override precedence over Debian's generated autologin setting.
- [x] Guard mounted-volume updates: same whole removable disk, expected labels,
  generated file comparison, sync and eject. Routine updates do not reformat.
- [x] Build the existing Go app for Linux amd64.
- [x] Identify SDL mismatch: bindings need 3.4; Debian's 3.2 fails symbol loading.
- [x] Build and validate SDL 3.4 against Debian with KMSDRM/EGL/GLES/ALSA.

## 3. Local validation and card deployment

- [x] Boot stock signed EFI loader, custom GRUB menu, and copied Debian OS in
  a local VM with matching FAT/HFS+ partition layout.
- [x] Exercise startup hook: offline package install, key-only SSH, RAM overlay,
  read-only RAM OS medium, writable HFS+ data, persistent reports.
- [x] Cold-boot automatically with corrected hook and Internet disabled; verify
  SSH host identity, root tty1 autologin, service completion, mount behavior.
- [x] Apply the tested base payload to the physical SD card and compare files.
- [x] Safely eject the physical card for its MacBook hardware test.

- [x] Add the finished SDL 3.4 runtime and checker to the reinserted physical
  card; verify every copied file and the library SHA-256, then safely eject.
  The existing startup script installs this library into RAM on every boot.

## 4. Physical first boot (requires moving the card)

- [x] Boot the external EFI entry using Option/Alt; no desktop or manual setup.
- [x] Verify Wi-Fi association, static address, route/DNS, key-only SSH.
- [x] Confirm MacBookPro12,1 identity, Broadcom firmware, Apple keyboard,
  Intel Iris graphics, ALSA audio, and automatic hardware report.
- [x] Verify RAM-backed root, no swap, read-only OS medium, writable data.
- [x] Stage 16×32 console font on tty1–tty6 after Debian console setup,
  matching Alpine's measured character size. Verify copied card scripts.
- [x] Confirm physical Debian resolution/text grid matches Alpine's
  2560×1600 / 160 columns × 50 rows, plus root autologin and backlight.

## 5. App and repeat boot

- [x] Run the staged `audio-utility` launcher over SSH; verify the physical
  display, touchpad/input, OpenGL ES renderer, and restoration of the console.
- [ ] Check actual audio output; enumeration alone is insufficient.
- [ ] Clean shutdown and cold boot again; repeat SSH/storage checks.
- [x] Record physical acceptance results and any remaining hardware differences.

A local emulator cannot establish Broadcom Wi-Fi association or physical Intel
KMSDRM/audio behavior. Do not mark physical acceptance complete from VM results.
See `docs/local-preparation.md` for reproducible package/runtime builds and
`docs/first-boot.md` for deployment and hardware validation.

## 6. Approved conservative trim

- [x] Record original Debian and Alpine resource profiles.
- [x] Pin explicit removals: printing, enterprise NIC/storage firmware,
  unrelated ASpeed/AMD/NVIDIA graphics firmware and umbrella metapackages.
- [x] Purge in a disposable local Linux filesystem; verify exact package diff,
  retained features, clean dpkg audit and unchanged kernel files.
- [x] Script native compression from a Linux metadata-preserving tar archive.
- [x] Boot-test the smaller root with the offline machine payload.
- [x] Back up and replace the physical Debian card's compressed root over SSH.
- [x] Reboot and verify Wi-Fi, SSH, Bluetooth, Intel graphics/SDL, console and RAM root.
- [x] Reprofile idle, running app and post-exit memory/CPU; record measured savings.
