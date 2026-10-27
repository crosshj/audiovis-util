# MacBookPro12,1 Debian app target

Build the GPU accelerated Audio Utility dashboard for Debian amd64:

```sh
machine/macbookpro12,1-debian/app/build.sh
```

The output is `build/macbookpro12,1-debian/audio-utility-debian-amd64.tar.gz`.
It contains the product binary, bundled UI and fonts/icons, Ultralight runtime,
GPU bridge, and shaders. The Debian base provides KMS/DRM, Mesa, Wi-Fi, and SSH;
the app bundle does not install packages or alter boot configuration.

Stage the bundle in target RAM and run the interaction/rendering smoke test:

```sh
machine/macbookpro12,1-debian/app/deploy-test.sh
machine/macbookpro12,1-debian/base/macbook-ssh.sh \
  'cd /tmp/audio-utility-debian && ./run-target.sh --full --self-test --capture --seconds 8'
```

The launcher requires Debian, refuses to compete with another running app,
uses the KMSDRM `opengles2` path, and restores the prior virtual terminal when
it exits. It runs manually; automatic launch remains a separate machine-base
decision.
