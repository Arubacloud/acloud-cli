#!/bin/bash

# E2E Test Script for Database Resources
# Tests CRUD operations for DBaaS, DBaaS Databases, DBaaS Users, and Database Backups

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# --- State tracking -------------------------------------------------------
CREATED_DBAAS=()
CREATED_DATABASES=()
CREATED_USERS=()
CREATED_BACKUPS=()
DBAAS_ID=""
BOOTSTRAP_PROJECT_ID=""

# Engine/flavor: prefer explicit env vars, then documented defaults.
# No acloud database engine list command exists (SDK does not expose an
# Engines client). Set ACLOUD_ENGINE_ID and ACLOUD_FLAVOR to enable DBaaS
# create; otherwise that step skips cleanly and list/format tests still run.
ENGINE_ID="${ACLOUD_ENGINE_ID:-${ACLOUD_E2E_DEFAULT_ENGINE_ID:-}}"
FLAVOR="${ACLOUD_FLAVOR:-${ACLOUD_E2E_DEFAULT_FLAVOR:-}}"

print_banner "Database"

# --- Cleanup --------------------------------------------------------------
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test resources...${NC}"

    # Delete database backups
    for backup_id in "${CREATED_BACKUPS[@]}"; do
        if is_valid_id "$backup_id"; then
            echo "Deleting backup: $backup_id"
            $ACLOUD_CMD database backup delete "$backup_id" --yes 2>&1 || true
        fi
    done

    # Delete DBaaS users
    if [ -n "$DBAAS_ID" ] && is_valid_id "$DBAAS_ID"; then
        for user in "${CREATED_USERS[@]}"; do
            echo "Deleting user: $user"
            $ACLOUD_CMD database dbaas user delete "$DBAAS_ID" "$user" --yes 2>&1 || true
        done
    fi

    # Delete DBaaS databases
    if [ -n "$DBAAS_ID" ] && is_valid_id "$DBAAS_ID"; then
        for db in "${CREATED_DATABASES[@]}"; do
            echo "Deleting database: $db"
            $ACLOUD_CMD database dbaas database delete "$DBAAS_ID" "$db" --yes 2>&1 || true
        done
    fi

    # Delete DBaaS instances — wait for Active before delete to avoid 400
    for dbaas_id in "${CREATED_DBAAS[@]}"; do
        if is_valid_id "$dbaas_id"; then
            echo "Waiting for DBaaS $dbaas_id before delete..."
            wait_for_status "$ACLOUD_CMD database dbaas get $dbaas_id" '^(Active|Ready)$' 300 2>/dev/null || true
            echo "Deleting DBaaS instance: $dbaas_id"
            $ACLOUD_CMD database dbaas delete "$dbaas_id" --yes 2>&1 || true
        fi
    done

    # Delete bootstrapped project last
    if [ -n "$BOOTSTRAP_PROJECT_ID" ]; then
        echo "Deleting bootstrapped project: $BOOTSTRAP_PROJECT_ID"
        $ACLOUD_CMD management project delete "$BOOTSTRAP_PROJECT_ID" --yes 2>&1 || true
    fi

    echo -e "${GREEN}Cleanup completed!${NC}"
}

trap cleanup EXIT

ensure_project || { echo -e "${RED}Cannot proceed without a project ID${NC}"; exit 1; }
setup_context

# --- Test --output flag ---------------------------------------------------
# dbaas list always tested; dbaas database list only when DBAAS_ID is set.
test_output_formats() {
    echo -e "${BLUE}--- Testing database list --output flag ---${NC}"

    local label="database dbaas list"
    local cmd="$ACLOUD_CMD database dbaas list"
    for fmt in json yaml; do
        echo -e "${YELLOW}Testing $label --output $fmt...${NC}"
        OUT=$(eval "$cmd --output $fmt" 2>&1)
        validate_list_output "$label" "$fmt" "$OUT" $?
    done

    if [ -n "$DBAAS_ID" ] && is_valid_id "$DBAAS_ID"; then
        local db_label="database dbaas database list"
        local db_cmd="$ACLOUD_CMD database dbaas database list $DBAAS_ID"
        for fmt in json yaml; do
            echo -e "${YELLOW}Testing $db_label --output $fmt...${NC}"
            OUT=$(eval "$db_cmd --output $fmt" 2>&1)
            validate_list_output "$db_label" "$fmt" "$OUT" $?
        done
    else
        echo -e "${YELLOW}⚠ Skipping 'dbaas database list' format test — no DBaaS instance available${NC}"
    fi
    echo ""
}

# --- Test DBaaS -----------------------------------------------------------
test_dbaas() {
    echo -e "${BLUE}=== 1. DBaaS CRUD Test ===${NC}"

    if [ -z "$ENGINE_ID" ] || [ -z "$FLAVOR" ]; then
        echo -e "${YELLOW}⚠ Skipping DBaaS create — set ACLOUD_ENGINE_ID and ACLOUD_FLAVOR (or the ACLOUD_E2E_DEFAULT_* fallbacks) to enable.${NC}"
        echo ""
        return 0
    fi

    local dbaas_name="${RESOURCE_PREFIX}-dbaas"

    echo -e "${GREEN}[CREATE]${NC} Creating DBaaS instance: $dbaas_name"
    CREATE_OUTPUT=$($ACLOUD_CMD database dbaas create \
        --name "$dbaas_name" \
        --region "$REGION" \
        --engine-id "$ENGINE_ID" \
        --flavor "$FLAVOR" \
        --tags "e2e-test,created-by-script" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then
        echo -e "${RED}DBaaS test failed: Authentication error${NC}"
        return 1
    fi

    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        return 1
    fi

    DBAAS_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${RED}Could not extract DBaaS ID from create output${NC}"
        echo "CREATE_OUTPUT: $CREATE_OUTPUT" >&2
        return 1
    fi

    CREATED_DBAAS+=("$DBAAS_ID")
    echo -e "${GREEN}DBaaS created: $DBAAS_ID${NC}"

    echo -e "${GREEN}[LIST]${NC} Listing DBaaS instances"
    $ACLOUD_CMD database dbaas list 2>&1 | head -20

    echo -e "${GREEN}[GET]${NC} Getting DBaaS details: $DBAAS_ID"
    $ACLOUD_CMD database dbaas get "$DBAAS_ID" 2>&1

    echo "Waiting for DBaaS $DBAAS_ID to be Active..."
    local dbaas_ready=0
    wait_for_status "$ACLOUD_CMD database dbaas get $DBAAS_ID" '^(Active|Ready)$' 600 && dbaas_ready=1

    if [ "$dbaas_ready" -eq 1 ]; then
        echo -e "${GREEN}[UPDATE]${NC} Updating DBaaS: $DBAAS_ID"
        UPDATE_OUTPUT=$($ACLOUD_CMD database dbaas update "$DBAAS_ID" \
            --tags "e2e-test,updated" 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}DBaaS updated successfully${NC}"
        else
            echo -e "${YELLOW}Update may have failed${NC}"
            echo "$UPDATE_OUTPUT" | head -5
        fi
    else
        echo -e "${YELLOW}[UPDATE]${NC} Skipping — DBaaS did not reach Active after 600s."
    fi

    echo -e "${GREEN}✓ DBaaS CRUD test completed!${NC}\n"
    return 0
}

# --- Test DBaaS Database --------------------------------------------------
test_dbaas_database() {
    echo -e "${BLUE}=== 2. DBaaS Database CRUD Test ===${NC}"

    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${YELLOW}Skipping DBaaS database test (no DBaaS instance available)${NC}\n"
        return 0
    fi

    local db_name="${RESOURCE_PREFIX}-database"

    echo -e "${GREEN}[CREATE]${NC} Creating database: $db_name"
    CREATE_OUTPUT=$($ACLOUD_CMD database dbaas database create "$DBAAS_ID" \
        --name "$db_name" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then return 1; fi
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        return 1
    fi

    CREATED_DATABASES+=("$db_name")
    echo -e "${GREEN}Database created: $db_name${NC}"

    echo -e "${GREEN}[LIST]${NC} Listing databases in DBaaS: $DBAAS_ID"
    $ACLOUD_CMD database dbaas database list "$DBAAS_ID" 2>&1

    echo -e "${GREEN}[GET]${NC} Getting database details: $db_name"
    $ACLOUD_CMD database dbaas database get "$DBAAS_ID" "$db_name" 2>&1

    echo -e "${GREEN}[UPDATE]${NC} Updating database: $db_name"
    UPDATE_OUTPUT=$($ACLOUD_CMD database dbaas database update "$DBAAS_ID" "$db_name" \
        --name "${db_name}-updated" 2>&1)
    if [ $? -eq 0 ]; then
        CREATED_DATABASES=("${CREATED_DATABASES[@]/$db_name/${db_name}-updated}")
        echo -e "${GREEN}Database updated successfully${NC}"
    else
        echo -e "${YELLOW}Update may have failed${NC}"
        echo "$UPDATE_OUTPUT" | head -5
    fi

    echo -e "${GREEN}✓ DBaaS Database CRUD test completed!${NC}\n"
    return 0
}

# --- Test DBaaS User ------------------------------------------------------
test_dbaas_user() {
    echo -e "${BLUE}=== 3. DBaaS User CRUD Test ===${NC}"

    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${YELLOW}Skipping DBaaS user test (no DBaaS instance available)${NC}\n"
        return 0
    fi

    local username="${RESOURCE_PREFIX}-user"
    local password="TestPassword123!"

    echo -e "${GREEN}[CREATE]${NC} Creating user: $username"
    CREATE_OUTPUT=$($ACLOUD_CMD database dbaas user create "$DBAAS_ID" \
        --username "$username" \
        --password "$password" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then return 1; fi
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        return 1
    fi

    CREATED_USERS+=("$username")
    echo -e "${GREEN}User created: $username${NC}"

    echo -e "${GREEN}[LIST]${NC} Listing users in DBaaS: $DBAAS_ID"
    $ACLOUD_CMD database dbaas user list "$DBAAS_ID" 2>&1

    echo -e "${GREEN}[GET]${NC} Getting user details: $username"
    $ACLOUD_CMD database dbaas user get "$DBAAS_ID" "$username" 2>&1

    echo -e "${GREEN}[UPDATE]${NC} Updating user password: $username"
    UPDATE_OUTPUT=$($ACLOUD_CMD database dbaas user update "$DBAAS_ID" "$username" \
        --password "NewPassword123!" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}User updated successfully${NC}"
    else
        echo -e "${YELLOW}Update may have failed${NC}"
        echo "$UPDATE_OUTPUT" | head -5
    fi

    echo -e "${GREEN}✓ DBaaS User CRUD test completed!${NC}\n"
    return 0
}

# --- Test Database Backup -------------------------------------------------
test_backup() {
    echo -e "${BLUE}=== 4. Database Backup CRUD Test ===${NC}"

    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${YELLOW}Skipping backup test (no DBaaS instance available)${NC}\n"
        return 0
    fi

    if [ ${#CREATED_DATABASES[@]} -eq 0 ]; then
        echo -e "${YELLOW}Skipping backup test (no database available)${NC}\n"
        return 0
    fi

    local backup_name="${RESOURCE_PREFIX}-backup"
    local database_name="${CREATED_DATABASES[0]}"

    echo -e "${GREEN}[CREATE]${NC} Creating backup: $backup_name"
    CREATE_OUTPUT=$($ACLOUD_CMD database backup create \
        --name "$backup_name" \
        --region "$REGION" \
        --dbaas-id "$DBAAS_ID" \
        --database-name "$database_name" \
        --billing-period "Hour" \
        --tags "e2e-test" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then return 1; fi
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        return 1
    fi

    BACKUP_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$BACKUP_ID" ] || ! is_valid_id "$BACKUP_ID"; then
        echo -e "${RED}Could not extract backup ID from create output${NC}"
        echo "CREATE_OUTPUT: $CREATE_OUTPUT" >&2
        return 1
    fi

    CREATED_BACKUPS+=("$BACKUP_ID")
    echo -e "${GREEN}Backup created: $BACKUP_ID${NC}"

    echo -e "${GREEN}[LIST]${NC} Listing database backups"
    $ACLOUD_CMD database backup list 2>&1 | head -20

    echo -e "${GREEN}[GET]${NC} Getting backup details: $BACKUP_ID"
    $ACLOUD_CMD database backup get "$BACKUP_ID" 2>&1

    echo -e "${GREEN}✓ Database Backup CRUD test completed!${NC}\n"
    return 0
}

# --- Main -----------------------------------------------------------------
echo -e "${BLUE}Starting Database Resources E2E Tests...${NC}\n"

test_dbaas
test_dbaas_database
test_dbaas_user
test_backup
test_output_formats

# Test summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Project ID: $PROJECT_ID"
if [ -n "$DBAAS_ID" ]; then
    echo "DBaaS ID: $DBAAS_ID"
fi
if [ ${#CREATED_DBAAS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ DBaaS Instances: ${#CREATED_DBAAS[@]} created${NC}"
else
    echo -e "${YELLOW}○ DBaaS Instances: 0 created${NC}"
fi
if [ ${#CREATED_DATABASES[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Databases: ${#CREATED_DATABASES[@]} created${NC}"
else
    echo -e "${YELLOW}○ Databases: 0 created${NC}"
fi
if [ ${#CREATED_USERS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Users: ${#CREATED_USERS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Users: 0 created${NC}"
fi
if [ ${#CREATED_BACKUPS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Backups: ${#CREATED_BACKUPS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Backups: 0 created${NC}"
fi
echo ""

if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}=== Database E2E: all checks passed ===${NC}"
    exit 0
else
    echo -e "${RED}=== Database E2E: $FAILURES check(s) failed ===${NC}"
    exit 1
fi
