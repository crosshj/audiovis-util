# Alpine SDL app

This is the existing SDL3 application built for the Alpine/musl machine target.
It remains separate from the Debian GPU dashboard in `src/`. Its gesture
recognizer is a local copy, so the Alpine app is an independent Go module and
does not depend on `src/`.

Build it from the repository root with:

```sh
machine/macbookpro12,1-alpine/app/build.sh
```

The build produces the `audio-utility-alpine-amd64` release binary for
MacBookPro12,1. Alpine provides the SDL3 runtime through its machine base.
