# audiovis-util

This repository contains software for a local-network workstation, starting
with an audio utility app. The app is its own full-screen environment: it owns
the display, pointer, and user interaction. It does not run inside a
conventional desktop or window manager.

The app runs on a prepared machine base that provides Linux, hardware support,
Wi-Fi, and remote development access. Machine-specific setup is kept separate
from the app so the software can eventually run on other supported machines.

## Repository layout

- [`src/`](src/README.md): the Debian GPU dashboard product app.
- [`src-alpine/`](src-alpine/README.md): the Alpine SDL/musl app.
- [`docs/app-roadmap.md`](docs/app-roadmap.md): product and app roadmap.
- [`machine/macbookpro12,1-alpine/`](machine/macbookpro12,1-alpine/README.md): existing Alpine base for the Early 2015 13-inch MacBook Pro.
- [`machine/macbookpro12,1-debian/`](machine/macbookpro12,1-debian/README.md): Debian setup path for the same MacBook.
- [`machine/macbookpro12,1-alpine/app/`](machine/macbookpro12,1-alpine/app/README.md): the
  x86_64 Alpine build and KMSDRM test/deployment path for that target.
- [`machine/macbookpro12,1-alpine/roadmap.md`](machine/macbookpro12,1-alpine/roadmap.md): the
  original machine/base roadmap, preserved separately from the app roadmap.
- [`docs/archive/roadmap-next-desktop-draft.md`](docs/archive/roadmap-next-desktop-draft.md):
  an earlier app roadmap draft, retained for history. Its desktop-environment
  proposal is superseded by the current app-as-environment direction.

## Local machine configuration

Copy [`machine/macbookpro12,1-alpine/.env.example`](machine/macbookpro12,1-alpine/.env.example)
to `machine/macbookpro12,1-alpine/.env` and fill in that machine's local network and
access settings. The base builder reads it and writes generated configuration
under the ignored `build/` directory. Do not commit `.env` or generated payloads.

The Debian variant has a separate
[`machine/macbookpro12,1-debian/.env.example`](machine/macbookpro12,1-debian/.env.example)
and setup flow. Its generated config is also written under ignored `build/`.

See the [MacBook base README](machine/macbookpro12,1-alpine/README.md) for what the
builder creates and how that payload fits into the existing Alpine boot base.

## Current state

The Debian product app is the GPU accelerated dashboard in `src/app/`; it has
passed a fullscreen target interaction/rendering smoke test. Alpine retains its
independent SDL/musl app in `src-alpine/`. GitHub Actions builds separate Debian
and Alpine outputs and attaches both to `v*` releases. Audio controls still use
sample data; see the app and machine roadmaps for remaining work.
