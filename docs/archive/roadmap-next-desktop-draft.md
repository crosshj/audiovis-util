# Audio Utility MacBook roadmap

## Objective

Turn the Early 2015 MacBook Pro into a remotely programmable, RAM-resident
Linux appliance with a quiet graphical boot, a visible Wi-Fi status, and the
Audio Utility app running from its Git checkout.

The repository is the source of truth for the app, boot configuration, splash
art, and deployment scripts. The SD card is the boot and recovery target. Make
and review changes in the repository, then deploy them to the running MacBook
over SSH; do not hand-edit separate copies on the card.

## Target hardware

- MacBook Pro (Retina, 13-inch, Early 2015), `MacBookPro12,1`
- Intel Core i5 2.7 GHz, 8 GB RAM, Intel Iris Graphics 6100
- Broadcom BCM43602 Wi-Fi
- macOS High Sierra 10.13.6 remains installed on the internal APFS volume

## Verified foundation

- Alpine Linux boots from the SD card and runs from RAM.
- The card has one EFI boot path and three partitions: EFI, `ALPINE`, and
  `audio-utility`.
- `wlan0` connects to `gastropod` at `192.168.1.38`; gateway, DNS, and internet
  checks passed.
- The keyboard backlight works.
- SSH is available with the staged key, and a persistent SSH session is the
  normal way to inspect and update the live machine.
- Reports and startup logs save to the writable `audio-utility` partition.
- The full setup script lives once on `ALPINE`; the Alpine overlay keeps a small
  launcher. Keep this single-copy arrangement.
- Missing optional BCM43602 calibration files are not a blocker; Wi-Fi works.

## Desired boot experience

1. After selecting the card's EFI Boot entry, hide routine boot text and show a
   white music-note logo on black with a progress indicator.
2. When startup completes, open a full-screen graphical Audio Utility app.
3. Show Wi-Fi state in the upper-right tray area. Indicate connecting, connected,
   and disconnected states; clicking it opens connection details and controls.
4. Provide a clickable Agent Activity view that shows the live commands and
   output being run on the MacBook, with a way to return to the main app.
5. Keep the app source in a Git checkout on persistent storage and run the app
   from that checkout.
6. Once networking is available, fetch and fast-forward the configured GitHub
   branch before launching the app. If GitHub is temporarily unreachable, start
   the last successfully fetched version and show that it is offline or stale.
7. Keep the current automatic hardware report, readable startup status, and safe
   shutdown instructions available for troubleshooting.

## Implementation sequence

### 1. Make the repository deployable

- Configure the correct GitHub remote and branch.
- Put the app, splash assets, Alpine configuration, package lists, and deploy
  scripts under version control.
- Add one repeatable build/deploy command that copies a reviewed build to the SD
  card over the existing SSH connection.
- Keep the current boot path available until a replacement completes a cold-boot
  check.

### 2. Choose persistent storage

- Prototype first on the existing `audio-utility` SD partition, which is mounted
  read/write on Alpine.
- The internal disk currently contains an APFS/macOS installation and is not
  mounted by Alpine. A dedicated Linux partition there would require resizing the
  APFS container. Preserve the macOS installation; do not repartition it until
  the storage plan is explicitly approved.
- Once the location is chosen, keep the source checkout, app state, and logs on
  that persistent data volume. Keep the OS itself RAM-resident.

### 3. Build the graphical session

- Use a small, established graphical desktop and its system tray rather than
  writing a Wi-Fi stack. Evaluate XFCE with NetworkManager's tray applet against
  the current Alpine release.
- Run the Audio Utility app automatically in the graphical session.
- Make the Agent Activity button reveal a live terminal/activity view, not a
  separate unauthenticated command server.
- Measure memory and confirm graphics, keyboard, touchpad, audio, and Wi-Fi
  behavior on the actual MacBook.

### 4. Add the boot splash and update flow

- Suppress routine kernel and service text while retaining a recovery option for
  diagnostics.
- Show the logo and progress through early boot, then transition cleanly into the
  graphical app.
- Start the Git update only after Wi-Fi is connected; launch the app from the
  resulting checkout and retain the previous usable version if fetching fails.

### 5. Validate and document

- Cold boot from the SD card and confirm the complete splash-to-app sequence.
- Verify Wi-Fi status updates, detail controls, Agent Activity visibility,
  GitHub refresh, offline startup, report persistence, and safe shutdown.
- Confirm the internal macOS volume remains intact and unmodified.
- Document recovery by booting the current known-good Alpine path.

## Constraints

- Do not create additional EFI boot entries or duplicate full setup scripts.
- Do not install to or overwrite the internal macOS volume.
- Keep Alpine diskless: the OS runs from RAM; persistent data is stored separately.
- Do not add an unauthenticated remote command endpoint. Use the existing SSH
  key and the visible Agent Activity view.
- Do not eject or reflash the SD card for routine changes; deploy over SSH while
  the MacBook is running.

## GitHub and storage facts to resolve

The current local checkout has no Git remote configured and contains no app
implementation yet. Confirm the intended GitHub repository/branch before adding
automatic updates. Decide whether `audio-utility` on the SD card is an acceptable
prototype location or whether the eventual source checkout must live on a new
partition on the MacBook's internal drive.
