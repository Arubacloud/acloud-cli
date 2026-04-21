#!/bin/bash

# E2E Test Script for Database Resources
# Tests CRUD operations for DBaaS, DBaaS Databases, DBaaS Users, and Database Backups

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# Cleanup tracking
CREATED_DBAAS=()
CREATED_DATABASES=()
CREATED_USERS=()
CREATED_BACKUPS=()
DBAAS_ID=""
ENGINE_ID="${ACLOUD_ENGINE_ID:-}"  # Optional: DBaaS engine ID
FLAVOR="${ACLOUD_FLAVOR:-}"  # Optional: DBaaS flavor

print_banner "Database"

# Test --output flag for database list commands
test_output_formats() {
    echo -e "${BLUE}--- Testing database list --output flag ---${NC}"

    for resource_cmd in "dbaas list" "dbaas database list --dbaas-id dummy-id"; do
        local label="database $resource_cmd"
        echo -e "${YELLOW}Testing $label --output json...${NC}"
        JSON_OUTPUT=$($ACLOUD_CMD database $resource_cmd --project-id "$PROJECT_ID" --output json 2>&1)
        JSON_EXIT=$?
        if [ $JSON_EXIT -ne 0 ]; then
            echo -e "${YELLOW}⚠ $label --output json: command failed (may need valid dbaas-id)${NC}"
        elif echo "$JSON_OUTPUT" | grep -qF "No "; then
            echo -e "${YELLOW}⚠ $label --output json: no resources — format validation skipped${NC}"
        elif is_valid_json "$JSON_OUTPUT"; then
            echo -e "${GREEN}✓ $label --output json: valid JSON${NC}"
        else
            echo -e "${RED}✗ $label --output json: output is not valid JSON${NC}"
        fi

        echo -e "${YELLOW}Testing $label --output yaml...${NC}"
        YAML_OUTPUT=$($ACLOUD_CMD database $resource_cmd --project-id "$PROJECT_ID" --output yaml 2>&1)
        YAML_EXIT=$?
        if [ $YAML_EXIT -ne 0 ]; then
            echo -e "${YELLOW}⚠ $label --output yaml: command failed (may need valid dbaas-id)${NC}"
        elif echo "$YAML_OUTPUT" | grep -qF "No "; then
            echo -e "${YELLOW}⚠ $label --output yaml: no resources — format validation skipped${NC}"
        elif echo "$YAML_OUTPUT" | grep -qE '^[a-zA-Z].*:|^- '; then
            echo -e "${GREEN}✓ $label --output yaml: output looks like YAML${NC}"
        else
            echo -e "${RED}✗ $label --output yaml: output does not look like YAML${NC}"
        fi
    done
    echo ""
}

# Cleanup function
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
    
    # Delete DBaaS instances
    for dbaas_id in "${CREATED_DBAAS[@]}"; do
        if is_valid_id "$dbaas_id"; then
            echo "Deleting DBaaS instance: $dbaas_id"
            $ACLOUD_CMD database dbaas delete "$dbaas_id" --yes 2>&1 || true
        fi
    done
    
    echo -e "${GREEN}Cleanup completed!${NC}"
}

# Trap to ensure cleanup runs on exit
trap cleanup EXIT

setup_context

# Test function for DBaaS
test_dbaas() {
    echo -e "${BLUE}=== 1. DBaaS CRUD Test ===${NC}"
    
    if [ -z "$ENGINE_ID" ] || [ -z "$FLAVOR" ]; then
        echo -e "${YELLOW}Skipping DBaaS test (ENGINE_ID and FLAVOR not set)${NC}"
        echo "Set ACLOUD_ENGINE_ID and ACLOUD_FLAVOR to test DBaaS"
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
        echo -e "${RED}DBaaS test failed${NC}"
        return 1
    fi
    
    DBAAS_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${RED}Could not extract DBaaS ID from create output${NC}"
        echo "CREATE_OUTPUT: $CREATE_OUTPUT" >&2
        echo -e "${RED}DBaaS test failed${NC}"
        return 1
    fi
    
    CREATED_DBAAS+=("$DBAAS_ID")
    echo -e "${GREEN}DBaaS created: $DBAAS_ID${NC}"
    
    echo -e "${GREEN}[LIST]${NC} Listing DBaaS instances"
    $ACLOUD_CMD database dbaas list 2>&1 | head -20
    
    echo -e "${GREEN}[GET]${NC} Getting DBaaS details: $DBAAS_ID"
    $ACLOUD_CMD database dbaas get "$DBAAS_ID" 2>&1
    
    echo -e "${GREEN}[UPDATE]${NC} Updating DBaaS: $DBAAS_ID"
    UPDATE_OUTPUT=$($ACLOUD_CMD database dbaas update "$DBAAS_ID" \
        --tags "e2e-test,updated" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}DBaaS updated successfully${NC}"
    else
        echo -e "${YELLOW}Update may have failed or resource is in InCreation state${NC}"
        echo "$UPDATE_OUTPUT" | head -5
    fi
    
    echo -e "${GREEN}DBaaS test completed successfully${NC}\n"
    return 0
}

# Test function for DBaaS Database
test_dbaas_database() {
    echo -e "${BLUE}=== 2. DBaaS Database CRUD Test ===${NC}"
    
    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${YELLOW}Skipping DBaaS database test (no DBaaS instance available)${NC}"
        return 0
    fi
    
    local db_name="${RESOURCE_PREFIX}-database"
    
    echo -e "${GREEN}[CREATE]${NC} Creating database: $db_name"
    CREATE_OUTPUT=$($ACLOUD_CMD database dbaas database create "$DBAAS_ID" \
        --name "$db_name" 2>&1)
    exit_code=$?
    
    if ! check_auth_error "$CREATE_OUTPUT"; then
        echo -e "${RED}Database test failed: Authentication error${NC}"
        return 1
    fi
    
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        echo -e "${RED}Database test failed${NC}"
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
    
    echo -e "${GREEN}Database test completed successfully${NC}\n"
    return 0
}

# Test function for DBaaS User
test_dbaas_user() {
    echo -e "${BLUE}=== 3. DBaaS User CRUD Test ===${NC}"
    
    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${YELLOW}Skipping DBaaS user test (no DBaaS instance available)${NC}"
        return 0
    fi
    
    local username="${RESOURCE_PREFIX}-user"
    local password="TestPassword123!"
    
    echo -e "${GREEN}[CREATE]${NC} Creating user: $username"
    CREATE_OUTPUT=$($ACLOUD_CMD database dbaas user create "$DBAAS_ID" \
        --username "$username" \
        --password "$password" 2>&1)
    exit_code=$?
    
    if ! check_auth_error "$CREATE_OUTPUT"; then
        echo -e "${RED}User test failed: Authentication error${NC}"
        return 1
    fi
    
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        echo -e "${RED}User test failed${NC}"
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
    
    echo -e "${GREEN}User test completed successfully${NC}\n"
    return 0
}

# Test function for Database Backup
test_backup() {
    echo -e "${BLUE}=== 4. Database Backup CRUD Test ===${NC}"
    
    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${YELLOW}Skipping backup test (no DBaaS instance available)${NC}"
        return 0
    fi
    
    if [ ${#CREATED_DATABASES[@]} -eq 0 ]; then
        echo -e "${YELLOW}Skipping backup test (no database available)${NC}"
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
    
    if ! check_auth_error "$CREATE_OUTPUT"; then
        echo -e "${RED}Backup test failed: Authentication error${NC}"
        return 1
    fi
    
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        echo -e "${RED}Backup test failed${NC}"
        return 1
    fi
    
    BACKUP_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$BACKUP_ID" ] || ! is_valid_id "$BACKUP_ID"; then
        echo -e "${RED}Could not extract backup ID from create output${NC}"
        echo "CREATE_OUTPUT: $CREATE_OUTPUT" >&2
        echo -e "${RED}Backup test failed${NC}"
        return 1
    fi
    
    CREATED_BACKUPS+=("$BACKUP_ID")
    echo -e "${GREEN}Backup created: $BACKUP_ID${NC}"
    
    echo -e "${GREEN}[LIST]${NC} Listing database backups"
    $ACLOUD_CMD database backup list 2>&1 | head -20
    
    echo -e "${GREEN}[GET]${NC} Getting backup details: $BACKUP_ID"
    $ACLOUD_CMD database backup get "$BACKUP_ID" 2>&1
    
    echo -e "${GREEN}Backup test completed successfully${NC}\n"
    return 0
}

# Run tests
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
echo "○ DBaaS Instances: ${#CREATED_DBAAS[@]} created"
echo "○ Databases: ${#CREATED_DATABASES[@]} created"
echo "○ Users: ${#CREATED_USERS[@]} created"
echo "○ Backups: ${#CREATED_BACKUPS[@]} created"
echo ""

echo -e "${GREEN}=== All Database Tests Completed! ===${NC}"

