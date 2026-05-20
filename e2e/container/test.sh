#!/bin/bash

# E2E Test Script for Container Resources
# Tests CRUD operations for KaaS clusters and Container Registry.
# Fully self-contained: bootstraps VPC/Subnet/SG/EIP/BlockStorage when
# env-var URIs are absent and tears them down in reverse order on exit.

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# --- State tracking -------------------------------------------------------
CREATED_CLUSTERS=()
CREATED_REGISTRIES=()

# Bootstrapped dep IDs (empty = pre-supplied; do not delete on exit)
BOOTSTRAP_PROJECT_ID=""
BOOTSTRAP_VPC_ID=""
BOOTSTRAP_SUBNET_ID=""
BOOTSTRAP_SG_ID=""
BOOTSTRAP_PUBIP_ID=""
BOOTSTRAP_BLOCK_ID=""

# Resolved dep URIs (populated by ensure_* functions)
VPC_URI="${ACLOUD_VPC_URI:-}"
SUBNET_URI="${ACLOUD_SUBNET_URI:-}"
SG_URI="${ACLOUD_SECURITY_GROUP_URI:-}"
PUBIP_URI="${ACLOUD_PUBLIC_IP_URI:-}"
BLOCK_URI="${ACLOUD_BLOCK_STORAGE_URI:-}"

print_banner "Container"

# --- Cleanup --------------------------------------------------------------
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test resources...${NC}"

    # Registries first (use VPC/subnet/SG/EIP/block)
    for id in "${CREATED_REGISTRIES[@]}"; do
        if is_valid_id "$id"; then
            echo "Waiting for registry $id before delete..."
            wait_for_status "$ACLOUD_CMD container containerregistry get $id" '^(Active|Ready)$' 300 2>/dev/null || true
            echo "Deleting container registry: $id"
            $ACLOUD_CMD container containerregistry delete "$id" --yes 2>&1 || true
        fi
    done

    # KaaS clusters (use VPC/subnet)
    for id in "${CREATED_CLUSTERS[@]}"; do
        if is_valid_id "$id"; then
            echo "Waiting for cluster $id before delete..."
            wait_for_status "$ACLOUD_CMD container kaas get $id" '^(Active|Ready)$' 600 2>/dev/null || true
            echo "Deleting KaaS cluster: $id"
            $ACLOUD_CMD container kaas delete "$id" --yes 2>&1 || true
        fi
    done

    # Bootstrapped infra in reverse dep order
    if [ -n "$BOOTSTRAP_BLOCK_ID" ]; then
        echo "Deleting bootstrapped block storage: $BOOTSTRAP_BLOCK_ID"
        $ACLOUD_CMD storage blockstorage delete "$BOOTSTRAP_BLOCK_ID" --yes 2>&1 || true
    fi
    if [ -n "$BOOTSTRAP_PUBIP_ID" ]; then
        echo "Deleting bootstrapped elastic IP: $BOOTSTRAP_PUBIP_ID"
        $ACLOUD_CMD network elasticip delete "$BOOTSTRAP_PUBIP_ID" --yes 2>&1 || true
    fi
    if [ -n "$BOOTSTRAP_SG_ID" ] && [ -n "$BOOTSTRAP_VPC_ID" ]; then
        echo "Deleting bootstrapped security group: $BOOTSTRAP_SG_ID"
        $ACLOUD_CMD network securitygroup delete "$BOOTSTRAP_VPC_ID" "$BOOTSTRAP_SG_ID" --yes 2>&1 || true
    fi
    if [ -n "$BOOTSTRAP_SUBNET_ID" ] && [ -n "$BOOTSTRAP_VPC_ID" ]; then
        echo "Deleting bootstrapped subnet: $BOOTSTRAP_SUBNET_ID"
        wait_for_status "$ACLOUD_CMD network subnet get $BOOTSTRAP_VPC_ID $BOOTSTRAP_SUBNET_ID" '^(Active|Ready)$' 60 2>/dev/null || true
        $ACLOUD_CMD network subnet delete "$BOOTSTRAP_VPC_ID" "$BOOTSTRAP_SUBNET_ID" --yes 2>&1 || true
    fi
    if [ -n "$BOOTSTRAP_VPC_ID" ]; then
        echo "Deleting bootstrapped VPC: $BOOTSTRAP_VPC_ID"
        local vpc_del_elapsed=0
        while [ "$vpc_del_elapsed" -lt 180 ]; do
            $ACLOUD_CMD network vpc delete "$BOOTSTRAP_VPC_ID" --yes 2>&1 && break
            sleep 15
            vpc_del_elapsed=$((vpc_del_elapsed + 15))
        done
    fi

    # Delete bootstrapped project last (after all child resources)
    if [ -n "$BOOTSTRAP_PROJECT_ID" ]; then
        echo "Deleting bootstrapped project: $BOOTSTRAP_PROJECT_ID"
        $ACLOUD_CMD management project delete "$BOOTSTRAP_PROJECT_ID" --yes 2>&1 || true
    fi

    echo -e "${GREEN}Cleanup completed!${NC}"
}

trap cleanup EXIT

# --- Dep resolvers -------------------------------------------------------

ensure_vpc() {
    if [ -n "$VPC_URI" ]; then
        echo "  → using pre-supplied VPC: $VPC_URI"
        return 0
    fi
    echo "Bootstrapping VPC for container suite..."
    local out
    out=$($ACLOUD_CMD network vpc create \
        --name "${RESOURCE_PREFIX}-container-vpc" \
        --region "$REGION" 2>&1) || { echo -e "${RED}VPC create failed: $out${NC}"; return 1; }
    local vpc_id
    vpc_id=$(extract_id "$out")
    if [ -z "$vpc_id" ] || ! is_valid_id "$vpc_id"; then
        echo -e "${RED}Could not extract VPC ID: $out${NC}"; return 1
    fi
    BOOTSTRAP_VPC_ID="$vpc_id"
    echo "  → waiting for VPC $vpc_id to be Active..."
    wait_for_status "$ACLOUD_CMD network vpc get $vpc_id" '^(Active|Ready)$' 300 || {
        echo -e "${RED}VPC did not become Active${NC}"; return 1
    }
    local uri_line
    uri_line=$($ACLOUD_CMD network vpc get "$vpc_id" 2>/dev/null | grep -i "^URI:" | awk '{print $2}')
    VPC_URI="${uri_line:-/projects/$PROJECT_ID/providers/Aruba.Network/vpcs/$vpc_id}"
    echo "  → VPC $vpc_id ready"
}

ensure_subnet() {
    if [ -n "$SUBNET_URI" ]; then
        echo "  → using pre-supplied Subnet: $SUBNET_URI"
        return 0
    fi
    local vpc_id="${VPC_URI##*/}"
    echo "Bootstrapping Subnet for container suite..."
    local _ts="${RESOURCE_PREFIX##*-}"
    local cidr="10.$(( (_ts % 200) + 10 )).0.0/24"
    local out
    out=$($ACLOUD_CMD network subnet create "$vpc_id" \
        --name "${RESOURCE_PREFIX}-container-subnet" \
        --cidr "$cidr" \
        --dhcp-enabled \
        --region "$REGION" 2>&1) || { echo -e "${RED}Subnet create failed: $out${NC}"; return 1; }
    local subnet_id
    subnet_id=$(extract_id "$out" "$vpc_id")
    if [ -z "$subnet_id" ] || ! is_valid_id "$subnet_id"; then
        echo -e "${RED}Could not extract Subnet ID: $out${NC}"; return 1
    fi
    BOOTSTRAP_SUBNET_ID="$subnet_id"
    echo "  → waiting for Subnet $subnet_id to be Active..."
    wait_for_status "$ACLOUD_CMD network subnet get $vpc_id $subnet_id" '^(Active|Ready)$' 180 || true
    local uri_line
    uri_line=$($ACLOUD_CMD network subnet get "$vpc_id" "$subnet_id" 2>/dev/null | grep -i "^URI:" | awk '{print $2}')
    SUBNET_URI="${uri_line:-/projects/$PROJECT_ID/providers/Aruba.Network/subnets/$subnet_id}"
    echo "  → Subnet $subnet_id ready"
}

ensure_security_group() {
    if [ -n "$SG_URI" ]; then
        echo "  → using pre-supplied Security Group: $SG_URI"
        return 0
    fi
    local vpc_id="${VPC_URI##*/}"
    echo "Bootstrapping Security Group for container suite..."
    local out
    out=$($ACLOUD_CMD network securitygroup create "$vpc_id" \
        --name "${RESOURCE_PREFIX}-container-sg" \
        --region "$REGION" 2>&1) || { echo -e "${RED}SG create failed: $out${NC}"; return 1; }
    local sg_id
    sg_id=$(extract_id "$out" "$vpc_id")
    if [ -z "$sg_id" ] || ! is_valid_id "$sg_id"; then
        echo -e "${RED}Could not extract SG ID: $out${NC}"; return 1
    fi
    BOOTSTRAP_SG_ID="$sg_id"
    echo "  → waiting for Security Group $sg_id to be Active..."
    wait_for_status "$ACLOUD_CMD network securitygroup get $vpc_id $sg_id" '^(Active|Ready)$' 180 || true
    local uri_line
    uri_line=$($ACLOUD_CMD network securitygroup get "$vpc_id" "$sg_id" 2>/dev/null | grep -i "^URI:" | awk '{print $2}')
    SG_URI="${uri_line:-/projects/$PROJECT_ID/providers/Aruba.Network/securityGroups/$sg_id}"
    echo "  → Security Group $sg_id ready"
}

ensure_public_ip() {
    if [ -n "$PUBIP_URI" ]; then
        echo "  → using pre-supplied Elastic IP: $PUBIP_URI"
        return 0
    fi
    echo "Bootstrapping Elastic IP for container suite..."
    local out
    out=$($ACLOUD_CMD network elasticip create \
        --name "${RESOURCE_PREFIX}-container-eip" \
        --region "$REGION" \
        --billing-period Hour 2>&1) || { echo -e "${RED}EIP create failed: $out${NC}"; return 1; }
    local eip_id
    eip_id=$(extract_id "$out")
    if [ -z "$eip_id" ] || ! is_valid_id "$eip_id"; then
        echo -e "${RED}Could not extract EIP ID: $out${NC}"; return 1
    fi
    BOOTSTRAP_PUBIP_ID="$eip_id"
    local uri_line
    uri_line=$($ACLOUD_CMD network elasticip get "$eip_id" 2>/dev/null | grep -i "^URI:" | awk '{print $2}')
    PUBIP_URI="${uri_line:-/projects/$PROJECT_ID/providers/Aruba.Network/elasticIps/$eip_id}"
    echo "  → Elastic IP $eip_id ready"
}

ensure_block_storage() {
    if [ -n "$BLOCK_URI" ]; then
        echo "  → using pre-supplied Block Storage: $BLOCK_URI"
        return 0
    fi
    echo "Bootstrapping Block Storage for container suite..."
    local out
    out=$($ACLOUD_CMD storage blockstorage create \
        --name "${RESOURCE_PREFIX}-container-block" \
        --region "$REGION" \
        --size 10 \
        --type Standard \
        --billing-period Hour \
        --tags "e2e-test,container" 2>&1) || { echo -e "${RED}Block storage create failed: $out${NC}"; return 1; }
    local block_id
    block_id=$(extract_id "$out")
    if [ -z "$block_id" ] || ! is_valid_id "$block_id"; then
        echo -e "${RED}Could not extract block storage ID: $out${NC}"; return 1
    fi
    BOOTSTRAP_BLOCK_ID="$block_id"
    echo "  → waiting for block storage $block_id to be Active..."
    wait_for_status "$ACLOUD_CMD storage blockstorage get $block_id" '^(Active|Available)$' 300 || {
        echo -e "${RED}Block storage did not become Active${NC}"; return 1
    }
    local uri_line
    uri_line=$($ACLOUD_CMD storage blockstorage get "$block_id" 2>/dev/null | grep -i "^URI:" | awk '{print $2}')
    BLOCK_URI="${uri_line:-/projects/$PROJECT_ID/providers/Aruba.Storage/blockStorages/$block_id}"
    echo "  → Block storage $block_id ready"
}

# --- Test KaaS -----------------------------------------------------------
test_kaas() {
    local cluster_name="${RESOURCE_PREFIX}-kaas"
    local version="${ACLOUD_K8S_VERSION:-1.28.0}"
    echo -e "${BLUE}=== 1. KaaS CRUD Test ===${NC}"

    ensure_vpc    || { echo -e "${RED}VPC bootstrap failed — skipping KaaS${NC}"; return 1; }
    ensure_subnet || return 1

    # KaaS node-pool instance/zone are catalog values with no discovery API.
    # Set ACLOUD_NODE_POOL_INSTANCE and ACLOUD_NODE_POOL_ZONE to enable create.
    local node_pool_instance="${ACLOUD_NODE_POOL_INSTANCE:-}"
    local node_pool_zone="${ACLOUD_NODE_POOL_ZONE:-}"

    if [ -z "$node_pool_instance" ] || [ -z "$node_pool_zone" ]; then
        echo -e "${YELLOW}⚠ Skipping KaaS CREATE — set ACLOUD_NODE_POOL_INSTANCE and ACLOUD_NODE_POOL_ZONE${NC}"
        echo -e "${YELLOW}  (no catalog API exists yet; see docs/out-of-scope)${NC}"
        echo ""
        echo -e "${GREEN}[LIST]${NC} Listing KaaS clusters..."
        $ACLOUD_CMD container kaas list 2>&1 | head -10
        echo ""
        return 0
    fi

    local node_cidr_address="${ACLOUD_NODE_CIDR:-10.0.0.0/16}"
    local node_cidr_name="${ACLOUD_NODE_CIDR_NAME:-node-cidr}"
    local security_group_name="${ACLOUD_SECURITY_GROUP_NAME:-kaas-sg}"
    local node_pool_name="${ACLOUD_NODE_POOL_NAME:-default-pool}"
    local node_pool_nodes="${ACLOUD_NODE_POOL_NODES:-1}"

    echo -e "${GREEN}[CREATE]${NC} Creating KaaS cluster: $cluster_name"
    CREATE_OUTPUT=$($ACLOUD_CMD container kaas create \
        --name "$cluster_name" \
        --region "$REGION" \
        --vpc-uri "$VPC_URI" \
        --subnet-uri "$SUBNET_URI" \
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

    if ! check_auth_error "$CREATE_OUTPUT"; then return 1; fi
    if [ $CREATE_EXIT -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi

    CLUSTER_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$CLUSTER_ID" ] || ! is_valid_id "$CLUSTER_ID"; then
        echo -e "${RED}Could not extract cluster ID${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi
    CREATED_CLUSTERS+=("$CLUSTER_ID")
    echo -e "${GREEN}KaaS cluster created: $CLUSTER_ID${NC}\n"

    echo -e "${GREEN}[LIST]${NC} Listing KaaS clusters..."
    $ACLOUD_CMD container kaas list 2>&1 | head -10
    echo ""

    echo "Waiting for cluster $CLUSTER_ID to be Active..."
    local cluster_ready=0
    wait_for_status "$ACLOUD_CMD container kaas get $CLUSTER_ID" '^(Active|Ready)$' 900 && cluster_ready=1
    echo ""

    echo -e "${GREEN}[GET]${NC} Getting cluster details..."
    $ACLOUD_CMD container kaas get "$CLUSTER_ID" 2>&1
    echo ""

    if [ "$cluster_ready" -eq 1 ]; then
        echo -e "${GREEN}[UPDATE]${NC} Updating KaaS cluster..."
        $ACLOUD_CMD container kaas update "$CLUSTER_ID" \
            --name "${cluster_name}-updated" \
            --tags "e2e-test,kaas,updated" 2>&1 || true
        echo ""
    else
        echo -e "${YELLOW}[UPDATE]${NC} Skipping — cluster did not reach Active after 900s."
        echo ""
    fi

    echo -e "${GREEN}✓ KaaS CRUD test completed!${NC}\n"
}

# --- Test Container Registry --------------------------------------------
test_containerregistry() {
    local registry_name="${RESOURCE_PREFIX}-registry"
    echo -e "${BLUE}=== 2. Container Registry CRUD Test ===${NC}"

    ensure_vpc             || { echo -e "${RED}VPC bootstrap failed — skipping registry${NC}"; return 1; }
    ensure_subnet          || return 1
    ensure_security_group  || return 1
    ensure_public_ip       || return 1
    ensure_block_storage   || return 1

    echo -e "${GREEN}[CREATE]${NC} Creating Container Registry: $registry_name"
    CREATE_OUTPUT=$($ACLOUD_CMD container containerregistry create \
        --name "$registry_name" \
        --region "$REGION" \
        --public-ip-uri "$PUBIP_URI" \
        --vpc-uri "$VPC_URI" \
        --subnet-uri "$SUBNET_URI" \
        --security-group-uri "$SG_URI" \
        --block-storage-uri "$BLOCK_URI" \
        --admin-username "${ACLOUD_REGISTRY_ADMIN_USER:-admin}" \
        --concurrent-users "${ACLOUD_REGISTRY_CONCURRENT_USERS:-10}" \
        --billing-period Hour \
        --tags "e2e-test,registry" 2>&1)
    CREATE_EXIT=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then return 1; fi
    if [ $CREATE_EXIT -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi

    REGISTRY_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$REGISTRY_ID" ] || ! is_valid_id "$REGISTRY_ID"; then
        echo -e "${RED}Could not extract registry ID${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi
    CREATED_REGISTRIES+=("$REGISTRY_ID")
    echo -e "${GREEN}Container Registry created: $REGISTRY_ID${NC}\n"

    echo -e "${GREEN}[LIST]${NC} Listing Container Registries..."
    $ACLOUD_CMD container containerregistry list 2>&1 | head -10
    echo ""

    echo "Waiting for registry $REGISTRY_ID to be Active..."
    local registry_ready=0
    wait_for_status "$ACLOUD_CMD container containerregistry get $REGISTRY_ID" '^(Active|Ready)$' 600 && registry_ready=1
    echo ""

    echo -e "${GREEN}[GET]${NC} Getting registry details..."
    $ACLOUD_CMD container containerregistry get "$REGISTRY_ID" 2>&1
    echo ""

    if [ "$registry_ready" -eq 1 ]; then
        echo -e "${GREEN}[UPDATE]${NC} Updating Container Registry..."
        $ACLOUD_CMD container containerregistry update "$REGISTRY_ID" \
            --name "${registry_name}-updated" \
            --tags "e2e-test,registry,updated" \
            --concurrent-users 20 2>&1 || true
        echo ""
    else
        echo -e "${YELLOW}[UPDATE]${NC} Skipping — registry did not reach Active after 600s."
        echo ""
    fi

    echo -e "${GREEN}✓ Container Registry CRUD test completed!${NC}\n"
}

# --- Test --output flag --------------------------------------------------
test_output_formats() {
    echo -e "${BLUE}--- Testing container list --output flag ---${NC}"
    for resource_cmd in "kaas list" "containerregistry list"; do
        local label="container $resource_cmd"
        local cmd="$ACLOUD_CMD container $resource_cmd"
        for fmt in json yaml; do
            echo -e "${YELLOW}Testing $label --output $fmt...${NC}"
            OUT=$(eval "$cmd --output $fmt" 2>&1)
            validate_list_output "$label" "$fmt" "$OUT" $?
        done
    done
    echo ""
}

# --- Main ----------------------------------------------------------------
ensure_project || { echo -e "${RED}Cannot proceed without a project ID${NC}"; exit 1; }
setup_context

echo -e "${BLUE}Starting Container Resources E2E Tests...${NC}\n"

test_kaas               || FAILURES=$((FAILURES + 1))
test_containerregistry  || FAILURES=$((FAILURES + 1))
test_output_formats

echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Project ID: $PROJECT_ID"
if [ ${#CREATED_CLUSTERS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ KaaS Clusters: ${#CREATED_CLUSTERS[@]} created${NC}"
else
    echo -e "${YELLOW}○ KaaS Clusters: 0 created${NC}"
fi
if [ ${#CREATED_REGISTRIES[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Container Registries: ${#CREATED_REGISTRIES[@]} created${NC}"
else
    echo -e "${YELLOW}○ Container Registries: 0 created${NC}"
fi
echo ""

if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}=== Container E2E: all checks passed ===${NC}"
    exit 0
else
    echo -e "${RED}=== Container E2E: $FAILURES check(s) failed ===${NC}"
    exit 1
fi
