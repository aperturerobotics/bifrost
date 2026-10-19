#!/usr/bin/env python3
"""Copies libdeflate's gzip decoder into this package as flat files.

Usage: vendor-libdeflate.py <libdeflate source root>

go mod vendor copies only a package's own directory, so every source and
header sits beside the Go binding, and each include names its flat file.
"""
import os
import re
import shutil
import sys

ENTRY = [
    "lib/deflate_decompress.c",
    "lib/gzip_decompress.c",
    "lib/crc32.c",
    "lib/utils.c",
    "lib/x86/cpu_features.c",
    "lib/arm/cpu_features.c",
]
CONSTRAINT = "//go:build libdeflate && cgo\n\n"
INCLUDE = re.compile(r'^(\s*#\s*include\s+)"([^"]+)"', re.M)


def flat(rel):
    """Returns the flat file name for a path under the source root."""
    if rel.startswith("lib/"):
        rel = rel[len("lib/"):]
    return rel.replace("/", "_")


def resolve(root, rel, name):
    """Resolves an include from the including file, then from lib/."""
    for base in (os.path.dirname(rel), "lib"):
        cand = os.path.normpath(os.path.join(base, name))
        if os.path.exists(os.path.join(root, cand)):
            return cand
    raise SystemExit(f"{rel}: cannot resolve {name}")


def main():
    root, out = sys.argv[1], os.path.dirname(os.path.abspath(__file__))
    todo, done = list(ENTRY), set()
    while todo:
        rel = todo.pop()
        if rel in done:
            continue
        done.add(rel)
        with open(os.path.join(root, rel)) as f:
            src = f.read()

        def sub(m):
            dep = resolve(root, rel, m.group(2))
            todo.append(dep)
            return f'{m.group(1)}"{flat(dep)}"'

        src = INCLUDE.sub(sub, src)
        if rel.endswith(".c"):
            src = CONSTRAINT + src
        with open(os.path.join(out, flat(rel)), "w") as f:
            f.write(src)
    shutil.copy(os.path.join(root, "COPYING"), os.path.join(out, "COPYING"))
    print("\n".join(sorted(flat(r) for r in done)))


main()
