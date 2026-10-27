# Audio Utility app and OS roadmap

## Objective

Build a local-network audio utility workstation whose app is the operating
environment. The app owns the display, pointer, and input and draws the complete
user experience itself. Do not build on GNOME, XFCE, a window manager, or a
conventional desktop shell.

## Milestones

### 1. Define the app/base interface

- The base-level SDL3/KMSDRM/OpenGL shader-triangle smoke test passes on the
  MacBookPro12,1. Build the first minimal Go/SDL3 full-screen app on this path,
  then verify app startup, keyboard and pointer input, and shutdown behavior.
- Identify what the app needs from Linux for audio, networking, storage, and
  startup.
- Keep a device-independent gesture layer between machine input adapters and
  UI controls; controls consume pointer, press/release, and drag events.
- Keep device enablement and machine recovery in the machine-specific base.
- Record any graphics runtime requirements in the MacBook base definition.

### 2. Establish the full-screen app shell

- Provide a manual `audio-utility` command first. Keep automatic app startup
  disabled until manual launch, recovery, and updates are reliable.
- Render the full screen and manage pointer/input without an intervening desktop.
- Provide clear startup, failure, and shutdown states.

### 3. Add workstation and system surfaces

- Build the initial audio utility workflow.
- Show Wi-Fi state and useful connection details in the app.
- Provide an Agent Activity view that visibly presents work and output on the
  machine, using the established authenticated development connection.

### 4. Add binary update and persistence behavior

- Keep app source in the repository and distribute machine-compatible binaries
  as GitHub release assets. The machine needs the binary, not a source checkout.
- Keep the installed binary on persistent storage provided by the machine base.
- Pull compatible release binaries from GitHub when networking is ready.
- Publish a separate compatible release asset for each machine folder target.
- If GitHub is unreachable, start the last successfully fetched version and
  make its update status visible in the app.

### 5. Validate on supported hardware

- Measure resource use and verify graphics, pointer, keyboard, Wi-Fi, audio,
  safe shutdown, and recovery on each target machine.
- Keep machine-specific results and setup steps in that machine's folder.

## Prioritized next steps

### 1. Prove the GitHub build and release path

- Push the repository and verify GitHub Actions builds the MacBookPro12,1 Go
  binary from the checked-in build script.
- Verify a tagged release contains a downloadable x86_64 Alpine/musl binary,
  then install and run that downloaded binary on the target.
- Stop building the app over SSH once release downloads are verified; SSH
  should remain a development and recovery channel.
- Preserve the current target binary as a known-good fallback until a release
  binary has been downloaded and run successfully.

### 2. Make manual launch and app updates dependable

- Install a root-shell `audio-utility` command that opens the fullscreen app and
  returns to the caller when the app exits. Do not auto-launch it at boot yet.
- Have the base check GitHub for a compatible release during boot and expose a
  manual update check in the app when the UI can report its status.
- Download to a temporary file, verify version and integrity, keep the last
  known-good binary, and replace it atomically. Restart the app after a
  successful update; if the app is running, use a separate updater/launcher so
  it does not try to replace its own executable in place.
- Keep a usable offline path: no network or a failed update must leave the
  currently installed app launchable.
- Record checksums and compatible Alpine release/package versions so another
  SD card can reproduce the working graphics runtime, not merely download
  whatever package versions happen to be current at rebuild time.

### 3. Build useful app behavior and reusable controls

- Prioritize real audio input/output, playback, volume control, and clear device
  status over visual polish.
- Build a small set of reusable UI controls around the gesture layer, including
  sliders that release correctly when the original click finger lifts.
- Add keyboard navigation, useful Wi-Fi status, and the remote Agent Activity
  surface as the app's underlying functions become available.
- Treat fonts, artwork, spacing, and detailed visual refinement as a separate
  polish pass after the core workflows work.

### 4. Measure and improve the appliance startup experience

- Record cold-boot-to-app-ready time and distinguish Linux startup, network
  readiness, update checking, and app startup.
- Add a quiet, branded boot screen with progress and visible failure/recovery
  states; keep detailed diagnostics available off-screen or in logs.
- Only after manual launch and recovery are dependable, make app-on-boot an
  explicit base option and then lock down the normal boot path.

### 5. Treat internal-disk installation as an optional later project

- First design a readable plan showing the current Apple partitions, proposed
  sizes, and recovery path. Let the user choose how much internal storage to
  allocate.
- Preserve macOS and its recovery partitions. Do not resize, format, or install
  Alpine to the internal disk until the plan is reviewed and explicitly chosen.
- Compare SD-card boot with an optional dedicated Alpine/app partition only
  after the SD-card appliance path is stable.

## Current state

The app has an SDL3 windowed UI preview in `src/` for development on the host
Mac and a machine-targeted Linux x86_64 build. GitHub Actions calls the same
machine build script, uploads the app artifact, and publishes it for `v*` tags.
The release workflow is present but has not yet been verified by pushing a tag
and installing the resulting release binary on the target. The current target
binary was built on Alpine and copied over SSH. The manual `audio-utility`
root-shell command is in the SD-card overlay; automatic launch remains
disabled. The overlay and runtime package paths have been inspected, but a cold
reboot is still needed to verify the complete startup path end to end.
The SD card stores the compiled app binary and runtime packages, not a source
checkout. Its writable data partition contains the app binary, its VT launcher,
and 31 SDL3/Mesa runtime APKs. The Alpine overlay contains the global manual
command and base configuration. On boot, the base mounts the data partition and
installs the graphics runtime into RAM; the binary stays on the data
partition. These copies match the current working tree, but a second-card
rebuild and cold-boot persistence check remain outstanding.
The first target build is on the writable `audio-utility` partition. The input
layer translates SDL and trackpad events into UI gestures, including held-click
drags that continue when another finger is lifted and repositioned. A
30-second fullscreen run on MacBookPro12,1 selected SDL's `opengles2` renderer
at 2560×1600 and returned cleanly to the original virtual terminal. The audio
surfaces still use sample data; audio device control, real Wi-Fi status,
keyboard navigation, full pointer navigation, and the Agent Activity view
remain future work. The cursor-size slider was exercised on the target and
reported each selected pixel size.
Volume and playback-position sliders use the same drag lifecycle, but remain
visual controls until connected to an audio engine. The target has the SDL3/
Mesa runtime APK set on the writable data partition; the boot script loads it
into RAM. Confirm this with a cold reboot before treating boot persistence as
fully verified.
