# Debian resource profile and Alpine comparison — 2026-10-27

**Later update:** the approved trim was deployed and reprofiled. See
[trimmed Debian results](TRIMMED-RESOURCE-PROFILE.md): idle RAM decreased by
213.0 MiB and app-running RAM by 201.6 MiB. The original measurements below
are preserved for comparison.

**Debian idle → app running:** total CPU **0.13% → 5.35%**;
unavailable RAM **1,808.4 → 1,926.0 MiB**
(**+117.6 MiB**). The app averaged **88.7 MiB RSS**.

Compared with the [Alpine profile](../macbookpro12,1-alpine/RESOURCE-PROFILE.md),
Debian's runtime-ready idle baseline used **1,162.2 MiB more unavailable
RAM**. Incremental app memory was **20.0 MiB higher**.
Idle CPU was very low on both. Debian's total running CPU was lower in this
sample, but its app process CPU was higher; these individual interactive runs
do not establish a consistent performance winner.

## Measured comparison

Alpine's comparison column uses its **runtime-loaded idle** baseline, not its
smaller headless state before SDL/Mesa was installed. Figures are means except
total system CPU and GPU busy time, calculated across the complete interval.
`pp` means percentage points. Memory is MiB, not decimal MB.

| Metric | Alpine | Debian | Debian − Alpine |
| --- | ---: | ---: | ---: |
| Idle total CPU | 0.17% | 0.13% | -0.04 pp |
| App-running total CPU | 6.25% | 5.35% | -0.90 pp |
| Idle unavailable RAM | 646.20 MiB | 1,808.40 MiB | +1,162.20 MiB |
| App-running unavailable RAM | 743.80 MiB | 1,925.97 MiB | +1,182.17 MiB |
| Additional RAM while app runs | 97.60 MiB | 117.56 MiB | +19.96 MiB |
| App RSS | 79.74 MiB | 88.73 MiB | +8.99 MiB |
| App PSS | 79.13 MiB | 85.19 MiB | +6.05 MiB |
| App CPU (one core = 100%) | 18.23% | 20.60% | +2.37 pp |
| App GPU render-engine busy | 13.16% | 12.20% | -0.96 pp |
| App GPU resident memory | 69.89 MiB | 69.61 MiB | -0.27 MiB |
| Post-exit total CPU | 0.18% | 0.13% | -0.05 pp |
| Post-exit unavailable RAM | 622.01 MiB | 1,813.46 MiB | +1,191.44 MiB |

Total CPU uses **all four logical CPUs as 100%**. App CPU uses **one logical CPU
as 100%**, so those two rows have different denominators. System CPU also
includes kernel work and sampling/SSH overhead.

## Why Debian's baseline is larger

Debian copies its Live medium into RAM with `toram`. After profiling,
`/run/live/medium` occupied **1,344.6 MiB** of tmpfs, including roughly
**1,225 MiB of compressed root filesystem**, plus initramfs and kernel files.
The writable RAM overlay occupied another **49.8 MiB**. These are observed
filesystem allocations, not figures to add to the memory table: they are
already represented in system memory accounting.

Alpine instead assembled its RAM root from packages. Its runtime-loaded
baseline was about 646 MiB unavailable RAM; its bare headless baseline before
loading SDL/Mesa was about 393 MiB. Debian also has a different kernel,
library stack, and service set. The RAM-resident Live medium is a major
structural difference; this measurement does not isolate every contributor.

Debian's running services included SSH/WPA, systemd services, Avahi, Bluetooth,
cron, CUPS/cups-browsed, and smartmontools. They were left unchanged for this
profile. See the [captured context](docs/resource-profiles/2026-10-27-context.txt).

## Debian detail

- App RSS: **88.04–89.21 MiB**;
  PSS: **84.49–85.67 MiB**.
- App CPU interval range: **18.72–22.55% of one core**;
  **13 threads** during sampling.
- App i915 render-engine busy: **12.20%**. Copy, video, and video-enhance
  counters were zero. These are per-app DRM-client counters, not a power meter.
- No swap. No internal-disk reads/writes in any window. No SD reads; **8 KiB
  of SD writes** occurred during the app window, zero in idle/post-exit windows.
  Debian's data partition is writable; Alpine's was read-only. The write source
  was not traced. Startup/install I/O outside these windows is not included.
- The app exited with status **0**, returned to **tty1**, and post-exit CPU
  returned to **0.13%**. Post-exit unavailable RAM was **1,813.5 MiB**,
  about **5.1 MiB** above the pre-launch sample. A single run is not
  a memory-leak test; caches and transient allocations affect these numbers.

“Unavailable RAM” is `MemTotal − MemAvailable`. It includes kernel reservations
and unreclaimable allocations and is not the sum of process RSS. GPU buffers,
RSS/PSS, tmpfs, and cached pages overlap; do not add those figures together.
Go's reserved virtual address space is not treated as physical RAM use.

## Hardware, software, and workload

Both profiles used the same MacBookPro12,1: Intel Core i5-5257U, two physical
cores / four logical CPUs, 8 GB installed RAM, and Intel Iris 6100 graphics.
The app used fullscreen KMSDRM with the `opengles2` renderer at **2560×1600**,
legacy pageflip (`SDL_KMSDRM_ATOMIC=0`), and the physical trackpad.

| Component | Alpine run | Debian run |
| --- | --- | --- |
| Distribution | Alpine 3.24.3 | Debian 13.7 |
| Kernel | 6.18.58-0-lts | 6.12.107+deb13-amd64 |
| SDL | 3.4.10 | locally built 3.4.0 |
| Mesa | 26.1.6 | 25.0.7 |
| Root | tmpfs package root | squashfs + tmpfs overlay, medium copied to RAM |
| Data volume | HFS+ read-only | HFS+ read-write |

Existing staged binaries were used; no app rebuild, package installation, or
configuration changes were made during profiling. Their hashes differ:

- Alpine: `40cfc8125f3f9c7feb7e79d6fd34b2d8ba1e6aaef0e5e494862e782b64b58573`
- Debian: `64560d4ec03ba700d3812a8449f486f0d60791a598c33e8e76f36cb924ef6bd5`

Both logs show cursor-size changes during measurement. The interaction was
not an identical scripted replay, and Debian had already completed an earlier
app smoke test. This compares the configured systems as observed, not isolated
distribution or compiler performance. Audio playback, power, thermals, and
frame timing were not measured. The nonfatal evdev touchscreen-name warning
appeared on both; trackpad setup and app rendering succeeded.

## Method and artifacts

The same [read-only sampler](../macbookpro12,1-alpine/base/scripts/profile-resources.sh)
and [local summarizer](../macbookpro12,1-alpine/base/scripts/summarize-resources.py)
were used for both systems. Idle and running measurements contain 31 samples
about two seconds apart. The app ran for 90 seconds; sampling started after a
10-second warmup. After exit, a five-second settling period preceded the
30-second follow-up.

| Debian phase | UTC start | Samples | Measured interval |
| --- | --- | ---: | ---: |
| Idle, runtime loaded | 2026-10-27T18:28:59Z | 31 | 60.30 s |
| App running | 2026-10-27T18:30:40Z | 31 | 61.02 s |
| After app exit | 2026-10-27T18:32:03Z | 16 | 30.15 s |

Counters: `/proc/stat`, `/proc/meminfo`, `/proc/PID/stat`, `status`,
`smaps_rollup`, DRM `fdinfo` (deduplicated by client ID), and `/proc/diskstats`.
Every running sample contained the app process; idle/post-exit samples did not.
All raw measurements were streamed to the development Mac.

- [JSON summary](docs/resource-profiles/2026-10-27-summary.json)
- [Idle samples](docs/resource-profiles/2026-10-27-idle.txt)
- [App samples](docs/resource-profiles/2026-10-27-app-running.txt)
- [Post-exit samples](docs/resource-profiles/2026-10-27-after-exit.txt)
- [App log](docs/resource-profiles/2026-10-27-app.log)

Repeat an individual window from the repository root:

```sh
machine/macbookpro12,1-debian/base/macbook-ssh.sh \
  'APP_PATH=/media/audiovis-util/apps/macbookpro12,1/audio-utility sh -s -- 31 2' \
  < machine/macbookpro12,1-alpine/base/scripts/profile-resources.sh > /tmp/debian-profile.txt
python3 machine/macbookpro12,1-alpine/base/scripts/summarize-resources.py /tmp/debian-profile.txt
```
