# Ultralight CPU / GPU comparison

**Current implementation is GPU-only.** Following the comparison below, the
CPU renderer, renderer-selection flags, F2/button switching, CPU smoke test and
comparison runner were removed. The measurements and validation logs below
are retained as historical evidence for that decision.

The GPU-only build passed an eight-second native-resolution hardware test with
animation, pointer/text assertions, Go callbacks and cursor synchronization:
[validation log](validation/gpu-only.log),
[framebuffer capture](validation/gpu-only-screen.png). The capture run includes
PNG/readback overhead and is not a replacement performance benchmark. Build,
Linux `go vet`, shell syntax and whitespace checks passed. The updated demo was
left running without renderer-selection arguments or switching controls.

## Run the GPU demo

```sh
machine/macbookpro12,1-debian/base/macbook-ssh.sh \
  'cd /tmp/ultralight-eval && ./run-target.sh --full --animate-html --seconds 0'
```

The demo requires GPU acceleration. The status line shows FPS and CPU usage
(100% = one logical core). Toggle HTML animation with the button; Escape returns
to the console. See the [current build/run instructions](../../../src/README.md).

## Matched measurements — 2026-10-27

Physical MacBookPro12,1, Intel Iris 6100, Debian 13.7, SDL 3.4.0,
Mesa 25.0.7, GLES 3.2, Ultralight 1.4.0.864dc81. Both modes request a GLES 3
context. Each sample lasted 20 seconds at 2560×1600, using the same binary and
page. Benchmark mode ignored physical input except Escape, and screenshots were
disabled. The former comparison runner generated these samples before its removal.

| Workload | HTML renderer | Mean FPS | CPU / one core | Peak process RSS | Frames over 25 ms |
| --- | --- | ---: | ---: | ---: | ---: |
| Static | CPU | 59.42 | 13.25% | 192.21 MiB | 1 |
| Static | GPU | 59.43 | 14.10% | 155.97 MiB | 1 |
| CSS + native animation | CPU | 59.36 | 66.43% | 190.64 MiB | 6 |
| CSS + native animation | GPU | 59.38 | 34.44% | 159.16 MiB | 1 |

For the animated case, GPU mode reduced measured process CPU by **48%** and peak
process RSS by **31.48 MiB**. Static CPU cost was similar. Neither result implies
that every page benefits by the same amount.

CPU mode uploaded the full HTML bitmap **610 times**, averaging **12.09 ms** per
upload call. GPU mode made **zero full HTML bitmap uploads**: it submitted
**21,027 draw commands** across **596 command-list updates**, with about
**14.16 MB of glyph/image texture uploads** over the run. Geometry uploads also
occur and are not included in that texture-byte counter. Its CPU-side draw
submission averaged **0.89 ms per display frame**. These timings have different
denominators and are not a claim of a 12.09/0.89 overall speedup.

The call to Ultralight's render function averaged about **2.44 ms per frame in
both modes**. GPU mode removes the expensive full-surface transfer while layout,
JavaScript, command generation and other CPU work remain. HTML updates were
approximately 30 Hz, even though display presentation was approximately 59 Hz.

Raw logs: [CPU static](validation/comparison/cpu-static.log),
[GPU static](validation/comparison/gpu-static.log),
[CPU animated](validation/comparison/cpu-animated.log),
[GPU animated](validation/comparison/gpu-animated.log).

CPU measurements cover the timed application loop, while peak RSS includes
initialization. RSS is process memory, not total machine RAM or separately staged
SDK files in tmpfs. GL submission timings are CPU timings, not GPU timer queries.
Frames over 25 ms are not exact missed-vblank counts. These short runs do not
establish long-term memory behavior or audio scheduling reliability.

## Implementation and provenance

- `gpu.go` implements the Ultralight 1.4 GPUDriver C callbacks in Go: textures,
  render buffers, vertex/index geometry, clipping, uniforms and draw lists.
- Resource/vertex/uniform conventions were checked against the official
  [AppCore 1.4 reference driver](https://github.com/ultralight-ux/AppCore/blob/52d4b918250ed4410c5e0cf2de25b668b4327e90/src/linux/gl/GPUDriverGL.cpp).
  This is a GLES implementation using SDL's existing context; it does not bring
  in GLFW window creation or an AppCore event loop.
- The four GLSL shaders come from the checksum-pinned SDK archive already used
  for the CPU evaluation. Build-time extraction adapts GLSL 150 to GLSL ES 300,
  adds precision declarations and fixes one integer literal requiring explicit
  float conversion. Generated SDK shader files stay outside Git.
- `native/gpu_bridge.c` only adapts struct-valued callbacks and forwards the SDK
  projection helper. Twelve ABI size/offset checks must pass before rendering.
- Zig 0.15.2 cross-compiles the approximately 10 KiB Linux bridge on the Mac.
  Compiler download and SDK archive are checksum-pinned by the build scripts.
  The target needs no compiler or additional Debian packages.
- GPU mode produces an RGBA texture directly; CPU mode uploads BGRA pixels.
  The comparison used distinct composition paths for that difference, retaining
  premultiplied-alpha blending. SDL retains display/input ownership.

Validation passed for [CPU → GPU](validation/cpu-to-gpu-switch.log) and
[GPU → CPU](validation/gpu-to-cpu-switch.log), including injected pointer clicks,
text entry, JavaScript/Go callbacks and SDL pointer synchronization in each
process. GPU output was inspected from an actual framebuffer capture. Mode
switch tests invoke the JavaScript switch callback; physical F2 is provided
through the SDL keyboard event path.

The GPU implementation currently uses single-sample targets. Edge antialiasing
can therefore differ from CPU output; MSAA, broad CSS correctness, complex
clipping/effects, clipboard/IME, two-finger scrolling and long-running audio-load
tests remain future work. The SDK's signal-handler warning and release-specific
licensing considerations from the initial evaluation still apply.

## Recommendation

GPU rendering was selected, and the CPU implementation has been removed. The GPU path is worth pursuing for
native-resolution animated interfaces: this page looks substantially the same
and uses about half the CPU in the controlled animation test. The existing main
app and Debian card have not been replaced or modified; this remains an isolated,
RAM-staged evaluation.
