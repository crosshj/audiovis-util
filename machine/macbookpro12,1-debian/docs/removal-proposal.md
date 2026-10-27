# Conservative removal proposal — 2026-10-27

Approved by the user, including the AMD/NVIDIA graphics firmware candidates.
The exact implementation is tracked in [trimmed-root.md](trimmed-root.md) and
`base/trim/remove-packages.txt`. Preserve potentially useful features, including
Bluetooth and peripheral support. The original proposal below records scope.

## Recommended first pass

| Group | Explicit packages | Installed size |
| --- | --- | ---: |
| Printing | cups, cups-browsed, cups-client, cups-common, cups-core-drivers, cups-daemon, cups-filters, cups-filters-core-drivers, cups-ipp-utils, cups-ppdc, cups-server-common, ipp-usb | 19.7 MiB |
| Enterprise networking/storage firmware | firmware-netronome, firmware-marvell-prestera, firmware-bnx2, firmware-bnx2x, firmware-cavium, firmware-myricom, firmware-netxen, firmware-qlogic | 218.3 MiB |
| Server graphics firmware | firmware-ast | 0.04 MiB |

These target printing and specialized hardware absent from this MacBook, while
retaining ordinary USB/network/audio peripheral support. Approximately 238 MiB
of installed package contents are covered by this first pass.

## Additional hardware-specific candidate

`firmware-amd-graphics` and `firmware-nvidia-graphics`: approximately 152.8 MiB
combined. Neither serves the built-in Intel Iris 6100. Removing them gives up
their firmware support if an AMD/NVIDIA external GPU is used later, so keep
this separate from the conservative first pass.

## Keep

- Bluetooth, bluez, bluez-firmware, and potentially shared wireless firmware.
- Broadcom Wi-Fi; Intel graphics, microcode, and relevant audio/Intel firmware.
- Other ordinary Wi-Fi/Ethernet firmware for possible USB adapters.
- ALSA, PipeWire/PulseAudio components, USB audio, codecs and camera/capture support.
- SDL, Mesa OpenGL/EGL/GLES and Vulkan, DRM/GBM, input and console support.
- Avahi/mDNS (network discovery), network diagnostics, SSH and certificates.
- Storage/filesystem tools, SMART monitoring and recovery utilities.
- Compilers, headers and development/debugging tools for future on-device work.
- General fonts, Unicode support, input methods and accessibility support.
- Scanning support: not implicitly included in the decision to remove printing.

Bulk foreign spell-check dictionaries, word lists and translated manuals are
possible later content reductions, but are not part of this first proposal.
Do not assume all non-English or accessibility functionality is unwanted.

## Dependency simulation

A read-only `apt-get --simulate remove` covering all three first-pass groups
plus AMD/NVIDIA firmware proposed 25 removals: the 23 explicitly named packages
and `firmware-linux` / `firmware-linux-nonfree`. Those two are umbrella firmware
metapackages, not the kernel or the required device firmware. No other packages
were selected for removal in that simulation.

APT separately listed some orphan candidates. Do not run blanket `autoremove`:
shared libraries, certificates, accessibility components and future-use tools
must be reviewed individually. The exact inventory and simulation output are
retained in ignored `build/macbookpro12,1-debian/audit/package-removal-*` files.

## Expected RAM effect and implementation boundary

Sizes above are dpkg's **uncompressed installed sizes**, not predicted RAM
savings. The full candidate set is approximately 391 MiB before compression.
Most files are inside the compressed Live root already copied into RAM.
Deleting them in the running overlay would not shrink that lower filesystem.

After approval, make an explicit package removal manifest, produce a smaller
`filesystem.squashfs` locally, and replace that OS file on the Debian card.
Preserve the partition layout, RAM-only operation, networking, console and app
setup. Validate boot and hardware, then repeat the same resource profile to
measure the actual compressed-image/RAM reduction. Do not alter the Alpine
reference system.
