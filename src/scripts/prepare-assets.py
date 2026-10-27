#!/usr/bin/env python3
"""Verify vendored assets and bundle only the selected SVGs for offline loading."""
import hashlib
import json
from pathlib import Path
import shutil
import sys

source = Path(__file__).resolve().parent.parent / 'ui' / 'assets'
output = Path(sys.argv[1]) / 'ui' / 'assets'
for entry in json.loads((source / 'sources.json').read_text()):
    path = source / entry['path']
    if hashlib.sha256(path.read_bytes()).hexdigest() != entry['sha256']:
        raise SystemExit(f'Asset checksum mismatch: {path}')
shutil.copytree(source, output, dirs_exist_ok=True)
icons = {path.stem: path.read_text() for path in sorted((source / 'icons').glob('*.svg'))}
(output / 'icons.js').write_text('window.dashboardIcons=' + json.dumps(icons) + ';\n')
