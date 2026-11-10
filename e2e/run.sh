#!/usr/bin/env bash
# run.sh <protocol>...: boots the server binary in a scratch dir with auth off,
# joins with bin/client-<protocol> for each protocol and prints the result.
set -uo pipefail
here=$(cd "$(dirname "$0")" && pwd)
. "$here/lib.sh"
work=${E2E_WORK:-$here/build/srv}
port=${E2E_PORT:-19232}
mkdir -p "$work"
(cd "$here/.." && go build -o "$work/dfmv" .) || exit 1
cd "$work"
start() {
  rm -f server.log
  ./dfmv </dev/null >server.log 2>&1 &
  pid=$!
  for _ in $(seq 1 80); do grep -q 'Server is up' server.log 2>/dev/null && return 0; sleep 0.25; done
  echo "server did not start:"; tail -20 server.log; kill $pid 2>/dev/null; exit 1
}
stop() { kill $pid 2>/dev/null; wait $pid 2>/dev/null; }
if [ ! -f config.toml ]; then start; stop; fi
sed -i '' -e "s/Address = \":19132\"/Address = \":$port\"/" config.toml
# E2E_AUTH=1 keeps Xbox Live on: the offline e2e clients then only get in
# below LegacyOfflineBelow.
if [ "${E2E_AUTH:-0}" = 1 ]; then
  sed -i '' -e 's/AuthEnabled = false/AuthEnabled = true/' config.toml
else
  sed -i '' -e 's/AuthEnabled = true/AuthEnabled = false/' config.toml
fi
for p in "$@"; do "$here/gen.sh" "$p" >/dev/null || { echo "client build failed for $p"; exit 1; }; done
start
rc=0
for p in "$@"; do
  echo "=== client $p"
  before=$(wc -l <server.log)
  E2E_RAKNET_PROTOCOL=$(e2e_raknet "$p") "$here/bin/client-$p" -expect "$(e2e_expect "$p")" -addr "127.0.0.1:$port" -name "e2e$p" -for "${E2E_FOR:-6s}" </dev/null || rc=1
  sleep 0.5
  if tail -n +"$((before + 1))" server.log | sed 's/\x1b\[[0-9;]*m//g' | grep -iE 'error|warn|panic|unhandled|unknown packet' ; then
    echo "SERVER-ERRORS for $p"; rc=1
  fi
done
stop
exit $rc
