# MacBook first hardware test

Prepared for the Intel MacBookPro12,1 using Alpine Linux Standard 3.24.3 x86_64.
Alpine has booted several times. The keyboard backlight now works, and reports
confirm the Broadcom driver, `wlan0`, audio devices, and Apple SMC are available.
The latest full report confirms Wi-Fi association, IPv4, gateway, public IP, DNS,
and HTTPS all passed on the configured access point. An earlier startup log showed that the offline
APK repository path was wrong, so the Wi-Fi service was not installed on that boot;
the path is now corrected. The kernel still reports missing MacBookPro12,1 specific
Broadcom calibration files, but the latest connection test succeeded. The startup
log and reports go to `audio-utility`.
Only boot files, offline packages, the test script, and Alpine's configuration
overlay belong on `ALPINE`. If `audio-utility` cannot be mounted writable, checks
continue and logs fall back to `/tmp`; the scripts do not use the boot partition
for reports.

## SD card

The original `Untitled` card was identified as a 63,864,569,856-byte removable
USB device, `/dev/disk5` at preparation time. Device numbers can change.

The card uses a GUID partition map:

- EFI system partition: the removable-media EFI bootloader.
- `ALPINE`: approximately 1 GB FAT32, holding Alpine's boot files and offline tools.
- `audio-utility`: the remaining space, HFS+ without journaling, holding instructions,
  startup logs, and hardware reports after testing.

FAT32 and macOS's ExFAT tools limit volume names to 11 characters. The separate
HFS+ volume preserves the requested `audio-utility` name. Alpine's stock kernel
configuration includes the HFS+ module.

The source image is from [Alpine's official downloads](https://alpinelinux.org/downloads/):

```text
https://dl-cdn.alpinelinux.org/alpine/v3.24/releases/x86_64/alpine-standard-3.24.3-x86_64.iso
SHA-256: 1629ffe7308d4f3d3eeeed3a97d09ad62293f6d37ef3e483a7306077fdac12be
```

The checksum was verified against the official `.sha256` file. The stock kernel,
initramfs, module/firmware image, and package repository are copied from this ISO.
The EFI loader's embedded volume-label search is changed from the ISO label to
`ALPINE`, with its embedded configuration kept at the same byte length. The loader
is stored only in the EFI system partition at `/EFI/BOOT/BOOTX64.EFI`.
No custom kernel or root filesystem is built.

All 25 offline diagnostic APK signatures and their signed payload hashes were
verified against the public keys shipped in the stock image. The copied stock
files and offline APKs were also compared by SHA-256 after writing.

The preparation follows Alpine's [manual file-copy media layout](https://wiki.alpinelinux.org/wiki/Create_a_Bootable_Device#Manually_copying_Alpine_files).
Only EFI boot is configured; a legacy BIOS bootloader is unnecessary for this Mac.

## Boot the MacBook

1. Safely eject the whole card before moving it to the MacBook.
2. Shut down the 2015 MacBook and insert the card.
3. Press power and immediately hold **Option / Alt** until the startup picker appears.
4. Select the card's **EFI Boot** entry, then choose `Linux lts` if GRUB appears.
5. Wait for visible startup progress and the concise hardware report summary.
   The full startup log and report are saved under `audio-utility/hardware-reports`.
6. The main console logs in automatically as `root`. `/etc/motd` shows the latest
   network results and shutdown steps.
7. Type `poweroff` and wait for the MacBook to halt before removing the card.

If the card does not appear in the startup picker, try the same card in a USB
card reader. That tests a possible internal-reader boot limitation separately
from Linux hardware support.

## Automatic hardware report

The card runs `hardware-test.sh` automatically during Alpine startup. It installs
PCI, USB, ALSA audio, and Wi-Fi diagnostic tools from signed local APKs into RAM,
then records device IDs, attached drivers, firmware messages, wireless interfaces
and capabilities, sound devices, input devices, graphics/camera device nodes,
RAM, modules, and mounted filesystems.

It saves reports and startup logs to `audio-utility/hardware-reports`. If that
partition cannot be mounted writable, the report remains in RAM for the current
boot. The `ALPINE` boot volume is not used for report or log output.

After startup, `/` remains in RAM, `ALPINE` stays mounted read-only, and
`audio-utility` stays mounted read-write. The module image remains available
through the read-only boot volume. Configuration is committed to `ALPINE`
before it is remounted read-only; routine reports and logs go to the data
volume instead.

The report also waits for Wi-Fi association and an IPv4 address, then checks the LAN gateway,
public IP reachability, DNS resolution, and an HTTPS fetch. A failed ping alone
does not prove the Internet is unavailable, since some networks block ICMP; the
DNS and HTTPS results provide separate checks. Detection is evidence to inspect,
not a blanket compatibility verdict. Camera and audio device enumeration does
not establish usable video or sound.

## Wi-Fi and keyboard backlight setup

On boot, a setup hook installs signed Wi-Fi and OpenSSH packages, including the
OpenRC service scripts, from the card's local repository. It configures `wlan0`
using the SSID, address, and gateway rendered from the machine-specific `.env`, starts the
supplicant and networking service, and retries the connection on each boot. It
also enables the Apple SMC keyboard-backlight driver and saves the configuration
with Alpine local backup.
The Wi-Fi settings are rendered from the ignored machine-specific `.env` into the local
base payload. The generated payload and Alpine overlay contain local network
configuration and must not be checked in.
The main console uses root autologin for this dedicated hardware test card; other
virtual consoles keep the normal login prompt.
The overlay keeps only a small startup launcher; the full setup script lives
once on `ALPINE` alongside the report and storage scripts.
After Wi-Fi connects, SSH is available at the host configured in `.env` using the
matching Ed25519 key on the development Mac. Password login and forwarding are
disabled. Use `machine/macbookpro12,1-alpine/base/scripts/macbook-ssh.sh` from the
development Mac; it reads the host from `.env` and uses the matching local key.

The boot output and MOTD show the SSH address.
Alpine's local backup is written as `localhost.apkovl.tar.gz` on `ALPINE`. The
hook prints progress during startup, runs the hardware report, and saves the
report and completed startup log to `audio-utility`. OpenRC's `local` service
has verbose output enabled so progress and the final summary appear during
boot. The same results are shown in `/etc/motd` after login.

After startup finishes, `/etc/motd` and the hardware report show the connection
checks. The planned RAM-resident flow exposes `hardware-test` for a manual
rerun; check the project's README for its deployment and cold-boot status. You
can also check interactively:

```sh
ip address show wlan0
iw dev wlan0 link
ip route
ping -c 3 1.1.1.1
nslookup alpinelinux.org
```

Also ping the LAN gateway shown by `ip route`. The script saves its configuration
to the SD card using `lbu`; no internal MacBook disk is selected or modified.

For this test, do not run `setup-alpine` or `setup-disk` or select the internal
drive for installation. End the session with `poweroff`, then move the card back
to the development Mac to read the reports.
