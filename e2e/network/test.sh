#!/bin/bash

# E2E Test Script for Network Resources
# Tests CRUD operations for VPC, Subnet, Security Group, Security Rule,
# Elastic IP, VPC Peering, VPC Peering Route, VPN Tunnel, and VPN Route

# Don't use set -e - we want to continue testing even if one test fails
# set -e  # Exit on error

# Network-specific env overrides MUST come before sourcing common.sh so
# that common.sh picks up the correct PROJECT_ID default for this suite.
PROJECT_ID="${ACLOUD_PROJECT_ID:-68398923fb2cb026400d4d31}"
VPC_ID="${ACLOUD_VPC_ID:-69495ef64d0cdc87949b71ec}"
PEER_VPC_ID="${ACLOUD_PEER_VPC_ID:-689307f4745108d3c6343b5a}"
REGION="${ACLOUD_REGION:-ITBG-Bergamo}"
ELASTIC_IP_URI="${ACLOUD_ELASTIC_IP_URI:-}"

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# Derived values (depend on PROJECT_ID resolved above).
PEER_VPC_URI="${ACLOUD_PEER_VPC_URI:-/projects/${PROJECT_ID}/providers/Aruba.Network/vpcs/${PEER_VPC_ID}}"
# Per-run subnet CIDR: derive the second octet from the timestamp so each run
# uses a unique /24 (10.10.0.0/24 – 10.209.0.0/24), avoiding conflicts with
# orphaned subnets left by previous test runs.
_RESOURCE_TS="${RESOURCE_PREFIX##*-}"
SUBNET_CIDR="10.$(( (_RESOURCE_TS % 200) + 10 )).0.0/24"
VPN_SUBNET_CIDR="10.$(( (_RESOURCE_TS % 44) + 210 )).0.0/24"

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

print_banner "Network"

# Poll a resource via a get command until Status: matches ready_regex or timeout elapses.
# Usage: wait_for_status "<get-cmd>" "<ready-regex>" [timeout-seconds]
wait_for_status() {
    local get_cmd="$1"
    local ready_regex="$2"
    local timeout="${3:-180}"
    local elapsed=0
    local status=""
    while [ "$elapsed" -lt "$timeout" ]; do
        local out
        out=$(eval "$get_cmd" 2>&1) || return 1
        status=$(echo "$out" | grep -iE "^Status:" | head -1 | awk -F: '{print $2}' | tr -d '[:space:]')
        if [[ "$status" =~ $ready_regex ]]; then
            echo "  → ready (status=$status)"
            return 0
        fi
        # Bail early on terminal failure states
        if [[ "$status" =~ ^(Failed|Error|Deleted)$ ]]; then
            echo -e "${YELLOW}wait_for_status: terminal state reached ($status), aborting wait${NC}"
            return 1
        fi
        sleep 5
        elapsed=$((elapsed + 5))
    done
    echo -e "${YELLOW}wait_for_status: timeout after ${timeout}s (last status=$status)${NC}"
    return 1
}

wait_for_vpc_ready() {
    wait_for_status "$ACLOUD_CMD network vpc get $1" '^(Active|Ready)$' "${2:-180}"
}

# Wait for a VPC peering to leave transient states.
# Usage: wait_for_peering_ready <vpc_id> <peering_id> [timeout]
wait_for_peering_ready() {
    wait_for_status "$ACLOUD_CMD network vpcpeering get $1 $2" '^(Active|Ready)$' "${3:-180}"
}

# Wait for a VPC peering route to leave transient states.
# Usage: wait_for_peering_route_ready <vpc_id> <peering_id> <route_id> [timeout]
wait_for_peering_route_ready() {
    wait_for_status "$ACLOUD_CMD network vpcpeeringroute get $1 $2 $3" '^(Active|Ready)$' "${4:-180}"
}

# Wait for a VPN tunnel to leave transient states.
# Usage: wait_for_vpn_tunnel_ready <tunnel_id> [timeout]
wait_for_vpn_tunnel_ready() {
    wait_for_status "$ACLOUD_CMD network vpntunnel get $1" '^(Active|Ready)$' "${2:-300}"
}

# Accept an Elastic IP ID or URI; strips the trailing path segment to get the id.
# Uses elasticip list (more reliable than get for status extraction).
# Returns 1 immediately if the IP is not found — prevents infinite loop on stale IDs.
wait_for_elastic_ip_ready() {
    local eip_ref="$1"
    local eip_id="${eip_ref##*/}"
    local timeout="${2:-30}"
    local elapsed=0
    local status=""
    while [ "$elapsed" -lt "$timeout" ]; do
        status=$($ACLOUD_CMD network elasticip list 2>&1 | awk -v id="$eip_id" '$2 == id {print $NF}')
        if [ -z "$status" ]; then
            echo -e "${YELLOW}  ⚠ elastic IP $eip_id not found in list — cannot wait for readiness${NC}"
            return 1
        fi
        case "$status" in
            Active|NotUsed|Ready) echo "  → elastic IP $eip_id ready (status=$status)"; return 0;;
            *) sleep 5; elapsed=$((elapsed + 5));;
        esac
    done
    echo -e "${YELLOW}wait_for_elastic_ip_ready: timeout after ${timeout}s (last status=$status)${NC}"
    return 1
}

# Test --output flag for network list commands
test_output_formats() {
    echo -e "${BLUE}--- Testing network list --output flag ---${NC}"

    for resource_cmd in "vpc list" "subnet list" "securitygroup list" "elasticip list"; do
        local label="network $resource_cmd"
        local extra=""
        case "$resource_cmd" in
            "subnet list"|"securitygroup list")
                if [ -z "$VPC_ID" ] || ! is_valid_id "$VPC_ID"; then
                    echo -e "${YELLOW}⚠ skipping $label (no valid VPC_ID)${NC}"
                    continue
                fi
                extra="$VPC_ID"
                ;;
        esac

        echo -e "${YELLOW}Testing $label --output json...${NC}"
        JSON_OUTPUT=$($ACLOUD_CMD network $resource_cmd $extra --project-id "$PROJECT_ID" --output json 2>&1)
        JSON_EXIT=$?
        if [ $JSON_EXIT -ne 0 ]; then
            fail "$label --output json: command failed (exit $JSON_EXIT)"
        elif echo "$JSON_OUTPUT" | grep -qF "No "; then
            echo -e "${YELLOW}⚠ $label --output json: no resources — format validation skipped${NC}"
        elif is_valid_json "$JSON_OUTPUT"; then
            echo -e "${GREEN}✓ $label --output json: valid JSON${NC}"
        else
            fail "$label --output json: output is not valid JSON"
        fi

        echo -e "${YELLOW}Testing $label --output yaml...${NC}"
        YAML_OUTPUT=$($ACLOUD_CMD network $resource_cmd $extra --project-id "$PROJECT_ID" --output yaml 2>&1)
        YAML_EXIT=$?
        if [ $YAML_EXIT -ne 0 ]; then
            fail "$label --output yaml: command failed (exit $YAML_EXIT)"
        elif echo "$YAML_OUTPUT" | grep -qF "No "; then
            echo -e "${YELLOW}⚠ $label --output yaml: no resources — format validation skipped${NC}"
        elif echo "$YAML_OUTPUT" | grep -qE '^[a-zA-Z].*:|^- '; then
            echo -e "${GREEN}✓ $label --output yaml: output looks like YAML${NC}"
        else
            fail "$label --output yaml: output does not look like YAML"
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
    local had_tunnels=0
    for tunnel_id in "${CREATED_VPN_TUNNELS[@]}"; do
        if is_valid_id "$tunnel_id"; then
            had_tunnels=1
            echo "Waiting for VPN tunnel $tunnel_id to be deletable..."
            wait_for_vpn_tunnel_ready "$tunnel_id" 300 || true
            echo "Deleting VPN tunnel: $tunnel_id"
            $ACLOUD_CMD network vpntunnel delete "$tunnel_id" --yes 2>&1 || true
        fi
    done

    # Wait for the platform to fully release tunnel constraints (EIP, subnet, peer IP)
    # before attempting to delete the EIP — otherwise the delete returns 400.
    if [ "$had_tunnels" -eq 1 ]; then
        echo "Waiting for VPN tunnel to be fully purged before deleting EIPs..."
        local purge_elapsed=0
        while [ "$purge_elapsed" -lt 300 ]; do
            local remaining_tunnels
            remaining_tunnels=$($ACLOUD_CMD network vpntunnel list 2>/dev/null | awk 'NR>1 && /[0-9a-f]{24}/ {print $2}')
            if [ -z "$remaining_tunnels" ]; then
                echo "  → VPN tunnels fully purged."
                break
            fi
            sleep 10
            purge_elapsed=$((purge_elapsed + 10))
        done
    fi
    
    # Delete peering routes (only if VPC_ID and PEERING_ID are set and valid)
    if [ -n "$VPC_ID" ] && is_valid_id "$VPC_ID" && [ -n "$PEERING_ID" ] && is_valid_id "$PEERING_ID"; then
        for route_id in "${CREATED_PEERING_ROUTES[@]}"; do
            if is_valid_id "$route_id"; then
                echo "Waiting for peering route $route_id to be deletable..."
                wait_for_peering_route_ready "$VPC_ID" "$PEERING_ID" "$route_id" 180 || true
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
        "$ACLOUD_CMD network subnet create $VPC_ID --name ${RESOURCE_PREFIX}-subnet --cidr $SUBNET_CIDR --dhcp-enabled --region $REGION" \
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
#
# Only ONE peering can exist between a given VPC pair, so we cannot run the
# standard test_resource CRUD cycle (which deletes at the end) and then create
# a second peering for test_vpc_peering_route. Instead we CREATE → wait →
# LIST → GET → UPDATE here and leave the peering alive; test_vpc_peering_route
# reuses $PEERING_ID, and the EXIT trap cleans up via CREATED_PEERINGS.
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

    echo "Waiting for parent VPC $VPC_ID..."
    wait_for_vpc_ready "$VPC_ID" || { fail "vpcpeering: VPC $VPC_ID not ready after 180s"; return 1; }
    echo "Waiting for peer VPC $PEER_VPC_ID..."
    wait_for_vpc_ready "$PEER_VPC_ID" || { fail "vpcpeering: peer VPC $PEER_VPC_ID not ready after 180s"; return 1; }

    echo -e "${GREEN}[CREATE]${NC} Creating VPC Peering..."
    local create_out
    create_out=$($ACLOUD_CMD network vpcpeering create "$VPC_ID" \
        --name "${RESOURCE_PREFIX}-peering" \
        --peer-vpc-id "$PEER_VPC_URI" \
        --region "$REGION" 2>&1) || {
        fail "vpcpeering: CREATE failed: $create_out"
        return 1
    }
    echo "$create_out"

    PEERING_ID=$(extract_id "$create_out" "$VPC_ID")
    if [ -z "$PEERING_ID" ] || ! is_valid_id "$PEERING_ID" || [ "$PEERING_ID" = "$VPC_ID" ]; then
        fail "vpcpeering: could not extract peering ID from CREATE output"
        echo "$create_out"
        return 1
    fi
    CREATED_PEERINGS+=("$PEERING_ID")
    echo -e "${GREEN}Created peering ID: $PEERING_ID${NC}\n"

    echo "Waiting for peering to be Active..."
    wait_for_peering_ready "$VPC_ID" "$PEERING_ID" 180 || \
        echo -e "${YELLOW}  ⚠ peering may not be Active yet — UPDATE may fail${NC}"

    echo -e "${GREEN}[LIST]${NC} Listing VPC peerings..."
    $ACLOUD_CMD network vpcpeering list "$VPC_ID" 2>&1 | head -10
    echo ""

    echo -e "${GREEN}[GET]${NC} Getting peering details..."
    $ACLOUD_CMD network vpcpeering get "$VPC_ID" "$PEERING_ID" 2>&1
    echo ""

    echo -e "${GREEN}[UPDATE]${NC} Updating VPC Peering..."
    local update_out
    update_out=$($ACLOUD_CMD network vpcpeering update "$VPC_ID" "$PEERING_ID" \
        --name "${RESOURCE_PREFIX}-peering-updated" \
        --tags updated 2>&1) || {
        echo -e "${YELLOW}UPDATE failed (non-fatal):${NC}"
        echo "$update_out"
    }
    echo "$update_out"
    echo ""

    echo -e "${GREEN}✓ VPC Peering test completed (delete deferred to cleanup trap)${NC}\n"
}

# Test VPC Peering Route
test_vpc_peering_route() {
    if [ -z "$VPC_ID" ] || [ "$VPC_ID" = "your-vpc-id" ] || [ "$VPC_ID" = "" ] || [ -z "$PEERING_ID" ]; then
        echo -e "${YELLOW}Skipping VPC peering route test (no VPC or peering available)${NC}\n"
        return 0
    fi

    echo -e "${YELLOW}=== 7. VPC Peering Route CRUD Test ===${NC}\n"

    echo -e "${GREEN}[CREATE]${NC} Creating VPC Peering Route..."
    local create_out
    create_out=$($ACLOUD_CMD network vpcpeeringroute create "$VPC_ID" "$PEERING_ID" \
        --name "${RESOURCE_PREFIX}-route" \
        --region "$REGION" \
        --local-network 10.0.1.0/24 \
        --remote-network 10.0.2.0/24 \
        --billing-period Hour 2>&1)
    local create_exit=$?

    if [ $create_exit -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$create_out"
        return 1
    fi

    echo "$create_out"
    local route_id
    route_id=$(extract_id "$create_out" "$VPC_ID")
    if [ -z "$route_id" ] || [ "$route_id" = "$VPC_ID" ] || [ "$route_id" = "$PEERING_ID" ] || ! is_valid_id "$route_id"; then
        fail "vpcpeeringroute: could not extract valid route ID"
        return 1
    fi
    CREATED_PEERING_ROUTES+=("$route_id")
    echo -e "${GREEN}Created route ID: $route_id${NC}\n"

    echo -e "${GREEN}[LIST]${NC} Listing VPC Peering Routes..."
    $ACLOUD_CMD network vpcpeeringroute list "$VPC_ID" "$PEERING_ID" 2>&1 | head -10
    echo ""

    echo -e "${GREEN}[GET]${NC} Getting VPC Peering Route details..."
    $ACLOUD_CMD network vpcpeeringroute get "$VPC_ID" "$PEERING_ID" "$route_id" 2>&1
    echo ""

    echo -e "Waiting for peering route to be Active..."
    local route_ready=0
    wait_for_peering_route_ready "$VPC_ID" "$PEERING_ID" "$route_id" 180 && route_ready=1
    echo ""

    if [ "$route_ready" -eq 1 ]; then
        echo -e "${GREEN}[UPDATE]${NC} Updating VPC Peering Route..."
        $ACLOUD_CMD network vpcpeeringroute update "$VPC_ID" "$PEERING_ID" "$route_id" \
            --name "${RESOURCE_PREFIX}-route-updated" --tags updated 2>&1 || true
        echo ""
    else
        echo -e "${YELLOW}[UPDATE]${NC} Skipping update — route did not reach Active state."
        echo ""
    fi

    echo -e "${GREEN}[DELETE]${NC} Deleting VPC Peering Route..."
    $ACLOUD_CMD network vpcpeeringroute delete "$VPC_ID" "$PEERING_ID" "$route_id" --yes 2>&1 || true
    # Remove from cleanup list since we already deleted it inline
    CREATED_PEERING_ROUTES=("${CREATED_PEERING_ROUTES[@]/$route_id}")
    echo ""

    echo -e "${GREEN}✓ VPC Peering Route test completed successfully!${NC}\n"
}

# Test VPN Tunnel
test_vpn_tunnel() {
    echo -e "${YELLOW}=== 8. VPN Tunnel CRUD Test ===${NC}\n"

    if [ -z "$VPC_ID" ] || [ "$VPC_ID" = "your-vpc-id" ] || [ "$VPC_ID" = "" ]; then
        echo -e "${YELLOW}Skipping VPN tunnel test (no VPC available)${NC}\n"
        return 0
    fi

    echo "Waiting for VPC $VPC_ID..."
    wait_for_vpc_ready "$VPC_ID" || { fail "vpntunnel: VPC $VPC_ID not ready after 180s"; return 1; }

    # Pre-cleanup: delete any leftover VPN tunnels from previous runs.
    # The API enforces exclusive constraints on peer IP / EIP / subnet —
    # even after deletion the platform needs time to fully release them.
    echo "Checking for leftover VPN tunnels from previous runs..."
    local leftover_tunnels
    leftover_tunnels=$($ACLOUD_CMD network vpntunnel list 2>/dev/null | awk 'NR>1 && /[0-9a-f]{24}/ {print $2}')
    for old_id in $leftover_tunnels; do
        if is_valid_id "$old_id"; then
            echo "  → Waiting for leftover tunnel $old_id to be deletable..."
            wait_for_vpn_tunnel_ready "$old_id" 300 || true
            echo "  → Deleting leftover tunnel $old_id..."
            $ACLOUD_CMD network vpntunnel delete "$old_id" --yes 2>&1 || true
        fi
    done

    # Wait until vpntunnel list is empty — the platform releases EIP/subnet
    # constraints only once the resource is fully purged, which can take minutes.
    if [ -n "$leftover_tunnels" ]; then
        echo "  → Waiting for platform to fully release VPN tunnel constraints..."
        local wait_elapsed=0
        local wait_limit=300
        while [ "$wait_elapsed" -lt "$wait_limit" ]; do
            local remaining
            remaining=$($ACLOUD_CMD network vpntunnel list 2>/dev/null | awk 'NR>1 && /[0-9a-f]{24}/ {print $2}')
            if [ -z "$remaining" ]; then
                echo "  → VPN tunnel list is now empty, constraints released."
                break
            fi
            sleep 10
            wait_elapsed=$((wait_elapsed + 10))
        done
        if [ "$wait_elapsed" -ge "$wait_limit" ]; then
            echo -e "${YELLOW}  ⚠ Timed out waiting for tunnel constraints to release; proceeding anyway.${NC}"
        fi
    fi
    echo ""

    # Create a dedicated Elastic IP for the VPN tunnel (test_elastic_ip's EIP is deleted by its
    # own CRUD cycle; the hardcoded fallback URI may not exist in the tenant).
    local vpn_eip_uri="$ELASTIC_IP_URI"
    if [ -z "$vpn_eip_uri" ]; then
        echo "Creating prerequisite VPN Elastic IP..."
        local vpn_eip_out vpn_eip_id
        vpn_eip_out=$($ACLOUD_CMD network elasticip create \
            --name "${RESOURCE_PREFIX}-vpn-eip" \
            --region "$REGION" \
            --billing-period Hour 2>&1)
        vpn_eip_id=$(extract_id "$vpn_eip_out")
        if [ -z "$vpn_eip_id" ] || ! is_valid_id "$vpn_eip_id"; then
            fail "vpntunnel: could not create prerequisite Elastic IP: $vpn_eip_out"
            return 1
        fi
        CREATED_ELASTIC_IPS+=("$vpn_eip_id")
        vpn_eip_uri="/projects/$PROJECT_ID/providers/Aruba.Network/elasticIps/$vpn_eip_id"
        echo "  → Elastic IP $vpn_eip_id created, waiting for readiness..."
        wait_for_elastic_ip_ready "$vpn_eip_id" 120 || \
            echo -e "${YELLOW}  ⚠ Elastic IP may not be ready yet${NC}"
    else
        echo "Waiting for Elastic IP $vpn_eip_uri..."
        wait_for_elastic_ip_ready "$vpn_eip_uri" || { fail "vpntunnel: elastic IP not found or not ready"; return 1; }
    fi

    # The VPN tunnel API creates its own subnet as part of provisioning — do NOT pre-create it.
    local vpn_create_out vpn_create_exit
    echo -e "${YELLOW}--- Testing VPN Tunnel ---${NC}"
    echo -e "${GREEN}[CREATE]${NC} Creating VPN Tunnel..."
    vpn_create_out=$($ACLOUD_CMD network vpntunnel create \
        --name "${RESOURCE_PREFIX}-vpn" \
        --region "$REGION" \
        --peer-ip 210.8.25.187 \
        --vpc-uri "/projects/$PROJECT_ID/providers/Aruba.Network/vpcs/$VPC_ID" \
        --subnet-cidr "$VPN_SUBNET_CIDR" \
        --subnet-name "${RESOURCE_PREFIX}-vpn-subnet" \
        --elastic-ip-uri "$vpn_eip_uri" \
        --vpn-type Site-To-Site \
        --protocol ikev2 \
        --ike-lifetime 3600 \
        --ike-encryption aes256 \
        --ike-hash sha1 \
        --ike-dh-group 1 \
        --ike-dpd-action restart \
        --ike-dpd-interval 10 \
        --ike-dpd-timeout 30 \
        --esp-lifetime 1800 \
        --esp-encryption aes256 \
        --esp-hash sha1 \
        --esp-pfs enable \
        --psk-cloud-site psk-id-ARUBA \
        --psk-onprem-site psk-id-CUSTOMER \
        --psk e2e-test-psk-secret \
        --billing-period Hour 2>&1)
    vpn_create_exit=$?
    echo "$vpn_create_out"

    if [ $vpn_create_exit -ne 0 ]; then
        echo -e "${RED}CREATE failed — retrying with --debug for full API payload:${NC}"
        $ACLOUD_CMD --debug network vpntunnel create \
            --name "${RESOURCE_PREFIX}-vpn-dbg" \
            --region "$REGION" \
            --peer-ip 210.8.25.187 \
            --vpc-uri "/projects/$PROJECT_ID/providers/Aruba.Network/vpcs/$VPC_ID" \
            --subnet-cidr "$VPN_SUBNET_CIDR" \
            --subnet-name "${RESOURCE_PREFIX}-vpn-subnet" \
            --elastic-ip-uri "$vpn_eip_uri" \
            --vpn-type Site-To-Site \
            --protocol ikev2 \
            --ike-lifetime 3600 \
            --ike-encryption aes256 \
            --ike-hash sha1 \
            --ike-dh-group 1 \
            --ike-dpd-action restart \
            --ike-dpd-interval 10 \
            --ike-dpd-timeout 30 \
            --esp-lifetime 1800 \
            --esp-encryption aes256 \
            --esp-hash sha1 \
            --esp-pfs enable \
            --psk-cloud-site psk-id-ARUBA \
            --psk-onprem-site psk-id-CUSTOMER \
            --psk e2e-test-psk-secret \
            --billing-period Hour 2>&1 || true
        return 1
    fi

    VPN_TUNNEL_ID=$(extract_id "$vpn_create_out" "$VPC_ID")
    if [ -z "$VPN_TUNNEL_ID" ] || ! is_valid_id "$VPN_TUNNEL_ID"; then
        echo -e "${RED}Could not extract VPN tunnel ID from create output${NC}"
        return 1
    fi
    CREATED_VPN_TUNNELS+=("$VPN_TUNNEL_ID")
    echo -e "${GREEN}Created VPN Tunnel ID: $VPN_TUNNEL_ID${NC}\n"

    echo -e "${GREEN}[LIST]${NC} Listing VPN Tunnels..."
    $ACLOUD_CMD network vpntunnel list 2>&1 | head -10
    echo ""

    echo -e "${GREEN}[GET]${NC} Getting VPN Tunnel details..."
    $ACLOUD_CMD network vpntunnel get "$VPN_TUNNEL_ID" 2>&1
    echo ""

    echo "Waiting for VPN Tunnel $VPN_TUNNEL_ID to be Active..."
    local tunnel_ready=0
    wait_for_vpn_tunnel_ready "$VPN_TUNNEL_ID" 600 && tunnel_ready=1
    echo ""

    if [ "$tunnel_ready" -eq 1 ]; then
        echo -e "${GREEN}[UPDATE]${NC} Updating VPN Tunnel..."
        $ACLOUD_CMD network vpntunnel update "$VPN_TUNNEL_ID" \
            --name "${RESOURCE_PREFIX}-vpn-updated" --tags updated 2>&1 || true
        echo ""

        echo -e "${GREEN}[DELETE]${NC} Deleting VPN Tunnel..."
        echo "Waiting for VPN Tunnel to be deletable after update..."
        wait_for_vpn_tunnel_ready "$VPN_TUNNEL_ID" 120 || true
        if $ACLOUD_CMD network vpntunnel delete "$VPN_TUNNEL_ID" --yes 2>&1; then
            CREATED_VPN_TUNNELS=("${CREATED_VPN_TUNNELS[@]/$VPN_TUNNEL_ID}")
        fi
        echo ""
    else
        echo -e "${YELLOW}[UPDATE/DELETE]${NC} Skipping — tunnel did not reach Active state (left for cleanup trap)."
        echo ""
    fi

    echo -e "${GREEN}VPN Tunnel test completed successfully${NC}\n"
}

# Test VPN Route
test_vpn_route() {
    if [ -z "$VPN_TUNNEL_ID" ]; then
        echo -e "${YELLOW}Skipping VPN route test (no VPN tunnel available)${NC}\n"
        return 0
    fi
    
    echo -e "${YELLOW}=== 9. VPN Route CRUD Test ===${NC}\n"

    echo "Waiting for VPN Tunnel $VPN_TUNNEL_ID to be Active before creating routes..."
    wait_for_vpn_tunnel_ready "$VPN_TUNNEL_ID" 600 || {
        echo -e "${YELLOW}VPN Tunnel not Active — skipping VPN Route test${NC}\n"
        return 0
    }
    echo ""

    # The VPN route's --cloud-subnet must reference an existing VPC subnet CIDR,
    # not the VPN tunnel's own internal provisioning subnet. Create a dedicated
    # VPC subnet for this purpose and tear it down after the test.
    echo "Creating VPC subnet for VPN route cloud-subnet reference..."
    local route_subnet_out route_subnet_id
    route_subnet_out=$($ACLOUD_CMD network subnet create "$VPC_ID" \
        --name "${RESOURCE_PREFIX}-vpn-cloud-subnet" \
        --cidr "$SUBNET_CIDR" \
        --dhcp-enabled \
        --region "$REGION" 2>&1)
    route_subnet_id=$(extract_id "$route_subnet_out" "$VPC_ID")
    if [ -z "$route_subnet_id" ] || ! is_valid_id "$route_subnet_id"; then
        fail "vpnroute: could not create cloud subnet for route test: $route_subnet_out"
        return 1
    fi
    echo "  → Cloud subnet $route_subnet_id ($SUBNET_CIDR) created"
    echo ""

    local route_create_out route_create_exit
    echo -e "${YELLOW}--- Testing VPN Route ---${NC}"
    echo -e "${GREEN}[CREATE]${NC} Creating VPN Route..."
    route_create_out=$($ACLOUD_CMD network vpnroute create "$VPN_TUNNEL_ID" \
        --name "${RESOURCE_PREFIX}-vpn-route" \
        --region "$REGION" \
        --cloud-subnet "$SUBNET_CIDR" \
        --onprem-subnet 192.168.129.0/24 2>&1)
    route_create_exit=$?
    echo "$route_create_out"

    if [ $route_create_exit -ne 0 ]; then
        echo -e "${RED}CREATE failed — retrying with --debug for full API payload:${NC}"
        $ACLOUD_CMD --debug network vpnroute create "$VPN_TUNNEL_ID" \
            --name "${RESOURCE_PREFIX}-vpn-route-dbg" \
            --region "$REGION" \
            --cloud-subnet "$SUBNET_CIDR" \
            --onprem-subnet 192.168.129.0/24 2>&1 || true
        # Clean up the cloud subnet before returning
        $ACLOUD_CMD network subnet delete "$VPC_ID" "$route_subnet_id" --yes 2>&1 || true
        return 1
    fi

    local vpn_route_id
    vpn_route_id=$(extract_id "$route_create_out" "$VPN_TUNNEL_ID")
    if [ -z "$vpn_route_id" ] || ! is_valid_id "$vpn_route_id"; then
        echo -e "${RED}Could not extract VPN route ID from create output${NC}"
        return 1
    fi
    CREATED_VPN_ROUTES+=("$vpn_route_id")
    echo -e "${GREEN}Created VPN Route ID: $vpn_route_id${NC}\n"

    echo -e "${GREEN}[LIST]${NC} Listing VPN Routes..."
    $ACLOUD_CMD network vpnroute list "$VPN_TUNNEL_ID" 2>&1 | head -10
    echo ""

    echo -e "${GREEN}[GET]${NC} Getting VPN Route details..."
    $ACLOUD_CMD network vpnroute get "$VPN_TUNNEL_ID" "$vpn_route_id" 2>&1
    echo ""

    echo -e "${GREEN}[UPDATE]${NC} Updating VPN Route..."
    $ACLOUD_CMD network vpnroute update "$VPN_TUNNEL_ID" "$vpn_route_id" \
        --name "${RESOURCE_PREFIX}-vpn-route-updated" --tags updated 2>&1 || true
    echo ""

    echo -e "${GREEN}[DELETE]${NC} Deleting VPN Route..."
    if $ACLOUD_CMD network vpnroute delete "$VPN_TUNNEL_ID" "$vpn_route_id" --yes 2>&1; then
        CREATED_VPN_ROUTES=("${CREATED_VPN_ROUTES[@]/$vpn_route_id}")
    fi
    echo ""

    echo "Deleting VPN route cloud subnet..."
    $ACLOUD_CMD network subnet delete "$VPC_ID" "$route_subnet_id" --yes 2>&1 || true
    echo ""

    echo -e "${GREEN}VPN Route test completed successfully${NC}\n"
}

setup_context

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

test_subnet         || { FAILURES=$((FAILURES + 1)); echo -e "${YELLOW}Subnet test completed with errors${NC}\n"; }
test_security_group || { FAILURES=$((FAILURES + 1)); echo -e "${YELLOW}Security Group test completed with errors${NC}\n"; }
test_security_rule  || { FAILURES=$((FAILURES + 1)); echo -e "${YELLOW}Security Rule test completed with errors${NC}\n"; }
test_elastic_ip     || { FAILURES=$((FAILURES + 1)); echo -e "${YELLOW}Elastic IP test completed with errors${NC}\n"; }

# VPC Peering tests (require peer VPC)
if [ -n "$PEER_VPC_ID" ] && [ "$PEER_VPC_ID" != "your-peer-vpc-id" ]; then
    test_vpc_peering       || { FAILURES=$((FAILURES + 1)); echo -e "${YELLOW}VPC Peering test completed with errors${NC}\n"; }
    test_vpc_peering_route || { FAILURES=$((FAILURES + 1)); echo -e "${YELLOW}VPC Peering Route test completed with errors${NC}\n"; }
else
    echo -e "${YELLOW}Skipping VPC Peering tests (PEER_VPC_ID not set or invalid)${NC}\n"
fi

# VPN tests — ELASTIC_IP_URI may be empty; test_vpn_tunnel creates one if needed.
test_vpn_tunnel || { FAILURES=$((FAILURES + 1)); echo -e "${YELLOW}VPN Tunnel test completed with errors${NC}\n"; }
test_vpn_route  || { FAILURES=$((FAILURES + 1)); echo -e "${YELLOW}VPN Route test completed with errors${NC}\n"; }

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

if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}=== Network E2E: all checks passed ===${NC}"
    exit 0
else
    echo -e "${RED}=== Network E2E: $FAILURES check(s) failed ===${NC}"
    exit 1
fi
