#!/usr/bin/env python3
"""addversion.py <protocol> <version> [shares-data-with]: registers a legacy
protocol in multiversion/legacy and v630 with an empty delta file."""
import os, re, sys

proto, version = int(sys.argv[1]), sys.argv[2]
root = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "multiversion")

def edit(path, f):
    with open(path) as fh:
        s = fh.read()
    t = f(s)
    assert t != s, path
    with open(path, "w") as fh:
        fh.write(t)

p = os.path.join(root, "legacy", "protocol.go")
edit(p, lambda s: re.sub(r"var order = \[\]int32\{([^}]*)\}", lambda m: "var order = []int32{%s, %d}" % (m.group(1), proto), s))
edit(p, lambda s: re.sub(r"(var versions = map\[int32\]string\{\n(?:\t[^\n]*\n)*)", lambda m: m.group(1) + '\t%d: "%s",\n' % (proto, version), s, count=1))
edit(p, lambda s: re.sub(r"var deltas = \[\]delta\{([^}]*)\}", lambda m: "var deltas = []delta{%s, delta%d}" % (m.group(1), proto), s))
edit(os.path.join(root, "v630", "registry.go"), lambda s: s.replace("var versions = map[uint32]string{", 'var versions = map[uint32]string{%d: "%s", ' % (proto, version), 1))
d = os.path.join(root, "legacy", "v%d.go" % proto)
if not os.path.exists(d):
    with open(d, "w") as fh:
        fh.write("package legacy\n\n// %s (%d).\nvar delta%d = delta{proto: %d}\n" % (version, proto, proto, proto))
print("registered", proto, version)
