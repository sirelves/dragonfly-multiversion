#!/usr/bin/env bash
# gen-fixtures.sh <protocol> <gophertunnel tag>: encodes the conformance
# ItemStackRequest fixtures with that tag and merges them into item_stack.json.
set -euo pipefail
cd "$(dirname "$0")"
. ./lib.sh
proto=$1 tag=${2:-$(e2e_tag "$1")}
dir=build/fixtures-$proto
mkdir -p "$dir"
cp fixtures/main.go "$dir/main.go"
cp "${FIXTURE_ACTIONS:-fixtures/actions.go.txt}" "$dir/actions.go"
cp "${FIXTURE_DAMAGE:-fixtures/damage.go.txt}" "$dir/damage.go"
if [ ! -f "$dir/go.mod" ]; then
  (cd "$dir" && go mod init fixtures >/dev/null 2>&1 && go get "github.com/sandertv/gophertunnel@$tag" >/dev/null 2>&1 && go mod tidy >/dev/null 2>&1)
fi
rev=$(git -C "${GT_HISTORY:?set GT_HISTORY to a gophertunnel clone}" rev-parse "$tag^{commit}")
file=../multiversion/itemdata/runtime_item_states_$proto.json
[ -f "$file" ] || file=../multiversion/itemdata/runtime_item_states_$((proto - 1)).json  # 568 shares 567
sword=$(python3 -c "import json,sys;print(next(i['id'] for i in json.load(open('$file')) if i['name']=='minecraft:stone_sword'))")
json=$(cd "$dir" && go run . -stone-sword "$sword" -revision "$rev")
python3 - "$proto" "$json" <<'PY'
import json, sys
out = json.loads(sys.argv[2])
for kind in ("item_stack", "damage"):
    path = "../multiversion/conformance/testdata/%s.json" % kind
    data = json.load(open(path))
    data[sys.argv[1]] = out[kind]
    json.dump(data, open(path, "w"), indent=2, sort_keys=True)
    open(path, "a").write("\n")
PY
echo "fixtures for $proto ($tag, stone_sword=$sword)"
