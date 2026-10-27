# Project: Repurpose 2015 MacBook Pro as a RAM-Resident Programmable Machine

## Objective

Convert an Early 2015 MacBook Pro into a minimal, remotely programmable Linux machine.

The operating system should boot from removable media, load into RAM, and operate without requiring continued reads or writes to the boot device.

The machine will eventually run arbitrary custom software developed by coding agents and delivered over the network. Its first expected use is as a programmable music workstation connected to physical synthesizers, but **do not design specifically for music software**.

The immediate task is to establish a reliable boot and networking foundation.

## Target hardware

- MacBook Pro (Retina, 13-inch, Early 2015)
- Model identifier: MacBookPro12,1
- Intel Core i5, 2.7 GHz
- 8 GB RAM
- Intel Iris Graphics 6100
- Broadcom BCM43602 Wi-Fi (expected PCI ID `14e4:43ba`; verify)
- Currently running macOS High Sierra 10.13.6

## Architectural principles

1. Minimize operating system overhead.
2. Preserve as much CPU and RAM as reasonably possible for user programs.
3. Prefer standard Linux capabilities over custom infrastructure.
4. Avoid desktop environments and unnecessary background services.
5. Do not install anything on or overwrite the internal macOS drive.
6. Treat removable media as a boot image, not a working filesystem.
7. Keep the initial system simple and disposable.
8. Expect future software development and deployment over the LAN.
9. Favor flexibility over premature architecture or optimization.
10. Do not build security, orchestration, sandboxing, or fault-tolerance frameworks.

The user accepts that failed experiments may require physically power-cycling the machine.

## Phase 1 — Prepare boot media

Use the latest stable Alpine Linux x86_64 release.

Reference:

https://wiki.alpinelinux.org/wiki/Installation

https://wiki.alpinelinux.org/wiki/Diskless_Mode

https://wiki.alpinelinux.org/wiki/Create_a_Bootable_Device

Tasks:

- Identify the removable media being used.
- Download the appropriate Alpine boot image from an official source.
- Verify its checksum.
- Prepare bootable media compatible with the MacBook's Intel EFI firmware.
- Use an appropriately supported media creation method.
- Ensure the internal macOS installation remains untouched.
- Provide clear instructions for selecting the external boot device using the Mac's Option/Alt startup menu.

Start with the simplest bootable image. Do not customize a Linux distribution before verifying that the stock environment boots.

If the MacBook cannot boot directly from its internal SD reader, investigate the limitation and use a USB-connected card reader or USB boot media for testing.

**Deliverable:** Bootable removable media and a successful Alpine Linux console on the MacBook.

## Phase 2 — Identify hardware and establish Wi-Fi

Wi-Fi is the immediate priority.

The MacBook is expected to have a Broadcom BCM43602 chipset, supported by Linux's `brcmfmac` driver with appropriate firmware.

Do not assume the firmware is present or that the chipset identity is correct.

Investigate:

- Actual wireless chipset and PCI ID.
- Kernel driver attachment.
- Firmware files requested by the driver.
- Any firmware loading failures.
- Availability of the wireless interface.
- Successful association with a WPA2/WPA3 access point.
- DHCP address assignment.
- LAN connectivity.
- DNS and external connectivity.

Initial diagnostic commands may include:

```sh
lspci -nn -k
ip link
dmesg | grep -Ei 'brcm|firmware|wifi|wlan'
lsmod | grep brcm
```

Install diagnostic tools if necessary.

Prefer the in-kernel `brcmfmac` driver with the appropriate Broadcom firmware. Do not immediately introduce proprietary out-of-tree drivers.

Consult:

https://wiki.alpinelinux.org/wiki/Wi-Fi

https://wireless.docs.kernel.org/en/latest/en/users/drivers/brcm80211.html

If Wi-Fi does not work with the stock Alpine image:

1. Identify exactly what is missing.
2. Investigate the appropriate Alpine firmware packages.
3. Obtain firmware through a temporary network connection or stage it on removable media.
4. Load the firmware and driver.
5. Confirm the device becomes operational.
6. Document the requirements for including the fix in the permanent boot image.

If board-specific NVRAM or calibration data is required, verify its source and compatibility before installing it.

Do not abandon Alpine over an initial missing-firmware problem. However, do not become committed to Alpine if hardware compatibility requires substantial maintenance or compromises our goals.

**Deliverable:** Working Wi-Fi using the MacBook's built-in wireless hardware.

## Phase 3 — Establish remote access

Once Wi-Fi works:

- Configure the machine to connect to the local network.
- Install and enable a minimal SSH server.
- Establish SSH access from the primary development computer.
- Use the current LAN address `192.168.1.38` for the MacBook and `192.168.1.5`
  for the development Mac; connect directly over SSH with the staged Ed25519 key.
- Verify basic remote commands work reliably.

The machine will remain on a trusted local network.

Avoid elaborate authentication infrastructure. SSH keys are sufficient.

**Deliverable:** The coding agent can interact with the Linux machine remotely through SSH.

## Phase 4 — Make the environment rebootable and RAM-resident

Storage clarification: “diskless” means the Alpine root runs in RAM and normal
operation avoids unnecessary writes to the SD card. Keep `ALPINE` mounted
read-only after configuration is committed, and keep `audio-utility` mounted
read-write for reports and logs. The earlier unmount wording is superseded; an
unaltered copy is preserved at `docs/archive/machine-roadmap-unmount-wording.md`.

The initial Alpine installation may be temporary. The working configuration must eventually survive reboot.

Use Alpine's diskless mode and configuration persistence mechanisms where appropriate.

Ensure:

- Wi-Fi firmware is available during boot.
- Network configuration survives reboot.
- SSH starts automatically.
- Necessary packages and modules are available without downloading them again.
- The working operating environment is loaded into RAM.
- The `ALPINE` volume remains mounted read-only after boot configuration is saved.
- The `audio-utility` volume remains mounted read-write for reports and logs.
- Ordinary runtime logs and temporary files do not cause writes to the boot volume.

A small persistent boot configuration is acceptable. Normal execution should
not write to the boot volume after setup; the data volume remains mounted for
reports and logs.

If updating the boot image is necessary, it may be done deliberately, rather than continuously.

Validate using a complete shutdown and cold boot, not just a service restart.

**Deliverable:** Power on → boot removable media → load Linux into RAM → connect to Wi-Fi → accept SSH connections.

## Phase 5 — Measure and document

Once the preceding phases succeed, collect:

- Idle RAM usage.
- Total available RAM.
- Idle CPU activity.
- Running services.
- Active kernel modules.
- Mounted filesystems.
- Network connectivity.
- Boot-media access after startup.

Identify obvious unnecessary resource consumption, but do not aggressively strip services or drivers at the expense of useful hardware support.

Create a concise project README containing:

- How the boot media was prepared.
- How the MacBook boots from it.
- Any Broadcom firmware requirements.
- How to connect over SSH.
- How to regenerate or update the boot image.
- How configuration persistence works.
- How to recover from a broken running environment.

## Explicitly out of scope

Do not implement:

- Custom HTTP APIs.
- Program upload or execution protocols.
- Custom application frameworks.
- DAW or MIDI software.
- Graphics interfaces.
- Automatic Git deployment.
- Container infrastructure.
- Sandboxes.
- Advanced monitoring.
- Custom Linux kernels.
- Autonomous repair mechanisms.

These decisions will be made as practical needs emerge.

## Working style

Proceed incrementally.

Prioritize Phase 1 and Phase 2. Do not spend time speculating about future software architecture.

Inspect the actual hardware and report findings rather than assuming compatibility.

Prefer established tools and documented practices.

Do not modify the MacBook's internal drive.

Ask for user intervention only when physical interaction is required, such as inserting media, selecting the boot device, entering Wi-Fi credentials, or checking the MacBook screen.

## Definition of success

The project reaches its initial milestone when:

**I can power on my old MacBook Pro, boot Alpine Linux from removable media, have it operate from RAM, automatically join my Wi-Fi network, and connect to it remotely over SSH.**

At that point, stop.

Future capabilities will be developed against this foundation as needed.
