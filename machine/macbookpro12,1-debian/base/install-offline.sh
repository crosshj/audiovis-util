#!/bin/sh
# Install a verified local package closure without starting services mid-install.
set -eu
cache=${1:?Usage: install-offline.sh /path/to/packages}
(cd "$cache" && sha256sum -c SHA256SUMS)
policy=/usr/sbin/policy-rc.d
saved=
if [ -e "$policy" ]; then
    saved=$(mktemp)
    cp -p "$policy" "$saved"
fi
restore() {
    if [ -n "$saved" ]; then
        cp -p "$saved" "$policy"
        rm -f "$saved"
    else
        rm -f "$policy"
    fi
}
trap restore EXIT HUP INT TERM
printf '#!/bin/sh\nexit 101\n' > "$policy"
chmod 755 "$policy"
DEBIAN_FRONTEND=noninteractive dpkg --install "$cache"/*.deb
