#!/bin/sh
# Read-only sampler. Stream stdout back to the development Mac; install nothing.
# Usage: sh profile-resources.sh [samples=31] [interval_seconds=2]
set -eu
samples=${1:-31}
interval=${2:-2}
app=${APP_PATH:-/media/audio-utility/apps/macbookpro12,1/audio-utility}
printf 'DATE '; date -u +%Y-%m-%dT%H:%M:%SZ
printf 'KERNEL '; uname -r
printf 'MODEL '; cat /sys/class/dmi/id/product_name
printf 'HZ '; getconf CLK_TCK
printf 'CPUS '; grep -c '^processor' /proc/cpuinfo
printf 'CPU_MODEL '; sed -n 's/^model name[[:space:]]*: //p' /proc/cpuinfo | head -1
printf 'OS '; (. /etc/os-release; printf '%s %s\n' "$ID" "$VERSION_ID")
printf 'APP_SHA256 '; sha256sum "$app"
printf 'MOUNTS_BEGIN\n'; cat /proc/mounts; printf 'MOUNTS_END\n'
i=0
while [ "$i" -lt "$samples" ]; do
    printf 'SAMPLE %s ' "$i"; cat /proc/uptime
    awk '/^cpu / || /^ctxt / || /^processes / || /^procs_running / || /^procs_blocked /' /proc/stat
    awk '/^(MemTotal|MemFree|MemAvailable|Buffers|Cached|Shmem|Slab|SReclaimable|SUnreclaim|AnonPages|SwapTotal|SwapFree):/' /proc/meminfo
    printf 'LOAD '; cat /proc/loadavg
    for pid in $(pidof audio-utility 2>/dev/null || true); do
        printf 'APP_STAT '; cat "/proc/$pid/stat"
        awk '/^(VmRSS|VmHWM|VmSize|RssAnon|RssFile|RssShmem|Threads):/ {print "APP_STATUS", $0}' "/proc/$pid/status"
        if [ -r "/proc/$pid/smaps_rollup" ]; then
            awk '/^(Rss|Pss|Private_Clean|Private_Dirty|Swap):/ {print "APP_SMAPS", $0}' "/proc/$pid/smaps_rollup"
        fi
        for fd in /proc/"$pid"/fdinfo/*; do
            [ -r "$fd" ] || continue
            awk '/^drm-/ {print "APP_DRM", FILENAME, $0}' "$fd"
        done
    done
    awk '$3 ~ /^(sda|sdb)$/ {print "DISK", $0}' /proc/diskstats
    printf 'END_SAMPLE\n'
    i=$((i+1))
    [ "$i" -ge "$samples" ] || sleep "$interval"
done
