#!/usr/bin/env bash
# gtdiff.sh <old tag> <new tag>: wire-relevant gophertunnel changes between two
# tags (protocol package, comments dropped).
set -euo pipefail
. "$(dirname "$0")/lib.sh"; cd "$(e2e_gt_history)"
git diff --stat "$1" "$2" -- minecraft/protocol | tail -n +1
git diff -U1 "$1" "$2" -- minecraft/protocol ':!minecraft/protocol/info.go' | grep -E '^(\+\+\+|[-+])' | grep -vE '^[-+]\s*//|^[-+]\s*$'
