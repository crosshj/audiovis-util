# System architecture

The system has two separately developed layers.

## Machine base

The base boots Linux, supports the target machine's devices, brings up Wi-Fi,
and provides the runtime and remote access needed for development. Its boot,
firmware, network, and recovery details belong under `machine/<machine-id>/`.
The current supported base is Alpine on `MacBookPro12,1`.

## Audio Utility app

The app is the user's environment. It renders the display, owns the pointer and
input, presents connectivity and system status, and hosts the workstation's
audio tools. It should not assume a desktop environment or window manager.
Product code belongs in `src/`; product decisions and milestones belong in
`docs/`.

## Display and input path

The Go app is the user-facing display environment. It runs directly on a Linux
virtual terminal and uses SDL3's KMS/DRM video backend to take control of the
display and input devices. It does not depend on X11, Wayland, a window manager,
or a desktop session. The graphics path uses the kernel's DRM/KMS device,
EGL/GBM, and OpenGL; SDL handles display modes and keyboard/pointer events.

The MacBookPro12,1 has passed a fullscreen SDL3 KMSDRM/OpenGL smoke test: SDL
acquired the built-in 2560×1600 display, Mesa reported Intel Iris Graphics 6100
as the renderer, and a GLSL shader triangle was rendered and swapped
successfully. The driver reports OpenGL 4.6 Core and GLSL 4.60. The base keeps
the signed SDL3/Mesa runtime packages on the writable data volume and installs
them into RAM at boot. The visible test stays manually launchable with
`graphics-test`; it is not the future Go app and does not run automatically.
The test selects SDL's legacy KMS pageflip path because the atomic commit path
was rejected by this Mac's display driver. The test checks the framebuffer's
center pixel (`RGBA 242,204,51`) and the active DRM plane is owned by SDL, so
shader output and panel scanout are both verified. Its virtual-terminal
launcher returns to the original VT after the test ends. Keyboard and pointer
input still need their own verification.

## Input and gesture boundary

Machine adapters translate SDL mouse events and raw trackpad contacts into a
shared input stream. The gesture recognizer turns that stream into pointer
movement, press/release, and drag start/move/end events. UI controls consume
those events rather than reading SDL or trackpad slots directly. A held click
is tracked separately from finger contacts: another finger can move the
pointer during a drag, lifting and repositioning that finger keeps the drag
active, and lifting the contact that began the click releases the drag even if
SDL omits its button-up event. The development Mac's windowed preview uses SDL
pointer events through the same UI gesture path.

The first Go app build has also run fullscreen on this target. It loaded the
base's system `libSDL3.so.0`, selected the `opengles2` renderer, rendered at
2560×1600 for 30 seconds, and returned cleanly to the original VT. The binary
is stored under `/media/audio-utility/apps/macbookpro12,1/`; the test launcher
is in `machine/macbookpro12,1/app/`. This checks the Go/SDL deployment path,
but the UI is still sample content and keyboard, pointer, Wi-Fi UI, and audio
device behavior remain unverified.

Each supported machine describes its release target in its machine folder.
Release automation builds a compatible app asset per machine, rather than
assuming one binary and runtime will work everywhere.

The interface between the layers should stay small and explicit: the base
provides device access, networking, audio/input/graphics support, persistence,
and a way to start the app. The app should not contain machine-specific boot
logic. Conversely, machine setup should not grow product UI or application
features.
