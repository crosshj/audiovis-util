# System architecture

The system has two separately developed layers.

## Machine base

The base boots Linux, supports the target machine's devices, brings up Wi-Fi,
and provides the runtime and remote access needed for development. Its boot,
firmware, network, and recovery details belong under `machine/<machine-id>/`.
The supported bases are Alpine and Debian on `MacBookPro12,1`.

## Audio Utility app

The app is the user's environment. It renders the display, owns the pointer and
input, presents connectivity and system status, and hosts the workstation's
audio tools. It should not assume a desktop environment or window manager.
The Debian product app lives in `src/`; the established Alpine SDL app is
preserved in `src-alpine/`. Product decisions and milestones belong in `docs/`.

## Display and input path

The user-facing Debian app is a fullscreen Go dashboard using Ultralight's GPU
renderer on an SDL-owned GLES context. It runs directly on a Linux virtual
terminal through KMS/DRM and does not depend on a window manager or desktop
session. The established Alpine output remains a separate SDL3 app built for
musl; the two machine targets deliberately have independent app binaries.

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

Each app target translates SDL mouse events and raw trackpad contacts into its
local event stream. A gesture recognizer turns that stream into pointer
movement, press/release, and drag start/move/end events. In Debian, the
recognizer and trackpad adapter live together in `src/internal/system/`; Alpine keeps an
independent copy in its own source module. UI controls consume gestures rather
than reading SDL or trackpad slots directly. A held click
is tracked separately from finger contacts: another finger can move the
pointer during a drag, lifting and repositioning that finger keeps the drag
active, and lifting the contact that began the click releases the drag even if
SDL omits its button-up event. The development Mac's windowed preview uses SDL
pointer events through the same UI gesture path.

The Debian Ultralight dashboard has passed an on-device accelerated-renderer
and interaction smoke test on this target. The Alpine SDL app remains separately
built and deployed through `machine/macbookpro12,1-alpine/app/`. The dashboard
controls are prototype content and do not yet operate audio devices.

Each supported machine describes its release target in its machine folder.
Release automation publishes two machine-compatible assets: a musl Alpine
binary and a Debian amd64 dashboard bundle with its graphics runtime.

The interface between the layers should stay small and explicit: the base
provides device access, networking, audio/input/graphics support, persistence,
and a way to start the app. The app should not contain machine-specific boot
logic. Conversely, machine setup should not grow product UI or application
features.
