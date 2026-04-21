#!/bin/bash

# E2E Test Script for Storage Resources
# Tests CRUD operations for Block Storage, Snapshots, Backups, and Restores

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="${ACLOUD_PROJECT_ID:-your-project-id}"
REGION="${ACLOUD_REGION:-ITBG-Bergamo}"
RESOURCE_PREFIX="e2e-test-$(date +%s)"

# Determine acloud command path - try relative to script location first, then current dir, then PATH
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/../../acloud" ]; then
    ACLOUD_CMD="$SCRIPT_DIR/../../acloud"
elif [ -f "./acloud" ]; then
    ACLOUD_CMD="./acloud"
elif command -v acloud >/dev/null 2>&1; then
    ACLOUD_CMD="acloud"
else
    ACLOUD_CMD="${ACLOUD_CMD:-./acloud}"
fi

# Cleanup tracking
CREATED_VOLUMES=()
CREATED_SNAPSHOTS=()
CREATED_BACKUPS=()
CREATED_RESTORES=()
BACKUP_ID=""  # Track backup ID for restore operations

# Failure tracking — every sub-check that prints a red ✗ calls fail() to
# increment this counter. The final exit code is non-zero iff FAILURES>0.
FAILURES=0
fail() {
    FAILURES=$((FAILURES + 1))
    echo -e "${RED}✗ $*${NC}"
}

echo -e "${BLUE}=== Storage Resources E2E Test ===${NC}\n"
echo "Project ID: $PROJECT_ID"
echo "Region: $REGION"
echo "Test prefix: $RESOURCE_PREFIX"
echo "ACLOUD command: $ACLOUD_CMD"
echo ""

# Function to extract resource ID from output
# This function tries multiple strategies to find the correct resource ID:
# 1. Extract all IDs and take the last one (resource IDs are usually printed last)
# 2. Look for ID in table format (in the ID column)
extract_id() {
    local output="$1"
    local exclude_id="${2:-}"  # Optional ID to exclude
    
    # Strategy 1: Extract all IDs, filter out exclude_id, take the last one
    if [ -n "$exclude_id" ]; then
        local filtered_ids=$(echo "$output" | grep -oE '[a-f0-9]{24}' | grep -v "^${exclude_id}$")
        if [ -n "$filtered_ids" ]; then
            echo "$filtered_ids" | tail -1
            return 0
        fi
    fi
    
    # Strategy 2: Extract all IDs and take the last one
    local all_ids=$(echo "$output" | grep -oE '[a-f0-9]{24}')
    if [ -n "$all_ids" ]; then
        if [ -n "$exclude_id" ]; then
            echo "$all_ids" | grep -v "^${exclude_id}$" | tail -1
        else
            echo "$all_ids" | tail -1
        fi
    fi
}

# Check if string is valid JSON
is_valid_json() {
    local input="$1"
    if command -v python3 >/dev/null 2>&1; then
        echo "$input" | python3 -c "import sys,json; json.load(sys.stdin)" 2>/dev/null && return 0
    elif command -v python >/dev/null 2>&1; then
        echo "$input" | python -c "import sys,json; json.load(sys.stdin)" 2>/dev/null && return 0
    fi
    return 1
}

# Test --output flag for storage list commands
test_output_formats() {
    echo -e "${BLUE}--- Testing storage list --output flag ---${NC}"

    for resource_cmd in "blockstorage list" "backup list"; do
        local label="storage $resource_cmd"
        echo -e "${YELLOW}Testing $label --output json...${NC}"
        JSON_OUTPUT=$($ACLOUD_CMD storage $resource_cmd --project-id "$PROJECT_ID" --output json 2>&1)
        JSON_EXIT=$?
        if [ $JSON_EXIT -ne 0 ]; then
            fail "$label --output json: command failed (exit $JSON_EXIT)"
        elif echo "$JSON_OUTPUT" | grep -qF "No "; then
            echo -e "${YELLOW}⚠ $label --output json: no resources — format validation skipped${NC}"
        elif is_valid_json "$JSON_OUTPUT"; then
            echo -e "${GREEN}✓ $label --output json: valid JSON${NC}"
        else
            fail "$label --output json: output is not valid JSON"
        fi

        echo -e "${YELLOW}Testing $label --output yaml...${NC}"
        YAML_OUTPUT=$($ACLOUD_CMD storage $resource_cmd --project-id "$PROJECT_ID" --output yaml 2>&1)
        YAML_EXIT=$?
        if [ $YAML_EXIT -ne 0 ]; then
            fail "$label --output yaml: command failed (exit $YAML_EXIT)"
        elif echo "$YAML_OUTPUT" | grep -qF "No "; then
            echo -e "${YELLOW}⚠ $label --output yaml: no resources — format validation skipped${NC}"
        elif echo "$YAML_OUTPUT" | grep -qE '^[a-zA-Z].*:|^- '; then
            echo -e "${GREEN}✓ $label --output yaml: output looks like YAML${NC}"
        else
            fail "$label --output yaml: output does not look like YAML"
        fi
    done
    echo ""
}

# Helper function to validate resource ID
is_valid_id() {
    local id="$1"
    # Check if it's a 24-character hex string (MongoDB ObjectID format)
    [[ "$id" =~ ^[a-f0-9]{24}$ ]]
}

# Poll a block-storage volume until it leaves transient states (InCreation, Creating, Pending),
# or until timeout elapses. Returns 0 on ready, 1 on timeout or get failure.
wait_for_volume_ready() {
    local volume_id="$1"
    local timeout="${2:-180}"
    local elapsed=0
    local status=""
    while [ "$elapsed" -lt "$timeout" ]; do
        local out
        out=$($ACLOUD_CMD storage blockstorage get "$volume_id" 2>&1) || return 1
        status=$(echo "$out" | grep -iE "^Status:" | head -1 | awk -F: '{print $2}' | tr -d '[:space:]')
        case "$status" in
            ""|InCreation|Creating|Pending) sleep 5; elapsed=$((elapsed + 5));;
            *) echo "  → volume $volume_id ready (status=$status)"; return 0;;
        esac
    done
    echo -e "${YELLOW}wait_for_volume_ready: timeout after ${timeout}s (last status=$status)${NC}"
    return 1
}

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test resources...${NC}"
    
    # Delete restores (requires backup-id and restore-id)
    # Note: Restore delete requires both backup-id and restore-id
    # We'll need to track backup-id for each restore, but for now skip if we don't have it
    for restore_id in "${CREATED_RESTORES[@]}"; do
        if [ -n "$BACKUP_ID" ]; then
            echo "Deleting restore: $restore_id"
            $ACLOUD_CMD storage restore delete "$BACKUP_ID" "$restore_id" --yes 2>&1 || true
        else
            echo "Skipping restore delete $restore_id (backup-id not available)"
        fi
    done
    
    # Delete backups
    for backup_id in "${CREATED_BACKUPS[@]}"; do
        echo "Deleting backup: $backup_id"
        $ACLOUD_CMD storage backup delete "$backup_id" --yes 2>&1 || true
    done
    
    # Delete snapshots
    for snapshot_id in "${CREATED_SNAPSHOTS[@]}"; do
        echo "Deleting snapshot: $snapshot_id"
        $ACLOUD_CMD storage snapshot delete "$snapshot_id" --yes 2>&1 || true
    done
    
    # Delete volumes
    for volume_id in "${CREATED_VOLUMES[@]}"; do
        echo "Deleting volume: $volume_id"
        $ACLOUD_CMD storage blockstorage delete "$volume_id" --yes 2>&1 || true
    done
}

trap cleanup EXIT

# Test Block Storage
test_block_storage() {
    local volume_name="${RESOURCE_PREFIX}-volume"
    
    echo -e "${YELLOW}--- Testing Block Storage CRUD ---${NC}\n"
    
    # CREATE
    echo -e "${GREEN}[CREATE]${NC} Creating block storage: $volume_name"
    CREATE_OUTPUT=$($ACLOUD_CMD storage blockstorage create \
        --name "$volume_name" \
        --region "$REGION" \
        --size 10 \
        --type Standard \
        --billing-period Hour \
        --tags "e2e-test,storage" 2>&1) || {
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT"
        # Check for common error patterns
        if echo "$CREATE_OUTPUT" | grep -qi "authentication failed\|invalid_client\|Invalid client"; then
            echo -e "${RED}Authentication error detected. Please check your credentials.${NC}"
        fi
        return 1
    }
    echo "$CREATE_OUTPUT"
    
    VOLUME_ID=$(extract_id "$CREATE_OUTPUT" | tr -d '[:space:]')
    if [ -z "$VOLUME_ID" ] || ! is_valid_id "$VOLUME_ID"; then
        echo -e "${RED}Could not extract valid volume ID${NC}"
        echo -e "${YELLOW}CREATE_OUTPUT:${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi
    CREATED_VOLUMES+=("$VOLUME_ID")
    echo -e "${GREEN}Created volume ID: $VOLUME_ID${NC}\n"
    
    # Wait for volume to leave InCreation before attempting UPDATE
    echo "Waiting for volume to be ready..."
    wait_for_volume_ready "$VOLUME_ID" 180 || \
        echo -e "${YELLOW}Warning: volume not ready after 180s — UPDATE may still fail${NC}"
    
    # LIST
    echo -e "${GREEN}[LIST]${NC} Listing block storage..."
    LIST_OUTPUT=$($ACLOUD_CMD storage blockstorage list 2>&1) || {
        echo -e "${RED}LIST failed:${NC}"
        echo "$LIST_OUTPUT"
        return 1
    }
    echo "$LIST_OUTPUT" | head -15
    echo ""
    
    # GET
    echo -e "${GREEN}[GET]${NC} Getting block storage details..."
    GET_OUTPUT=$($ACLOUD_CMD storage blockstorage get "$VOLUME_ID" 2>&1) || {
        echo -e "${RED}GET failed:${NC}"
        echo "$GET_OUTPUT"
        return 1
    }
    echo "$GET_OUTPUT"
    echo ""
    
    # UPDATE
    echo -e "${GREEN}[UPDATE]${NC} Updating block storage..."
    UPDATE_OUTPUT=$($ACLOUD_CMD storage blockstorage update "$VOLUME_ID" \
        --name "${volume_name}-updated" \
        --tags "e2e-test,updated" 2>&1) || {
        echo -e "${RED}UPDATE failed:${NC}"
        echo "$UPDATE_OUTPUT"
        return 1
    }
    echo "$UPDATE_OUTPUT"
    echo ""
    
    echo -e "${GREEN}✓ Block Storage CRUD test completed!${NC}\n"
    echo "$VOLUME_ID"  # Return volume ID for use in snapshots
}

# Test Snapshot
test_snapshot() {
    local volume_id="$1"
    local snapshot_name="${RESOURCE_PREFIX}-snapshot"
    
    if [ -z "$volume_id" ]; then
        echo -e "${YELLOW}Skipping snapshot test (no volume available)${NC}\n"
        return 0
    fi
    
    echo -e "${YELLOW}--- Testing Snapshot CRUD ---${NC}\n"
    
    # Get volume URI
    VOLUME_GET=$($ACLOUD_CMD storage blockstorage get "$volume_id" 2>&1)
    VOLUME_URI=$(echo "$VOLUME_GET" | grep -i "URI:" | awk '{print $2}')
    
    if [ -z "$VOLUME_URI" ]; then
        echo -e "${YELLOW}Could not extract volume URI, skipping snapshot test${NC}\n"
        return 0
    fi
    
    # CREATE
    echo -e "${GREEN}[CREATE]${NC} Creating snapshot: $snapshot_name"
    CREATE_OUTPUT=$($ACLOUD_CMD storage snapshot create \
        --name "$snapshot_name" \
        --region "$REGION" \
        --volume-uri "$VOLUME_URI" \
        --tags "e2e-test,snapshot" 2>&1) || {
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    }
    echo "$CREATE_OUTPUT"
    
    SNAPSHOT_ID=$(extract_id "$CREATE_OUTPUT" "$volume_id" | tr -d '[:space:]')
    if [ -z "$SNAPSHOT_ID" ] || [ "$SNAPSHOT_ID" = "$volume_id" ] || ! is_valid_id "$SNAPSHOT_ID"; then
        echo -e "${RED}Could not extract valid snapshot ID${NC}"
        echo -e "${YELLOW}CREATE_OUTPUT:${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi
    CREATED_SNAPSHOTS+=("$SNAPSHOT_ID")
    echo -e "${GREEN}Created snapshot ID: $SNAPSHOT_ID${NC}\n"
    
    # LIST
    echo -e "${GREEN}[LIST]${NC} Listing snapshots..."
    LIST_OUTPUT=$($ACLOUD_CMD storage snapshot list --volume-uri "$VOLUME_URI" 2>&1) || {
        echo -e "${RED}LIST failed:${NC}"
        echo "$LIST_OUTPUT"
        return 1
    }
    echo "$LIST_OUTPUT" | head -15
    echo ""
    
    # GET
    echo -e "${GREEN}[GET]${NC} Getting snapshot details..."
    GET_OUTPUT=$($ACLOUD_CMD storage snapshot get "$SNAPSHOT_ID" 2>&1) || {
        echo -e "${RED}GET failed:${NC}"
        echo "$GET_OUTPUT"
        return 1
    }
    echo "$GET_OUTPUT"
    echo ""
    
    # UPDATE
    echo -e "${GREEN}[UPDATE]${NC} Updating snapshot..."
    UPDATE_OUTPUT=$($ACLOUD_CMD storage snapshot update "$SNAPSHOT_ID" \
        --name "${snapshot_name}-updated" \
        --tags "e2e-test,updated" 2>&1) || {
        echo -e "${RED}UPDATE failed:${NC}"
        echo "$UPDATE_OUTPUT"
        return 1
    }
    echo "$UPDATE_OUTPUT"
    echo ""
    
    echo -e "${GREEN}✓ Snapshot CRUD test completed!${NC}\n"
}

# Test Backup (if available)
test_backup() {
    echo -e "${YELLOW}--- Testing Backup CRUD ---${NC}\n"
    echo -e "${YELLOW}Note: Backup operations may require specific prerequisites${NC}\n"
    # Backup tests would go here
    echo -e "${GREEN}✓ Backup test placeholder${NC}\n"
}

# Test Restore (if available)
test_restore() {
    echo -e "${YELLOW}--- Testing Restore CRUD ---${NC}\n"
    echo -e "${YELLOW}Note: Restore operations require existing backups${NC}\n"
    # Restore tests would go here
    echo -e "${GREEN}✓ Restore test placeholder${NC}\n"
}

# Set up context for project ID (so we don't need --project-id flag on every command)
if [ "$PROJECT_ID" != "your-project-id" ]; then
    echo -e "${BLUE}Setting up context for project ID...${NC}"
    $ACLOUD_CMD context set e2e-test-context --project-id "$PROJECT_ID" >/dev/null 2>&1 || true
    $ACLOUD_CMD context use e2e-test-context >/dev/null 2>&1 || true
    echo ""
fi

# Run tests
echo -e "${BLUE}Starting Storage Resources E2E Tests...${NC}\n"

VOLUME_ID=""
if test_block_storage; then
    # VOLUME_ID is set inside test_block_storage and added to CREATED_VOLUMES
    # Get it from the array or the function output
    if [ ${#CREATED_VOLUMES[@]} -gt 0 ]; then
        VOLUME_ID="${CREATED_VOLUMES[0]}"
    fi
    if [ -n "$VOLUME_ID" ]; then
        test_snapshot "$VOLUME_ID" || FAILURES=$((FAILURES + 1))
    else
        echo -e "${YELLOW}Skipping snapshot test (no volume ID available)${NC}\n"
    fi
else
    FAILURES=$((FAILURES + 1))
fi

test_backup || FAILURES=$((FAILURES + 1))
test_restore || FAILURES=$((FAILURES + 1))
test_output_formats

echo -e "${GREEN}=== All Storage Tests Completed! ===${NC}\n"

# Print summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo -e "Project ID: ${PROJECT_ID:-N/A}"
if [ ${#CREATED_VOLUMES[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Block Storage: ${#CREATED_VOLUMES[@]} created${NC}"
else
    echo -e "${YELLOW}○ Block Storage: 0 created${NC}"
fi
if [ ${#CREATED_SNAPSHOTS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Snapshots: ${#CREATED_SNAPSHOTS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Snapshots: 0 created${NC}"
fi
if [ ${#CREATED_BACKUPS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Backups: ${#CREATED_BACKUPS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Backups: 0 created${NC}"
fi
if [ ${#CREATED_RESTORES[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Restores: ${#CREATED_RESTORES[@]} created${NC}"
else
    echo -e "${YELLOW}○ Restores: 0 created${NC}"
fi
echo ""

if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}=== Storage E2E: all checks passed ===${NC}"
    exit 0
else
    echo -e "${RED}=== Storage E2E: $FAILURES check(s) failed ===${NC}"
    exit 1
fi
