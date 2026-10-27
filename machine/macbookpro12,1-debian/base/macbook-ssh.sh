#!/bin/sh
# Open an interactive shell on the Debian MacBook, or run a remote command.
set -eu

host_override=${MACBOOK_SSH_HOST:-}
user_override=${MACBOOK_SSH_USER:-}
machine_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
if [ -r "$machine_dir/.env" ]; then
	set -a
	. "$machine_dir/.env"
	set +a
fi
ssh_key=${SSH_PRIVATE_KEY_FILE:-$HOME/.ssh/audio-utility_ed25519}
ssh_host=${host_override:-${MACBOOK_SSH_HOST:-}}
ssh_user=${user_override:-${MACBOOK_SSH_USER:-root}}
[ -n "$ssh_host" ] || { echo 'Set MACBOOK_SSH_HOST or pass it in the environment.' >&2; exit 1; }
known_hosts=$machine_dir/../../build/macbookpro12,1-debian/known_hosts
[ -r "$known_hosts" ] || { echo 'Build the Debian card payload first to establish its SSH host identity.' >&2; exit 1; }

ssh_args=
if [ "$#" -eq 0 ]; then
	ssh_args=-tt
fi

# Intentional word splitting supplies -tt only for interactive use.
# shellcheck disable=SC2086
exec ssh $ssh_args \
	-i "$ssh_key" \
	-p "${MACBOOK_SSH_PORT:-22}" \
	-o HostKeyAlias=macbookpro12-1-debian \
	-o "UserKnownHostsFile=$known_hosts" \
	-o StrictHostKeyChecking=yes \
	-o IdentitiesOnly=yes \
	-o ConnectTimeout=8 \
	-o ServerAliveInterval=15 \
	-o ServerAliveCountMax=3 \
	"$ssh_user@$ssh_host" "$@"
