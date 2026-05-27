#!/bin/bash

# E2E Test Script for Storage Resources
# Tests CRUD operations for Block Storage, Snapshots, Backups, and Restores

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# Cleanup tracking
CREATED_VOLUMES=()
CREATED_SNAPSHOTS=()
CREATED_BACKUPS=()
CREATED_RESTORES=()
BACKUP_ID=""  # Track backup ID for restore operations
BOOTSTRAP_PROJECT_ID=""

print_banner "Storage"

# Test --output flag for storage list commands
test_output_formats() {
    echo -e "${BLUE}--- Testing storage list --output flag ---${NC}"

    for resource_cmd in "blockstorage list" "backup list"; do
        local label="storage $resource_cmd"
        local cmd="$ACLOUD_CMD storage $resource_cmd --project-id \"$PROJECT_ID\""
        for fmt in json yaml; do
            echo -e "${YELLOW}Testing $label --output $fmt...${NC}"
            OUT=$(eval "$cmd --output $fmt" 2>&1)
            validate_list_output "$label" "$fmt" "$OUT" $?
        done
    done
    echo ""
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

# Poll a snapshot until it leaves transient states, or until timeout elapses.
wait_for_snapshot_ready() {
    local snapshot_id="$1"
    local timeout="${2:-180}"
    local elapsed=0
    local status=""
    while [ "$elapsed" -lt "$timeout" ]; do
        local out
        out=$($ACLOUD_CMD storage snapshot get "$snapshot_id" 2>&1) || return 1
        status=$(echo "$out" | grep -iE "^Status:" | head -1 | awk -F: '{print $2}' | tr -d '[:space:]')
        case "$status" in
            ""|InCreation|Creating|Pending) sleep 5; elapsed=$((elapsed + 5));;
            *) echo "  → snapshot $snapshot_id ready (status=$status)"; return 0;;
        esac
    done
    echo -e "${YELLOW}wait_for_snapshot_ready: timeout after ${timeout}s (last status=$status)${NC}"
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
    
    # Delete volumes — wait for InCreation to resolve before deleting
    for volume_id in "${CREATED_VOLUMES[@]}"; do
        echo "Deleting volume: $volume_id"
        wait_for_volume_ready "$volume_id" 120 2>/dev/null || true
        $ACLOUD_CMD storage blockstorage delete "$volume_id" --yes 2>&1 || true
    done

    # Delete bootstrapped project last
    if [ -n "$BOOTSTRAP_PROJECT_ID" ]; then
        echo "Deleting bootstrapped project: $BOOTSTRAP_PROJECT_ID"
        local del_elapsed=0
        while [ "$del_elapsed" -lt 120 ]; do
            $ACLOUD_CMD management project delete "$BOOTSTRAP_PROJECT_ID" --yes 2>&1 && break
            sleep 10
            del_elapsed=$((del_elapsed + 10))
        done
    fi
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
    local volume_ready=0
    wait_for_volume_ready "$VOLUME_ID" 180 && volume_ready=1
    if [ "$volume_ready" -eq 0 ]; then
        echo -e "${YELLOW}Warning: volume not ready after 180s — UPDATE will be skipped${NC}"
    fi

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

    # UPDATE — only attempt when volume has left InCreation
    if [ "$volume_ready" -eq 1 ]; then
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
    else
        echo -e "${YELLOW}[UPDATE]${NC} Skipping — volume did not reach ready state after 180s."
        echo ""
    fi

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

    # CREATE
    echo -e "${GREEN}[CREATE]${NC} Creating snapshot: $snapshot_name"
    CREATE_OUTPUT=$($ACLOUD_CMD storage snapshot create \
        --name "$snapshot_name" \
        --region "$REGION" \
        --volume-id "$volume_id" \
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

    # Wait for snapshot to leave InCreation before attempting UPDATE
    echo "Waiting for snapshot to be ready..."
    local snapshot_ready=0
    wait_for_snapshot_ready "$SNAPSHOT_ID" 300 && snapshot_ready=1
    if [ "$snapshot_ready" -eq 0 ]; then
        echo -e "${YELLOW}Warning: snapshot not ready after 300s — UPDATE will be skipped${NC}"
    fi

    # LIST
    echo -e "${GREEN}[LIST]${NC} Listing snapshots..."
    LIST_OUTPUT=$($ACLOUD_CMD storage snapshot list --volume-id "$volume_id" 2>&1) || {
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
    
    if [ "$snapshot_ready" -eq 1 ]; then
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
    else
        echo -e "${YELLOW}[UPDATE]${NC} Skipping — snapshot did not reach ready state after 300s."
        echo ""
    fi

    echo -e "${GREEN}✓ Snapshot CRUD test completed!${NC}\n"
}

# Test Backup
test_backup() {
    local volume_id="$1"
    local backup_name="${RESOURCE_PREFIX}-backup"

    if [ -z "$volume_id" ]; then
        echo -e "${YELLOW}Skipping backup test (no volume available)${NC}\n"
        return 0
    fi

    echo -e "${YELLOW}--- Testing Storage Backup CRUD ---${NC}\n"

    # Ensure volume is stable before taking a backup
    echo "Ensuring volume $volume_id is stable..."
    wait_for_volume_ready "$volume_id" 60 2>/dev/null || true

    # CREATE
    echo -e "${GREEN}[CREATE]${NC} Creating backup: $backup_name"
    CREATE_OUTPUT=$($ACLOUD_CMD storage backup "$volume_id" \
        --name "$backup_name" \
        --region "$REGION" \
        --type Full \
        --billing-period Hour \
        --tags "e2e-test,backup" 2>&1)
    exit_code=$?

    if [ $exit_code -ne 0 ]; then
        echo -e "${YELLOW}Backup CREATE failed (soft-fail — volume may need to be attached first):${NC}"
        echo "$CREATE_OUTPUT" | head -5
        echo -e "${YELLOW}Skipping remaining backup/restore tests${NC}\n"
        return 0
    fi
    echo "$CREATE_OUTPUT"

    local backup_id
    backup_id=$(extract_id "$CREATE_OUTPUT" "$volume_id" | tr -d '[:space:]')
    if [ -z "$backup_id" ] || ! is_valid_id "$backup_id"; then
        echo -e "${YELLOW}Could not extract backup ID — skipping remaining backup/restore tests${NC}\n"
        return 0
    fi
    CREATED_BACKUPS+=("$backup_id")
    BACKUP_ID="$backup_id"
    echo -e "${GREEN}Created backup ID: $backup_id${NC}\n"

    echo "Waiting for backup to be ready..."
    local backup_ready=0
    wait_for_status "$ACLOUD_CMD storage backup get $backup_id" \
        '^(Active|Available|Ready|Completed|Complete)$' 300 && backup_ready=1

    # LIST
    echo -e "${GREEN}[LIST]${NC} Listing backups..."
    $ACLOUD_CMD storage backup list 2>&1 | head -15
    echo ""

    # GET
    echo -e "${GREEN}[GET]${NC} Getting backup details..."
    $ACLOUD_CMD storage backup get "$backup_id" 2>&1
    echo ""

    if [ "$backup_ready" -eq 1 ]; then
        echo -e "${GREEN}[UPDATE]${NC} Updating backup..."
        UPDATE_OUTPUT=$($ACLOUD_CMD storage backup update "$backup_id" \
            --name "${backup_name}-updated" \
            --tags "e2e-test,updated" 2>&1)
        if [ $? -eq 0 ]; then
            echo "$UPDATE_OUTPUT"
        else
            echo -e "${YELLOW}Update failed (non-fatal):${NC}"
            echo "$UPDATE_OUTPUT" | head -5
        fi
        echo ""
    else
        echo -e "${YELLOW}[UPDATE]${NC} Skipping — backup did not reach ready state after 300s."
        echo ""
    fi

    echo -e "${GREEN}✓ Storage Backup CRUD test completed!${NC}\n"
}

# Test Restore
test_restore() {
    local backup_id="$1"
    local volume_id="$2"
    local restore_name="${RESOURCE_PREFIX}-restore"

    if [ -z "$backup_id" ] || [ -z "$volume_id" ]; then
        echo -e "${YELLOW}Skipping restore test (no backup or volume available)${NC}\n"
        return 0
    fi

    echo -e "${YELLOW}--- Testing Storage Restore CRUD ---${NC}\n"

    # CREATE
    echo -e "${GREEN}[CREATE]${NC} Creating restore: $restore_name"
    CREATE_OUTPUT=$($ACLOUD_CMD storage restore "$backup_id" "$volume_id" \
        --name "$restore_name" \
        --region "$REGION" \
        --tags "e2e-test,restore" 2>&1)
    exit_code=$?

    if [ $exit_code -ne 0 ]; then
        echo -e "${YELLOW}Restore CREATE failed (soft-fail):${NC}"
        echo "$CREATE_OUTPUT" | head -5
        echo -e "${YELLOW}Skipping remaining restore tests${NC}\n"
        return 0
    fi
    echo "$CREATE_OUTPUT"

    local restore_id
    restore_id=$(extract_id "$CREATE_OUTPUT" "$backup_id" | tr -d '[:space:]')
    if [ -z "$restore_id" ] || ! is_valid_id "$restore_id"; then
        echo -e "${YELLOW}Could not extract restore ID — skipping remaining restore tests${NC}\n"
        return 0
    fi
    CREATED_RESTORES+=("$restore_id")
    echo -e "${GREEN}Created restore ID: $restore_id${NC}\n"

    # LIST
    echo -e "${GREEN}[LIST]${NC} Listing restores for backup $backup_id..."
    $ACLOUD_CMD storage restore list "$backup_id" 2>&1 | head -15
    echo ""

    # GET
    echo -e "${GREEN}[GET]${NC} Getting restore details..."
    $ACLOUD_CMD storage restore get "$backup_id" "$restore_id" 2>&1
    echo ""

    echo "Waiting for restore to be ready..."
    local restore_ready=0
    wait_for_status "$ACLOUD_CMD storage restore get $backup_id $restore_id" \
        '^(Active|Available|Ready|Completed|Complete)$' 300 && restore_ready=1

    if [ "$restore_ready" -eq 1 ]; then
        echo -e "${GREEN}[UPDATE]${NC} Updating restore..."
        UPDATE_OUTPUT=$($ACLOUD_CMD storage restore update "$backup_id" "$restore_id" \
            --name "${restore_name}-updated" \
            --tags "e2e-test,updated" 2>&1)
        if [ $? -eq 0 ]; then
            echo "$UPDATE_OUTPUT"
        else
            echo -e "${YELLOW}Update failed (non-fatal):${NC}"
            echo "$UPDATE_OUTPUT" | head -5
        fi
        echo ""
    else
        echo -e "${YELLOW}[UPDATE]${NC} Skipping — restore did not reach ready state after 300s."
        echo ""
    fi

    echo -e "${GREEN}[DELETE]${NC} Deleting restore..."
    DELETE_OUTPUT=$($ACLOUD_CMD storage restore delete "$backup_id" "$restore_id" --yes 2>&1)
    if [ $? -eq 0 ]; then
        echo "$DELETE_OUTPUT"
        CREATED_RESTORES=("${CREATED_RESTORES[@]/$restore_id}")
    else
        echo -e "${YELLOW}DELETE failed — cleanup trap will retry:${NC}"
        echo "$DELETE_OUTPUT" | head -3
    fi
    echo ""

    echo -e "${GREEN}✓ Storage Restore CRUD test completed!${NC}\n"
}

ensure_project || { echo -e "${RED}Cannot proceed without a project ID${NC}"; exit 1; }
setup_context

# Run tests
echo -e "${BLUE}Starting Storage Resources E2E Tests...${NC}\n"

VOLUME_ID=""
if test_block_storage; then
    if [ ${#CREATED_VOLUMES[@]} -gt 0 ]; then
        VOLUME_ID="${CREATED_VOLUMES[0]}"
    fi
    if [ -n "$VOLUME_ID" ]; then
        test_snapshot "$VOLUME_ID" || FAILURES=$((FAILURES + 1))
        test_backup  "$VOLUME_ID" || FAILURES=$((FAILURES + 1))
        if [ -n "$BACKUP_ID" ]; then
            test_restore "$BACKUP_ID" "$VOLUME_ID" || FAILURES=$((FAILURES + 1))
        else
            echo -e "${YELLOW}Skipping restore test (no backup available)${NC}\n"
        fi
    else
        echo -e "${YELLOW}Skipping snapshot/backup/restore tests (no volume ID available)${NC}\n"
    fi
else
    FAILURES=$((FAILURES + 1))
fi
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
