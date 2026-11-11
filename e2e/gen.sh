#!/usr/bin/env bash
# gen.sh <protocol> <gophertunnel tag>: builds bin/client-<protocol> against that tag.
set -euo pipefail
cd "$(dirname "$0")"
. ./lib.sh
proto=$1 tag=${2:-$(e2e_tag "$1")}
legacy=0
dir=build/$proto
[ "$tag" = "$(e2e_tag "$1")" ] || dir=build/$proto-$tag # its go.mod pins the tag
mkdir -p "$dir" bin
(cd /tmp && go mod download "github.com/sandertv/gophertunnel@$tag" >/dev/null 2>&1) || true
cp client/main.go "$dir/main.go"
cp client/chunk.go "$dir/chunk.go"
if grep -q 'func (d Dialer) DialTimeout' "$(go env GOMODCACHE)/github.com/sandertv/gophertunnel@$tag/minecraft/dial.go" 2>/dev/null; then
  cp client/dial_timeout.go.txt "$dir/dial.go"
else
  cp client/dial_plain.go.txt "$dir/dial.go"
fi
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
# gophertunnel before v1.13 links crypto/cipher.xorBytes, which Go removed;
# build against a copy that carries its own.
gcm=$(go env GOMODCACHE)/github.com/sandertv/gophertunnel@$tag/minecraft/protocol/packet/gcm.go
if grep -q 'go:linkname xorBytes' "$gcm" 2>/dev/null; then
  fixed=$PWD/build/gt-$tag-xor
  if [ ! -d "$fixed" ]; then
    cp -r "$(go env GOMODCACHE)/github.com/sandertv/gophertunnel@$tag" "$fixed" && chmod -R u+w "$fixed"
    python3 - "$fixed/minecraft/protocol/packet/gcm.go" <<'PY'
import re, sys
p = sys.argv[1]; s = open(p).read()
s = re.sub(r"//go:linkname xorBytes crypto/cipher.xorBytes\n(?://[^\n]*\n)*func xorBytes\(dst, a, b \[\]byte\) int\n",
           "func xorBytes(dst, a, b []byte) int {\n\tn := len(a)\n\tif len(b) < n {\n\t\tn = len(b)\n\t}\n\tfor i := 0; i < n; i++ {\n\t\tdst[i] = a[i] ^ b[i]\n\t}\n\treturn n\n}\n", s)
s = s.replace('_ "unsafe"\n', "").replace('\t_ "unsafe"\n', "")
open(p, "w").write(s)
PY
  fi
  (cd "$dir" && go mod edit -replace "github.com/sandertv/gophertunnel=$fixed" && go mod tidy >/dev/null 2>&1)
fi
if [ "$legacy" = 1 ]; then
  # Old go-raknet dialers lack RakNet security cookies; use the patched copy.
  rk=../../raknet-client
  case $(cd "$dir" && go list -m -f '{{.Version}}' github.com/sandertv/go-raknet 2>/dev/null) in
    v1.7.[0-9]) rk=../../raknet-client-v17 ;; # the listener API of the 1.16.0 tags
  esac
  (cd "$dir" && go mod edit -dropreplace github.com/sandertv/go-raknet && go mod tidy >/dev/null 2>&1; \
    case $(go list -m -f '{{.Version}}' github.com/sandertv/go-raknet) in v1.7.[0-9]) rk=../../raknet-client-v17 ;; esac; \
    go mod edit -replace github.com/sandertv/go-raknet=$rk && go mod tidy >/dev/null)
fi
(cd "$dir" && go build -o "../../bin/client-$proto" .)
echo "built bin/client-$proto ($tag)"
