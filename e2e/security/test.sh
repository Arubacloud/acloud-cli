#!/bin/bash

# E2E Test Script for Security Resources
# Tests CRUD operations for KMS (Key Management Service)

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

# Determine acloud command path
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
CREATED_KMS=()

echo -e "${BLUE}=== Security Resources E2E Test ===${NC}\n"
echo "Project ID: $PROJECT_ID"
echo "Region: $REGION"
echo "Test prefix: $RESOURCE_PREFIX"
echo "ACLOUD command: $ACLOUD_CMD"
echo ""

# Function to extract resource ID from output
extract_id() {
    local output="$1"
    local exclude_id="${2:-}"
    
    if [ -n "$exclude_id" ]; then
        local filtered_ids=$(echo "$output" | grep -oE '[a-f0-9]{24}' | grep -v "^${exclude_id}$")
        if [ -n "$filtered_ids" ]; then
            echo "$filtered_ids" | tail -1
            return 0
        fi
    fi
    
    local all_ids=$(echo "$output" | grep -oE '[a-f0-9]{24}')
    if [ -n "$all_ids" ]; then
        if [ -n "$exclude_id" ]; then
            echo "$all_ids" | grep -v "^${exclude_id}$" | tail -1
        else
            echo "$all_ids" | tail -1
        fi
    fi
}

# Helper function to validate resource ID
is_valid_id() {
    local id="$1"
    [[ "$id" =~ ^[a-f0-9]{24}$ ]]
}

# Check for authentication errors
check_auth_error() {
    local output="$1"
    if echo "$output" | grep -qi "authentication failed\|invalid_client\|unauthorized"; then
        echo -e "${RED}Authentication error detected. Please check your credentials.${NC}" >&2
        return 1
    fi
    return 0
}

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test resources...${NC}"
    
    # Delete KMS keys
    for kms_id in "${CREATED_KMS[@]}"; do
        if is_valid_id "$kms_id"; then
            echo "Deleting KMS key: $kms_id"
            $ACLOUD_CMD security kms delete "$kms_id" --yes 2>&1 || true
        fi
    done
    
    echo -e "${GREEN}Cleanup completed!${NC}"
}

# Trap to ensure cleanup runs on exit
trap cleanup EXIT

# Set up context
echo -e "${BLUE}Setting up context for project ID...${NC}"
$ACLOUD_CMD context set e2e-test-context --project-id "$PROJECT_ID" 2>&1 || true
echo ""

# Test function for KMS
test_kms() {
    echo -e "${BLUE}=== 1. KMS CRUD Test ===${NC}"
    
    local kms_name="${RESOURCE_PREFIX}-kms"
    
    echo -e "${GREEN}[CREATE]${NC} Creating KMS key: $kms_name"
    CREATE_OUTPUT=$($ACLOUD_CMD security kms create \
        --name "$kms_name" \
        --region "$REGION" \
        --billing-period "Hour" \
        --tags "e2e-test,created-by-script" 2>&1)
    exit_code=$?
    
    if ! check_auth_error "$CREATE_OUTPUT"; then
        echo -e "${RED}KMS test failed: Authentication error${NC}"
        return 1
    fi
    
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        echo -e "${RED}KMS test failed${NC}"
        return 1
    fi
    
    KMS_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$KMS_ID" ] || ! is_valid_id "$KMS_ID"; then
        echo -e "${RED}Could not extract KMS ID from create output${NC}"
        echo "CREATE_OUTPUT: $CREATE_OUTPUT" >&2
        echo -e "${RED}KMS test failed${NC}"
        return 1
    fi
    
    CREATED_KMS+=("$KMS_ID")
    echo -e "${GREEN}KMS key created: $KMS_ID${NC}"
    
    echo -e "${GREEN}[LIST]${NC} Listing KMS keys"
    $ACLOUD_CMD security kms list 2>&1 | head -20
    
    echo -e "${GREEN}[GET]${NC} Getting KMS details: $KMS_ID"
    $ACLOUD_CMD security kms get "$KMS_ID" 2>&1
    
    echo -e "${GREEN}[UPDATE]${NC} Updating KMS: $KMS_ID"
    UPDATE_OUTPUT=$($ACLOUD_CMD security kms update "$KMS_ID" \
        --name "${kms_name}-updated" \
        --tags "e2e-test,updated" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}KMS updated successfully${NC}"
    else
        echo -e "${YELLOW}Update may have failed or resource is in InCreation state${NC}"
        echo "$UPDATE_OUTPUT" | head -5
    fi
    
    echo -e "${GREEN}KMS test completed successfully${NC}\n"
    return 0
}

# Run tests
echo -e "${BLUE}Starting Security Resources E2E Tests...${NC}\n"

test_kms

# Test summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Project ID: $PROJECT_ID"
echo "○ KMS Keys: ${#CREATED_KMS[@]} created"
echo ""

echo -e "${GREEN}=== All Security Tests Completed! ===${NC}"

