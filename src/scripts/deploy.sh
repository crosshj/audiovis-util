#!/bin/sh
# Build and stage the Debian product bundle in target RAM for a smoke test.
set -eu
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo=$(CDPATH= cd -- "$here/../../.." && pwd)
exec "$repo/machine/macbookpro12,1-debian/app/deploy-test.sh" "$@"
