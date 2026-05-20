#!/bin/bash

# E2E Test Script for Compute Resources
# Tests CRUD operations for Cloud Servers and Key Pairs.
# Fully self-contained: bootstraps VPC/Subnet/SG when env-var URIs are absent
# and tears them down in reverse order on exit.

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# --- State tracking -------------------------------------------------------
CREATED_SERVERS=()
CREATED_KEYPAIRS=()
EPHEMERAL_KEY_PATH=""
EPHEMERAL_KEY2_PATH=""

# Bootstrapped dep IDs (empty = pre-supplied; do not delete on exit)
BOOTSTRAP_PROJECT_ID=""
BOOTSTRAP_VPC_ID=""
BOOTSTRAP_SUBNET_ID=""
BOOTSTRAP_SG_ID=""
BOOTSTRAP_BOOT_DISK_ID=""

# Resolved dep URIs (populated by ensure_* functions)
VPC_URI="${ACLOUD_VPC_URI:-}"
SUBNET_URI="${ACLOUD_SUBNET_URI:-}"
SG_URI="${ACLOUD_SECURITY_GROUP_URI:-}"
BOOT_DISK_URI="${ACLOUD_BOOT_DISK_URI:-}"

print_banner "Compute"

# --- Cleanup --------------------------------------------------------------
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test resources...${NC}"

    # Cloud servers first (they reference SG/subnet/boot-disk)
    for id in "${CREATED_SERVERS[@]}"; do
        if is_valid_id "$id"; then
            echo "Waiting for server $id before delete..."
            wait_for_status "$ACLOUD_CMD compute cloudserver get $id" '^(Active|PoweredOff)$' 120 2>/dev/null || true
            echo "Deleting cloud server: $id"
            $ACLOUD_CMD compute cloudserver delete "$id" --yes 2>&1 || true
        fi
    done

    # Keypairs
    for name in "${CREATED_KEYPAIRS[@]}"; do
        echo "Deleting keypair: $name"
        $ACLOUD_CMD compute keypair delete "$name" --yes 2>&1 || true
    done

    # Bootstrapped infra in reverse dep order
    if [ -n "$BOOTSTRAP_BOOT_DISK_ID" ]; then
        echo "Deleting bootstrapped boot disk: $BOOTSTRAP_BOOT_DISK_ID"
        $ACLOUD_CMD storage blockstorage delete "$BOOTSTRAP_BOOT_DISK_ID" --yes 2>&1 || true
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
        while [ "$vpc_del_elapsed" -lt 120 ]; do
            $ACLOUD_CMD network vpc delete "$BOOTSTRAP_VPC_ID" --yes 2>&1 && break
            sleep 15
            vpc_del_elapsed=$((vpc_del_elapsed + 15))
        done
    fi

    # Ephemeral key files
    [ -n "$EPHEMERAL_KEY_PATH" ]  && rm -f "$EPHEMERAL_KEY_PATH"  "${EPHEMERAL_KEY_PATH}.pub"
    [ -n "$EPHEMERAL_KEY2_PATH" ] && rm -f "$EPHEMERAL_KEY2_PATH" "${EPHEMERAL_KEY2_PATH}.pub"

    # Delete bootstrapped project last (after all child resources)
    # Retry because VPC deletion is async and the project may still report
    # child resources for a short window after the VPC delete call returns.
    if [ -n "$BOOTSTRAP_PROJECT_ID" ]; then
        echo "Deleting bootstrapped project: $BOOTSTRAP_PROJECT_ID"
        local proj_del_elapsed=0
        while [ "$proj_del_elapsed" -lt 120 ]; do
            $ACLOUD_CMD management project delete "$BOOTSTRAP_PROJECT_ID" --yes 2>&1 && break
            sleep 15
            proj_del_elapsed=$((proj_del_elapsed + 15))
        done
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
    echo "Bootstrapping VPC for compute suite..."
    local out
    out=$($ACLOUD_CMD network vpc create \
        --name "${RESOURCE_PREFIX}-compute-vpc" \
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
    echo "Bootstrapping Subnet for compute suite..."
    local _ts="${RESOURCE_PREFIX##*-}"
    local cidr="10.$(( (_ts % 200) + 10 )).0.0/24"
    local out
    out=$($ACLOUD_CMD network subnet create "$vpc_id" \
        --name "${RESOURCE_PREFIX}-compute-subnet" \
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

ensure_boot_disk() {
    if [ -n "$BOOT_DISK_URI" ]; then
        echo "  → using pre-supplied boot disk: $BOOT_DISK_URI"
        return 0
    fi
    echo "Bootstrapping boot disk for compute suite (image LU20-001)..."
    local out
    out=$($ACLOUD_CMD storage blockstorage create \
        --name "${RESOURCE_PREFIX}-compute-bootdisk" \
        --size 20 \
        --region "$REGION" \
        --zone "${ACLOUD_ZONE:-ITBG-1}" \
        --image "LU20-001" \
        --set-bootable \
        --billing-period Hour \
        --tags "e2e-test" 2>&1) || { echo -e "${RED}Boot disk create failed: $out${NC}"; return 1; }
    local disk_id
    disk_id=$(extract_id "$out")
    if [ -z "$disk_id" ] || ! is_valid_id "$disk_id"; then
        echo -e "${RED}Could not extract boot disk ID: $out${NC}"; return 1
    fi
    BOOTSTRAP_BOOT_DISK_ID="$disk_id"
    echo "  → waiting for boot disk $disk_id to be Active..."
    wait_for_status "$ACLOUD_CMD storage blockstorage get $disk_id" '^(Active|NotUsed)$' 300 || {
        echo -e "${RED}Boot disk did not become Active${NC}"; return 1
    }
    local uri_line
    uri_line=$($ACLOUD_CMD storage blockstorage get "$disk_id" 2>/dev/null | grep -i "^URI:" | awk '{print $2}')
    # API returns camelCase blockStorages in the URI; fall back to that if grep found nothing
    BOOT_DISK_URI="${uri_line:-/projects/$PROJECT_ID/providers/Aruba.Storage/blockStorages/$disk_id}"
    echo "  → boot disk $disk_id ready (URI: $BOOT_DISK_URI)"
}

ensure_security_group() {
    if [ -n "$SG_URI" ]; then
        echo "  → using pre-supplied Security Group: $SG_URI"
        return 0
    fi
    local vpc_id="${VPC_URI##*/}"
    echo "Bootstrapping Security Group for compute suite..."
    local out
    out=$($ACLOUD_CMD network securitygroup create "$vpc_id" \
        --name "${RESOURCE_PREFIX}-compute-sg" \
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

# --- Test Cloud Server ---------------------------------------------------
test_cloudserver() {
    local server_name="${RESOURCE_PREFIX}-server"
    echo -e "${BLUE}=== 1. Cloud Server CRUD Test ===${NC}"

    echo "Resolving compute dependencies..."
    ensure_vpc            || { echo -e "${RED}VPC bootstrap failed — skipping cloudserver${NC}"; return 1; }
    ensure_subnet         || return 1
    ensure_security_group || return 1
    ensure_boot_disk      || return 1

    echo -e "${GREEN}[CREATE]${NC} Creating cloud server: $server_name"
    echo "  (zone=${ACLOUD_ZONE:-ITBG-1}, flavor=${ACLOUD_FLAVOR:-small}, boot-disk-uri=$BOOT_DISK_URI)"
    CREATE_OUTPUT=$($ACLOUD_CMD --debug compute cloudserver create \
        --name "$server_name" \
        --region "$REGION" \
        --zone "${ACLOUD_ZONE:-ITBG-1}" \
        --flavor "${ACLOUD_FLAVOR:-small}" \
        --vpc-uri "$VPC_URI" \
        --subnet-uri "$SUBNET_URI" \
        --security-group-uri "$SG_URI" \
        --boot-disk-uri "$BOOT_DISK_URI" \
        --billing-period Hour \
        --tags "e2e-test,compute" 2>&1)
    CREATE_EXIT=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then return 1; fi
    if [ $CREATE_EXIT -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi

    SERVER_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$SERVER_ID" ] || ! is_valid_id "$SERVER_ID"; then
        echo -e "${RED}Could not extract server ID${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi
    CREATED_SERVERS+=("$SERVER_ID")
    echo -e "${GREEN}Cloud server created: $SERVER_ID${NC}\n"

    echo -e "${GREEN}[LIST]${NC} Listing cloud servers..."
    $ACLOUD_CMD compute cloudserver list 2>&1 | head -10
    echo ""

    echo "Waiting for server $SERVER_ID to be Active..."
    local server_ready=0
    wait_for_status "$ACLOUD_CMD compute cloudserver get $SERVER_ID" '^(Active|Running)$' 300 && server_ready=1
    echo ""

    echo -e "${GREEN}[GET]${NC} Getting cloud server details..."
    $ACLOUD_CMD compute cloudserver get "$SERVER_ID" 2>&1
    echo ""

    if [ "$server_ready" -eq 1 ]; then
        echo -e "${GREEN}[UPDATE]${NC} Updating cloud server..."
        $ACLOUD_CMD compute cloudserver update "$SERVER_ID" \
            --name "${server_name}-updated" 2>&1 || true
        echo ""

        echo -e "${GREEN}[POWER-OFF]${NC} Powering off server..."
        $ACLOUD_CMD compute cloudserver power-off "$SERVER_ID" 2>&1 || true
        echo ""

        echo -e "${GREEN}[POWER-ON]${NC} Powering on server..."
        $ACLOUD_CMD compute cloudserver power-on "$SERVER_ID" 2>&1 || true
        echo ""
    else
        echo -e "${YELLOW}[UPDATE/POWER OPS]${NC} Skipping — server did not reach Active after 300s."
        echo ""
    fi

    echo -e "${GREEN}✓ Cloud Server CRUD test completed!${NC}\n"
}

# --- Test Key Pair -------------------------------------------------------
test_keypair() {
    local keypair_name="${RESOURCE_PREFIX}-keypair"
    echo -e "${BLUE}=== 2. Key Pair CRUD Test ===${NC}"

    local pubkey
    if [ -n "${ACLOUD_PUBLIC_KEY:-}" ]; then
        pubkey="$ACLOUD_PUBLIC_KEY"
    else
        EPHEMERAL_KEY_PATH="${TMPDIR:-/tmp}/acloud-e2e-key-$$"
        pubkey=$(gen_ephemeral_pubkey "$EPHEMERAL_KEY_PATH") || {
            echo -e "${YELLOW}⚠ Skipping keypair CREATE — ssh-keygen unavailable and ACLOUD_PUBLIC_KEY unset${NC}"
            echo ""
            echo -e "${GREEN}[LIST]${NC} Listing keypairs..."
            $ACLOUD_CMD compute keypair list 2>&1 | head -10
            echo ""
            return 0
        }
    fi

    echo -e "${GREEN}[CREATE]${NC} Creating keypair: $keypair_name"
    CREATE_OUTPUT=$($ACLOUD_CMD compute keypair create \
        --name "$keypair_name" \
        --region "$REGION" \
        --public-key "$pubkey" 2>&1)
    CREATE_EXIT=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then return 1; fi
    if [ $CREATE_EXIT -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi
    KEYPAIR_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$KEYPAIR_ID" ] || ! is_valid_id "$KEYPAIR_ID"; then
        echo -e "${RED}Could not extract keypair ID${NC}"
        echo "$CREATE_OUTPUT"
        return 1
    fi
    CREATED_KEYPAIRS+=("$KEYPAIR_ID")
    echo -e "${GREEN}Keypair created: $keypair_name (ID: $KEYPAIR_ID)${NC}\n"

    echo -e "${GREEN}[LIST]${NC} Listing keypairs..."
    $ACLOUD_CMD compute keypair list 2>&1 | head -10
    echo ""

    echo -e "${GREEN}[GET]${NC} Getting keypair details..."
    $ACLOUD_CMD compute keypair get "$KEYPAIR_ID" 2>&1
    echo ""

    # UPDATE — generate a second ephemeral key
    EPHEMERAL_KEY2_PATH="${TMPDIR:-/tmp}/acloud-e2e-key2-$$"
    local update_pubkey
    update_pubkey=$(gen_ephemeral_pubkey "$EPHEMERAL_KEY2_PATH") && {
        echo -e "${GREEN}[UPDATE]${NC} Updating keypair with new public key..."
        $ACLOUD_CMD compute keypair update "$keypair_name" \
            --public-key "$update_pubkey" 2>&1 || true
        echo ""
    } || {
        echo -e "${YELLOW}[UPDATE]${NC} Skipping — could not generate second key"
        echo ""
    }

    echo -e "${GREEN}✓ Key Pair CRUD test completed!${NC}\n"
}

# --- Test --output flag --------------------------------------------------
test_output_formats() {
    echo -e "${BLUE}--- Testing compute list --output flag ---${NC}"
    for resource in "cloudserver" "keypair"; do
        local label="compute $resource list"
        local cmd="$ACLOUD_CMD compute $resource list"
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

echo -e "${BLUE}Starting Compute Resources E2E Tests...${NC}\n"

test_cloudserver || FAILURES=$((FAILURES + 1))
test_keypair     || FAILURES=$((FAILURES + 1))
test_output_formats

echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Project ID: $PROJECT_ID"
if [ ${#CREATED_SERVERS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Cloud Servers: ${#CREATED_SERVERS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Cloud Servers: 0 created${NC}"
fi
if [ ${#CREATED_KEYPAIRS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Keypairs: ${#CREATED_KEYPAIRS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Keypairs: 0 created${NC}"
fi
echo ""

if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}=== Compute E2E: all checks passed ===${NC}"
    exit 0
else
    echo -e "${RED}=== Compute E2E: $FAILURES check(s) failed ===${NC}"
    exit 1
fi
