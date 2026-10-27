#!/bin/sh
# Boot the stock ISO in a disposable local emulator. No physical disks attached.
set -eu
iso=${1:?Usage: prepare-local-vm.sh /path/to/verified-standard.iso}
repo=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
output=$repo/build/macbookpro12,1-debian
mkdir -p "$output/vm-source"
# Extract the entire live directory so ISO hardlink targets resolve correctly.
if [ ! -r "$output/vm-source/live/vmlinuz" ]; then
    bsdtar -xf "$iso" -C "$output/vm-source" live
fi
qemu=${QEMU:-qemu-system-x86_64}
set -- -machine q35 -accel tcg -cpu max -m 3072 -smp 2 \
    -display none -serial stdio -monitor none -no-reboot \
    -kernel "$output/vm-source/live/vmlinuz" \
    -initrd "$output/vm-source/live/initrd.img" \
    -append 'boot=live components init=/bin/bash console=ttyS0 net.ifnames=0' \
    -cdrom "$iso" -nic user,model=virtio-net-pci \
    -virtfs "local,path=$repo,mount_tag=work,security_model=none,id=work"
if [ -n "${QEMU_DATA_DIR:-}" ]; then set -- -L "$QEMU_DATA_DIR" "$@"; fi
exec "$qemu" "$@"
