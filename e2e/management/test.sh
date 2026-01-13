#!/bin/bash

# E2E Test Script for Management Resources
# Tests CRUD operations for Projects

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME_PREFIX="e2e-test-$(date +%s)"

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

echo -e "${BLUE}=== Management Resources E2E Test ===${NC}\n"
echo "Test prefix: $PROJECT_NAME_PREFIX"
echo "ACLOUD command: $ACLOUD_CMD"
echo ""

# Function to extract resource ID from output
extract_id() {
    local output="$1"
    echo "$output" | grep -oE '[a-f0-9]{24}' | head -1
}

# Function to test Project CRUD
test_project() {
    local project_name="${PROJECT_NAME_PREFIX}-project"
    
    echo -e "${YELLOW}--- Testing Project CRUD ---${NC}\n"
    
    # CREATE
    echo -e "${GREEN}[CREATE]${NC} Creating project: $project_name"
    CREATE_OUTPUT=$($ACLOUD_CMD management project create \
        --name "$project_name" \
        --description "E2E test project" \
        --tags "e2e,test,management" 2>&1) || {
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT"
        # Check for common error patterns
        if echo "$CREATE_OUTPUT" | grep -qi "authentication failed\|invalid_client\|Invalid client"; then
            echo -e "${RED}Authentication error detected. Please check your credentials.${NC}"
        fi
        return 1
    }
    echo "$CREATE_OUTPUT"
    
    PROJECT_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$PROJECT_ID" ]; then
        echo -e "${RED}Could not extract project ID from create output${NC}"
        # Check for common error patterns
        if echo "$CREATE_OUTPUT" | grep -qi "authentication failed\|invalid_client\|Invalid client"; then
            echo -e "${RED}Authentication error detected. Please check your credentials.${NC}"
        fi
        return 1
    fi
    echo -e "${GREEN}Created project ID: $PROJECT_ID${NC}\n"
    
    # LIST
    echo -e "${GREEN}[LIST]${NC} Listing projects..."
    LIST_OUTPUT=$($ACLOUD_CMD management project list 2>&1) || {
        echo -e "${RED}LIST failed:${NC}"
        echo "$LIST_OUTPUT"
        return 1
    }
    echo "$LIST_OUTPUT" | head -15
    echo ""
    
    # GET
    echo -e "${GREEN}[GET]${NC} Getting project details..."
    GET_OUTPUT=$($ACLOUD_CMD management project get "$PROJECT_ID" 2>&1) || {
        echo -e "${RED}GET failed:${NC}"
        echo "$GET_OUTPUT"
        return 1
    }
    echo "$GET_OUTPUT"
    echo ""
    
    # UPDATE
    echo -e "${GREEN}[UPDATE]${NC} Updating project..."
    UPDATE_OUTPUT=$($ACLOUD_CMD management project update "$PROJECT_ID" \
        --description "Updated E2E test project" \
        --tags "e2e,test,updated" 2>&1) || {
        echo -e "${RED}UPDATE failed:${NC}"
        echo "$UPDATE_OUTPUT"
        return 1
    }
    echo "$UPDATE_OUTPUT"
    echo ""
    
    # DELETE
    echo -e "${GREEN}[DELETE]${NC} Deleting project..."
    DELETE_OUTPUT=$($ACLOUD_CMD management project delete "$PROJECT_ID" --yes 2>&1) || {
        echo -e "${RED}DELETE failed:${NC}"
        echo "$DELETE_OUTPUT"
        return 1
    }
    echo "$DELETE_OUTPUT"
    echo ""
    
    echo -e "${GREEN}✓ Project CRUD test completed successfully!${NC}\n"
}

# Run tests
echo -e "${BLUE}Starting Management Resources E2E Tests...${NC}\n"

TEST_PASSED=false
if test_project; then
    TEST_PASSED=true
fi

echo -e "${GREEN}=== All Management Tests Completed! ===${NC}\n"

# Print summary
echo -e "${BLUE}=== Test Summary ===${NC}"
if [ "$TEST_PASSED" = true ]; then
    echo -e "${GREEN}✓ Project CRUD: Passed${NC}"
    if [ -n "$PROJECT_ID" ]; then
        echo -e "  Project ID: $PROJECT_ID (deleted)"
    fi
else
    echo -e "${RED}✗ Project CRUD: Failed${NC}"
fi
echo ""

if [ "$TEST_PASSED" = true ]; then
    exit 0
else
    exit 1
fi

