#!/usr/bin/env bash
# check.sh: the validation gate before a commit. Builds, vets the touched
# packages and runs the multiversion and server tests; exits non-zero on any
# failure.
set -uo pipefail
cd "$(dirname "$0")"
go build ./... || exit 1
go vet ./multiversion/legacy ./server || exit 1
out=$(go test ./multiversion/... ./server/ 2>&1)
echo "$out" | grep -E '^(ok|FAIL|---)' | grep -v 'no test files'
if echo "$out" | grep -qE '^(FAIL|--- FAIL|panic:)'; then exit 1; fi
