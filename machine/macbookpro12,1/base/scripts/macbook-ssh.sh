#!/bin/sh
# Open an interactive shell on the Alpine MacBook, or run a remote command.
set -eu

machine_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
if [ -r "$machine_dir/.env" ]; then
	set -a
	. "$machine_dir/.env"
	set +a
fi
: "${MACBOOK_SSH_HOST:?Set MACBOOK_SSH_HOST in machine/macbookpro12,1/.env}"

ssh_args=
if [ "$#" -eq 0 ]; then
	ssh_args=-tt
fi

# Intentional word splitting supplies -tt only for interactive use.
# shellcheck disable=SC2086
exec ssh $ssh_args \
	-i "$HOME/.ssh/audio-utility_ed25519" \
	-o IdentitiesOnly=yes \
	-o ConnectTimeout=8 \
	-o ServerAliveInterval=15 \
	-o ServerAliveCountMax=3 \
	root@"$MACBOOK_SSH_HOST" "$@"
