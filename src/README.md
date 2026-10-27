# Workstation app

This directory contains the Audio Utility product software, independent of any
one machine's boot process. The app will provide the full user environment and
own the screen, pointer, and input. It will not depend on GNOME, XFCE, a window
manager, or another desktop shell.

## Development preview

Run the current SDL3 windowed UI preview on the development Mac:

```sh
go run .
```

The preview requests a 1280×800 point window with a high-density backing
surface, matching the target Mac's 2560×1600 pixels on a 2× Retina display.

The preview follows [`../design/mockup.png`](../design/mockup.png): audio
output and input panels, a spectrum display, audio tool tiles, and playback
controls. The displayed levels and spectrum are visual sample data; controls
do not operate audio devices yet. It uses the bundled Inter font and a
pure-Go font rasterizer; the font's SIL Open Font License is included at
[`assets/fonts/OFL.txt`](assets/fonts/OFL.txt). Close the window to quit.

The top bar has a draggable cursor-size slider. It changes the SDL mouse cursor
immediately and logs each selected size in pixels. The default is 60 px, the
center mark on the 32–60–88 px scale.

UI controls receive normalized pointer and drag gestures from the input layer.
SDL events on the development Mac and raw trackpad contacts on the MacBook
target use the same control path; the windowed preview keeps macOS's native
pointer behavior.

The same app can run fullscreen on Linux with SDL's KMSDRM video driver and
OpenGL ES renderer. The MacBookPro12,1 target build and display test are
documented in [`../machine/macbookpro12,1/app/README.md`](../machine/macbookpro12,1/app/README.md).
The audio levels, spectrum, and controls are still prototype content, not live
hardware behavior. See
[`../docs/app-roadmap.md`](../docs/app-roadmap.md) and
[`../docs/architecture.md`](../docs/architecture.md) for the intended
milestones and the boundary between the app and a prepared machine base.
