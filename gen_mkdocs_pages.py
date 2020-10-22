#!/usr/bin/env python3
import os
import sys
import json

if len(sys.argv) != 2:
    print(f'usage: {sys.argv[0]} <path to generated reference docs>', file=sys.stderr)
    sys.exit(1)

path = sys.argv[1]

nav = []
for f in os.listdir(path):
    if not f.endswith('.md'):
        continue

    nav.append({f[:-3].replace('_', ' '): f})

nav = list(sorted(nav, key=lambda i: next(iter(i))))

json.dump({'nav': nav}, sys.stdout, indent=2)
