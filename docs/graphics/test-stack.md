# Audio-Utility: Ultralight Graphics Stack Evaluation

> Historical evaluation brief. The implemented demo is now GPU-only; CPU
> rendering and mode switching were removed after comparison. See
> [current instructions](../../src/README.md) and
> [recorded results](test-stack-gpu.md).

## Objective

Evaluate Ultralight as the HTML/CSS/JavaScript interface engine for `audio-utility`.

Repository: https://github.com/crosshj/audiovis-util

We want a minimal Linux appliance that boots directly into our application, supports custom HTML interfaces, and retains full native GPU capabilities for shaders, 3D rendering, and real-time audio visualizations.

**Do not redesign the application yet. Prove the stack works first.**
**Do not replace the current app with this test yet.  We'll rewrite it using this stack later if the stack works.**
**Build the test app on dev machine and push to target machine for testing.  Do not hook this into github action yet.  It's okay to add dependency libs to the running target, but keep track of that so we can add it to base build once we see this is working.**

## Intended architecture

```text
HTML + CSS + JavaScript
         |
      Ultralight
         |
    Native C API / cgo
         |
    Go Application
         |
      SDL3
         |
    OpenGL / GLES
         |
      DRM / KMS
         |
    Intel Iris 6100
```

Ultralight should provide HTML layout, CSS styling, JavaScript execution, DOM interaction, and conventional interface controls.

Go remains responsible for application logic, hardware integration, networking, MIDI, and audio coordination.

SDL3 retains ownership of the window, graphics context, display, and input events.

OpenGL remains available for custom GPU rendering independently of Ultralight.

C/C++ should be used only where necessary for native integration.

## Target environment

- Early 2015 MacBook Pro, MacBookPro12,1
- Intel Core i5
- 8 GB RAM
- Intel Iris Graphics 6100
- 2560×1600 Retina display
- Debian Linux, minimal installation
- No conventional desktop environment
- Go + SDL3
- Direct KMSDRM display backend
- Existing OpenGL ES rendering

Inspect the repository and actual target environment before assuming these components are all working under Debian.

Preserve existing functionality.

## Phase 1 — Compatibility investigation

Before implementation:

1. Inspect the current Go, SDL3, and OpenGL architecture.
2. Identify the exact Ultralight SDK version and supported Linux requirements.
3. Verify compatibility with Debian's glibc version.
4. Identify required shared libraries and runtime resources.
5. Determine whether the SDK supports this CPU's instruction set.
6. Identify the appropriate C API and JavaScriptCore interfaces.
7. Verify whether Ultralight can operate without X11, Wayland, or AppCore's windowing functionality.
8. Review redistribution, licensing, and SDK packaging requirements.

Use current official documentation:

- https://docs.ultralig.ht/
- https://docs.ultralig.ht/llms.txt
- https://docs.ultralig.ht/docs/integrating-with-games
- https://docs.ultralig.ht/docs/using-the-c-api
- https://docs.ultralig.ht/docs/gpu-renderer-overview

Do not assume that successful linking means runtime compatibility.

## Phase 2 — Minimal CPU-rendering experiment

Create an isolated proof of concept using Ultralight's CPU renderer.

Do not introduce a custom GPU driver yet.

The experiment should:

1. Initialize Ultralight directly, without giving it control of the application window.
2. Configure the necessary font and filesystem handlers.
3. Load a local HTML document.
4. Render that document into an Ultralight CPU pixel surface.
5. Transfer the resulting pixels to an SDL3 texture.
6. Display the texture through our existing KMSDRM application.
7. Forward SDL3 mouse, keyboard, text-input, and scrolling events to Ultralight.
8. Advance Ultralight's update/render lifecycle correctly.
9. Upload only changed regions when practical.
10. Cleanly release resources on exit.

The test page should contain:

- A transparent toolbar with Wi-Fi status, YYMMDD date, and time.
- A responsive flexbox layout.
- A button.
- A text input.
- A slider.
- A dropdown/select control.
- A scrollable panel.
- A CSS animation.
- A JavaScript-updated counter.

Use bundled fonts and local assets. Do not depend on an internet connection to display the interface.

### Acceptance criteria

All controls must be interactive using the MacBook's keyboard and trackpad.

Text entry, focus changes, pointer interaction, and scrolling must work.

CSS layout must render correctly at the MacBook's native resolution.

The application must run without X11, Wayland, GNOME, KDE, or a conventional desktop session.

## Phase 3 — JavaScript-to-Go integration

Establish a minimal bidirectional communication mechanism.

Demonstrate:

- JavaScript invokes a Go function when a button is clicked.
- Go returns a value that JavaScript displays.
- Go changes a value displayed in the HTML interface.
- JavaScript receives a recurring status update without reloading the page.

Use the supported Ultralight/JavaScriptCore interfaces.

Keep native bindings thin. Avoid introducing an unnecessary messaging framework.

The objective is to validate that future UI features can be implemented primarily using HTML, CSS, and JavaScript while Go handles native operations.

## Phase 4 — Mixed HTML and native graphics

Verify that Ultralight can coexist with our existing graphics capabilities.

Demonstrate:

- Ultralight displaying a partially transparent HTML overlay.
- A native OpenGL-rendered animation behind or beside the HTML interface.
- Correct alpha blending and compositing.
- Native graphics updating independently of the HTML interface.
- No unnecessary HTML rerendering when only the native scene changes.
- Correct input routing between the native scene and interactive HTML elements.

Use a simple animated shader or rotating 3D object as the native rendering demonstration.

Do not build an extensive graphics framework.

## Phase 5 — Performance measurements

Measure performance on the actual MacBook rather than relying solely on desktop development hardware.

Record:

- Idle application memory consumption.
- CPU utilization with a static HTML interface.
- CPU utilization while interacting with controls.
- CPU utilization during CSS animation.
- CPU utilization during native OpenGL animation.
- Frame times and dropped frames.
- Time spent rendering Ultralight content.
- Time spent uploading textures.
- Memory growth over an extended interactive session.

Test at 2560×1600 and at a lower internal UI rendering resolution scaled to the display.

The Retina display may make CPU rendering and texture uploads comparatively expensive. Determine whether lower-resolution UI rendering provides a meaningful performance improvement.

Pay particular attention to whether animated HTML content competes with future real-time audio workloads.

## Phase 6 — GPU-rendering feasibility

Only after the CPU-rendering prototype works, investigate Ultralight's GPU renderer.

Ultralight supports a custom GPUDriver that translates rendering commands into native graphics operations.

Research:

- Whether the provided OpenGL reference driver can be reused.
- Whether it supports our existing OpenGL ES context.
- Whether desktop OpenGL would be preferable or necessary.
- Which shader versions and GPU capabilities are required.
- Whether accelerated Ultralight views can share our existing rendering context.
- Complexity of integrating the driver through C/C++ and Go.
- Expected performance benefit compared with CPU rendering.

Do not implement a custom GPUDriver unless there is evidence that the CPU renderer is insufficient.

If GPU support requires substantial changes to the working rendering stack, document those changes before proceeding.

## Implementation constraints

- Preserve the existing SDL3/KMSDRM functionality.
- Avoid introducing a desktop environment.
- Avoid Chromium, Electron, or a separate browser process.
- Keep native dependencies explicit and reproducible.
- Do not embed large SDK binaries directly in Git without reviewing licensing and repository-size implications.
- Prefer building native dependencies in GitHub Actions rather than on the MacBook.
- Preserve the ability to deploy the resulting application as a ready-to-run artifact.
- Avoid blocking audio-sensitive execution paths with UI rendering or JavaScript callbacks.
- Keep the proof of concept isolated so it can be discarded without affecting the existing application.
- Do not build a DAW, sequencer, plugin host, or elaborate UI component library during this evaluation.

## Expected deliverables

1. A working Ultralight CPU-rendering prototype.
2. Interactive HTML/CSS/JavaScript displayed through SDL3/KMSDRM.
3. Working bidirectional JavaScript/Go communication.
4. Demonstration of Ultralight composited with native OpenGL graphics.
5. Actual CPU, RAM, and rendering performance measurements.
6. Documentation of dependencies, deployment requirements, and compatibility problems.
7. A recommendation on whether to adopt Ultralight and whether GPU rendering deserves further work.

## Decision criteria

We should adopt Ultralight if it provides:

- Familiar web-style UI development.
- Reasonably low CPU and RAM overhead.
- Reliable interaction.
- Easy integration with Go.
- Compatibility with direct KMSDRM rendering.
- No requirement for a conventional desktop environment.
- Continued independent access to native OpenGL rendering.

The ultimate objective is not the smallest theoretical rendering engine.

**It is the most productive graphics and application environment we can build on this hardware without sacrificing meaningful computing resources or imposing unnecessary future development friction.**

Stop after the evaluation and report findings before restructuring the main application.