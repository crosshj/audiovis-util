# Audio Utility

This repository will contain the software for a local-network audio utility
workstation. The workstation app is its own full-screen environment: it owns the
display, pointer, and user interaction. It does not run inside a conventional
desktop or window manager.

The app runs on a prepared machine base that provides Linux, hardware support,
Wi-Fi, and remote development access. Machine-specific setup is kept separate
from the app so the software can eventually run on other supported machines.

## Repository layout

- [`src/`](src/README.md): the workstation app and its runtime.
- [`docs/app-roadmap.md`](docs/app-roadmap.md): product and app roadmap.
- [`machine/macbookpro12,1/`](machine/macbookpro12,1/README.md): Alpine base
  configuration for the Early 2015 13-inch MacBook Pro.
- [`machine/macbookpro12,1/app/`](machine/macbookpro12,1/app/README.md): the
  x86_64 Alpine build and KMSDRM test/deployment path for that target.
- [`machine/macbookpro12,1/roadmap.md`](machine/macbookpro12,1/roadmap.md): the
  original machine/base roadmap, preserved separately from the app roadmap.
- [`docs/archive/roadmap-next-desktop-draft.md`](docs/archive/roadmap-next-desktop-draft.md):
  an earlier app roadmap draft, retained for history. Its desktop-environment
  proposal is superseded by the current app-as-environment direction.

## Local machine configuration

Copy [`machine/macbookpro12,1/.env.example`](machine/macbookpro12,1/.env.example)
to `machine/macbookpro12,1/.env` and fill in that machine's local network and
access settings. The base builder reads it and writes generated configuration
under the ignored `build/` directory. Do not commit `.env` or generated payloads.

See the [MacBook base README](machine/macbookpro12,1/README.md) for what the
builder creates and how that payload fits into the existing Alpine boot base.

## Current state

The Go app has a windowed UI preview and a Linux x86_64 build that has run on
the MacBookPro12,1 through SDL's KMSDRM `opengles2` renderer. The audio UI still
uses sample data. GitHub Actions builds the machine target, uploads it as an
artifact, and creates a release asset for `v*` tags. See the app and machine
roadmaps for remaining work.
