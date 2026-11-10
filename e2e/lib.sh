# Shared helpers for the e2e scripts.
e2e_tag()    { awk -v p="$1" '$1==p{print $2}' "$(dirname "${BASH_SOURCE[0]}")/versions.txt"; }
e2e_raknet() { awk -v p="$1" '$1==p{print $4}' "$(dirname "${BASH_SOURCE[0]}")/versions.txt"; }
e2e_protos() { awk '!/^#/{print $1}' "$(dirname "${BASH_SOURCE[0]}")/versions.txt"; }

# e2e_build_verifier <protocol> [tag]: builds bin/verify-<protocol>.
e2e_build_verifier() {
  local proto=$1 tag=${2:-$(e2e_tag "$1")} here M
  here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd); M=$(go env GOMODCACHE)
  local dir=$here/build/verify-$proto
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
    if grep -q 'shieldID int32' "$src/../reader.go"; then cp "$here/verify/rw_shield.go.txt" "$dir/rw.go"; else cp "$here/verify/rw_plain.go.txt" "$dir/rw.go"; fi
    if grep -q 'type Pool map\[uint32\]Packet' "$src/pool.go"; then cp "$here/verify/pool_inst.go.txt" "$dir/pool.go"; else cp "$here/verify/pool_func.go.txt" "$dir/pool.go"; fi
  else
    cp "$here/verify/era_buf.go.txt" "$dir/era.go"
  fi
  if [ ! -f "$dir/go.mod" ]; then
    (cd "$dir" && go mod init verify >/dev/null 2>&1 && go get "github.com/sandertv/gophertunnel@$tag" >/dev/null 2>&1 && go mod tidy >/dev/null 2>&1)
  fi
  (cd "$dir" && go build -o "$here/bin/verify-$proto" .) || return 1
}
