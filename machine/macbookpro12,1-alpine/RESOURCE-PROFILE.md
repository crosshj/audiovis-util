# Alpine resource baseline — 2026-10-27

**App-ready idle → app running:** total CPU **0.17% → 6.25%**, unavailable RAM **646.2 → 743.8 MiB** (**+97.6 MiB**). The app averaged **79.7 MiB RSS** and **18.2% of one CPU core**.

## System comparison

Values are sample means except total CPU, which uses counter differences over
the complete interval. All memory units are MiB (1,048,576 bytes).

| Metric | Headless base, no SDL loaded | Runtime loaded, app stopped | App running | After app exit |
| --- | ---: | ---: | ---: | ---: |
| Total CPU (all 4 CPUs) | 0.19% | 0.17% | 6.25% | 0.18% |
| Unavailable RAM (MiB) | 392.7 | 646.2 | 743.8 | 622.0 |
| Available RAM (MiB) | 7,455.6 | 7,202.1 | 7,104.5 | 7,226.3 |
| Free RAM (MiB) | 7,316.3 | 6,977.1 | 6,879.4 | 7,001.2 |
| Cached pages (MiB) | 387.6 | 714.0 | 785.7 | 716.1 |
| Shared/tmpfs pages (MiB) | 40.9 | 283.1 | 354.8 | 285.1 |
| Swap used | 0 | 0 | 0 | 0 |
| SD-card reads/writes during window | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| Internal-disk reads/writes during window | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |

The distinction between the first two columns matters: this boot initially
lacked `libSDL3.so.0`, and the app could not start. With explicit approval,
the existing signed SDL/Mesa APK cache was loaded **only into the RAM root**.
That added approximately **253.5 MiB** of unavailable RAM before the app
ran. It is runtime/package overhead, not the app's live process footprint.
Package installation reported 269.4 MiB total across 87 installed packages.
The OS and data card partitions stayed read-only; nothing was persisted.

“Unavailable RAM” means `MemTotal − MemAvailable`, a useful capacity measure
that includes kernel reservations and unreclaimable allocations. It is **not**
a sum of process RSS. Cached pages include tmpfs/shmem pages; the memory rows
are overlapping views and must not be added together. The installed runtime
remains in RAM after the app exits. No cache-dropping operation was performed.
Post-exit unavailable RAM fell to 622.0 MiB, below the pre-launch 646.2 MiB:
reclamation and transient allocations affect this system-level measure, so the
97.6 MiB delta is an observed window-to-window difference, not an exact fixed
allocation size or a memory-leak test.

## App and GPU

| Metric while app runs | Mean | Observed range |
| --- | ---: | ---: |
| App CPU, one core = 100% | 18.23% | 16.83–19.80% |
| App resident memory (RSS) | 79.74 MiB | 79.32–80.10 MiB |
| App proportional resident memory (PSS) | 79.13 MiB | 78.71–79.49 MiB |
| App threads | 13 | 13–13 |
| App GPU resident system memory | 69.89 MiB | 69.82–69.95 MiB |

The app's i915 **render-engine busy time was 13.16%** of elapsed time;
copy, video, and video-enhance engine counters were zero. These are this app's
DRM-client counters, not a whole-device power/utilization meter. GPU buffers
use shared system RAM and overlap other memory accounting; do not add GPU
resident memory to RSS or the system-memory total.

The app CPU number uses one logical CPU as 100%; divide by four for its share
of total machine capacity. Total system CPU includes rendering-related kernel
work and the small profiling/SSH overhead, not just the Go process.

## Hardware and workload

- MacBookPro12,1; Intel Core i5-5257U at 2.70 GHz, 2 physical cores / 4 logical CPUs.
- Kernel-visible RAM: 7848.3 MiB; no swap.
- Alpine 3.24.3; kernel `6.18.58-0-lts`; SDL 3.4.10; Mesa 26.1.6.
- Root `/` is tmpfs; `/.modloop` is read-only squashfs. `/media/usb` (Alpine
  FAT volume) and `/media/audio-utility` (HFS+) remained read-only.
- Existing staged Go binary; SHA-256:
  `40cfc8125f3f9c7feb7e79d6fd34b2d8ba1e6aaef0e5e494862e782b64b58573`.
- Fullscreen SDL KMSDRM, `opengles2` renderer, **2560×1600**, legacy pageflip
  (`SDL_KMSDRM_ATOMIC=0`). Trackpad input initialized successfully.
- The existing launcher ran for 90 seconds and returned to the original
  console with exit status 0. Sampling began after a 10-second warmup.
- Cursor-size changes were logged during the run: this is an **interactive
  observed workload**, not a controlled untouched-screen benchmark. Audio
  playback, sustained input stress, power draw, thermals, and frame timing
  were not measured. Do not infer those results from this profile.

## Sampling and reproducibility

The base, runtime-idle, and running windows each contain 31 samples spaced
about two seconds apart (60.23, 60.23, and
60.68 seconds respectively). After exit, a five-second settling
period preceded 16 samples over 30.12 seconds. UTC start times:

| Phase | UTC start |
| --- | --- |
| Base idle | 2026-10-27T18:04:10Z |
| Runtime idle | 2026-10-27T18:07:12Z |
| App running | 2026-10-27T18:08:53Z |
| After exit | 2026-10-27T18:10:33Z |

CPU comes from `/proc/stat` deltas, excluding idle and I/O wait from busy time.
App CPU comes from `/proc/PID/stat` user+system ticks (`CLK_TCK=100`). Memory
comes from `/proc/meminfo`, `/proc/PID/status`, and `smaps_rollup`. GPU counters
come from `/proc/PID/fdinfo`, deduplicated by DRM client ID. Disk I/O comes
from `/proc/diskstats`; installation/startup I/O outside the sample windows
is not included in the zero-I/O rows above.

Before profiling, the earlier 1,284,431,872-byte Debian transfer was removed
from `/tmp/debian-audit/root.squashfs` with explicit approval, so it does not
inflate these results. No profiling packages were installed. The sampler ran
from SSH stdin and all measurements were saved on the development Mac.

Reusable tools:

- [`profile-resources.sh`](base/scripts/profile-resources.sh): read-only sampler.
- [`summarize-resources.py`](base/scripts/summarize-resources.py): local analysis.
- [Raw samples, app logs, and JSON summary](docs/resource-profiles/).

Example from the repository root, with the app in the desired state:

```sh
machine/macbookpro12,1-alpine/base/scripts/macbook-ssh.sh 'sh -s -- 31 2' \
  < machine/macbookpro12,1-alpine/base/scripts/profile-resources.sh > /tmp/profile.txt
python3 machine/macbookpro12,1-alpine/base/scripts/summarize-resources.py /tmp/profile.txt
```

For Debian, use its SSH helper and pass
`APP_PATH=/media/audiovis-util/apps/macbookpro12,1/audio-utility` to the same
sampler. Compare **runtime-ready idle against runtime-ready idle**, then the
same app/display/workload and sampling duration. Keep first-boot package/root
storage overhead separate from the incremental cost of running the app.

The matching [Debian profile and side-by-side comparison](../macbookpro12,1-debian/RESOURCE-PROFILE.md) is now available.
