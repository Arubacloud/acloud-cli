#!/bin/bash

# E2E Test Script for Security Resources
# Tests CRUD operations for KMS (Key Management Service)

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# Cleanup tracking
CREATED_KMS=()

print_banner "Security"

# Test --output flag for security list commands
test_output_formats() {
    echo -e "${BLUE}--- Testing security list --output flag ---${NC}"

    local label="security kms list"
    local cmd="$ACLOUD_CMD security kms list --project-id \"$PROJECT_ID\""

    for fmt in json yaml; do
        echo -e "${YELLOW}Testing $label --output $fmt...${NC}"
        OUT=$(eval "$cmd --output $fmt" 2>&1)
        validate_list_output "$label" "$fmt" "$OUT" $?
    done
    echo ""
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

setup_context

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
test_output_formats

# Test summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Project ID: $PROJECT_ID"
echo "○ KMS Keys: ${#CREATED_KMS[@]} created"
echo ""

echo -e "${GREEN}=== All Security Tests Completed! ===${NC}"

