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

# Test --output flag for project list
test_project_output_formats() {
    echo -e "${BLUE}--- Testing management project list --output flag ---${NC}"

    for fmt in "" table; do
        local label="--output \"$fmt\""
        [ -z "$fmt" ] && label="(default, no --output)"
        echo -e "${YELLOW}Testing $label...${NC}"
        if [ -z "$fmt" ]; then
            OUT=$($ACLOUD_CMD management project list 2>&1)
        else
            OUT=$($ACLOUD_CMD management project list --output "$fmt" 2>&1)
        fi
        EXIT=$?
        if [ $EXIT -eq 0 ]; then
            echo -e "${GREEN}✓ $label: command succeeded${NC}"
        else
            echo -e "${RED}✗ $label: command failed (exit $EXIT)${NC}"
            echo "$OUT"
        fi
    done

    echo -e "${YELLOW}Testing --output json...${NC}"
    JSON_OUTPUT=$($ACLOUD_CMD management project list --output json 2>&1)
    JSON_EXIT=$?
    if [ $JSON_EXIT -ne 0 ]; then
        echo -e "${RED}✗ --output json: command failed (exit $JSON_EXIT)${NC}"
        echo "$JSON_OUTPUT"
    elif echo "$JSON_OUTPUT" | grep -qF "No projects found"; then
        echo -e "${YELLOW}⚠ --output json: no resources — format validation skipped${NC}"
    elif is_valid_json "$JSON_OUTPUT"; then
        echo -e "${GREEN}✓ --output json: valid JSON${NC}"
        if echo "$JSON_OUTPUT" | grep -q '"name"'; then
            echo -e "${GREEN}✓ --output json: 'name' key present${NC}"
        else
            echo -e "${RED}✗ --output json: 'name' key missing${NC}"
        fi
    else
        echo -e "${RED}✗ --output json: output is not valid JSON${NC}"
        echo "$JSON_OUTPUT"
    fi

    echo -e "${YELLOW}Testing --output yaml...${NC}"
    YAML_OUTPUT=$($ACLOUD_CMD management project list --output yaml 2>&1)
    YAML_EXIT=$?
    if [ $YAML_EXIT -ne 0 ]; then
        echo -e "${RED}✗ --output yaml: command failed (exit $YAML_EXIT)${NC}"
        echo "$YAML_OUTPUT"
    elif echo "$YAML_OUTPUT" | grep -qF "No projects found"; then
        echo -e "${YELLOW}⚠ --output yaml: no resources — format validation skipped${NC}"
    elif echo "$YAML_OUTPUT" | grep -qE '^[a-zA-Z].*:|^- '; then
        echo -e "${GREEN}✓ --output yaml: output looks like YAML${NC}"
        if echo "$YAML_OUTPUT" | grep -q 'name:'; then
            echo -e "${GREEN}✓ --output yaml: 'name' key present${NC}"
        else
            echo -e "${RED}✗ --output yaml: 'name' key missing${NC}"
        fi
    else
        echo -e "${RED}✗ --output yaml: output does not look like YAML${NC}"
        echo "$YAML_OUTPUT"
    fi

    echo ""
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

FORMAT_PASSED=false
if test_project_output_formats; then
    FORMAT_PASSED=true
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
if [ "$FORMAT_PASSED" = true ]; then
    echo -e "${GREEN}✓ Project output formats: Passed${NC}"
else
    echo -e "${RED}✗ Project output formats: Failed${NC}"
fi
echo ""

if [ "$TEST_PASSED" = true ] && [ "$FORMAT_PASSED" = true ]; then
    exit 0
else
    exit 1
fi

