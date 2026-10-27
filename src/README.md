# Debian product app

This Go module builds the Debian GPU dashboard. The executable lives in
`app/`, private Go packages live in `internal/renderer/` and
`internal/system/`, and the browser-style UI with its fonts and SVG icons lives
in `ui/`. The generic gesture recognizer is beside the trackpad adapter in
`internal/system/`. Alpine remains an independent SDL/musl module in
[`../src-alpine/`](../src-alpine/README.md), with its own gesture package.

## Build and run

From the repository root:

```sh
machine/macbookpro12,1-debian/app/build.sh
machine/macbookpro12,1-debian/app/deploy-test.sh
```

The build produces `build/macbookpro12,1-debian/audio-utility-debian-amd64.tar.gz`,
including the Go binary, bundled UI/fonts/icons, Ultralight runtime, GPU bridge,
and shaders. Run the target smoke test with:

```sh
machine/macbookpro12,1-debian/base/macbook-ssh.sh \
  'cd /tmp/audio-utility-debian && ./run-target.sh --full --self-test --capture --seconds 8'
```

The dashboard follows `../design/mockup.png`. Its controls are interactive
prototypes; audio-device control remains future work. CI produces the Debian
bundle and separate Alpine binary as assets for `v*` tags.
