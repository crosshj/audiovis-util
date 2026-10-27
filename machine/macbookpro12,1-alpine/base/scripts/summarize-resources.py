#!/usr/bin/env python3
"""Summarize profile-resources.sh output locally; never install on the target."""
import json
import pathlib
import statistics
import sys


def load(path):
    meta, samples, current = {}, [], None
    for line in pathlib.Path(path).read_text().splitlines():
        if line.startswith('SAMPLE '):
            _, number, uptime, *_ = line.split()
            current = {'number': int(number), 'uptime': float(uptime), 'mem': {}, 'app': {}, 'disks': {}, 'drm': {}}
            samples.append(current)
        elif line.startswith('END_SAMPLE'):
            current = None
        elif current is not None:
            if line.startswith('cpu '):
                current['cpu'] = list(map(int, line.split()[1:9]))
            elif line.startswith(('Mem', 'Buffers:', 'Cached:', 'Shmem:', 'Slab:', 'SReclaimable:', 'SUnreclaim:', 'AnonPages:', 'Swap')):
                key, value, *_ = line.split()
                current['mem'][key.rstrip(':')] = int(value)
            elif line.startswith('APP_STAT '):
                stat = line[len('APP_STAT '):]
                fields = stat.rsplit(')', 1)[1].split()
                current['app']['pid'] = int(stat.split()[0])
                current['app']['ticks'] = int(fields[11]) + int(fields[12])
            elif line.startswith(('APP_STATUS ', 'APP_SMAPS ')):
                section, key, value, *_ = line.split()
                current['app'][section + '_' + key.rstrip(':')] = int(value)
            elif line.startswith('DISK '):
                fields = line.split()[1:]
                current['disks'][fields[2]] = {'read_sectors': int(fields[5]), 'write_sectors': int(fields[9])}
            elif line.startswith('APP_DRM '):
                _, fd, key, value, *units = line.split()
                key = key.rstrip(':')
                info = current['drm'].setdefault(fd, {})
                info[key] = value
                if key.startswith('drm-resident-'):
                    scale = {'KiB': 1024, 'MiB': 1048576, 'bytes': 1}.get(units[0] if units else 'bytes', 1)
                    info[key + '_bytes'] = int(value) * scale
        elif line.startswith(('DATE ', 'KERNEL ', 'MODEL ', 'HZ ', 'CPUS ', 'CPU_MODEL ', 'OS ', 'APP_SHA256 ')):
            key, value = line.split(' ', 1)
            meta[key] = value
    return meta, samples


def stats(values):
    return {'mean': statistics.mean(values), 'min': min(values), 'max': max(values)} if values else None


def summarize(path):
    meta, s = load(path)
    if len(s) < 2:
        raise ValueError(f'{path}: need at least two samples')
    cpu, app_cpu = [], []
    for a, b in zip(s, s[1:]):
        delta = [y-x for x, y in zip(a['cpu'], b['cpu'])]
        cpu.append(100 * (sum(delta)-delta[3]-delta[4]) / sum(delta))
        if a['app'].get('pid') == b['app'].get('pid') and 'ticks' in a['app']:
            app_cpu.append(100 * (b['app']['ticks']-a['app']['ticks']) / int(meta['HZ']) / (b['uptime']-a['uptime']))
    delta = [y-x for x, y in zip(s[0]['cpu'], s[-1]['cpu'])]
    elapsed = s[-1]['uptime']-s[0]['uptime']
    result = {'source': str(path), 'metadata': meta, 'samples': len(s), 'elapsed_seconds': elapsed,
              'system_cpu_percent_all_cpus': 100 * (sum(delta)-delta[3]-delta[4]) / sum(delta),
              'system_cpu_intervals': stats(cpu), 'app_cpu_percent_one_core': stats(app_cpu)}
    result['memory_mib'] = {key: stats([x['mem'][key]/1024 for x in s]) for key in s[0]['mem']}
    result['memory_mib']['Unavailable'] = stats([(x['mem']['MemTotal']-x['mem']['MemAvailable'])/1024 for x in s])
    result['memory_mib']['Nonfree'] = stats([(x['mem']['MemTotal']-x['mem']['MemFree'])/1024 for x in s])
    result['app'] = {key: stats([x['app'][key] for x in s if key in x['app']])
                     for key in sorted({k for x in s for k in x['app']}) if key not in ('pid', 'ticks')}
    result['disk_io_mib'] = {disk: {key: (s[-1]['disks'][disk][key]-v)*512/1048576
                                    for key, v in counters.items()} for disk, counters in s[0]['disks'].items()}
    # fdinfo can repeat a DRM client on several file descriptors; deduplicate.
    def engines(sample):
        clients = {}
        for fd, info in sample['drm'].items():
            clients[info.get('drm-client-id', fd)] = info
        totals = {}
        for info in clients.values():
            for key, value in info.items():
                if key.startswith('drm-engine-') and not key.startswith('drm-engine-capacity-'):
                    totals[key] = totals.get(key, 0) + int(value)
        return totals
    first, last = engines(s[0]), engines(s[-1])
    result['app_gpu_engine_busy_percent'] = {key: 100*(last[key]-value)/(elapsed*1e9) for key, value in first.items() if key in last}
    gpu_memory = []
    for sample in s:
        clients = {}
        for fd, info in sample['drm'].items():
            clients[info.get('drm-client-id', fd)] = info
        if clients:
            gpu_memory.append(sum(v for info in clients.values() for k, v in info.items()
                                  if k.startswith('drm-resident-') and k.endswith('_bytes'))/1048576)
    result['app_gpu_resident_mib'] = stats(gpu_memory)
    return result

if __name__ == '__main__':
    print(json.dumps([summarize(path) for path in sys.argv[1:]], indent=2))
