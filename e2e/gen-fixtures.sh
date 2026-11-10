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
damage=fixtures/damage.go.txt
grep -q 'AttributeValue' "$(go env GOMODCACHE)/github.com/sandertv/gophertunnel@$tag/minecraft/protocol/attribute.go" || damage=fixtures/damage_old.go.txt
cp "${FIXTURE_DAMAGE:-$damage}" "$dir/damage.go"
src=$(go env GOMODCACHE)/github.com/sandertv/gophertunnel@$tag/minecraft/protocol
# AutoCraftRecipe gained TimesCrafted in 448 (CloudburstMC); gophertunnel
# v1.14.3, the 448 tag, forgot it, so that tag cannot produce the action.
if ! grep -A4 'type AutoCraftRecipeStackRequestAction struct' "$src/item_stack.go" | grep -q TimesCrafted; then
  if [ "$proto" = 448 ]; then
    sed -i '' '/"auto":/d' "$dir/actions.go"
  else
    sed -i '' 's/AutoCraftRecipeStackRequestAction{RecipeNetworkID: 123, TimesCrafted: 1}/AutoCraftRecipeStackRequestAction{RecipeNetworkID: 123}/' "$dir/actions.go"
  fi
fi
# Drop the fixtures of actions this tag does not have yet.
for t in $(grep -oE 'protocol\.[A-Za-z]+StackRequestAction\{' "$dir/actions.go" | sed 's/protocol\.//; s/{//' | sort -u); do
  grep -q "type $t struct" "$src"/*.go || sed -i '' "/protocol\.$t{/d" "$dir/actions.go"
done
# Before v1.21 the recipe actions named their ID RequestID and wrote it zigzag,
# while CloudburstMC writes it unsigned; 0 encodes the same either way.
for t in CraftRecipeStackRequestAction AutoCraftRecipeStackRequestAction CraftRecipeOptionalStackRequestAction CraftGrindstoneRecipeStackRequestAction; do
  if grep -q "type $t struct" "$src/item_stack.go" && ! grep -A4 "type $t struct" "$src/item_stack.go" | grep -q RecipeNetworkID; then
    sed -i '' "s/$t{RecipeNetworkID: 123, /$t{/; s/$t{RecipeNetworkID: 123}/$t{}/" "$dir/actions.go"
  fi
done
if ! grep -q 'FilterCause' "$src/item_stack.go"; then
  sed -i '' 's/, FilterCause: -1}}}/}}}/' "$dir/main.go"  # added in 554
fi
if [ ! -f "$dir/go.mod" ]; then
  (cd "$dir" && go mod init fixtures >/dev/null 2>&1 && go get "github.com/sandertv/gophertunnel@$tag" >/dev/null 2>&1 && go mod tidy >/dev/null 2>&1)
fi
# gophertunnel v1.12.4–v1.14.3 writes an empty item extra data (it encodes into
# a buffer, then writes the slice it started from). Fix that line in a copy of
# the tag so the fixtures carry the bytes a real client sends.
if sed -n '/^func (w \*Writer) Item(/,/^}/p' "$src/../protocol/writer.go" 2>/dev/null | grep -q 'w.ByteSlice(&extraData)'; then
  fixed=$PWD/build/gt-$tag
  if [ ! -d "$fixed" ]; then
    cp -r "$(go env GOMODCACHE)/github.com/sandertv/gophertunnel@$tag" "$fixed" && chmod -R u+w "$fixed"
    python3 - "$fixed/minecraft/protocol/writer.go" <<'PY'
import sys
p = sys.argv[1]; s = open(p).read()
i = s.index("func (w *Writer) Item(")
j = s.index("w.ByteSlice(&extraData)", i)
s = s[:j] + "extraData = buf.Bytes()\n\tw.ByteSlice(&extraData)" + s[j + len("w.ByteSlice(&extraData)"):]
open(p, "w").write(s)
PY
  fi
  (cd "$dir" && go mod edit -replace "github.com/sandertv/gophertunnel=$fixed" && go mod tidy >/dev/null 2>&1)
fi
rev=$(git -C "$(e2e_gt_history)" rev-parse "$tag^{commit}")
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
