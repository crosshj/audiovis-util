#!/bin/sh
set -eu
repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
build_script=$repo_root/machine/macbookpro12,1-debian/app/build.sh
ssh_script=$repo_root/machine/macbookpro12,1-debian/base/macbook-ssh.sh
bundle=$repo_root/build/macbookpro12,1-debian/audio-utility-debian-amd64.tar.gz
remote_dir=/media/audiovis-util/apps/macbookpro12,1

# The reference Alpine target must never receive Debian deployment writes.
"$ssh_script" '. /etc/os-release; [ "$ID" = debian ] || { echo "Refusing deployment: target is not Debian." >&2; exit 1; }'
"$build_script" "$repo_root/build/macbookpro12,1-debian"
"$ssh_script" '. /etc/os-release; [ "$ID" = debian ] && ! pgrep -x audio-utility && mkdir -p /tmp/audio-utility-debian && tar -xzf - -C /tmp/audio-utility-debian && chmod 755 /tmp/audio-utility-debian/audio-utility' < "$bundle"
echo 'Debian bundle staged in /tmp/audio-utility-debian. Run run-target.sh there.'
