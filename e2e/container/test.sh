#!/bin/bash

# E2E Test Script for Container Resources
# Tests CRUD operations for KaaS (Kubernetes as a Service) clusters and Container Registry

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# Cleanup tracking
CREATED_CLUSTERS=()
CREATED_REGISTRIES=()

print_banner "Container"
echo "Note: KaaS tests require the following environment variables:"
echo "  - ACLOUD_VPC_URI (required)"
echo "  - ACLOUD_SUBNET_URI (required)"
echo "  - ACLOUD_NODE_POOL_INSTANCE (required)"
echo "  - ACLOUD_NODE_POOL_ZONE (required)"
echo "  - ACLOUD_NODE_CIDR (optional, default: 10.0.0.0/16)"
echo "  - ACLOUD_NODE_CIDR_NAME (optional, default: node-cidr)"
echo "  - ACLOUD_SECURITY_GROUP_NAME (optional, default: kaas-sg)"
echo "  - ACLOUD_NODE_POOL_NAME (optional, default: default-pool)"
echo "  - ACLOUD_NODE_POOL_NODES (optional, default: 1)"
echo "  - ACLOUD_K8S_VERSION (optional, default: 1.28.0)"
echo ""
echo "Note: Container Registry tests require the following environment variables:"
echo "  - ACLOUD_PUBLIC_IP_URI (required)"
echo "  - ACLOUD_VPC_URI (required)"
echo "  - ACLOUD_SUBNET_URI (required)"
echo "  - ACLOUD_SECURITY_GROUP_URI (required)"
echo "  - ACLOUD_BLOCK_STORAGE_URI (required)"
echo ""

# Test --output flag for container list commands
test_output_formats() {
    echo -e "${BLUE}--- Testing container list --output flag ---${NC}"

    for resource_cmd in "kaas list" "containerregistry list"; do
        local label="container $resource_cmd"
        echo -e "${YELLOW}Testing $label --output json...${NC}"
        JSON_OUTPUT=$($ACLOUD_CMD container $resource_cmd --project-id "$PROJECT_ID" --output json 2>&1)
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
        YAML_OUTPUT=$($ACLOUD_CMD container $resource_cmd --project-id "$PROJECT_ID" --output yaml 2>&1)
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

# Function to test KaaS operations
test_kaas() {
    local cluster_name="${RESOURCE_PREFIX}-kaas"
    local version="${ACLOUD_K8S_VERSION:-1.28.0}"
    
    # Required environment variables for KaaS creation
    local vpc_uri="${ACLOUD_VPC_URI}"
    local subnet_uri="${ACLOUD_SUBNET_URI}"
    local node_cidr_address="${ACLOUD_NODE_CIDR:-10.0.0.0/16}"
    local node_cidr_name="${ACLOUD_NODE_CIDR_NAME:-node-cidr}"
    local security_group_name="${ACLOUD_SECURITY_GROUP_NAME:-kaas-sg}"
    local node_pool_name="${ACLOUD_NODE_POOL_NAME:-default-pool}"
    local node_pool_nodes="${ACLOUD_NODE_POOL_NODES:-1}"
    local node_pool_instance="${ACLOUD_NODE_POOL_INSTANCE}"
    local node_pool_zone="${ACLOUD_NODE_POOL_ZONE}"
    
    echo -e "${BLUE}--- Testing KaaS Operations ---${NC}"
    
    # Check if required variables are set
    if [ -z "$vpc_uri" ] || [ -z "$subnet_uri" ] || [ -z "$node_pool_instance" ] || [ -z "$node_pool_zone" ]; then
        echo -e "${YELLOW}⚠ Skipping KaaS create test - required environment variables not set${NC}"
        echo "Required: ACLOUD_VPC_URI, ACLOUD_SUBNET_URI, ACLOUD_NODE_POOL_INSTANCE, ACLOUD_NODE_POOL_ZONE"
        echo "Optional: ACLOUD_NODE_CIDR, ACLOUD_NODE_CIDR_NAME, ACLOUD_SECURITY_GROUP_NAME, ACLOUD_NODE_POOL_NAME, ACLOUD_NODE_POOL_NODES"
        return
    fi
    
    # Create
    echo -e "${YELLOW}Creating KaaS cluster...${NC}"
    CREATE_OUTPUT=$($ACLOUD_CMD container kaas create \
        --project-id "$PROJECT_ID" \
        --name "$cluster_name" \
        --region "$REGION" \
        --vpc-uri "$vpc_uri" \
        --subnet-uri "$subnet_uri" \
        --node-cidr-address "$node_cidr_address" \
        --node-cidr-name "$node_cidr_name" \
        --security-group-name "$security_group_name" \
        --kubernetes-version "$version" \
        --node-pool-name "$node_pool_name" \
        --node-pool-nodes "$node_pool_nodes" \
        --node-pool-instance "$node_pool_instance" \
        --node-pool-zone "$node_pool_zone" \
        --tags "e2e-test,kaas" 2>&1)
    CREATE_EXIT=$?
    
    if [ $CREATE_EXIT -eq 0 ]; then
        CLUSTER_ID=$(extract_id "$CREATE_OUTPUT")
        if [ -n "$CLUSTER_ID" ]; then
            CREATED_CLUSTERS+=("$CLUSTER_ID")
            echo -e "${GREEN}✓ KaaS cluster created: $CLUSTER_ID${NC}"
        else
            echo -e "${YELLOW}⚠ KaaS cluster creation may have succeeded but ID not found${NC}"
        fi
    else
        echo -e "${RED}✗ Failed to create KaaS cluster${NC}"
        echo "$CREATE_OUTPUT"
    fi
    
    # List
    echo -e "${YELLOW}Listing KaaS clusters...${NC}"
    LIST_OUTPUT=$($ACLOUD_CMD container kaas list --project-id "$PROJECT_ID" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ KaaS cluster list successful${NC}"
    else
        echo -e "${RED}✗ Failed to list KaaS clusters${NC}"
        echo "$LIST_OUTPUT"
    fi
    
    # Get (if we have an ID)
    if [ -n "$CLUSTER_ID" ]; then
        echo -e "${YELLOW}Getting KaaS cluster details...${NC}"
        GET_OUTPUT=$($ACLOUD_CMD container kaas get "$CLUSTER_ID" --project-id "$PROJECT_ID" 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ KaaS cluster get successful${NC}"
        else
            echo -e "${RED}✗ Failed to get KaaS cluster${NC}"
            echo "$GET_OUTPUT"
        fi
        
        # Update
        echo -e "${YELLOW}Updating KaaS cluster...${NC}"
        UPDATE_OUTPUT=$($ACLOUD_CMD container kaas update "$CLUSTER_ID" \
            --project-id "$PROJECT_ID" \
            --name "${cluster_name}-updated" \
            --tags "e2e-test,kaas,updated" 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ KaaS cluster update successful${NC}"
        else
            echo -e "${RED}✗ Failed to update KaaS cluster${NC}"
            echo "$UPDATE_OUTPUT"
        fi
        
        # Connect (optional - requires kubectl)
        if command -v kubectl >/dev/null 2>&1; then
            echo -e "${YELLOW}Testing KaaS connect...${NC}"
            CONNECT_OUTPUT=$($ACLOUD_CMD container kaas connect "$CLUSTER_ID" \
                --project-id "$PROJECT_ID" 2>&1)
            if [ $? -eq 0 ]; then
                echo -e "${GREEN}✓ KaaS connect successful${NC}"
            else
                echo -e "${YELLOW}⚠ KaaS connect failed (may be expected if cluster is not ready)${NC}"
                echo "$CONNECT_OUTPUT"
            fi
        else
            echo -e "${YELLOW}⚠ Skipping KaaS connect test - kubectl not found${NC}"
        fi
    fi
    
    echo ""
}

# Function to test Container Registry operations
test_containerregistry() {
    local registry_name="${RESOURCE_PREFIX}-registry"
    
    # Required environment variables for Container Registry creation
    local public_ip_uri="${ACLOUD_PUBLIC_IP_URI}"
    local vpc_uri="${ACLOUD_VPC_URI}"
    local subnet_uri="${ACLOUD_SUBNET_URI}"
    local security_group_uri="${ACLOUD_SECURITY_GROUP_URI}"
    local block_storage_uri="${ACLOUD_BLOCK_STORAGE_URI}"
    
    echo -e "${BLUE}--- Testing Container Registry Operations ---${NC}"
    
    # Check if required variables are set
    if [ -z "$public_ip_uri" ] || [ -z "$vpc_uri" ] || [ -z "$subnet_uri" ] || [ -z "$security_group_uri" ] || [ -z "$block_storage_uri" ]; then
        echo -e "${YELLOW}⚠ Skipping Container Registry create test - required environment variables not set${NC}"
        echo "Required: ACLOUD_PUBLIC_IP_URI, ACLOUD_VPC_URI, ACLOUD_SUBNET_URI, ACLOUD_SECURITY_GROUP_URI, ACLOUD_BLOCK_STORAGE_URI"
        return
    fi
    
    # Create
    echo -e "${YELLOW}Creating Container Registry...${NC}"
    CREATE_OUTPUT=$($ACLOUD_CMD container containerregistry create \
        --project-id "$PROJECT_ID" \
        --name "$registry_name" \
        --region "$REGION" \
        --public-ip-uri "$public_ip_uri" \
        --vpc-uri "$vpc_uri" \
        --subnet-uri "$subnet_uri" \
        --security-group-uri "$security_group_uri" \
        --block-storage-uri "$block_storage_uri" \
        --admin-username "${ACLOUD_REGISTRY_ADMIN_USER:-admin}" \
        --concurrent-users "${ACLOUD_REGISTRY_CONCURRENT_USERS:-10}" \
        --billing-period "Hour" \
        --tags "e2e-test,registry" 2>&1)
    CREATE_EXIT=$?
    
    if [ $CREATE_EXIT -eq 0 ]; then
        REGISTRY_ID=$(extract_id "$CREATE_OUTPUT")
        if [ -n "$REGISTRY_ID" ]; then
            CREATED_REGISTRIES+=("$REGISTRY_ID")
            echo -e "${GREEN}✓ Container Registry created: $REGISTRY_ID${NC}"
        else
            echo -e "${YELLOW}⚠ Container Registry creation may have succeeded but ID not found${NC}"
        fi
    else
        echo -e "${RED}✗ Failed to create Container Registry${NC}"
        echo "$CREATE_OUTPUT"
    fi
    
    # List
    echo -e "${YELLOW}Listing Container Registries...${NC}"
    LIST_OUTPUT=$($ACLOUD_CMD container containerregistry list --project-id "$PROJECT_ID" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Container Registry list successful${NC}"
    else
        echo -e "${RED}✗ Failed to list Container Registries${NC}"
        echo "$LIST_OUTPUT"
    fi
    
    # Get (if we have an ID)
    if [ -n "$REGISTRY_ID" ]; then
        echo -e "${YELLOW}Getting Container Registry details...${NC}"
        GET_OUTPUT=$($ACLOUD_CMD container containerregistry get "$REGISTRY_ID" --project-id "$PROJECT_ID" 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Container Registry get successful${NC}"
        else
            echo -e "${RED}✗ Failed to get Container Registry${NC}"
            echo "$GET_OUTPUT"
        fi
        
        # Update
        echo -e "${YELLOW}Updating Container Registry...${NC}"
        UPDATE_OUTPUT=$($ACLOUD_CMD container containerregistry update "$REGISTRY_ID" \
            --project-id "$PROJECT_ID" \
            --name "${registry_name}-updated" \
            --tags "e2e-test,registry,updated" \
            --concurrent-users 20 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Container Registry update successful${NC}"
        else
            echo -e "${RED}✗ Failed to update Container Registry${NC}"
            echo "$UPDATE_OUTPUT"
        fi
    fi
    
    echo ""
}

# Cleanup function
cleanup() {
    echo -e "${BLUE}--- Cleanup ---${NC}"
    
    # Delete Container Registries
    for registry_id in "${CREATED_REGISTRIES[@]}"; do
        echo -e "${YELLOW}Deleting Container Registry: $registry_id${NC}"
        $ACLOUD_CMD container containerregistry delete "$registry_id" --project-id "$PROJECT_ID" --yes 2>&1 >/dev/null
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Container Registry deleted: $registry_id${NC}"
        else
            echo -e "${RED}✗ Failed to delete Container Registry: $registry_id${NC}"
        fi
    done
    
    # Delete KaaS clusters
    for cluster_id in "${CREATED_CLUSTERS[@]}"; do
        echo -e "${YELLOW}Deleting KaaS cluster: $cluster_id${NC}"
        $ACLOUD_CMD container kaas delete "$cluster_id" --project-id "$PROJECT_ID" --yes 2>&1 >/dev/null
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ KaaS cluster deleted: $cluster_id${NC}"
        else
            echo -e "${RED}✗ Failed to delete KaaS cluster: $cluster_id${NC}"
        fi
    done
    
    echo ""
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Run tests
test_kaas
test_containerregistry
test_output_formats

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Created KaaS clusters: ${#CREATED_CLUSTERS[@]}"
echo "Created Container Registries: ${#CREATED_REGISTRIES[@]}"
echo ""
echo -e "${GREEN}E2E tests completed!${NC}"

