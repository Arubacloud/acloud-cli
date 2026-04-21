#!/bin/bash

# E2E Test Script for Compute Resources
# Tests CRUD operations for Cloud Servers and Key Pairs

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# Cleanup tracking
CREATED_SERVERS=()
CREATED_KEYPAIRS=()

print_banner "Compute"

# Test --output flag for compute list commands
test_output_formats() {
    echo -e "${BLUE}--- Testing compute list --output flag ---${NC}"

    for resource in "cloudserver" "keypair"; do
        if [ "$resource" = "cloudserver" ]; then
            CMD_PREFIX="$ACLOUD_CMD compute cloudserver list --project-id \"$PROJECT_ID\""
        else
            CMD_PREFIX="$ACLOUD_CMD compute keypair list --project-id \"$PROJECT_ID\""
        fi

        for fmt in table std standard table-json std-json table-yaml std-yaml json yaml; do
            OUT=$(eval "$CMD_PREFIX --output $fmt" 2>&1)
            EXIT=$?
            if [ $EXIT -eq 0 ]; then
                echo -e "${GREEN}✓ $resource --output $fmt: command succeeded${NC}"
            else
                echo -e "${RED}✗ $resource --output $fmt: command failed (exit $EXIT)${NC}"
                echo "$OUT"
            fi
        done

        echo -e "${YELLOW}Testing $resource --output table-json (flat JSON shape)...${NC}"
        if [ "$resource" = "cloudserver" ]; then
            TABLE_JSON_OUT=$($ACLOUD_CMD compute cloudserver list --project-id "$PROJECT_ID" --output table-json 2>&1)
        else
            TABLE_JSON_OUT=$($ACLOUD_CMD compute keypair list --project-id "$PROJECT_ID" --output table-json 2>&1)
        fi
        if echo "$TABLE_JSON_OUT" | grep -qF "No "; then
            echo -e "${YELLOW}⚠ $resource --output table-json: no resources — format validation skipped${NC}"
        elif is_valid_json "$TABLE_JSON_OUT"; then
            echo -e "${GREEN}✓ $resource --output table-json: valid JSON (flat shape)${NC}"
        else
            echo -e "${RED}✗ $resource --output table-json: not valid JSON${NC}"
            echo "$TABLE_JSON_OUT"
        fi

        echo -e "${YELLOW}Testing $resource --output json (full SDK response)...${NC}"
        if [ "$resource" = "cloudserver" ]; then
            JSON_OUT=$($ACLOUD_CMD compute cloudserver list --project-id "$PROJECT_ID" --output json 2>&1)
        else
            JSON_OUT=$($ACLOUD_CMD compute keypair list --project-id "$PROJECT_ID" --output json 2>&1)
        fi
        if echo "$JSON_OUT" | grep -qF "No "; then
            echo -e "${YELLOW}⚠ $resource --output json: no resources — format validation skipped${NC}"
        elif is_valid_json "$JSON_OUT"; then
            echo -e "${GREEN}✓ $resource --output json: valid JSON${NC}"
            if echo "$JSON_OUT" | grep -q '"values"'; then
                echo -e "${GREEN}✓ $resource --output json: 'values' key present (full SDK response)${NC}"
            else
                echo -e "${YELLOW}⚠ $resource --output json: 'values' key not found${NC}"
            fi
        else
            echo -e "${RED}✗ $resource --output json: not valid JSON${NC}"
            echo "$JSON_OUT"
        fi
    done
    echo ""
}

# Function to test Cloud Server operations
test_cloudserver() {
    local server_name="${RESOURCE_PREFIX}-server"
    local flavor="${ACLOUD_FLAVOR:-small}"
    local image="${ACLOUD_IMAGE:-your-image-id}"
    
    echo -e "${BLUE}--- Testing Cloud Server Operations ---${NC}"
    
    # Create
    echo -e "${YELLOW}Creating cloud server...${NC}"
    CREATE_OUTPUT=$($ACLOUD_CMD compute cloudserver create \
        --project-id "$PROJECT_ID" \
        --name "$server_name" \
        --region "$REGION" \
        --flavor "$flavor" \
        --image "$image" 2>&1)
    CREATE_EXIT=$?
    
    if [ $CREATE_EXIT -eq 0 ]; then
        SERVER_ID=$(extract_id "$CREATE_OUTPUT")
        if [ -n "$SERVER_ID" ]; then
            CREATED_SERVERS+=("$SERVER_ID")
            echo -e "${GREEN}✓ Cloud server created: $SERVER_ID${NC}"
        else
            echo -e "${YELLOW}⚠ Cloud server creation may have succeeded but ID not found${NC}"
        fi
    else
        echo -e "${RED}✗ Failed to create cloud server${NC}"
        echo "$CREATE_OUTPUT"
    fi
    
    # List
    echo -e "${YELLOW}Listing cloud servers...${NC}"
    LIST_OUTPUT=$($ACLOUD_CMD compute cloudserver list --project-id "$PROJECT_ID" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Cloud server list successful${NC}"
    else
        echo -e "${RED}✗ Failed to list cloud servers${NC}"
        echo "$LIST_OUTPUT"
    fi
    
    # Get (if we have an ID)
    if [ -n "$SERVER_ID" ]; then
        echo -e "${YELLOW}Getting cloud server details...${NC}"
        GET_OUTPUT=$($ACLOUD_CMD compute cloudserver get "$SERVER_ID" --project-id "$PROJECT_ID" 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Cloud server get successful${NC}"
        else
            echo -e "${RED}✗ Failed to get cloud server${NC}"
            echo "$GET_OUTPUT"
        fi
        
        # Update
        echo -e "${YELLOW}Updating cloud server...${NC}"
        UPDATE_OUTPUT=$($ACLOUD_CMD compute cloudserver update "$SERVER_ID" \
            --project-id "$PROJECT_ID" \
            --name "${server_name}-updated" 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Cloud server update successful${NC}"
        else
            echo -e "${RED}✗ Failed to update cloud server${NC}"
            echo "$UPDATE_OUTPUT"
        fi
        
        # Power operations (if server is in a state that allows it)
        echo -e "${YELLOW}Testing power-off...${NC}"
        POWER_OFF_OUTPUT=$($ACLOUD_CMD compute cloudserver power-off "$SERVER_ID" \
            --project-id "$PROJECT_ID" 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Cloud server power-off successful${NC}"
            # Wait a bit before powering on
            sleep 2
        else
            echo -e "${YELLOW}⚠ Power-off failed (may be expected if server is already off)${NC}"
            echo "$POWER_OFF_OUTPUT"
        fi
        
        echo -e "${YELLOW}Testing power-on...${NC}"
        POWER_ON_OUTPUT=$($ACLOUD_CMD compute cloudserver power-on "$SERVER_ID" \
            --project-id "$PROJECT_ID" 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Cloud server power-on successful${NC}"
        else
            echo -e "${YELLOW}⚠ Power-on failed (may be expected if server is already on)${NC}"
            echo "$POWER_ON_OUTPUT"
        fi
        
        # Set password (optional test)
        echo -e "${YELLOW}Testing set-password...${NC}"
        SET_PASSWORD_OUTPUT=$($ACLOUD_CMD compute cloudserver set-password "$SERVER_ID" \
            --project-id "$PROJECT_ID" \
            --password "TestPassword123!" 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Cloud server set-password successful${NC}"
        else
            echo -e "${YELLOW}⚠ Set-password failed (may not be supported for all server types)${NC}"
            echo "$SET_PASSWORD_OUTPUT"
        fi
        
        # Connect (requires Elastic IP)
        echo -e "${YELLOW}Testing connect...${NC}"
        CONNECT_OUTPUT=$($ACLOUD_CMD compute cloudserver connect "$SERVER_ID" \
            --project-id "$PROJECT_ID" \
            --user ubuntu 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Cloud server connect successful${NC}"
            echo "$CONNECT_OUTPUT"
        else
            echo -e "${YELLOW}⚠ Connect failed (server may not have an Elastic IP)${NC}"
            echo "$CONNECT_OUTPUT"
        fi
    fi
    
    echo ""
}

# Function to test Key Pair operations
test_keypair() {
    local keypair_name="${RESOURCE_PREFIX}-keypair"
    local public_key="${ACLOUD_PUBLIC_KEY:-ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com}"
    
    echo -e "${BLUE}--- Testing Key Pair Operations ---${NC}"
    
    # Create
    echo -e "${YELLOW}Creating keypair...${NC}"
    CREATE_OUTPUT=$($ACLOUD_CMD compute keypair create \
        --project-id "$PROJECT_ID" \
        --name "$keypair_name" \
        --public-key "$public_key" 2>&1)
    CREATE_EXIT=$?
    
    if [ $CREATE_EXIT -eq 0 ]; then
        echo -e "${GREEN}✓ Keypair created: $keypair_name${NC}"
        CREATED_KEYPAIRS+=("$keypair_name")
    else
        echo -e "${RED}✗ Failed to create keypair${NC}"
        echo "$CREATE_OUTPUT"
    fi
    
    # List
    echo -e "${YELLOW}Listing keypairs...${NC}"
    LIST_OUTPUT=$($ACLOUD_CMD compute keypair list --project-id "$PROJECT_ID" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Keypair list successful${NC}"
    else
        echo -e "${RED}✗ Failed to list keypairs${NC}"
        echo "$LIST_OUTPUT"
    fi
    
    # Get
    echo -e "${YELLOW}Getting keypair details...${NC}"
    GET_OUTPUT=$($ACLOUD_CMD compute keypair get "$keypair_name" --project-id "$PROJECT_ID" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Keypair get successful${NC}"
    else
        echo -e "${RED}✗ Failed to get keypair${NC}"
        echo "$GET_OUTPUT"
    fi
    
    # Update
    echo -e "${YELLOW}Updating keypair...${NC}"
    UPDATE_OUTPUT=$($ACLOUD_CMD compute keypair update "$keypair_name" \
        --project-id "$PROJECT_ID" \
        --public-key "${public_key}-updated" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Keypair update successful${NC}"
    else
        echo -e "${RED}✗ Failed to update keypair${NC}"
        echo "$UPDATE_OUTPUT"
    fi
    
    echo ""
}

# Cleanup function
cleanup() {
    echo -e "${BLUE}--- Cleanup ---${NC}"
    
    # Delete keypairs
    for keypair in "${CREATED_KEYPAIRS[@]}"; do
        echo -e "${YELLOW}Deleting keypair: $keypair${NC}"
        $ACLOUD_CMD compute keypair delete "$keypair" --project-id "$PROJECT_ID" --yes 2>&1 >/dev/null
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Keypair deleted: $keypair${NC}"
        else
            echo -e "${RED}✗ Failed to delete keypair: $keypair${NC}"
        fi
    done
    
    # Delete cloud servers
    for server_id in "${CREATED_SERVERS[@]}"; do
        echo -e "${YELLOW}Deleting cloud server: $server_id${NC}"
        $ACLOUD_CMD compute cloudserver delete "$server_id" --project-id "$PROJECT_ID" --yes 2>&1 >/dev/null
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Cloud server deleted: $server_id${NC}"
        else
            echo -e "${RED}✗ Failed to delete cloud server: $server_id${NC}"
        fi
    done
    
    echo ""
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Run tests
test_cloudserver
test_keypair
test_output_formats

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Created cloud servers: ${#CREATED_SERVERS[@]}"
echo "Created keypairs: ${#CREATED_KEYPAIRS[@]}"
echo ""
echo -e "${GREEN}E2E tests completed!${NC}"

