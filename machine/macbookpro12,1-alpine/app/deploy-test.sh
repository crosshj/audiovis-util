#!/bin/sh
# Build locally and stage the test binary on the writable audio-utility volume.
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
build_script=$repo_root/machine/macbookpro12,1-alpine/app/build.sh
ssh_script=$repo_root/machine/macbookpro12,1-alpine/base/scripts/macbook-ssh.sh
binary=$repo_root/build/macbookpro12,1/audio-utility
remote_dir=/media/audio-utility/apps/macbookpro12,1

"$build_script" "$binary"
"$ssh_script" "mount | grep -E '^/dev/[^ ]+ on /media/audio-utility type hfsplus \\(rw,' >/dev/null && mkdir -p '$remote_dir' && cat > '$remote_dir/audio-utility.new' && chmod 755 '$remote_dir/audio-utility.new' && mv '$remote_dir/audio-utility.new' '$remote_dir/audio-utility' && sync && ls -lh '$remote_dir/audio-utility'" < "$binary"
"$ssh_script" "cat > '$remote_dir/run-test.new' && chmod 755 '$remote_dir/run-test.new' && mv '$remote_dir/run-test.new' '$remote_dir/run-test.sh' && sync && echo 'Staged 30-second KMSDRM test launcher.'" < "$repo_root/machine/macbookpro12,1-alpine/app/run-test.sh"
printf 'Test build deployed. Run it on the target with:\n  %s "sh %s/run-test.sh"\n' \
	"$ssh_script" "$remote_dir"
