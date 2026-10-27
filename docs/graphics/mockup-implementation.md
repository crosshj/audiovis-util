# GPU dashboard mockup implementation

Implemented 2026-10-27 in `src/app`. Reference: the display UI
in `design/mockup.png`, excluding its photographed bezel, perspective and glare.

The product app uses the Ultralight GPUDriver / SDL3 / GLES path. Its logical
canvas is 1280×800; `--full` renders at the MacBook panel's native 2560×1600.
The layout reproduces the status strip, waveform brand, Output/Input/Audio Tools
cards, spectrum, and transport card. Static Inter fonts, charcoal surfaces,
outlined cards, blue spectrum bars, segmented input meter and custom slider
thumbs follow the reference. A solid background avoids a visible gradient seam
observed with this GPU adapter. A CSS `scaleX(-1)` on the wrench icon also
triggered missing card contents; it was removed. The replacement Heroicons wrench needs no transform. General transformed-layer compositing still needs investigation.
Single-sample SVG edges are not perfectly smooth.

## Assets

The approved icon library is [Bootstrap Icons](https://icons.getbootstrap.com/),
vendored SVGs at v1.13.1 commit `ce0e49dd063243118a115f17ad1fe1fe7576d552`.
There is no Bootstrap CSS or runtime dependency. The waveform brand uses CSS;
the Tone sine-wave is a small custom SVG to match the reference. Other icons
come from Bootstrap, except the closer-matching solid wrench from
[Heroicons v2.2.0](https://github.com/tailwindlabs/heroicons/blob/v2.2.0/optimized/24/solid/wrench.svg).
Its upper-right jaw and rounded handle with a small hole match the reference;
the SVG is used unmodified, with its MIT license included. Inter static WOFF2 files are pinned to commit
`353b61b9f4430d5f420d56605a6e7993e0941470`.

`ui/assets/sources.json` records source URLs and SHA-256 hashes. The build verifies
the files and produces `ui/assets/icons.js` from the SVG subset. MIT/OFL licenses
are included. Everything loads locally without a network request.

## Interaction and scope

All controls are div/CSS/JavaScript: no native input, select, button or textarea.
No accessibility feature work was added, per the requested scope.

- Volume and seek support dragging, including movement outside their bounds.
  Values clamp to 0–100 and pass synchronously through `nativeControl` to Go.
- Device dropdowns open custom option panels and send selected indices to Go.
- Play/pause changes the icon and advances an illustrative three-minute timeline.
  Previous/next move the seek position by ten percent.
- Mute and Tone/Record/Analyzer toggle demo state. Analyzer animates the bars.
- Escape closes an open dropdown before exiting the app.
- One pointer position controls HTML events, the drawn arrow and hidden SDL cursor.

Audio routing, tone generation, recording, device discovery and real spectrum
analysis are **not connected**. Audio device options, spectrum and meter are illustrative. Wi-Fi and battery
status now read the actual hardware (see the tray sections below). The normal clock uses the target system time; `--static`
uses the reference date/time. The UI exercises the stack without pretending to
be a completed audio engine. The dashboard is the Debian product app in `src/app/`; its test bundle is staged in `/tmp/audio-utility-debian`.

## Validation

Native-resolution test:

```sh
./run-target.sh --full --static --benchmark --self-test --capture --seconds 5
```

Actual Ultralight mouse down/move/up events tested both sliders from 25% to 75%,
play/pause, dropdown opening and option selection. Assertions checked the Go
values, absence of native form controls, font availability and SDL pointer
synchronization. The same interaction test passed at 1280×800. The full-resolution
test caught and fixed Ultralight's CSS zoom discrepancy: mouse client coordinates
need dividing by body zoom before comparison with element bounds.

- [Native-resolution test log](validation/mockup-full-test.log)
- [1280×800 test log](validation/mockup-half-test.log)
- [Final framebuffer screenshot](validation/mockup-screen.png)

The final full-resolution run reported 53.93 FPS, 24.02% of one CPU core and
197.74 MiB peak RSS. This short run includes screenshot readback/PNG encoding
and control tests; it is not a performance benchmark. The screenshot was visually
compared with the reference. Go vet, JavaScript syntax and asset checksum checks
passed. Existing evdev touchscreen-name and WebKit signal-handler warnings remain;
trackpad input and the GPU tests succeeded.

Build and launch instructions are in the experiment README. Interactive launch:

```sh
./run-target.sh --full --seconds 0
```

## Battery tray indicator

The battery icon reads real Linux `power_supply` capacity and status once per
second through Go. Its fill tracks charge, a bolt indicates charging, and low
charge (15% or less while not charging) uses an amber tint. Missing/unreadable
capacity displays an unknown dash instead of a fabricated percentage. The
initial hardware check reported BAT0 at 47%, discharging. Like the clock,
`--static` freezes the indicator after its initial reading.

## Reusable anchored popovers

`popover.js` supplies `popovers.attach(anchorElement, panelElement)`, returning
`open()` and `close()` methods. It mounts panels directly under body, above
cards and dropdowns, and permits one open popover at a time. Clicking the anchor
toggles it; outside mouse-down and the shared Escape handler dismiss it.
Positioning uses logical CSS coordinates, right-aligns near the anchor, clamps
to a 12px viewport margin and flips above the anchor when needed. Window resize
repositions the active panel; content that changes size can call
`popovers.reposition()`.

Shared `.popover`, `.popover-heading`, `.popover-row`, `.popover-secondary` and
`.popover-separator` styles define the 280px dark surface, rounded corners,
thin outline, shadow, padding and typography for future tray panels. The surface
is solid rather than relying on unverified backdrop-blur support. No CSS
transform is used for positioning. Panels can override width for their content.

Battery is the first consumer: live percentage, charging description, charge
bar and power source inferred from the kernel's charge status. No unsupported
time-remaining estimate is displayed. The tray icon highlights while open.
Actual pointer tests passed at both render scales for opening, anchored bounds,
Escape dismissal and outside-click dismissal. A hardware screenshot is saved at
[validation/battery-popover.png](validation/battery-popover.png).

## Wi-Fi popover

The Wi-Fi tray icon now uses the shared popover with an on/off switch, current
SSID and connection state. It offers no network selection or saved-profile
editing. The custom switch waits for the actual radio state and shows failures
instead of pretending the operation succeeded.

`wifi.go` polls Linux WLAN rfkill state and `wpa_cli status` every two seconds
on a worker goroutine. UI requests use a bounded channel; system calls and
subprocesses never block the rendering thread. Commands have a two-second timeout.
Only a radio whose rfkill type is `wlan` is toggled; Bluetooth is unaffected.
Enabling unblocks the radio and requests reconnection through the existing
wpa_supplicant configuration. The popup continues to work locally without SSH.
Hardware blocking and unavailable radios disable the switch.

The visual result is in [validation/wifi-popover.png](validation/wifi-popover.png).

Hardware toggle verification: the popup handler turned WLAN off at 21:02:57,
back on at 21:03:00, and the saved connection completed at 21:03:03. A temporary
local recovery timer was armed for the SSH-disrupting test; reconnection occurred
before it ran. The one-shot recovery subsequently ran at 21:03:09 and expired;
normal startup was restored, and no recovery timer remains active.
See [wifi-toggle-test.log](validation/wifi-toggle-test.log). Go vet and JavaScript
syntax checks passed.
