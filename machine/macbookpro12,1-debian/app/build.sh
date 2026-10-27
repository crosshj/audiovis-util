#!/bin/sh
# Build the Ultralight GPU app bundle for Debian amd64 on MacBookPro12,1.
set -eu
repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
output=${1:-$repo_root/build/macbookpro12,1-debian}
"$repo_root/src/scripts/build.sh" "$output"
