#!/usr/bin/env bash
# verify.sh <protocol>:<gophertunnel tag>...: encodes every serverbound and
# clientbound packet for each protocol on the server side and decodes it with
# the tag that shipped the protocol, and the other way around. Runs once with
# zero values and once with every field filled.
set -uo pipefail
here=$(cd "$(dirname "$0")" && pwd)
dump=$here/build/verify-dump
rm -rf "$dump"; mkdir -p "$dump/zero" "$dump/fill"
M=$(go env GOMODCACHE)
build() {
  local proto=$1 tag=$2 dir=$here/build/verify-$1
  mkdir -p "$dir"
  (cd /tmp && go mod download "github.com/sandertv/gophertunnel@$tag" >/dev/null 2>&1) || true
  local src=$M/github.com/sandertv/gophertunnel@$tag/minecraft/protocol/packet
  rm -f "$dir"/*.go
  cp "$here/verify/main.go" "$dir/"
  if grep -q 'Marshal(io protocol.IO)' "$src/packet.go"; then
    cp "$here/verify/era_io.go.txt" "$dir/era.go"
    if grep -q 'func NewClientPool' "$src/pool.go"; then cp "$here/verify/pools_split.go.txt" "$dir/pools.go"; else cp "$here/verify/pools_single.go.txt" "$dir/pools.go"; fi
    if grep -q 'enableLimits bool' "$src/../reader.go"; then cp "$here/verify/reader_limits.go.txt" "$dir/reader.go"; else cp "$here/verify/reader_plain.go.txt" "$dir/reader.go"; fi
  elif grep -q 'Unmarshal(r \*protocol.Reader)' "$src/packet.go"; then
    cp "$here/verify/era_rw.go.txt" "$dir/era.go"
    if grep -q 'type Pool map\[uint32\]Packet' "$src/pool.go"; then cp "$here/verify/pool_inst.go.txt" "$dir/pool.go"; else cp "$here/verify/pool_func.go.txt" "$dir/pool.go"; fi
  else
    cp "$here/verify/era_buf.go.txt" "$dir/era.go"
  fi
  if [ ! -f "$dir/go.mod" ]; then
    (cd "$dir" && go mod init verify >/dev/null 2>&1 && go get "github.com/sandertv/gophertunnel@$tag" >/dev/null 2>&1 && go mod tidy >/dev/null 2>&1)
  fi
  (cd "$dir" && go build -o "$here/bin/verify-$proto" .) || return 1
}
rc=0
for pt in "$@"; do build "${pt%%:*}" "${pt##*:}" || { echo "build failed for $pt"; rc=1; }; done
for mode in zero fill; do
  fillenv=""; [ $mode = fill ] && fillenv=1
  for pt in "$@"; do BLINDVERIFY_FILL=$fillenv "$here/bin/verify-${pt%%:*}" -recv "$dump/$mode/recv_${pt%%:*}.txt"; done
  (cd "$here/.." && TMDRAGONFLY_BLINDVERIFY=1 BLINDVERIFY_FILL=$fillenv DUMPDIR="$dump/$mode" \
    go test ./multiversion/conformance/ -run 'TestBlindVerify(Send|Recv)$' -count=1 -v 2>&1) > "$dump/$mode/test.log"
  for pt in "$@"; do
    p=${pt%%:*}
    echo "=== $p $mode server→client"
    "$here/bin/verify-$p" -send "$dump/$mode/send_$p.txt" || rc=1
    echo "=== $p $mode client→server"
    if grep -E "TestBlindVerifyRecv/$p\b" -A100000 "$dump/$mode/test.log" | sed -n '2,$p' | sed '/^    --- \|^=== RUN/q' | grep -E 'RECV' | grep -v 'NOT IN OUR CLIENT POOL'; then rc=1; else echo "RECV clean"; fi
  done
done
exit $rc
