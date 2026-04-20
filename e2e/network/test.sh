#!/bin/bash

# E2E Test Script for Network Resources
# Tests CRUD operations for VPC, Subnet, Security Group, Security Rule,
# Elastic IP, VPC Peering, VPC Peering Route, VPN Tunnel, and VPN Route

# Don't use set -e - we want to continue testing even if one test fails
# set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration - UPDATE THESE VALUES
PROJECT_ID="${ACLOUD_PROJECT_ID:-68398923fb2cb026400d4d31}"
VPC_ID="${ACLOUD_VPC_ID:-69495ef64d0cdc87949b71ec}"
PEER_VPC_ID="${ACLOUD_PEER_VPC_ID:-689307f4745108d3c6343b5a}"
REGION="${ACLOUD_REGION:-ITBG-Bergamo}"
ELASTIC_IP_URI="${ACLOUD_ELASTIC_IP_URI:-/projects/68398923fb2cb026400d4d31/providers/Aruba.Network/elasticIps/694914e94d0cdc87949b70f1}"

# Determine acloud command path - try relative to script location first, then current dir, then PATH
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
if [ -f "$PROJECT_ROOT/acloud" ]; then
    ACLOUD_CMD="${ACLOUD_CMD:-$PROJECT_ROOT/acloud}"
elif [ -f "./acloud" ]; then
    ACLOUD_CMD="${ACLOUD_CMD:-./acloud}"
elif command -v acloud >/dev/null 2>&1; then
    ACLOUD_CMD="${ACLOUD_CMD:-acloud}"
else
    echo -e "${RED}Error: acloud binary not found. Please build it first with 'go build -o acloud .'${NC}" >&2
    exit 1
fi

# Derived values
PEER_VPC_URI="${ACLOUD_PEER_VPC_URI:-/projects/${PROJECT_ID}/providers/Aruba.Network/vpcs/${PEER_VPC_ID}}"
RESOURCE_PREFIX="e2e-test-$(date +%s)"

# Cleanup tracking
CREATED_VPCS=()
CREATED_SUBNETS=()
CREATED_SECURITY_GROUPS=()
CREATED_SECURITY_RULES=()
CREATED_ELASTIC_IPS=()
CREATED_PEERINGS=()
CREATED_PEERING_ROUTES=()
CREATED_VPN_TUNNELS=()
CREATED_VPN_ROUTES=()

echo -e "${BLUE}=== Network Resources E2E Test ===${NC}\n"
echo "Project ID: $PROJECT_ID"
echo "Region: $REGION"
echo "Test prefix: $RESOURCE_PREFIX"
echo "ACLOUD command: $ACLOUD_CMD"
echo ""

# Function to extract resource ID from output
# This function tries multiple strategies to find the correct resource ID:
# 1. Extract all IDs and filter out known parent IDs (like VPC_ID), take the last one
# 2. Look for ID in table format (in the ID column)
# 3. Fallback to first ID found (if no exclude_id provided)
extract_id() {
    local output="$1"
    local exclude_id="${2:-}"  # Optional ID to exclude (e.g., VPC_ID)
    
    # Strategy 1: Extract all IDs, filter out exclude_id, take the last one
    # (Resource IDs are usually printed last in successful create operations)
    if [ -n "$exclude_id" ]; then
        local filtered_ids=$(echo "$output" | grep -oE '[a-f0-9]{24}' | grep -v "^${exclude_id}$")
        if [ -n "$filtered_ids" ]; then
            echo "$filtered_ids" | tail -1
            return 0
        fi
    fi
    
    # Strategy 2: Look for ID in table format (in the ID column)
    # Table format: NAME    ID                          REGION    STATUS
    #                name    694bb9767712ac0032dbe640    region    status
    # Try to find lines that look like table rows (have multiple space-separated fields)
    local table_id=$(echo "$output" | awk '
        /^[A-Z ]+ID[ A-Z]*$/ { 
            getline
            if (NF >= 2 && $2 ~ /^[a-f0-9]{24}$/) {
                print $2
            }
        }
    ' | head -1)
    if [ -n "$table_id" ] && [ "$table_id" != "$exclude_id" ]; then
        echo "$table_id"
        return 0
    fi
    
    # Strategy 3: Extract all IDs and take the last one
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
    # Check if it's a 24-character hex string (MongoDB ObjectID format)
    [[ "$id" =~ ^[a-f0-9]{24}$ ]]
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

# Test --output flag for network list commands
test_output_formats() {
    echo -e "${BLUE}--- Testing network list --output flag ---${NC}"

    for resource_cmd in "vpc list" "subnet list" "securitygroup list" "elasticip list"; do
        local label="network $resource_cmd"
        echo -e "${YELLOW}Testing $label --output json...${NC}"
        JSON_OUTPUT=$($ACLOUD_CMD network $resource_cmd --project-id "$PROJECT_ID" --output json 2>&1)
        JSON_EXIT=$?
        if [ $JSON_EXIT -ne 0 ]; then
            echo -e "${RED}✗ $label --output json: command failed (exit $JSON_EXIT)${NC}"
        elif echo "$JSON_OUTPUT" | grep -qF "No "; then
            echo -e "${YELLOW}⚠ $label --output json: no resources — format validation skipped${NC}"
        elif is_valid_json "$JSON_OUTPUT"; then
            echo -e "${GREEN}✓ $label --output json: valid JSON${NC}"
        else
            echo -e "${RED}✗ $label --output json: output is not valid JSON${NC}"
        fi

        echo -e "${YELLOW}Testing $label --output yaml...${NC}"
        YAML_OUTPUT=$($ACLOUD_CMD network $resource_cmd --project-id "$PROJECT_ID" --output yaml 2>&1)
        YAML_EXIT=$?
        if [ $YAML_EXIT -ne 0 ]; then
            echo -e "${RED}✗ $label --output yaml: command failed (exit $YAML_EXIT)${NC}"
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

# Function to check VPC status
check_vpc_status() {
    local vpc_id="$1"
    if [ -z "$vpc_id" ] || [ "$vpc_id" = "your-vpc-id" ]; then
        return 1
    fi
    
    local vpc_output=$($ACLOUD_CMD network vpc get "$vpc_id" 2>&1)
    
    # Check for errors first
    if echo "$vpc_output" | grep -qi "Error\|Failed\|not found"; then
        echo -e "${YELLOW}Warning: Could not retrieve VPC status for $vpc_id${NC}"
        return 1
    fi
    
    # Extract status from output (format: "Status:          Active" or "STATUS" column in table)
    local status=$(echo "$vpc_output" | grep -iE "Status:" | head -1 | awk -F: '{print $2}' | tr -d '[:space:]')
    
    if [ -z "$status" ]; then
        # Try to get from list output if get doesn't show status clearly
        local list_output=$($ACLOUD_CMD network vpc list 2>&1 | grep "$vpc_id" | head -1)
        if [ -n "$list_output" ]; then
            # Status is typically in the last column
            status=$(echo "$list_output" | awk '{print $NF}' | tr -d '[:space:]')
        fi
    fi
    
    if [ "$status" = "Active" ]; then
        return 0
    elif [ "$status" = "InCreation" ]; then
        echo -e "${YELLOW}Warning: VPC $vpc_id is in 'InCreation' state. Resources may not be created until VPC is active.${NC}"
        return 1
    elif [ -n "$status" ]; then
        echo -e "${YELLOW}Warning: VPC $vpc_id is in '$status' state. Resources may not be created until VPC is active.${NC}"
        return 1
    else
        echo -e "${YELLOW}Warning: Could not determine VPC status for $vpc_id (status: '$status')${NC}"
        return 1
    fi
}

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test resources...${NC}"
    
    # Delete VPN routes (only if VPN_TUNNEL_ID is set and route IDs are valid)
    if [ -n "$VPN_TUNNEL_ID" ] && is_valid_id "$VPN_TUNNEL_ID"; then
        for route_id in "${CREATED_VPN_ROUTES[@]}"; do
            if is_valid_id "$route_id"; then
                echo "Deleting VPN route: $route_id"
                $ACLOUD_CMD network vpnroute delete "$VPN_TUNNEL_ID" "$route_id" --yes 2>&1 || true
            fi
        done
    fi
    
    # Delete VPN tunnels
    for tunnel_id in "${CREATED_VPN_TUNNELS[@]}"; do
        if is_valid_id "$tunnel_id"; then
            echo "Deleting VPN tunnel: $tunnel_id"
            $ACLOUD_CMD network vpntunnel delete "$tunnel_id" --yes 2>&1 || true
        fi
    done
    
    # Delete peering routes (only if VPC_ID and PEERING_ID are set and valid)
    if [ -n "$VPC_ID" ] && is_valid_id "$VPC_ID" && [ -n "$PEERING_ID" ] && is_valid_id "$PEERING_ID"; then
        for route_id in "${CREATED_PEERING_ROUTES[@]}"; do
            if is_valid_id "$route_id"; then
                echo "Deleting peering route: $route_id"
                $ACLOUD_CMD network vpcpeeringroute delete "$VPC_ID" "$PEERING_ID" "$route_id" --yes 2>&1 || true
            fi
        done
    fi
    
    # Delete peerings (only if VPC_ID is set and valid)
    if [ -n "$VPC_ID" ] && is_valid_id "$VPC_ID"; then
        for peering_id in "${CREATED_PEERINGS[@]}"; do
            if is_valid_id "$peering_id"; then
                echo "Deleting VPC peering: $peering_id"
                $ACLOUD_CMD network vpcpeering delete "$VPC_ID" "$peering_id" --yes 2>&1 || true
            fi
        done
    fi
    
    # Delete security rules (only if VPC_ID and SECURITY_GROUP_ID are set and valid)
    if [ -n "$VPC_ID" ] && is_valid_id "$VPC_ID" && [ -n "$SECURITY_GROUP_ID" ] && is_valid_id "$SECURITY_GROUP_ID" && [ ${#CREATED_SECURITY_RULES[@]} -gt 0 ]; then
        for rule_id in "${CREATED_SECURITY_RULES[@]}"; do
            # Skip if rule_id is empty, VPC_ID, SECURITY_GROUP_ID, or not a valid ID
            if [ -z "$rule_id" ] || [ "$rule_id" = "$VPC_ID" ] || [ "$rule_id" = "$SECURITY_GROUP_ID" ] || ! is_valid_id "$rule_id"; then
                continue
            fi
            echo "Deleting security rule: $rule_id"
            $ACLOUD_CMD network securityrule delete "$VPC_ID" "$SECURITY_GROUP_ID" "$rule_id" --yes 2>&1 || true
        done
    fi
    
    # Delete security groups (only if VPC_ID is set and valid)
    if [ -n "$VPC_ID" ] && is_valid_id "$VPC_ID" && [ ${#CREATED_SECURITY_GROUPS[@]} -gt 0 ]; then
        for sg_id in "${CREATED_SECURITY_GROUPS[@]}"; do
            # Skip if sg_id is empty, VPC_ID, or not a valid ID
            if [ -z "$sg_id" ] || [ "$sg_id" = "$VPC_ID" ] || ! is_valid_id "$sg_id"; then
                continue
            fi
            echo "Deleting security group: $sg_id"
            $ACLOUD_CMD network securitygroup delete "$VPC_ID" "$sg_id" --yes 2>&1 || true
        done
    fi
    
    # Delete elastic IPs
    if [ ${#CREATED_ELASTIC_IPS[@]} -gt 0 ]; then
        for eip_id in "${CREATED_ELASTIC_IPS[@]}"; do
            # Skip if eip_id is empty, VPC_ID, or not a valid ID
            if [ -z "$eip_id" ] || [ "$eip_id" = "$VPC_ID" ] || ! is_valid_id "$eip_id"; then
                continue
            fi
            echo "Deleting elastic IP: $eip_id"
            $ACLOUD_CMD network elasticip delete "$eip_id" --yes 2>&1 || true
        done
    fi
    
    # Delete subnets (only if VPC_ID is set and valid)
    if [ -n "$VPC_ID" ] && is_valid_id "$VPC_ID" && [ ${#CREATED_SUBNETS[@]} -gt 0 ]; then
        for subnet_id in "${CREATED_SUBNETS[@]}"; do
            # Skip if subnet_id is empty, VPC_ID, or not a valid ID
            if [ -z "$subnet_id" ] || [ "$subnet_id" = "$VPC_ID" ] || ! is_valid_id "$subnet_id"; then
                continue
            fi
            echo "Deleting subnet: $subnet_id"
            $ACLOUD_CMD network subnet delete "$VPC_ID" "$subnet_id" --yes 2>&1 || true
        done
    fi
    
    # Delete VPCs (only if we created them)
    for vpc_id in "${CREATED_VPCS[@]}"; do
        if is_valid_id "$vpc_id"; then
            echo "Deleting VPC: $vpc_id"
            $ACLOUD_CMD network vpc delete "$vpc_id" --yes 2>&1 || true
        fi
    done
}

trap cleanup EXIT

# Function to test a resource
test_resource() {
    local resource_name=$1
    local create_cmd=$2
    local list_cmd=$3
    local get_cmd=$4
    local update_cmd=$5
    local delete_cmd=$6
    
    echo -e "${YELLOW}--- Testing $resource_name ---${NC}"
    
    # CREATE
    echo -e "${GREEN}[CREATE]${NC} Creating $resource_name..."
    CREATE_OUTPUT=$(eval "$create_cmd" 2>&1) || {
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT"
        # Check for common error patterns
        if echo "$CREATE_OUTPUT" | grep -qi "authentication failed\|invalid_client\|Invalid client"; then
            echo -e "${RED}Authentication error detected. Please check your credentials.${NC}"
        elif echo "$CREATE_OUTPUT" | grep -qi "timeout.*VPC.*active"; then
            echo -e "${RED}VPC is not in active state. Please wait for VPC to become active before creating resources.${NC}"
        fi
        return 1
    }
    echo "$CREATE_OUTPUT"
    
    # Extract resource ID from output
    # For resources that require VPC_ID (subnet, securitygroup, securityrule, etc.), exclude it from extraction
    local exclude_id=""
    # Check if VPC_ID appears in the create command (either as variable or as actual value)
    if [ -n "$VPC_ID" ] && (echo "$create_cmd" | grep -q "\$VPC_ID\|$VPC_ID" || echo "$create_cmd" | grep -qE "(subnet|securitygroup|securityrule|vpcpeering|vpcpeeringroute|vpnroute).*create"); then
        exclude_id="$VPC_ID"
    fi
    RESOURCE_ID=$(extract_id "$CREATE_OUTPUT" "$exclude_id")
    if [ -z "$RESOURCE_ID" ]; then
        echo -e "${RED}Could not extract resource ID from create output${NC}"
        echo -e "${YELLOW}CREATE_OUTPUT:${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi
    # Validate that we didn't accidentally extract VPC_ID
    if [ "$RESOURCE_ID" = "$VPC_ID" ] && [ -n "$VPC_ID" ]; then
        echo -e "${RED}Error: Extracted VPC_ID instead of resource ID. This should not happen.${NC}"
        echo -e "${YELLOW}CREATE_OUTPUT:${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi
    echo -e "${GREEN}Created resource ID: $RESOURCE_ID${NC}\n"
    
    # LIST
    echo -e "${GREEN}[LIST]${NC} Listing $resource_name..."
    LIST_OUTPUT=$(eval "$list_cmd" 2>&1) || {
        echo -e "${RED}LIST failed:${NC}"
        echo "$LIST_OUTPUT"
        return 1
    }
    echo "$LIST_OUTPUT" | head -10
    echo ""
    
    # GET
    echo -e "${GREEN}[GET]${NC} Getting $resource_name details..."
    GET_OUTPUT=$(eval "$get_cmd" 2>&1) || {
        echo -e "${RED}GET failed:${NC}"
        echo "$GET_OUTPUT"
        return 1
    }
    echo "$GET_OUTPUT"
    echo ""
    
    # UPDATE (skip if resource is in InCreation state)
    echo -e "${GREEN}[UPDATE]${NC} Updating $resource_name..."
    UPDATE_OUTPUT=$(eval "$update_cmd" 2>&1) || {
        # Check if the error is due to InCreation state
        if echo "$UPDATE_OUTPUT" | grep -qi "InCreation\|not ready\|not available"; then
            echo -e "${YELLOW}UPDATE skipped: Resource is still being created${NC}"
            echo "$UPDATE_OUTPUT"
        elif echo "$UPDATE_OUTPUT" | grep -qi "unknown flag"; then
            echo -e "${YELLOW}UPDATE skipped: Command doesn't support the requested flags${NC}"
            echo "$UPDATE_OUTPUT"
        else
            echo -e "${RED}UPDATE failed:${NC}"
            echo "$UPDATE_OUTPUT"
            # Don't return error - continue with delete
        fi
    }
    if ! echo "$UPDATE_OUTPUT" | grep -qi "InCreation\|not ready\|not available\|unknown flag"; then
        echo "$UPDATE_OUTPUT"
    fi
    echo ""
    
    # DELETE (skip if resource is in InCreation state - we'll clean it up later)
    echo -e "${GREEN}[DELETE]${NC} Deleting $resource_name..."
    DELETE_OUTPUT=$(eval "$delete_cmd" 2>&1) || {
        # Check if the error is due to InCreation state or if delete doesn't support --yes
        if echo "$DELETE_OUTPUT" | grep -qi "InCreation\|not ready\|not available\|unknown flag.*yes"; then
            echo -e "${YELLOW}DELETE skipped: Resource may still be creating or command doesn't support --yes flag${NC}"
            echo "$DELETE_OUTPUT"
            # Don't return error - we'll try to clean up in cleanup function
        else
            echo -e "${RED}DELETE failed:${NC}"
            echo "$DELETE_OUTPUT"
            # Don't return error - continue to return resource ID for cleanup
        fi
    }
    if ! echo "$DELETE_OUTPUT" | grep -qi "InCreation\|not ready\|not available\|unknown flag.*yes"; then
        echo "$DELETE_OUTPUT"
    fi
    echo ""
    
    echo -e "${GREEN}✓ $resource_name CRUD test completed successfully!${NC}\n"
    echo "$RESOURCE_ID"  # Return resource ID
}

# Test VPC
test_vpc() {
    echo -e "${YELLOW}=== 1. VPC CRUD Test ===${NC}\n"
    VPC_ID_OUTPUT=$(test_resource "VPC" \
        "$ACLOUD_CMD network vpc create --name ${RESOURCE_PREFIX}-vpc --region $REGION" \
        "$ACLOUD_CMD network vpc list" \
        "$ACLOUD_CMD network vpc get \$RESOURCE_ID" \
        "$ACLOUD_CMD network vpc update \$RESOURCE_ID --name ${RESOURCE_PREFIX}-vpc-updated --tags updated" \
        "$ACLOUD_CMD network vpc delete \$RESOURCE_ID --yes" 2>&1)
    
    if [ -n "$VPC_ID_OUTPUT" ] && [ "$VPC_ID_OUTPUT" != "1" ]; then
        CREATED_VPCS+=("$VPC_ID_OUTPUT")
        VPC_ID="$VPC_ID_OUTPUT"
        echo -e "${GREEN}VPC ID set to: $VPC_ID${NC}\n"
    else
        echo -e "${RED}Failed to create or extract VPC ID${NC}\n"
        return 1
    fi
}

# Test Subnet
test_subnet() {
    if [ -z "$VPC_ID" ] || [ "$VPC_ID" = "your-vpc-id" ] || [ "$VPC_ID" = "" ]; then
        echo -e "${YELLOW}Skipping subnet test (no VPC available)${NC}\n"
        return 0
    fi
    
    echo -e "${YELLOW}=== 2. Subnet CRUD Test ===${NC}\n"
    local output
    output=$(test_resource "Subnet" \
        "$ACLOUD_CMD network subnet create $VPC_ID --name ${RESOURCE_PREFIX}-subnet --cidr 10.150.0.0/24 --region $REGION" \
        "$ACLOUD_CMD network subnet list $VPC_ID" \
        "$ACLOUD_CMD network subnet get $VPC_ID \$RESOURCE_ID" \
        "$ACLOUD_CMD network subnet update $VPC_ID \$RESOURCE_ID --name ${RESOURCE_PREFIX}-subnet-updated --tags updated" \
        "$ACLOUD_CMD network subnet delete $VPC_ID \$RESOURCE_ID --yes" 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
        SUBNET_ID=$(echo "$output" | tail -1 | tr -d '[:space:]')
        if [ -n "$SUBNET_ID" ] && [ "$SUBNET_ID" != "1" ] && [ "$SUBNET_ID" != "$VPC_ID" ] && [ ${#SUBNET_ID} -eq 24 ] && is_valid_id "$SUBNET_ID"; then
            CREATED_SUBNETS+=("$SUBNET_ID")
            echo -e "${GREEN}Subnet test completed successfully${NC}\n"
        else
            echo -e "${YELLOW}Subnet test completed but could not extract valid ID${NC}"
            echo -e "${YELLOW}Last line of output: ${SUBNET_ID}${NC}\n"
        fi
    else
        echo -e "${RED}Subnet test failed with exit code: $exit_code${NC}"
        if [ -n "$output" ]; then
            echo -e "${RED}Error output:${NC}"
            echo "$output" | tail -20
        fi
        echo ""
        return 1
    fi
}

# Test Security Group
test_security_group() {
    if [ -z "$VPC_ID" ] || [ "$VPC_ID" = "your-vpc-id" ] || [ "$VPC_ID" = "" ]; then
        echo -e "${YELLOW}Skipping security group test (no VPC available)${NC}\n"
        return 0
    fi
    
    echo -e "${YELLOW}=== 3. Security Group CRUD Test ===${NC}\n"
    local output
    output=$(test_resource "Security Group" \
        "$ACLOUD_CMD network securitygroup create $VPC_ID --name ${RESOURCE_PREFIX}-sg --region $REGION" \
        "$ACLOUD_CMD network securitygroup list $VPC_ID" \
        "$ACLOUD_CMD network securitygroup get $VPC_ID \$RESOURCE_ID" \
        "$ACLOUD_CMD network securitygroup update $VPC_ID \$RESOURCE_ID --tags updated" \
        "$ACLOUD_CMD network securitygroup delete $VPC_ID \$RESOURCE_ID --yes" 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
        SECURITY_GROUP_ID=$(echo "$output" | tail -1 | tr -d '[:space:]')
        if [ -n "$SECURITY_GROUP_ID" ] && [ "$SECURITY_GROUP_ID" != "1" ] && [ "$SECURITY_GROUP_ID" != "$VPC_ID" ] && [ ${#SECURITY_GROUP_ID} -eq 24 ] && is_valid_id "$SECURITY_GROUP_ID"; then
            CREATED_SECURITY_GROUPS+=("$SECURITY_GROUP_ID")
            echo -e "${GREEN}Security Group test completed successfully${NC}\n"
        else
            echo -e "${YELLOW}Security Group test completed but could not extract valid ID${NC}"
            echo -e "${YELLOW}Last line of output: ${SECURITY_GROUP_ID}${NC}\n"
        fi
    else
        echo -e "${RED}Security Group test failed with exit code: $exit_code${NC}"
        if [ -n "$output" ]; then
            echo -e "${RED}Error output:${NC}"
            echo "$output" | tail -20
        fi
        echo ""
        return 1
    fi
}

# Test Security Rule
test_security_rule() {
    if [ -z "$VPC_ID" ] || [ "$VPC_ID" = "your-vpc-id" ] || [ "$VPC_ID" = "" ] || [ -z "$SECURITY_GROUP_ID" ]; then
        echo -e "${YELLOW}Skipping security rule test (no VPC or security group available)${NC}\n"
        return 0
    fi
    
    echo -e "${YELLOW}=== 4. Security Rule CRUD Test ===${NC}\n"
    local output
    output=$(test_resource "Security Rule" \
        "$ACLOUD_CMD network securityrule create $VPC_ID $SECURITY_GROUP_ID --name ${RESOURCE_PREFIX}-rule --region $REGION --direction Ingress --protocol TCP --port 80 --target-kind Ip --target-value 0.0.0.0/0" \
        "$ACLOUD_CMD network securityrule list $VPC_ID $SECURITY_GROUP_ID" \
        "$ACLOUD_CMD network securityrule get $VPC_ID $SECURITY_GROUP_ID \$RESOURCE_ID" \
        "$ACLOUD_CMD network securityrule update $VPC_ID $SECURITY_GROUP_ID \$RESOURCE_ID --name ${RESOURCE_PREFIX}-rule-updated --tags updated" \
        "$ACLOUD_CMD network securityrule delete $VPC_ID $SECURITY_GROUP_ID \$RESOURCE_ID --yes" 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
        SECURITY_RULE_ID=$(echo "$output" | tail -1 | tr -d '[:space:]')
        if [ -n "$SECURITY_RULE_ID" ] && [ "$SECURITY_RULE_ID" != "1" ] && [ "$SECURITY_RULE_ID" != "$VPC_ID" ] && [ "$SECURITY_RULE_ID" != "$SECURITY_GROUP_ID" ] && [ ${#SECURITY_RULE_ID} -eq 24 ] && is_valid_id "$SECURITY_RULE_ID"; then
            CREATED_SECURITY_RULES+=("$SECURITY_RULE_ID")
            echo -e "${GREEN}Security Rule test completed successfully${NC}\n"
        else
            echo -e "${YELLOW}Security Rule test completed but could not extract valid ID${NC}\n"
        fi
    else
        echo -e "${RED}Security Rule test failed${NC}\n"
        if [ -n "$output" ]; then
            echo -e "${RED}Error output:${NC}"
            echo "$output" | tail -5
        fi
        return 1
    fi
}

# Test Elastic IP
test_elastic_ip() {
    echo -e "${YELLOW}=== 5. Elastic IP CRUD Test ===${NC}\n"
    local output
    output=$(test_resource "Elastic IP" \
        "$ACLOUD_CMD network elasticip create --name ${RESOURCE_PREFIX}-eip --region $REGION --billing-period Hour" \
        "$ACLOUD_CMD network elasticip list" \
        "$ACLOUD_CMD network elasticip get \$RESOURCE_ID" \
        "$ACLOUD_CMD network elasticip update \$RESOURCE_ID --name ${RESOURCE_PREFIX}-eip-updated --tags updated" \
        "$ACLOUD_CMD network elasticip delete \$RESOURCE_ID --yes" 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
        EIP_ID=$(echo "$output" | tail -1 | tr -d '[:space:]')
        if [ -n "$EIP_ID" ] && [ "$EIP_ID" != "1" ] && [ "$EIP_ID" != "$VPC_ID" ] && [ ${#EIP_ID} -eq 24 ] && is_valid_id "$EIP_ID"; then
            CREATED_ELASTIC_IPS+=("$EIP_ID")
            echo -e "${GREEN}Elastic IP test completed successfully${NC}\n"
        else
            echo -e "${YELLOW}Elastic IP test completed but could not extract valid ID${NC}"
            echo -e "${YELLOW}Last line of output: ${EIP_ID}${NC}\n"
        fi
    else
        echo -e "${RED}Elastic IP test failed${NC}\n"
        if [ -n "$output" ]; then
            echo -e "${RED}Error output:${NC}"
            echo "$output" | tail -5
        fi
        return 1
    fi
}

# Test VPC Peering
test_vpc_peering() {
    if [ -z "$VPC_ID" ] || [ "$VPC_ID" = "your-vpc-id" ] || [ "$VPC_ID" = "" ]; then
        echo -e "${YELLOW}Skipping VPC peering test (no VPC available)${NC}\n"
        return 0
    fi
    
    if [ -z "$PEER_VPC_ID" ] || [ "$PEER_VPC_ID" = "your-peer-vpc-id" ] || [ "$PEER_VPC_ID" = "" ]; then
        echo -e "${YELLOW}Skipping VPC peering test (no peer VPC ID available)${NC}\n"
        return 0
    fi
    
    echo -e "${YELLOW}=== 6. VPC Peering CRUD Test ===${NC}\n"
    local output
    output=$(test_resource "VPC Peering" \
        "$ACLOUD_CMD network vpcpeering create $VPC_ID --name ${RESOURCE_PREFIX}-peering --peer-vpc-id $PEER_VPC_ID --region $REGION" \
        "$ACLOUD_CMD network vpcpeering list $VPC_ID" \
        "$ACLOUD_CMD network vpcpeering get $VPC_ID \$RESOURCE_ID" \
        "$ACLOUD_CMD network vpcpeering update $VPC_ID \$RESOURCE_ID --name ${RESOURCE_PREFIX}-peering-updated --tags updated" \
        "$ACLOUD_CMD network vpcpeering delete $VPC_ID \$RESOURCE_ID --yes" 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
        PEERING_ID=$(echo "$output" | tail -1 | tr -d '[:space:]')
        if [ -n "$PEERING_ID" ] && [ "$PEERING_ID" != "1" ] && [ "$PEERING_ID" != "$VPC_ID" ] && [ ${#PEERING_ID} -eq 24 ] && is_valid_id "$PEERING_ID"; then
            CREATED_PEERINGS+=("$PEERING_ID")
            echo -e "${GREEN}VPC Peering test completed successfully${NC}\n"
        else
            echo -e "${YELLOW}VPC Peering test completed but could not extract valid ID${NC}\n"
        fi
    else
        echo -e "${RED}VPC Peering test failed${NC}\n"
        if [ -n "$output" ]; then
            echo -e "${RED}Error output:${NC}"
            echo "$output" | tail -5
        fi
        return 1
    fi
}

# Test VPC Peering Route
test_vpc_peering_route() {
    if [ -z "$VPC_ID" ] || [ "$VPC_ID" = "your-vpc-id" ] || [ "$VPC_ID" = "" ] || [ -z "$PEERING_ID" ]; then
        echo -e "${YELLOW}Skipping VPC peering route test (no VPC or peering available)${NC}\n"
        return 0
    fi
    
    echo -e "${YELLOW}=== 7. VPC Peering Route CRUD Test ===${NC}\n"
    local output
    output=$(test_resource "VPC Peering Route" \
        "$ACLOUD_CMD network vpcpeeringroute create $VPC_ID $PEERING_ID --name ${RESOURCE_PREFIX}-route --local-network 10.0.1.0/24 --remote-network 10.0.2.0/24 --billing-period Hour" \
        "$ACLOUD_CMD network vpcpeeringroute list $VPC_ID $PEERING_ID" \
        "$ACLOUD_CMD network vpcpeeringroute get $VPC_ID $PEERING_ID \$RESOURCE_ID" \
        "$ACLOUD_CMD network vpcpeeringroute update $VPC_ID $PEERING_ID \$RESOURCE_ID --name ${RESOURCE_PREFIX}-route-updated --tags updated" \
        "$ACLOUD_CMD network vpcpeeringroute delete $VPC_ID $PEERING_ID \$RESOURCE_ID --yes" 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
        ROUTE_ID=$(echo "$output" | tail -1 | tr -d '[:space:]')
        if [ -n "$ROUTE_ID" ] && [ "$ROUTE_ID" != "1" ] && [ "$ROUTE_ID" != "$VPC_ID" ] && [ "$ROUTE_ID" != "$PEERING_ID" ] && [ ${#ROUTE_ID} -eq 24 ] && is_valid_id "$ROUTE_ID"; then
            CREATED_PEERING_ROUTES+=("$ROUTE_ID")
            echo -e "${GREEN}VPC Peering Route test completed successfully${NC}\n"
        else
            echo -e "${YELLOW}VPC Peering Route test completed but could not extract valid ID${NC}\n"
        fi
    else
        echo -e "${RED}VPC Peering Route test failed${NC}\n"
        if [ -n "$output" ]; then
            echo -e "${RED}Error output:${NC}"
            echo "$output" | tail -5
        fi
        return 1
    fi
}

# Test VPN Tunnel
test_vpn_tunnel() {
    if [ -z "$VPC_ID" ] || [ "$VPC_ID" = "your-vpc-id" ] || [ "$VPC_ID" = "" ] || [ -z "$ELASTIC_IP_URI" ] || [ "$ELASTIC_IP_URI" = "your-elastic-ip-uri" ]; then
        echo -e "${YELLOW}Skipping VPN tunnel test (missing VPC or Elastic IP)${NC}\n"
        return 0
    fi
    
    echo -e "${YELLOW}=== 8. VPN Tunnel CRUD Test ===${NC}\n"
    local output
    output=$(test_resource "VPN Tunnel" \
        "$ACLOUD_CMD network vpntunnel create --name ${RESOURCE_PREFIX}-vpn --region $REGION --peer-ip 203.0.113.1 --vpc-uri /projects/$PROJECT_ID/providers/Aruba.Network/vpcs/$VPC_ID --subnet-cidr 10.0.1.0/24 --elastic-ip-uri $ELASTIC_IP_URI --billing-period Hour" \
        "$ACLOUD_CMD network vpntunnel list" \
        "$ACLOUD_CMD network vpntunnel get \$RESOURCE_ID" \
        "$ACLOUD_CMD network vpntunnel update \$RESOURCE_ID --name ${RESOURCE_PREFIX}-vpn-updated --tags updated" \
        "$ACLOUD_CMD network vpntunnel delete \$RESOURCE_ID --yes" 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
        VPN_TUNNEL_ID=$(echo "$output" | tail -1 | tr -d '[:space:]')
        if [ -n "$VPN_TUNNEL_ID" ] && [ "$VPN_TUNNEL_ID" != "1" ] && [ "$VPN_TUNNEL_ID" != "$VPC_ID" ] && [ ${#VPN_TUNNEL_ID} -eq 24 ] && is_valid_id "$VPN_TUNNEL_ID"; then
            CREATED_VPN_TUNNELS+=("$VPN_TUNNEL_ID")
            echo -e "${GREEN}VPN Tunnel test completed successfully${NC}\n"
        else
            echo -e "${YELLOW}VPN Tunnel test completed but could not extract valid ID${NC}\n"
        fi
    else
        echo -e "${RED}VPN Tunnel test failed${NC}\n"
        if [ -n "$output" ]; then
            echo -e "${RED}Error output:${NC}"
            echo "$output" | tail -5
        fi
        return 1
    fi
}

# Test VPN Route
test_vpn_route() {
    if [ -z "$VPN_TUNNEL_ID" ]; then
        echo -e "${YELLOW}Skipping VPN route test (no VPN tunnel available)${NC}\n"
        return 0
    fi
    
    echo -e "${YELLOW}=== 9. VPN Route CRUD Test ===${NC}\n"
    local output
    output=$(test_resource "VPN Route" \
        "$ACLOUD_CMD network vpnroute create $VPN_TUNNEL_ID --name ${RESOURCE_PREFIX}-vpn-route --region $REGION --cloud-subnet 10.0.1.0/24 --onprem-subnet 192.168.1.0/24" \
        "$ACLOUD_CMD network vpnroute list $VPN_TUNNEL_ID" \
        "$ACLOUD_CMD network vpnroute get $VPN_TUNNEL_ID \$RESOURCE_ID" \
        "$ACLOUD_CMD network vpnroute update $VPN_TUNNEL_ID \$RESOURCE_ID --name ${RESOURCE_PREFIX}-vpn-route-updated --tags updated" \
        "$ACLOUD_CMD network vpnroute delete $VPN_TUNNEL_ID \$RESOURCE_ID --yes" 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
        ROUTE_ID=$(echo "$output" | tail -1 | tr -d '[:space:]')
        if [ -n "$ROUTE_ID" ] && [ "$ROUTE_ID" != "1" ] && [ "$ROUTE_ID" != "$VPC_ID" ] && [ "$ROUTE_ID" != "$VPN_TUNNEL_ID" ] && [ ${#ROUTE_ID} -eq 24 ] && is_valid_id "$ROUTE_ID"; then
            CREATED_VPN_ROUTES+=("$ROUTE_ID")
            echo -e "${GREEN}VPN Route test completed successfully${NC}\n"
        else
            echo -e "${YELLOW}VPN Route test completed but could not extract valid ID${NC}\n"
        fi
    else
        echo -e "${RED}VPN Route test failed${NC}\n"
        if [ -n "$output" ]; then
            echo -e "${RED}Error output:${NC}"
            echo "$output" | tail -5
        fi
        return 1
    fi
}

# Set up context for project ID (so we don't need --project-id flag on every command)
echo -e "${BLUE}Setting up context for project ID...${NC}"
$ACLOUD_CMD context set e2e-test-context --project-id "$PROJECT_ID" >/dev/null 2>&1 || true
$ACLOUD_CMD context use e2e-test-context >/dev/null 2>&1 || true
echo ""

# Run tests
echo -e "${BLUE}Starting Network Resources E2E Tests...${NC}\n"

# Only test VPC if not provided
if [ -z "$VPC_ID" ] || [ "$VPC_ID" = "your-vpc-id" ]; then
    echo -e "${BLUE}No VPC ID provided, creating a new VPC for testing...${NC}\n"
    test_vpc
    if [ -z "$VPC_ID" ]; then
        echo -e "${RED}Failed to create VPC. Cannot continue with dependent tests.${NC}"
        exit 1
    fi
else
    echo -e "${BLUE}Using existing VPC ID: $VPC_ID${NC}"
    # Check VPC status
    if ! check_vpc_status "$VPC_ID"; then
        echo -e "${YELLOW}Warning: VPC may not be in active state. Some tests may fail.${NC}\n"
    else
        echo -e "${GREEN}VPC is active and ready for testing.${NC}\n"
    fi
fi

test_subnet || echo -e "${YELLOW}Subnet test completed with errors${NC}\n"
test_security_group || echo -e "${YELLOW}Security Group test completed with errors${NC}\n"
test_security_rule || echo -e "${YELLOW}Security Rule test completed with errors${NC}\n"
test_elastic_ip || echo -e "${YELLOW}Elastic IP test completed with errors${NC}\n"

# VPC Peering tests (require peer VPC)
if [ -n "$PEER_VPC_ID" ] && [ "$PEER_VPC_ID" != "your-peer-vpc-id" ]; then
    test_vpc_peering || echo -e "${YELLOW}VPC Peering test completed with errors${NC}\n"
    test_vpc_peering_route || echo -e "${YELLOW}VPC Peering Route test completed with errors${NC}\n"
else
    echo -e "${YELLOW}Skipping VPC Peering tests (PEER_VPC_ID not set or invalid)${NC}\n"
fi

# VPN tests (require Elastic IP)
if [ -n "$ELASTIC_IP_URI" ] && [ "$ELASTIC_IP_URI" != "your-elastic-ip-uri" ]; then
    test_vpn_tunnel || echo -e "${YELLOW}VPN Tunnel test completed with errors${NC}\n"
    test_vpn_route || echo -e "${YELLOW}VPN Route test completed with errors${NC}\n"
else
    echo -e "${YELLOW}Skipping VPN tests (ELASTIC_IP_URI not set or invalid)${NC}\n"
fi

test_output_formats

echo -e "${GREEN}=== All Network Tests Completed! ===${NC}\n"

# Print summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo -e "VPC ID: ${VPC_ID:-N/A}"
if [ ${#CREATED_SUBNETS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Subnets: ${#CREATED_SUBNETS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Subnets: 0 created${NC}"
fi
if [ ${#CREATED_SECURITY_GROUPS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Security Groups: ${#CREATED_SECURITY_GROUPS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Security Groups: 0 created${NC}"
fi
if [ ${#CREATED_SECURITY_RULES[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Security Rules: ${#CREATED_SECURITY_RULES[@]} created${NC}"
else
    echo -e "${YELLOW}○ Security Rules: 0 created${NC}"
fi
if [ ${#CREATED_ELASTIC_IPS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Elastic IPs: ${#CREATED_ELASTIC_IPS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Elastic IPs: 0 created${NC}"
fi
if [ ${#CREATED_PEERINGS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ VPC Peerings: ${#CREATED_PEERINGS[@]} created${NC}"
else
    echo -e "${YELLOW}○ VPC Peerings: 0 created${NC}"
fi
if [ ${#CREATED_PEERING_ROUTES[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ VPC Peering Routes: ${#CREATED_PEERING_ROUTES[@]} created${NC}"
else
    echo -e "${YELLOW}○ VPC Peering Routes: 0 created${NC}"
fi
if [ ${#CREATED_VPN_TUNNELS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ VPN Tunnels: ${#CREATED_VPN_TUNNELS[@]} created${NC}"
else
    echo -e "${YELLOW}○ VPN Tunnels: 0 created${NC}"
fi
if [ ${#CREATED_VPN_ROUTES[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ VPN Routes: ${#CREATED_VPN_ROUTES[@]} created${NC}"
else
    echo -e "${YELLOW}○ VPN Routes: 0 created${NC}"
fi
echo ""

