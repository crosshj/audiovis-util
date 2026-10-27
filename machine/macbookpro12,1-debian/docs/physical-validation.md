# Physical Debian validation — 2026-10-27

Validated over the card's pinned-key SSH connection on MacBookPro12,1:

- Debian 13.7; automatic startup service completed successfully.
- Broadcom BCM43602 using brcmfmac; Wi-Fi associated and configured address present.
- Default route, DNS resolution, and HTTPS connectivity checked.
- Key-only root SSH with the retained host key; root tty1 autologin configured.
- Root overlay and its writable upper layer in RAM; OS medium read-only tmpfs;
  no swap. HFS+ data partition mounted read-write and write-probed successfully.
- Console framebuffer 2560×1600 and 160 columns × 50 rows, exactly matching Alpine.
- Keyboard backlight brightness 127; Intel Iris 6100 using i915.
- ALSA enumerates CS4208 analog/digital audio and HDMI devices. Audible output
  has not been tested.
- SDL 3.4.0 loads all Go binding symbols and includes KMSDRM/OpenGL ES.
- The staged Go app ran for 30 seconds at 2560×1600 with the `opengles2`
  renderer. Trackpad initialization and cursor-size changes were logged.
  It exited with status 0 and restored tty1.

The first app launch exposed a Debian-specific `openvt` argument parsing
issue: it interpreted the shell's `-c` as its own VT selector. Adding `--`
before `sh -c` fixed this. The correction is in both the repository launcher
and the card's persistent app launcher.

The app logged a nonfatal evdev touchscreen-name warning, also observed on
Alpine; trackpad input and rendering succeeded. Physical cold-boot repetition
and audible output testing remain outstanding. Raw checks are retained locally
under ignored `build/macbookpro12,1-debian/audit/physical-*` files.
