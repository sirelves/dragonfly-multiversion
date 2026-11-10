#!/usr/bin/env bash
# verify.sh <protocol>:<gophertunnel tag>...: encodes every serverbound and
# clientbound packet for each protocol on the server side and decodes it with
# the tag that shipped the protocol, and the other way around. Runs once with
# zero values and once with every field filled.
set -uo pipefail
here=$(cd "$(dirname "$0")" && pwd)
. "$here/lib.sh"
args=()
for a in "$@"; do
  if [[ $a == *:* ]]; then args+=("$a"); else args+=("$a:$(e2e_tag "$a")"); fi
done
set -- "${args[@]}"
dump=$here/build/verify-dump
rm -rf "$dump"; mkdir -p "$dump/zero" "$dump/fill"
rc=0
for pt in "$@"; do e2e_build_verifier "${pt%%:*}" "${pt##*:}" || { echo "build failed for $pt"; rc=1; }; done
{
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
    recv=$(awk -v p="$p" '$0 == "=== RUN   TestBlindVerifyRecv/" p {f=1; next} f && /^(=== RUN|--- |    --- )/ {f=0} f && /RECV/ && !/NOT IN OUR CLIENT POOL/' "$dump/$mode/test.log")
    if [ -n "$recv" ]; then echo "$recv"; rc=1; else echo "RECV clean"; fi
  done
done
} > "$dump/report.log" 2>&1
grep -a -E '^(===|SEND)|RECV clean' "$dump/report.log"
# Report only what the 1.20.0 baseline does not already explain.
for pt in "$@"; do
  p=${pt%%:*}
  for mode in zero fill; do
    for dir in server client; do
      head="=== $p $mode $( [ $dir = server ] && echo 'server→client' || echo 'client→server')"
      awk -v h="$head" '$0==h{f=1;next} /^=== /{f=0} f' "$dump/report.log" | grep -a -E '^BAD|RECV ' \
        | sed -E 's/.*(BAD|RECV) +([A-Za-z0-9]+).*/\2/' | sort -u | while read -r name; do
          awk -v k="$mode $dir $name" -v p="$p" 'index($0, k) == 1 { rest = substr($0, length(k) + 1); if (rest == "") f = 1; else { n = split(substr(rest, 2), l, ","); for (i = 1; i <= n; i++) if (l[i] == p) f = 1 } } END { exit !f }' "$here/verify-baseline.txt" || echo "NEW $p $mode $dir $name"
        done
    done
  done
done > "$dump/new.txt"
if [ -s "$dump/new.txt" ]; then cat "$dump/new.txt"; exit 1; fi
echo "no failures beyond the 1.20.0 baseline"
