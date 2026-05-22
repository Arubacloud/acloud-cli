#!/bin/bash

# E2E Test Script for Management Resources
# Tests CRUD operations for Projects

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# Cleanup tracking
CREATED_PROJECTS=()

print_banner "Management"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test resources...${NC}"

    for proj_id in "${CREATED_PROJECTS[@]}"; do
        if is_valid_id "$proj_id"; then
            echo "Deleting project: $proj_id"
            local del_elapsed=0
            while [ "$del_elapsed" -lt 120 ]; do
                $ACLOUD_CMD management project delete "$proj_id" --yes 2>&1 && break
                if $ACLOUD_CMD management project get "$proj_id" 2>&1 | grep -qi "not found\|404"; then break; fi
                sleep 10
                del_elapsed=$((del_elapsed + 10))
            done
        fi
    done

    echo -e "${GREEN}Cleanup completed!${NC}"
}

trap cleanup EXIT

# Test --output flag for project list
test_project_output_formats() {
    local fn_fail=0
    _fn_fail() { fn_fail=$((fn_fail + 1)); echo -e "${RED}✗ $*${NC}"; }

    echo -e "${BLUE}--- Testing management project list --output flag ---${NC}"

    for fmt in "" table std standard; do
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
            _fn_fail "$label: command failed (exit $EXIT)"
            echo "$OUT"
        fi
    done

    echo -e "${YELLOW}Testing --output table-json...${NC}"
    TABLE_JSON_OUTPUT=$($ACLOUD_CMD management project list --output table-json 2>&1)
    TABLE_JSON_EXIT=$?
    if [ $TABLE_JSON_EXIT -ne 0 ]; then
        _fn_fail "--output table-json: command failed (exit $TABLE_JSON_EXIT)"
        echo "$TABLE_JSON_OUTPUT"
    elif echo "$TABLE_JSON_OUTPUT" | grep -qF "No projects found"; then
        echo -e "${YELLOW}⚠ --output table-json: no resources — format validation skipped${NC}"
    elif is_valid_json "$TABLE_JSON_OUTPUT"; then
        echo -e "${GREEN}✓ --output table-json: valid JSON${NC}"
        if echo "$TABLE_JSON_OUTPUT" | grep -q '"name"'; then
            echo -e "${GREEN}✓ --output table-json: 'name' key present${NC}"
        else
            _fn_fail "--output table-json: 'name' key missing"
        fi
    else
        _fn_fail "--output table-json: output is not valid JSON"
        echo "$TABLE_JSON_OUTPUT"
    fi

    echo -e "${YELLOW}Testing --output table-yaml...${NC}"
    TABLE_YAML_OUTPUT=$($ACLOUD_CMD management project list --output table-yaml 2>&1)
    TABLE_YAML_EXIT=$?
    if [ $TABLE_YAML_EXIT -ne 0 ]; then
        _fn_fail "--output table-yaml: command failed (exit $TABLE_YAML_EXIT)"
        echo "$TABLE_YAML_OUTPUT"
    elif echo "$TABLE_YAML_OUTPUT" | grep -qF "No projects found"; then
        echo -e "${YELLOW}⚠ --output table-yaml: no resources — format validation skipped${NC}"
    elif echo "$TABLE_YAML_OUTPUT" | grep -qE '^[a-zA-Z].*:|^- '; then
        echo -e "${GREEN}✓ --output table-yaml: output looks like YAML${NC}"
        if echo "$TABLE_YAML_OUTPUT" | grep -q 'name:'; then
            echo -e "${GREEN}✓ --output table-yaml: 'name' key present${NC}"
        else
            _fn_fail "--output table-yaml: 'name' key missing"
        fi
    else
        _fn_fail "--output table-yaml: output does not look like YAML"
        echo "$TABLE_YAML_OUTPUT"
    fi

    echo -e "${YELLOW}Testing --output json (full SDK response)...${NC}"
    JSON_OUTPUT=$($ACLOUD_CMD management project list --output json 2>&1)
    JSON_EXIT=$?
    if [ $JSON_EXIT -ne 0 ]; then
        _fn_fail "--output json: command failed (exit $JSON_EXIT)"
        echo "$JSON_OUTPUT"
    elif echo "$JSON_OUTPUT" | grep -qF "No projects found"; then
        echo -e "${YELLOW}⚠ --output json: no resources — format validation skipped${NC}"
    elif is_valid_json "$JSON_OUTPUT"; then
        echo -e "${GREEN}✓ --output json: valid JSON${NC}"
        if echo "$JSON_OUTPUT" | grep -q '"values"'; then
            echo -e "${GREEN}✓ --output json: 'values' key present (full SDK response)${NC}"
        else
            echo -e "${YELLOW}⚠ --output json: 'values' key not found (empty list or different shape)${NC}"
        fi
    else
        _fn_fail "--output json: output is not valid JSON"
        echo "$JSON_OUTPUT"
    fi

    echo -e "${YELLOW}Testing --output yaml (full SDK response)...${NC}"
    YAML_OUTPUT=$($ACLOUD_CMD management project list --output yaml 2>&1)
    YAML_EXIT=$?
    if [ $YAML_EXIT -ne 0 ]; then
        _fn_fail "--output yaml: command failed (exit $YAML_EXIT)"
        echo "$YAML_OUTPUT"
    elif echo "$YAML_OUTPUT" | grep -qF "No projects found"; then
        echo -e "${YELLOW}⚠ --output yaml: no resources — format validation skipped${NC}"
    elif echo "$YAML_OUTPUT" | grep -qE '^[a-zA-Z].*:|^- '; then
        echo -e "${GREEN}✓ --output yaml: output looks like YAML (full SDK response)${NC}"
        if echo "$YAML_OUTPUT" | grep -q 'values:'; then
            echo -e "${GREEN}✓ --output yaml: 'values' key present (full SDK response)${NC}"
        else
            echo -e "${YELLOW}⚠ --output yaml: 'values' key not found (empty list or different shape)${NC}"
        fi
    else
        _fn_fail "--output yaml: output does not look like YAML"
        echo "$YAML_OUTPUT"
    fi

    echo ""
    return $fn_fail
}

# Function to test Project CRUD
test_project() {
    local project_name="${RESOURCE_PREFIX}-project"
    local crud_project_id=""

    echo -e "${YELLOW}--- Testing Project CRUD ---${NC}\n"

    # CREATE
    echo -e "${GREEN}[CREATE]${NC} Creating project: $project_name"
    CREATE_OUTPUT=$($ACLOUD_CMD management project create \
        --name "$project_name" \
        --description "E2E test project" \
        --tags "e2e-test,management" 2>&1)
    if [ $? -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi
    echo "$CREATE_OUTPUT"

    crud_project_id=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$crud_project_id" ] || ! is_valid_id "$crud_project_id"; then
        echo -e "${RED}Could not extract project ID from create output${NC}"
        return 1
    fi
    CREATED_PROJECTS+=("$crud_project_id")
    echo -e "${GREEN}Created project ID: $crud_project_id${NC}\n"

    # LIST
    echo -e "${GREEN}[LIST]${NC} Listing projects..."
    LIST_OUTPUT=$($ACLOUD_CMD management project list 2>&1)
    if [ $? -ne 0 ]; then
        echo -e "${RED}LIST failed:${NC}"
        echo "$LIST_OUTPUT"
        return 1
    fi
    echo "$LIST_OUTPUT" | head -15
    echo ""

    # GET
    echo -e "${GREEN}[GET]${NC} Getting project details..."
    GET_OUTPUT=$($ACLOUD_CMD management project get "$crud_project_id" 2>&1)
    if [ $? -ne 0 ]; then
        echo -e "${RED}GET failed:${NC}"
        echo "$GET_OUTPUT"
        return 1
    fi
    echo "$GET_OUTPUT"
    echo ""

    # UPDATE
    echo -e "${GREEN}[UPDATE]${NC} Updating project..."
    UPDATE_OUTPUT=$($ACLOUD_CMD management project update "$crud_project_id" \
        --description "Updated E2E test project" \
        --tags "e2e-test,updated" 2>&1)
    if [ $? -ne 0 ]; then
        echo -e "${RED}UPDATE failed:${NC}"
        echo "$UPDATE_OUTPUT"
        return 1
    fi
    echo "$UPDATE_OUTPUT"
    echo ""

    # DELETE inline — cleanup trap is the safety net
    echo -e "${GREEN}[DELETE]${NC} Deleting project..."
    DELETE_OUTPUT=$($ACLOUD_CMD management project delete "$crud_project_id" --yes 2>&1)
    if [ $? -eq 0 ]; then
        echo "$DELETE_OUTPUT"
        CREATED_PROJECTS=("${CREATED_PROJECTS[@]/$crud_project_id}")
    else
        echo -e "${YELLOW}DELETE may need a retry — cleanup trap will handle it:${NC}"
        echo "$DELETE_OUTPUT"
    fi
    echo ""

    echo -e "${GREEN}✓ Project CRUD test completed successfully!${NC}\n"
}

# Run tests
echo -e "${BLUE}Starting Management Resources E2E Tests...${NC}\n"

test_project               || FAILURES=$((FAILURES + 1))
test_project_output_formats || FAILURES=$((FAILURES + 1))

# Test summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Test prefix: $RESOURCE_PREFIX"
echo ""

if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}=== Management E2E: all checks passed ===${NC}"
    exit 0
else
    echo -e "${RED}=== Management E2E: $FAILURES check(s) failed ===${NC}"
    exit 1
fi
