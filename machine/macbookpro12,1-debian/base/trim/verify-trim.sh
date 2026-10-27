#!/bin/sh
# Run on a booted trimmed Debian root, with the two manifests as arguments.
set -eu
remove=${1:?Usage: verify-trim.sh remove-packages.txt keep-packages.txt}
keep=${2:?}
. /etc/os-release
[ "$ID" = debian ]
while IFS= read -r package; do
    state=$(dpkg-query -W -f='${db:Status-Abbrev}' "$package" 2>/dev/null || true)
    case "$state" in ii*) echo "Unexpected installed package: $package" >&2; exit 1;; esac
done < "$remove"
while IFS= read -r package; do
    state=$(dpkg-query -W -f='${db:Status-Abbrev}' "$package" 2>/dev/null || true)
    case "$state" in ii*) ;; *) echo "Required package missing: $package" >&2; exit 1;; esac
done < "$keep"
for unit in cups.service cups-browsed.service cups.socket cups.path ipp-usb.service; do
    [ "$(systemctl show "$unit" -p LoadState --value)" = not-found ] || {
        echo "Printing activation unit remains: $unit" >&2; exit 1;
    }
done
audit=$(dpkg --audit)
[ -z "$audit" ] || { printf '%s\n' "$audit" >&2; exit 1; }
echo 'PASS: approved removals, protected packages, absent printing services, and clean dpkg audit.'
