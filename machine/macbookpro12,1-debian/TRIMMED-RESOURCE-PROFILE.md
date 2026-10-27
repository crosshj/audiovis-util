# Debian after conservative trim — 2026-10-27

The applied trim reduced idle unavailable RAM by **213.0 MiB**
and app-running unavailable RAM by **201.6 MiB**.
The compressed OS file is **193.27 MiB smaller**. Printing services were removed;
Bluetooth, discovery, SMART monitoring, networking, SSH and the graphics/audio
stack were retained. The physical machine rebooted successfully.

## Before and after

| Metric | Original Debian | Trimmed Debian | Reduction |
| --- | ---: | ---: | ---: |
| Idle unavailable RAM | 1,808.40 MiB | 1,595.45 MiB | 212.96 MiB |
| App-running unavailable RAM | 1,925.97 MiB | 1,724.41 MiB | 201.56 MiB |
| Post-exit unavailable RAM | 1,813.46 MiB | 1,608.21 MiB | 205.24 MiB |

| CPU / app metric | Original Debian | Trimmed Debian |
| --- | ---: | ---: |
| Idle total CPU | 0.13% | 0.15% |
| App-running total CPU | 5.35% | 5.48% |
| Post-exit total CPU | 0.13% | 0.13% |
| App CPU (one core = 100%) | 20.60% | 20.39% |
| App RSS | 88.73 MiB | 89.47 MiB |
| App PSS | 85.19 MiB | 86.20 MiB |
| App GPU render-engine busy | 12.20% | 12.76% |
| Running services | 14 | 12 |

The app's incremental unavailable RAM was **129.0 MiB**.
It exited with status 0 and restored tty1. Post-exit unavailable RAM was
**12.8 MiB** above pre-launch idle; caches and transient
allocations affect this, and a single run is not a leak test. CPU results are
similar; these interactive samples do not establish a performance improvement.

There was **no swap and no internal-disk or SD-card I/O in any of the three
trimmed measurement windows**. Installation, boot and earlier validation I/O
are outside those windows.

## What changed

Exactly 25 packages were purged: printing, the approved enterprise NIC/storage
firmware, ASpeed/AMD/NVIDIA graphics firmware, and two firmware metapackages.
No blanket autoremove was used. `cups` and `cups-browsed` no longer run, and
printing activation units are absent. Bluetooth, Avahi and SMART services
were confirmed active after reboot. Peripheral pairing and audible output
were not tested by this resource profile.

The root file decreased from 1,284,431,872 to 1,081,774,080 bytes. Compression
also changed from XZ 128 KiB blocks to XZ 1 MiB blocks, so the reduction is not
attributable exclusively to removed packages. The RAM OS medium occupies
**1,151.38 MiB**, versus 1,344.6 MiB before; the writable overlay was
**47.37 MiB**, versus 49.8 MiB before. These allocations are already included
in memory accounting; do not add them to the table. The measured system
reduction additionally reflects removed services and other boot/cache
variation. It does not isolate the exact RAM cost of each service.

The original root is retained on the data partition for rollback, outside the
OS medium copied into RAM. See [build/update/rollback instructions](docs/trimmed-root.md)
and [validation evidence](docs/trim-validation/).

## Comparability and method

Same physical MacBookPro12,1, kernel, SDL 3.4.0, Mesa stack and Go binary as the
original Debian profile. The app SHA-256 remains
`64560d4ec03ba700d3812a8449f486f0d60791a598c33e8e76f36cb924ef6bd5`.
Both use KMSDRM/OpenGL ES at 2560×1600 with the physical trackpad. Console
resolution and text grid remain 2560×1600 and 160×50, matching Alpine.

A brief physical app smoke test preceded the trimmed idle measurement. Idle
and running windows each contain 31 samples approximately two seconds apart;
the app ran for 90 seconds with sampling after warmup. A 16-sample post-exit
window followed. User cursor-size changes appear in the app log, as they did
in the original profile; interaction was not an identical scripted replay.

Unavailable RAM is `MemTotal − MemAvailable`, not summed process RSS. System
CPU uses all four logical CPUs as 100%; app CPU uses one core as 100%. The same
read-only sampler and summarizer used for Alpine and original Debian were
used here. The Alpine runtime-loaded idle baseline remains 646.20 MiB, so
trimmed Debian still uses about **949.2 MiB more** at idle.
This conservative trim does not make the different RAM-root architectures
identical.

## Raw evidence

- [JSON summary](docs/resource-profiles/2026-10-27-trimmed-summary.json)
- [Idle samples](docs/resource-profiles/2026-10-27-trimmed-idle.txt)
- [App-running samples](docs/resource-profiles/2026-10-27-trimmed-app-running.txt)
- [Post-exit samples](docs/resource-profiles/2026-10-27-trimmed-after-exit.txt)
- [App log](docs/resource-profiles/2026-10-27-trimmed-app.log)
- [Boot checksum, mounts, services and console](docs/resource-profiles/2026-10-27-trimmed-context.txt)
- [Original Debian/Alpine comparison](RESOURCE-PROFILE.md)
