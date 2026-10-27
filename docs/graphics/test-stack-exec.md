# Ultralight: shortest path to a hardware decision

**Current status:** GPU-only rendering has been selected. CPU rendering and
mode switching have been removed. The plan and initial results below record
the evaluation history; see the [comparison report](test-stack-gpu.md) for the
measurements behind this decision and current run instructions.

Execute the evaluation in [test-stack.md](test-stack.md) as three small checkpoints.
Keep the existing app intact. Build locally, deploy over SSH to Debian, and record
every added dependency. Do not change Alpine or add CI integration.

## 1. Prove the Linux SDK runs

- Pin a Linux x86-64 SDK and record its source, checksum, license and resources.
- Inspect shared-library, glibc and CPU requirements.
- Build a tiny executable on the development machine and deploy it to Debian.
- Render local HTML into a CPU pixel buffer, with fonts and JavaScript working,
  without an X11 or Wayland session.
- Record actual runtime results before building a larger interface.

## 2. First meaningful interactive result

- Create an isolated Go test using the working SDL3/KMSDRM launch pattern.
- Begin with a 1280×800 HTML surface scaled to the 2560×1600 display.
- Show a transparent panel, button, text field and counter using local assets.
- Forward pointer, keyboard and text input.
- Have a button call Go, change a native background color, and display Go's reply.
- Have Go update HTML independently of clicks.
- Handle pixel format, row stride and premultiplied alpha correctly; upload only
  when the HTML surface changes. Restore the console on exit.

This proves display, interaction, transparency and bidirectional integration.

## 3. Native graphics and cost

- Animate a simple native shader behind mostly static HTML.
- Measure CPU, RAM, frame times, HTML rendering and texture-upload time.
- Repeat with full-resolution HTML and with a CSS animation.
- Determine whether CPU rendering is sufficient before implementing a GPUDriver.

The existing app uses SDL's 2D renderer with an OpenGL ES backend. It does not
yet establish application-owned GL shader rendering. Prefer an explicit GLES
context in the isolated shader test; verify compositing and input independently.

## Requirements and boundaries

The anticipated integration needs a pinned SDK, runtime libraries/resources,
thin native bindings, font/filesystem handlers and SDL event translation.
If cgo/native linking is required, use a local Linux build environment; the
existing CGO_ENABLED=0 macOS-to-Linux build cannot simply absorb it.

Defer the full toolbar, exhaustive controls, extensive dirty-region optimization,
long-duration testing and custom GPU driver until the initial checkpoints pass.
Review licensing early. A successful link is not proof of runtime compatibility,
and a visible page is not proof of acceptable audio-workload headroom.

## Execution record

- **Checkpoint 1 passed on physical Debian:** HTML rendered into a CPU surface,
  JavaScript executed, and JavaScript called Go synchronously and received a value.
  No display server was used. Debian 13.7 has glibc 2.41 and all required system
  libraries already installed. No Debian packages were added.
- **Checkpoint 2 implemented and exercised:** isolated Go binary, local HTML/font,
  transparent panel, text field, controls, Go status updates, pointer/text event
  assertions and physical button callbacks. Console restoration passed.
  The full-resolution self-test also passed after scaling CSS coordinates to
  surface pixels: [log](validation/full-self-test.log),
  [actual framebuffer capture](validation/full-screen.png).
- **Checkpoint 3 rendering passed:** explicit SDL-owned GLES context on Intel
  Iris 6100, native shader behind HTML, correct BGRA/premultiplied-alpha composite.
  Performance comparisons are recorded below and in `validation/`.
- Implementation and build/deploy/run instructions:
  [product app](../../../src/README.md).
- Existing Debian baseline remains in `machine/macbookpro12,1-debian/`.

### SDK selection and dependencies

Pinned **1.4.0.864dc81**, the CPU-surface API version used by this evaluation.
The current download page defaults to **2.0.0-beta.2**; these results do not
validate 2.0 or imply its API/features are interchangeable with 1.4.

Official archive:
`https://ultralig.ht/api/v1/sdk/download?platform=linux-x64&version=1.4.0`

Retrieved SHA-256:
`75e3079a9bcae834e98607af0d8ec6ef8f8144abc674ce961cf878fb2d816ddd`.
The legacy metadata provides no publisher checksum; the build script pins this
retrieved archive for subsequent reproducibility. SDK binaries stay out of Git.

Runtime: `libAppCore.so`, `libUltralight.so`, `libUltralightCore.so`,
`libWebCore.so`, ICU data and SDK resources. AppCore supplies font/filesystem
helpers only, but its binary links GTK/X11/Wayland client libraries. These are
already present; no desktop session is initialized. Dependencies and CPU details
are captured in [runtime-context.txt](validation/runtime-context.txt).

The actual MacBook CPU executed the tested paths successfully. This is stronger
evidence for this target than merely satisfying the linker; it is not an ISA
compatibility guarantee for every SDK feature or another machine.

The bundled 1.4 `license/LICENSE.txt` and `license/EULA.txt` are retained with the
private SDK/runtime. The archive's license text differs from current website
pricing terms. Internal evaluation is the scope here; distribution terms must
be resolved for the chosen release before shipping. No SDK binaries or licenses
were republished in the repository.

### Integration findings

- The existing `purego` dependency calls the C API and JavaScriptCore directly.
  The working prototype therefore cross-builds locally with `CGO_ENABLED=0`;
  neither a Linux compiler VM nor a custom C++ shim was necessary.
- Ultralight runs on the locked rendering OS thread. Go callbacks are tiny;
  future audio work must stay off this thread.
- The initial variable-font test showed abnormal spacing. Bundled static Go
  Regular replaces it; a broader font evaluation remains future work.
- Whole surfaces upload only when dirty. Native-only redraws do not require
  HTML texture uploads. Partial-region uploads are deferred.
- The SDK logs that it replaces a SIGUSR1 handler. This did not prevent the
  short runs; signal coexistence and long-running GC behavior need checking
  before adoption. SDL also logs a touchscreen-name warning; the explicit
  bcm5974 adapter successfully opens and handles physical clicks.
- Pending: full hands-on keyboard/focus/scroll validation, two-finger trackpad
  scrolling, clipboard/IME, native-scene hit routing, long-duration memory
  behavior and audio-under-load testing. The initial CPU checkpoint did not yet
  include a GPUDriver; the subsequent GPU comparison adds it.

### First controlled hardware results — 2026-10-27

All cases use the physical 2560×1600 display. Each controlled sample ran for
15 seconds with physical input ignored (Escape still exits). Static cases freeze
native shader time and status updates but still draw/present each frame. Animated
cases run both the native shader and the CSS animation, with Go status updates.

| HTML surface | Workload | Mean FPS | CPU, one core = 100% | Peak process RSS | HTML uploads |
| --- | --- | ---: | ---: | ---: | ---: |
| 1280×800 | Static | 59.26 | 13.82% | 161.61 MiB | 1 |
| 2560×1600 | Static | 59.26 | 13.69% | 189.19 MiB | 1 |
| 1280×800 | CSS + native animation | 59.28 | 27.13% | 161.33 MiB | 447 |
| 2560×1600 | CSS + native animation | 59.15 | 64.40% | 188.93 MiB | 455 |

Animated full-surface uploads averaged **1.75 ms** at 1280×800 versus **12.11 ms**
at 2560×1600. Ultralight render calls averaged **1.05 ms** versus **2.48 ms** per
display frame. The full-resolution animated sample had 4 frames over 25 ms;
the half-resolution sample had 1. CSS repaint cadence was approximately 30 Hz,
so 59 display FPS does **not** mean HTML animation itself rendered at 59 FPS.

The static cases uploaded once across approximately 890 frames, confirming that
the native rendering loop does not force HTML texture uploads.

Raw controlled logs: [half static](validation/benchmark-half-static.log),
[full static](validation/benchmark-full-static.log),
[half animated](validation/benchmark-half-animated.log),
[full animated](validation/benchmark-full-animated.log).
Earlier `half-*`/`full-*` measurements allowed physical interaction and are retained
as exploratory runs, not controlled baselines. The first interactive run recorded
three additional physical native-function clicks after its injected click.

CPU covers the timed application loop; peak RSS includes process initialization.
RSS excludes the separate cost of staging SDK files in tmpfs and is not total
machine RAM usage. Screenshot capture was disabled for these comparisons.
Upload timing measures the CPU call, not GPU timer-query completion; long calls
may include synchronization. Frames over 25 ms are not exact dropped-vblank
counts. Short runs do not establish leak freedom or audio scheduling reliability.

### Decision after these checkpoints

**The proposed CPU-rendered HTML + Go + SDL3 + native GLES stack is viable on
this Debian appliance with Ultralight 1.4.** It runs without a desktop session,
needs no additional Debian packages on the current image, and builds on the Mac
with the existing Go toolchain. This is an evaluation result, not approval to
rewrite the main application or a conclusion about Ultralight 2.0.

Continue with the smaller UI surface or smaller native-resolution panels first.
If full-screen animated HTML is needed, implement dirty-region uploads and
measure again before undertaking a GPUDriver. The current full-resolution case
still displays smoothly, but its upload cost leaves substantially less frame
budget. Resolve the listed input, lifetime/signal, licensing and audio-load
questions before adopting this stack for the main application.

### Handoff

Follow-up cursor fix: SDL's independently tracked native cursor had remained
visible alongside the custom pointer. It is now hidden, with SDL's internal
position synchronized to the same clamped trackpad position used by HTML input
and native drawing. The 8-second full-resolution self-test checked cursor state
agreement each frame and passed the click/text assertions:
[cursor-sync.log](validation/cursor-sync.log). SIGTERM now performs orderly
cleanup for demo restarts.

The native-resolution interactive demo was left running on tty2 from
`/tmp/ultralight-eval` after validation. Press Escape to exit and restore tty1.
It is not an autostart service and disappears on reboot. Build, Linux-targeted
`go vet`, shell syntax and whitespace checks passed. The existing app, Alpine
reference and Debian card contents were not changed by this evaluation.
