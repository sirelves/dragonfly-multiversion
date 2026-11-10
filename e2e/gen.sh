#!/usr/bin/env bash
# gen.sh <protocol> <gophertunnel tag>: builds bin/client-<protocol> against that tag.
set -euo pipefail
cd "$(dirname "$0")"
. ./lib.sh
proto=$1 tag=${2:-$(e2e_tag "$1")}
legacy=0
dir=build/$proto
mkdir -p "$dir" bin
(cd /tmp && go mod download "github.com/sandertv/gophertunnel@$tag" >/dev/null 2>&1) || true
cp client/main.go "$dir/main.go"
rm -f "$dir"/errlog*.go
if grep -q "ErrorLog \*slog.Logger" "$(go env GOMODCACHE)/github.com/sandertv/gophertunnel@$tag/minecraft/dial.go" 2>/dev/null; then
  cp client/errlog_slog.go.txt "$dir/errlog.go"
else
  cp client/errlog_log.go.txt "$dir/errlog.go"
  legacy=1
fi
if [ ! -f "$dir/go.mod" ]; then
  (cd "$dir" && go mod init e2eclient >/dev/null 2>&1 && go get "github.com/sandertv/gophertunnel@$tag" >/dev/null && go mod tidy >/dev/null)
fi
if [ "$legacy" = 1 ]; then
  # Old go-raknet dialers lack RakNet security cookies; use the patched copy.
  (cd "$dir" && go mod edit -replace github.com/sandertv/go-raknet=../../raknet-client && go mod tidy >/dev/null)
fi
(cd "$dir" && go build -o "../../bin/client-$proto" .)
echo "built bin/client-$proto ($tag)"
