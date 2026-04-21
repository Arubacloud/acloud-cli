#!/bin/bash
# Shared helpers for e2e/*/test.sh scripts.
# Sourced from each suite via: source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# Guard against double-sourcing.
[ -n "${_ACLOUD_E2E_COMMON_LOADED:-}" ] && return 0
_ACLOUD_E2E_COMMON_LOADED=1

# --- Colors --------------------------------------------------------------
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# --- Defaults ------------------------------------------------------------
# Only assign if the caller has not already set them (network overrides
# PROJECT_ID, adds VPC_ID, etc. before sourcing).
: "${PROJECT_ID:=${ACLOUD_PROJECT_ID:-your-project-id}}"
: "${REGION:=${ACLOUD_REGION:-ITBG-Bergamo}}"
: "${RESOURCE_PREFIX:=e2e-test-$(date +%s)}"

# --- acloud binary resolution -------------------------------------------
# Honour a pre-set $ACLOUD_CMD; otherwise look in the repo root (relative to
# this file), the current directory, then $PATH. Hard-fail if none found.
if [ -z "${ACLOUD_CMD:-}" ]; then
    _ACLOUD_COMMON_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    if [ -f "$_ACLOUD_COMMON_DIR/../acloud" ]; then
        ACLOUD_CMD="$_ACLOUD_COMMON_DIR/../acloud"
    elif [ -f "./acloud" ]; then
        ACLOUD_CMD="./acloud"
    elif command -v acloud >/dev/null 2>&1; then
        ACLOUD_CMD="acloud"
    else
        echo -e "${RED}Error: acloud binary not found. Build it first with 'make build'.${NC}" >&2
        return 1 2>/dev/null || exit 1
    fi
    unset _ACLOUD_COMMON_DIR
fi

# --- Failure tracking ----------------------------------------------------
# Scripts can call fail() on each red ✗ to feed a non-zero exit code.
FAILURES=0
fail() {
    FAILURES=$((FAILURES + 1))
    echo -e "${RED}✗ $*${NC}"
}

# --- Validators ----------------------------------------------------------
is_valid_json() {
    local input="$1"
    if command -v python3 >/dev/null 2>&1; then
        echo "$input" | python3 -c "import sys,json; json.load(sys.stdin)" 2>/dev/null && return 0
    elif command -v python >/dev/null 2>&1; then
        echo "$input" | python -c "import sys,json; json.load(sys.stdin)" 2>/dev/null && return 0
    fi
    return 1
}

is_valid_id() {
    local id="$1"
    [[ "$id" =~ ^[a-f0-9]{24}$ ]]
}

check_auth_error() {
    local output="$1"
    if echo "$output" | grep -qi "authentication failed\|invalid_client\|unauthorized"; then
        echo -e "${RED}Authentication error detected. Please check your credentials.${NC}" >&2
        return 1
    fi
    return 0
}

# --- ID extraction -------------------------------------------------------
# Extracts a 24-char hex ID from command output.
#   Strategy 1: if $2 (exclude_id) is supplied, return the last ID that is
#               not equal to it (filters out a known parent ID).
#   Strategy 2: parse a "... ID ..." header and read the value column below.
#   Strategy 3: return the last 24-char hex token in the output.
extract_id() {
    local output="$1"
    local exclude_id="${2:-}"

    if [ -n "$exclude_id" ]; then
        local filtered_ids
        filtered_ids=$(echo "$output" | grep -oE '[a-f0-9]{24}' | grep -v "^${exclude_id}$")
        if [ -n "$filtered_ids" ]; then
            echo "$filtered_ids" | tail -1
            return 0
        fi
    fi

    local table_id
    table_id=$(echo "$output" | awk '
        /^[A-Z ]+ID[ A-Z]*$/ {
            getline
            if (NF >= 2 && $2 ~ /^[a-f0-9]{24}$/) {
                print $2
            }
        }
    ' | head -1)
    if [ -n "$table_id" ] && [ "$table_id" != "$exclude_id" ]; then
        echo "$table_id"
        return 0
    fi

    local all_ids
    all_ids=$(echo "$output" | grep -oE '[a-f0-9]{24}')
    if [ -n "$all_ids" ]; then
        if [ -n "$exclude_id" ]; then
            echo "$all_ids" | grep -v "^${exclude_id}$" | tail -1
        else
            echo "$all_ids" | tail -1
        fi
    fi
}

# --- Reusable output-format validation ----------------------------------
# Given the result of running "<cmd> --output <fmt>", print ✓ / ⚠ / ✗ and
# increment FAILURES on ✗. Used by per-suite test_output_formats loops.
# Usage: validate_list_output <label> <json|yaml> <output> <exit_code>
validate_list_output() {
    local label="$1" fmt="$2" output="$3" exit_code="$4"
    if [ "$exit_code" -ne 0 ]; then
        fail "$label --output $fmt: command failed (exit $exit_code)"
        return 1
    fi
    if echo "$output" | grep -qF "No "; then
        echo -e "${YELLOW}⚠ $label --output $fmt: no resources — format validation skipped${NC}"
        return 0
    fi
    case "$fmt" in
        json)
            if is_valid_json "$output"; then
                echo -e "${GREEN}✓ $label --output $fmt: valid JSON${NC}"
                return 0
            fi
            fail "$label --output $fmt: output is not valid JSON"
            ;;
        yaml)
            if echo "$output" | grep -qE '^[a-zA-Z].*:|^- '; then
                echo -e "${GREEN}✓ $label --output $fmt: output looks like YAML${NC}"
                return 0
            fi
            fail "$label --output $fmt: output does not look like YAML"
            ;;
    esac
    return 1
}

# --- Banner --------------------------------------------------------------
# Usage: print_banner "Compute"  →  "=== Compute Resources E2E Test ==="
print_banner() {
    local title="$1"
    echo -e "${BLUE}=== ${title} Resources E2E Test ===${NC}\n"
    echo "Project ID: $PROJECT_ID"
    echo "Region: $REGION"
    echo "Test prefix: $RESOURCE_PREFIX"
    echo "ACLOUD command: $ACLOUD_CMD"
    echo ""
}

# --- Context setup -------------------------------------------------------
# Registers and activates the shared `e2e-test-context` so subsequent
# commands don't need --project-id. No-op if PROJECT_ID was left at the
# placeholder default.
setup_context() {
    if [ -n "$PROJECT_ID" ] && [ "$PROJECT_ID" != "your-project-id" ]; then
        echo -e "${BLUE}Setting up context for project ID...${NC}"
        $ACLOUD_CMD context set e2e-test-context --project-id "$PROJECT_ID" >/dev/null 2>&1 || true
        $ACLOUD_CMD context use e2e-test-context >/dev/null 2>&1 || true
        echo ""
    fi
}
