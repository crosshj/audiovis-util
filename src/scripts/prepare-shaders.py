#!/usr/bin/env python3
"""Extract pinned SDK shaders and adapt desktop GLSL 150 to GLES 300 es.
Generated SDK code remains in the ignored build/runtime directories.
"""
import pathlib, re, sys
source, output = map(pathlib.Path, sys.argv[1:])
output.mkdir(parents=True, exist_ok=True)
for name in ('v2f_c4f_t2f_vert', 'v2f_c4f_t2f_t2f_d28f_vert', 'fill_path_frag', 'fill_frag'):
    header = (source / ('shader_' + name + '.h')).read_text()
    parts = re.findall(r'R"\((.*?)\)"', header, flags=re.S)
    if not parts:
        raise SystemExit('No GLSL found in ' + name)
    shader = ''.join(parts).replace('#version 150', '#version 300 es\nprecision highp float;\nprecision highp int;')
    # GLES requires an explicit float here; desktop GLSL accepts this literal.
    shader = shader.replace('radius * 2 + 0.2', 'radius * 2.0 + 0.2')
    (output / (name + '.glsl')).write_text(shader)
