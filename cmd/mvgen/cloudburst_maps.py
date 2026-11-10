#!/usr/bin/env python3
"""cloudburst_maps.py <cloudburst-protocol dir> <IdotBedrockData dir> <out root> <protocol>...

BedrockData only has level_sound_id_map.json and particle_id_map.json from
1.18.10 (486) on. CloudburstMC/Protocol defines both tables for every codec as
TypeMap builder chains (insert/replace/shift/remove/update). This script
evaluates those chains, then names each entry the way BedrockData does by
joining on the 486 tables, where both sources exist, and writes the tables of
the requested protocols to <out root>/<protocol>/.
"""
import json, os, re, sys

cb, data, out = sys.argv[1], sys.argv[2], sys.argv[3]
protos = [int(p) for p in sys.argv[4:]]
codec = os.path.join(cb, "bedrock-codec/src/main/java/org/cloudburstmc/protocol/bedrock/codec")
TABLES = {"SOUND_EVENTS": "SoundEvent", "PARTICLE_TYPES": "ParticleType"}

def source(v):
    with open(os.path.join(codec, "v%d" % v, "Bedrock_v%d.java" % v)) as f:
        return f.read()

versions = sorted(int(d[1:]) for d in os.listdir(codec) if re.fullmatch(r"v\d+", d))
cache = {}

def table(name, v):
    """The id → enum name table `name` as codec v defines or inherits it."""
    if (name, v) in cache:
        return cache[(name, v)]
    src = source(v)
    m = re.search(r"\b%s\s*=\s*(.*?)\.build\(\);" % name, src, re.S)
    if not m:
        prev = [x for x in versions if x < v]
        res = table(name, prev[-1]) if prev else {}
    else:
        body = m.group(1)
        base = re.match(r"\s*(?:Bedrock_v(\d+)\.%s\.toBuilder\(\)|TypeMap\.builder\([^)]*\))" % name, body)
        assert base, (name, v, body[:80])
        res = dict(table(name, int(base.group(1)))) if base.group(1) else {}
        enum = TABLES[name]
        for call in re.finditer(r"\.(insert|replace|shift|remove|update)\(([^()]*)\)", body[base.end():]):
            op, args = call.group(1), [a.strip() for a in call.group(2).split(",")]
            if op in ("insert", "replace"):
                if not args[1].startswith(enum + "."):
                    continue  # insert(offset, otherMap) is not used for these tables
                res[int(eval(args[0]))] = args[1][len(enum) + 1:]
            elif op == "remove":
                res.pop(int(args[0]), None)
            elif op == "shift":
                start, amount = int(args[0]), int(args[1])
                moved = {k + amount: val for k, val in res.items() if k >= start}
                res = {k: val for k, val in res.items() if k < start}
                res.update(moved)
            elif op == "update":
                res.pop(int(args[0]), None)
                res[int(args[1])] = args[2][len(enum) + 1:]
    cache[(name, v)] = res
    return res

def bridge(name, jsonfile, at):
    """enum name → BedrockData name, from the tables both sources have at `at`."""
    with open(os.path.join(data, "v%d" % at, jsonfile)) as f:
        named = json.load(f)
    by_id = {}
    for n, i in named.items():
        by_id.setdefault(i, n)
    return {e: by_id[i] for i, e in table(name, at).items() if i in by_id and e != "UNDEFINED"}

files = {"SOUND_EVENTS": "level_sound_id_map.json", "PARTICLE_TYPES": "particle_id_map.json"}
names = {k: {} for k in files}
for at in (503, 486):  # 486 wins: closest to the old tables
    for k, f in files.items():
        names[k].update(bridge(k, f, at))

def build(k, v):
    return {names[k][e]: i for i, e in sorted(table(k, v).items()) if e in names[k]}

# Self-check: the evaluated tables must reproduce BedrockData where it exists.
for at in (486, 503, 527):
    for k, f in files.items():
        with open(os.path.join(data, "v%d" % at, f)) as fh:
            want = json.load(fh)
        got = build(k, at)
        bad = {n: (got[n], want[n]) for n in got if n in want and got[n] != want[n]}
        if bad:
            sys.exit("%s %d disagrees with BedrockData: %s" % (f, at, list(bad.items())[:5]))
        print("check %d %s: %d/%d entries agree" % (at, f, len(got), len(want)))

for p in protos:
    d = os.path.join(out, str(p))
    os.makedirs(d, exist_ok=True)
    for k, f in files.items():
        t = build(k, p)
        with open(os.path.join(d, f), "w") as fh:
            json.dump(dict(sorted(t.items())), fh, indent=4)
            fh.write("\n")
        print("%d %s: %d entries" % (p, f, len(t)))
