# Local preparation (never on the Alpine target)

The running Alpine machine is read-only reference material. Build packages in a
**disposable local Debian amd64 VM**, using the same verified Standard Live ISO
as the card. This preserves its exact installed package baseline. Do not use an
arbitrary newer Debian installation to calculate the missing dependencies.

With QEMU installed on the development machine:

```sh
base/prepare-local-vm.sh /path/to/debian-live-13.7.0-amd64-standard.iso
```

Paths below assume the repository is shared as `work`. At the VM's root prompt:

```sh
mount -t proc proc /proc
mount -t sysfs sysfs /sys
mkdir -p /work
mount -t 9p -o trans=virtio,version=9p2000.L work /work
ip link set eth0 up
ip address add 10.0.2.15/24 dev eth0
ip route add default via 10.0.2.2
printf 'nameserver 10.0.2.3\n' > /etc/resolv.conf
sh /work/machine/macbookpro12,1-debian/base/stage-packages.sh \
  /work/build/macbookpro12,1-debian
```

Run package staging **before installing anything** into that VM. APT validates
Debian's signed indexes and archive hashes. The resulting `packages/` contains
SSH's missing dependency closure; `runtime-packages/` contains graphics/audio
packages missing from the Live root. SHA-256 manifests detect damaged copies.
The boot installer uses `dpkg` only: it cannot fetch anything from the network.
It temporarily suppresses package-triggered service starts, then enables SSH
only after installing the key-only configuration.

## SDL version and application

The existing Go binding requires SDL 3.4; Debian trixie's SDL 3.2 fails with a
missing `SDL_AddAtomicU32` symbol. Keep the Go application unchanged and build
SDL 3.4 against Debian's libraries. Download the official source on the Mac:

```sh
mkdir -p build/macbookpro12,1-debian/sources
curl -fL -o build/macbookpro12,1-debian/sources/SDL3-3.4.0.tar.gz \
  https://github.com/libsdl-org/SDL/releases/download/release-3.4.0/SDL3-3.4.0.tar.gz
machine/macbookpro12,1-debian/app/build.sh
```

In the preparation VM (after staging the dependency closures):

```sh
apt-get install -y --no-install-recommends gcc make cmake pkgconf \
  libdrm-dev libgbm-dev libegl-dev libgles-dev libasound2-dev libudev-dev
sh /work/machine/macbookpro12,1-debian/base/build-sdl.sh \
  /work/build/macbookpro12,1-debian
```

`build-sdl.sh` verifies the release archive's pinned SHA-256 and builds a shared
SDL library with KMSDRM, EGL/GLES, ALSA, and no X11/Wayland dependency. Build
products go under ignored `build/macbookpro12,1-debian/`. Only runtime artifacts,
not the compiler or build dependencies, are copied to the card.

Back on the Mac, `base/build-config.sh` renders the private card payload,
including packages, stable SSH host keys, network settings, app, and SDL runtime
when present. Apply it using `copy-card-payload.sh`; no reformat is needed.
