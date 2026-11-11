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

# e2e_gt_history: a full gophertunnel clone (for tag commits and diffs),
# cloned into build/ on first use unless GT_HISTORY points at one.
e2e_gt_history() {
  local dir=${GT_HISTORY:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/build/gophertunnel}
  [ -d "$dir/.git" ] || git clone -q https://github.com/Sandertv/gophertunnel.git "$dir"
  echo "$dir"
}

# e2e_expect <protocol>: packets a client of that protocol must receive.
e2e_expect() {
  if [ "$1" -lt 534 ]; then echo AdventureSettings; else echo UpdateAbilities; fi
}

# e2e_cloudburst_cp: the classpath of CloudburstMC/Protocol's bedrock-codec,
# built from CLOUDBURST (a clone; cloned into build/ on first use) with Java 17.
e2e_cloudburst_cp() {
  local here dir cache
  here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
  dir=${CLOUDBURST:-$here/build/cloudburst-protocol}
  cache=$here/build/cloudburst.cp
  if [ ! -s "$cache" ]; then
    [ -d "$dir/.git" ] || git clone -q https://github.com/CloudburstMC/Protocol.git "$dir"
    (cd "$dir" && JAVA_HOME=${JAVA17_HOME:-$(/usr/libexec/java_home -v 17)} ./gradlew -q --init-script "$here/cloudburst/cp.init.gradle" :bedrock-codec:printCp) |
      sed -n 's/^CP=//p' >"$cache"
  fi
  cat "$cache"
}
