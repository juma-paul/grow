#!/usr/bin/env bash
# Eval harness: named checks against the real test suite.
set -uo pipefail

pass=0; fail=0
run () {
  local name="$1"; shift
  if "$@" >/tmp/eval.log 2>&1; then
    echo "PASS  $name"; pass=$((pass+1))
  else
    echo "FAIL  $name"; fail=$((fail+1)); tail -n 5 /tmp/eval.log | sed 's/^/      /'
  fi
}

run "builds"            go build ./...
run "vet-clean"         go vet ./...
run "simulator-tests"   go test ./internal/simulator/...
run "race-safe"         go test -race ./internal/simulator/...
run "cpython-parity"    go test ./internal/simulator -run TestReferenceSnapshot
run "frontend-types"    bash -c "cd frontend && npx tsc --noEmit"

echo "----"
echo "$pass passed, $fail failed"
[ "$fail" -eq 0 ]
