#!/usr/bin/env python3
"""cloudburst_entity.py <cloudburst-protocol dir> <out.go> <protocol>...

Entity metadata keys and entity flags move between protocols. CloudburstMC
defines both per codec (ENTITY_DATA as an EntityDataTypeMap, ENTITY_FLAGS as a
TypeMap). This evaluates them for the current protocol (2225, the numbering
gophertunnel and so Dragonfly write) and each requested protocol, and writes the
moves as Go tables for multiversion/v786.
"""
import os, re, sys

cb, out, protos = sys.argv[1], sys.argv[2], [int(p) for p in sys.argv[3:]]
codec = os.path.join(cb, "bedrock-codec/src/main/java/org/cloudburstmc/protocol/bedrock/codec")
versions = sorted(int(d[1:]) for d in os.listdir(codec) if re.fullmatch(r"v\d+", d))
CURRENT = 2225

def source(v):
    with open(os.path.join(codec, "v%d" % v, "Bedrock_v%d.java" % v)) as f:
        return f.read()

def chain(name, v):
    """The builder chain of `name` in codec v: (base version or None, body), or
    None when v inherits it."""
    m = re.search(r"\b%s\s*=\s*(.*?)\.build\(\);" % name, source(v), re.S)
    if not m:
        return None
    body = m.group(1)
    base = re.match(r"\s*(?:Bedrock_v(\d+)\.%s\s*\.toBuilder\(\)|\w+\.builder\([^)]*\))" % name, body)
    assert base, (name, v, body[:80])
    return (int(base.group(1)) if base.group(1) else None), body[base.end():]

def calls(body):
    depth, start, i = 0, None, 0
    while i < len(body):
        if body.startswith(".", i) and depth == 0:
            m = re.match(r"\.(\w+)\(", body[i:])
            if m:
                name, j, depth2, k = m.group(1), i + m.end(), 1, i + m.end()
                while depth2:
                    depth2 += {"(": 1, ")": -1}.get(body[k], 0)
                    k += 1
                yield name, [a.strip() for a in split_args(body[j:k - 1])]
                i = k
                continue
        i += 1

def split_args(s):
    args, depth, cur = [], 0, ""
    for c in s:
        if c == "," and depth == 0:
            args.append(cur); cur = ""
            continue
        depth += {"(": 1, ")": -1, "<": 0, ">": 0}.get(c, 0)
        cur += c
    if cur.strip():
        args.append(cur)
    return args

cache = {}
def entity_data(v):
    """type name → key ID for codec v."""
    if ("D", v) in cache:
        return cache[("D", v)]
    c = chain("ENTITY_DATA", v)
    if c is None:
        prev = [x for x in versions if x < v]
        res = dict(entity_data(prev[-1]))
    else:
        base, body = c
        res = dict(entity_data(base)) if base else {}
        for op, a in calls(body):
            if op == "insert":
                res[a[0].split(".")[-1]] = int(a[1])
            elif op == "remove":
                idx = int(a[0])
                res = {t: i for t, i in res.items() if i != idx}
            elif op == "shift":
                start, delta = int(a[0]), int(a[1])
                length = int(a[2]) if len(a) > 2 else 10**6
                res = {t: (i + delta if start <= i < start + length else i) for t, i in res.items()}
            elif op == "update" and len(a) == 3:  # update(type, old, new) style, if any
                pass
    cache[("D", v)] = res
    return res

def flags(v):
    if ("F", v) in cache:
        return cache[("F", v)]
    c = chain("ENTITY_FLAGS", v)
    if c is None:
        prev = [x for x in versions if x < v]
        res = dict(flags(prev[-1]))
    else:
        base, body = c
        res = dict(flags(base)) if base else {}
        for op, a in calls(body):
            if op in ("insert", "replace") and a[1].startswith("EntityFlag."):
                res = {e: i for e, i in res.items() if i != int(a[0])}
                res[a[1].split(".")[-1]] = int(a[0])
            elif op == "remove":
                res = {e: i for e, i in res.items() if i != int(a[0])}
            elif op == "shift":
                start, delta = int(a[0]), int(a[1])
                res = {e: (i + delta if i >= start else i) for e, i in res.items()}
    cache[("F", v)] = res
    return res

def remap(cur, old):
    """current index → old index for every current index that differs; -1 when
    the old protocol has nothing for it. Several names may share an index
    (aliases); one of them known to the old protocol decides it."""
    by_old, by_cur = {}, {}
    for name, i in old.items():
        by_old.setdefault(i, []).append(name)
    for name, i in cur.items():
        by_cur.setdefault(i, []).append(name)
    top = max(old.values())
    out = {}
    for i, names in sorted(by_cur.items()):
        known = [n for n in names if n in old]
        if known:
            if old[known[0]] != i:
                out[i] = old[known[0]]
            continue
        occupants = by_old.get(i, [])
        if any(o in cur for o in occupants):
            out[i] = -1  # the slot belongs to an entry that moved there
        elif occupants:
            pass  # renamed since: unchanged
        elif i > top:
            out[i] = -1  # newer than the old protocol
        # else: a gap CloudburstMC never named: unchanged
    return out

cur_d, cur_f = entity_data(CURRENT), flags(CURRENT)
lines = ["// Code generated by cmd/mvgen/cloudburst_entity.py from CloudburstMC/Protocol. DO NOT EDIT.", "",
         "package v786", "",
         "// legacyEntityKeys maps, per protocol, the entity metadata keys Dragonfly",
         "// writes (gophertunnel's, protocol %d) to that protocol's keys. Keys not" % CURRENT,
         "// listed are unchanged; -1 means the protocol has no such key.",
         "var legacyEntityKeys = map[uint32]map[uint32]int32{"]
flag_lines = ["// legacyEntityFlags maps, per protocol, entity flag indices of protocol %d to" % CURRENT,
              "// that protocol's. Flags not listed are unchanged; -1 means it has no such flag.",
              "var legacyEntityFlags = map[uint32]map[uint8]int16{"]
for p in protos:
    d, f = entity_data(p), flags(p)
    moves = remap(cur_d, d)
    lines.append("\t%d: {%s}," % (p, ", ".join("%d: %d" % (k, moves[k]) for k in sorted(moves))))
    fm = remap(cur_f, f)
    flag_lines.append("\t%d: {%s}," % (p, ", ".join("%d: %d" % (k, fm[k]) for k in sorted(fm))))
    print(p, "keys moved:", len(moves), "flags moved:", len(fm))
lines.append("}")
flag_lines.append("}")
with open(out, "w") as fh:
    fh.write("\n".join(lines + [""] + flag_lines) + "\n")
