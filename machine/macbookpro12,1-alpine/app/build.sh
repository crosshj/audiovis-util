#!/bin/sh
# Build the app for Alpine Linux on the MacBookPro12,1 (x86_64).
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
output=${1:-$repo_root/build/macbookpro12,1/audio-utility}
mkdir -p "$(dirname -- "$output")"

build_app() (
	cd "$repo_root/src-alpine"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
		go build -trimpath -ldflags='-s -w -buildid=' -o "$output" .
)

if [ -f /etc/alpine-release ]; then
	build_app
elif [ -n "${CC:-}" ]; then
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC="$CC" build_app
elif command -v docker >/dev/null 2>&1; then
	case "$output" in
		"$repo_root"/*) output_relative=${output#"$repo_root"/} ;;
		*) echo 'With Docker, the output path must be inside the repository.' >&2; exit 1 ;;
	esac
	docker run --rm -v "$repo_root:/workspace" -w /workspace \
		golang:1.26-alpine3.24 sh -eu -c \
		"apk add --no-cache build-base >/dev/null && machine/macbookpro12,1-alpine/app/build.sh '/workspace/$output_relative'"
else
	echo 'A musl build environment is required. Run inside Alpine, set CC to an x86_64-linux-musl cross-compiler, or install Docker.' >&2
	exit 1
fi

chmod 755 "$output"
printf 'Built MacBookPro12,1 app: %s\n' "$output"
if command -v file >/dev/null 2>&1; then
	file "$output"
fi
