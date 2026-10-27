#!/usr/bin/env python3
"""Compare Linux archive ownership/modes to the compressed filesystem listing."""
import os
import re
import stat
import subprocess
import sys
import tarfile

archive, image = sys.argv[1:]
tool = os.environ.get("UNSQUASHFS", "unsquashfs")
listing = subprocess.check_output([tool, "-lln", image], text=True)
entries = {}
for line in listing.splitlines():
    match = re.match(r"(\S+)\s+(\d+)/(\d+)\s+\d+\s+\S+\s+\S+\s+squashfs-root(?:/(.*))?$", line)
    if match:
        mode, uid, gid, name = match.groups()
        name = (name or "").split(" -> ", 1)[0]
        entries[name] = (mode, int(uid), int(gid))
assert entries[""] == ("drwxr-xr-x", 0, 0), entries.get("")
checked = 0
with tarfile.open(archive) as source:
    members = {m.name.removeprefix("./").rstrip("/"): m for m in source}
    for name, member in members.items():
        if name in ("", "."):
            continue
        mode = member.mode
        if member.islnk():
            target = members[member.linkname.removeprefix("./").rstrip("/")]
            mode = target.mode
        kind = stat.S_IFDIR if member.isdir() else stat.S_IFLNK if member.issym() else stat.S_IFREG
        expected = (stat.filemode(kind | mode), member.uid, member.gid)
        assert entries.get(name) == expected, (name, expected, entries.get(name))
        checked += 1
print(f"PASS: root ownership/mode and {checked} archive entry ownership/modes match.")
