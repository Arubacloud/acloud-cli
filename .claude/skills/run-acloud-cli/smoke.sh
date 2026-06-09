#!/bin/bash
# smoke.sh — build acloud and run representative read-only commands.
# Run from the repo root: bash .claude/skills/run-acloud-cli/smoke.sh
# All commands are read-only (list/get/help/version) — safe to run anytime.
set -euo pipefail

export PATH="$HOME/go-dist/go/bin:$PATH"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "$REPO_ROOT"

PASS=0; FAIL=0
ok()   { echo "  ✓ $*"; PASS=$((PASS+1)); }
fail() { echo "  ✗ $*"; FAIL=$((FAIL+1)); }

run() {
    local label="$1"; shift
    local out exit_code=0
    out=$(./acloud "$@" 2>&1) || exit_code=$?
    if [ "$exit_code" -eq 0 ]; then ok "$label"; else fail "$label (exit $exit_code): $out"; fi
    echo "$out"
}

echo "=== Build ==="
make build
echo ""

echo "=== Smoke: core ==="
run "version"           --version
run "help"              --help

echo ""
echo "=== Smoke: config ==="
run "config show"       config show

echo ""
echo "=== Smoke: context ==="
run "context list"      context list

echo ""
echo "=== Smoke: management (table)" management project list
./acloud management project list | head -5
ok "management project list (table)"

echo ""
echo "=== Smoke: output formats ==="
out_json=$(./acloud management project list --output json 2>&1)
if echo "$out_json" | python3 -c "import sys,json; json.load(sys.stdin)" 2>/dev/null; then
    ok "management project list --output json (valid JSON)"
else
    fail "management project list --output json (not valid JSON)"
fi

out_yaml=$(./acloud management project list --output yaml 2>&1)
if echo "$out_yaml" | grep -qE '^[a-zA-Z].*:|^- '; then
    ok "management project list --output yaml (looks like YAML)"
else
    fail "management project list --output yaml (unexpected format)"
fi

echo ""
echo "=== Smoke: per-resource list (read-only, may return empty) ==="
for sub_cmd in \
    "storage blockstorage list" \
    "compute cloudserver list" \
    "container kaas list" \
    "container containerregistry list" \
    "database dbaas list" \
    "security kms list"
do
    exit_code=0
    ./acloud $sub_cmd 2>&1 || exit_code=$?
    if [ "$exit_code" -eq 0 ]; then ok "$sub_cmd"; else fail "$sub_cmd (exit $exit_code)"; fi
done

# network vpc list returns HTTP 404 (not an empty list) when the active project
# has no VPCs — API quirk. Treat exit 0 or a 404 message as both acceptable.
vpc_out=$(./acloud network vpc list 2>&1); vpc_exit=$?
if [ "$vpc_exit" -eq 0 ] || echo "$vpc_out" | grep -q "404\|Not Found\|not found"; then
    ok "network vpc list (empty/404 accepted)"
else
    fail "network vpc list: unexpected output (exit $vpc_exit): $vpc_out"
fi

echo ""
echo "=== Unit tests (no live credentials) ==="
ACLOUD_TEST_SKIP_CLIENT=true make test-skip-client 2>&1 | tail -3

echo ""
echo "============================================"
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
