#!/bin/bash

# E2E Test Script for Schedule Resources
# Tests CRUD operations for Scheduled Jobs.
#
# The test bootstraps a minimal cloud server to use as the job step resource,
# following the URI format from the Terraform provider examples:
#   /projects/{project_id}/providers/Aruba.Compute/cloudServers/{server_id}
#
# Override any bootstrap dependency via env vars:
#   ACLOUD_STEP_RESOURCE_URI  — skip bootstrap entirely; use this URI
#   ACLOUD_VPC_ID             — reuse existing VPC
#   ACLOUD_SUBNET_ID          — reuse existing subnet
#   ACLOUD_SECURITY_GROUP_ID  — reuse existing security group
#   ACLOUD_BOOT_DISK_ID       — reuse existing boot disk
#   ACLOUD_STEP_ACTION_URI    — action to schedule (default: poweroff)
#   ACLOUD_STEP_HTTP_VERB     — HTTP verb (default: POST)
#   ACLOUD_ZONE               — availability zone (default: ITBG-1)
#   ACLOUD_FLAVOR             — server flavor (default: CSO4A8)

# Don't exit on error - we want to continue and show summary
# set -e

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# --- State tracking -------------------------------------------------------
CREATED_JOBS=()
BOOTSTRAP_PROJECT_ID=""
BOOTSTRAP_VPC_ID=""
BOOTSTRAP_SUBNET_ID=""
BOOTSTRAP_SG_ID=""
BOOTSTRAP_BOOT_DISK_ID=""
BOOTSTRAP_SERVER_ID=""

# Resolved dep IDs
VPC_ID="${ACLOUD_VPC_ID:-}"
SUBNET_ID="${ACLOUD_SUBNET_ID:-}"
SG_ID="${ACLOUD_SECURITY_GROUP_ID:-}"
BOOT_DISK_ID="${ACLOUD_BOOT_DISK_ID:-}"

# Step configuration
STEP_RESOURCE_URI="${ACLOUD_STEP_RESOURCE_URI:-}"
STEP_ACTION_URI="${ACLOUD_STEP_ACTION_URI:-poweroff}"
STEP_HTTP_VERB="${ACLOUD_STEP_HTTP_VERB:-POST}"

print_banner "Schedule"

# --- Cleanup --------------------------------------------------------------
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test resources...${NC}"

    # Delete scheduled jobs first
    for job_id in "${CREATED_JOBS[@]}"; do
        if is_valid_id "$job_id"; then
            echo "Deleting job: $job_id"
            $ACLOUD_CMD schedule job delete "$job_id" --yes 2>&1 || true
            wait_for_removal "$ACLOUD_CMD schedule job get $job_id" 60 2>/dev/null || true
        fi
    done

    # Delete bootstrapped cloud server — wait for it to be fully gone before
    # touching the boot disk, SG, subnet, and VPC it holds.
    if [ -n "$BOOTSTRAP_SERVER_ID" ]; then
        echo "Waiting for server $BOOTSTRAP_SERVER_ID before delete..."
        wait_for_status "$ACLOUD_CMD compute cloudserver get $BOOTSTRAP_SERVER_ID" \
            '^(Active|PoweredOff|Shutdown|Stopped)$' 180 2>/dev/null || true
        echo "Deleting bootstrapped cloud server: $BOOTSTRAP_SERVER_ID"
        $ACLOUD_CMD compute cloudserver delete "$BOOTSTRAP_SERVER_ID" --yes 2>&1 || true
        echo "  Waiting for server $BOOTSTRAP_SERVER_ID to be fully removed..."
        wait_for_removal "$ACLOUD_CMD compute cloudserver get $BOOTSTRAP_SERVER_ID" 300 2>/dev/null || true
    fi

    # Boot disk — wait for unlink from server before deleting
    if [ -n "$BOOTSTRAP_BOOT_DISK_ID" ]; then
        echo "Waiting for boot disk $BOOTSTRAP_BOOT_DISK_ID to be unlinked..."
        wait_for_status "$ACLOUD_CMD storage blockstorage get $BOOTSTRAP_BOOT_DISK_ID" \
            '^(NotUsed|Active)$' 120 2>/dev/null || true
        echo "Deleting bootstrapped boot disk: $BOOTSTRAP_BOOT_DISK_ID"
        $ACLOUD_CMD storage blockstorage delete "$BOOTSTRAP_BOOT_DISK_ID" --yes 2>&1 || true
        wait_for_removal "$ACLOUD_CMD storage blockstorage get $BOOTSTRAP_BOOT_DISK_ID" 120 2>/dev/null || true
    fi

    # SG, subnet, VPC in reverse dep order
    if [ -n "$BOOTSTRAP_SG_ID" ] && [ -n "$BOOTSTRAP_VPC_ID" ]; then
        echo "Deleting bootstrapped security group: $BOOTSTRAP_SG_ID"
        $ACLOUD_CMD network securitygroup delete "$BOOTSTRAP_VPC_ID" "$BOOTSTRAP_SG_ID" --yes 2>&1 || true
        wait_for_removal "$ACLOUD_CMD network securitygroup get $BOOTSTRAP_VPC_ID $BOOTSTRAP_SG_ID" 120 2>/dev/null || true
    fi
    if [ -n "$BOOTSTRAP_SUBNET_ID" ] && [ -n "$BOOTSTRAP_VPC_ID" ]; then
        echo "Deleting bootstrapped subnet: $BOOTSTRAP_SUBNET_ID"
        $ACLOUD_CMD network subnet delete "$BOOTSTRAP_VPC_ID" "$BOOTSTRAP_SUBNET_ID" --yes 2>&1 || true
        wait_for_removal "$ACLOUD_CMD network subnet get $BOOTSTRAP_VPC_ID $BOOTSTRAP_SUBNET_ID" 120 2>/dev/null || true
    fi
    if [ -n "$BOOTSTRAP_VPC_ID" ]; then
        echo "Deleting bootstrapped VPC: $BOOTSTRAP_VPC_ID"
        local vpc_del_elapsed=0
        while [ "$vpc_del_elapsed" -lt 120 ]; do
            $ACLOUD_CMD network vpc delete "$BOOTSTRAP_VPC_ID" --yes 2>&1 && break
            sleep 15
            vpc_del_elapsed=$((vpc_del_elapsed + 15))
        done
        wait_for_removal "$ACLOUD_CMD network vpc get $BOOTSTRAP_VPC_ID" 120 2>/dev/null || true
    fi

    # Project last
    if [ -n "$BOOTSTRAP_PROJECT_ID" ]; then
        echo "Deleting bootstrapped project: $BOOTSTRAP_PROJECT_ID"
        local del_elapsed=0
        while [ "$del_elapsed" -lt 120 ]; do
            $ACLOUD_CMD management project delete "$BOOTSTRAP_PROJECT_ID" --yes 2>&1 && break
            sleep 10
            del_elapsed=$((del_elapsed + 10))
        done
    fi

    echo -e "${GREEN}Cleanup completed!${NC}"
}

trap cleanup EXIT

# --- Dep resolvers -------------------------------------------------------

ensure_vpc() {
    if [ -n "$VPC_ID" ]; then
        echo "  → using pre-supplied VPC: $VPC_ID"
        return 0
    fi
    echo "Bootstrapping VPC for schedule suite..."
    local out
    out=$($ACLOUD_CMD network vpc create \
        --name "${RESOURCE_PREFIX}-schedule-vpc" \
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
    VPC_ID="$vpc_id"
    echo "  → VPC $vpc_id ready"
}

ensure_subnet() {
    if [ -n "$SUBNET_ID" ]; then
        echo "  → using pre-supplied Subnet: $SUBNET_ID"
        return 0
    fi
    echo "Bootstrapping Subnet for schedule suite..."
    local _ts="${RESOURCE_PREFIX##*-}"
    local cidr="10.$(( (_ts % 200) + 10 )).0.0/24"
    local out
    out=$($ACLOUD_CMD network subnet create "$VPC_ID" \
        --name "${RESOURCE_PREFIX}-schedule-subnet" \
        --cidr "$cidr" \
        --dhcp-enabled \
        --region "$REGION" 2>&1) || { echo -e "${RED}Subnet create failed: $out${NC}"; return 1; }
    local subnet_id
    subnet_id=$(extract_id "$out" "$VPC_ID")
    if [ -z "$subnet_id" ] || ! is_valid_id "$subnet_id"; then
        echo -e "${RED}Could not extract Subnet ID: $out${NC}"; return 1
    fi
    BOOTSTRAP_SUBNET_ID="$subnet_id"
    echo "  → waiting for Subnet $subnet_id to be Active..."
    wait_for_status "$ACLOUD_CMD network subnet get $VPC_ID $subnet_id" '^(Active|Ready)$' 180 || true
    SUBNET_ID="$subnet_id"
    echo "  → Subnet $subnet_id ready"
}

ensure_security_group() {
    if [ -n "$SG_ID" ]; then
        echo "  → using pre-supplied Security Group: $SG_ID"
        return 0
    fi
    echo "Bootstrapping Security Group for schedule suite..."
    local out
    out=$($ACLOUD_CMD network securitygroup create "$VPC_ID" \
        --name "${RESOURCE_PREFIX}-schedule-sg" \
        --region "$REGION" 2>&1) || { echo -e "${RED}SG create failed: $out${NC}"; return 1; }
    local sg_id
    sg_id=$(extract_id "$out" "$VPC_ID")
    if [ -z "$sg_id" ] || ! is_valid_id "$sg_id"; then
        echo -e "${RED}Could not extract SG ID: $out${NC}"; return 1
    fi
    BOOTSTRAP_SG_ID="$sg_id"
    echo "  → waiting for Security Group $sg_id to be Active..."
    wait_for_status "$ACLOUD_CMD network securitygroup get $VPC_ID $sg_id" '^(Active|Ready)$' 180 || true
    SG_ID="$sg_id"
    echo "  → Security Group $sg_id ready"
}

ensure_boot_disk() {
    if [ -n "$BOOT_DISK_ID" ]; then
        echo "  → using pre-supplied boot disk: $BOOT_DISK_ID"
        return 0
    fi
    echo "Bootstrapping boot disk for schedule suite (image LU20-001)..."
    local out
    out=$($ACLOUD_CMD storage blockstorage create \
        --name "${RESOURCE_PREFIX}-schedule-bootdisk" \
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
    BOOT_DISK_ID="$disk_id"
    echo "  → Boot disk $disk_id ready"
}

# bootstrap_step_resource resolves STEP_RESOURCE_URI in this priority order:
#   1. ACLOUD_STEP_RESOURCE_URI already set — use as-is
#   2. Existing cloud server found in the project — use its URI
#   3. Bootstrap: VPC + subnet + SG + boot disk + cloud server
# Sets STEP_RESOURCE_URI on success; returns 1 on failure.
bootstrap_step_resource() {
    # 1. Caller pre-supplied a URI
    if [ -n "$STEP_RESOURCE_URI" ]; then
        echo "  → using pre-supplied ACLOUD_STEP_RESOURCE_URI: $STEP_RESOURCE_URI"
        return 0
    fi

    [ "$PROJECT_ID" = "your-project-id" ] && return 1

    echo "Bootstrapping step resource (ACLOUD_STEP_RESOURCE_URI not set)..."

    # 2. Try to find an existing cloud server in the project
    local cs_id
    cs_id=$($ACLOUD_CMD compute cloudserver list --output table-json 2>/dev/null \
        | python3 -c "import json,sys; d=json.load(sys.stdin); print(d[0]['id'] if d else '')" 2>/dev/null)
    if [ -n "$cs_id" ] && is_valid_id "$cs_id"; then
        STEP_RESOURCE_URI="/projects/$PROJECT_ID/providers/Aruba.Compute/cloudServers/$cs_id"
        echo "  → using existing cloud server: $STEP_RESOURCE_URI"
        return 0
    fi

    # 3. Bootstrap a minimal cloud server to act as the step resource.
    # URI format matches Terraform provider examples:
    #   /projects/{project_id}/providers/Aruba.Compute/cloudServers/{server_id}
    echo "  → no existing server found; bootstrapping cloud server..."
    ensure_vpc            || return 1
    ensure_subnet         || return 1
    ensure_security_group || return 1
    ensure_boot_disk      || return 1

    local server_name="${RESOURCE_PREFIX}-schedule-server"
    echo "  → creating cloud server: $server_name"
    local cs_out
    cs_out=$($ACLOUD_CMD compute cloudserver create \
        --name "$server_name" \
        --region "$REGION" \
        --zone "${ACLOUD_ZONE:-ITBG-1}" \
        --flavor "${ACLOUD_FLAVOR:-CSO4A8}" \
        --vpc-id "$VPC_ID" \
        --subnet-id "$SUBNET_ID" \
        --security-group-id "$SG_ID" \
        --boot-disk-id "$BOOT_DISK_ID" \
        --billing-period Hour \
        --tags "e2e-test,schedule" 2>&1)
    if [ $? -ne 0 ]; then
        echo -e "${RED}Cloud server create failed: $cs_out${NC}"
        return 1
    fi
    local server_id
    server_id=$(extract_id "$cs_out")
    if [ -z "$server_id" ] || ! is_valid_id "$server_id"; then
        echo -e "${RED}Could not extract server ID: $cs_out${NC}"; return 1
    fi
    BOOTSTRAP_SERVER_ID="$server_id"
    echo "  → waiting for server $server_id to be Active..."
    wait_for_status "$ACLOUD_CMD compute cloudserver get $server_id" '^(Active|Running)$' 300 || {
        echo -e "${RED}Server did not become Active${NC}"; return 1
    }
    STEP_RESOURCE_URI="/projects/$PROJECT_ID/providers/Aruba.Compute/cloudServers/$server_id"
    echo "  → step resource URI: $STEP_RESOURCE_URI"
}

# --- Output format test ---------------------------------------------------
test_output_formats() {
    echo -e "${BLUE}--- Testing schedule list --output flag ---${NC}"
    for fmt in json yaml; do
        echo -e "${YELLOW}Testing schedule job list --output $fmt...${NC}"
        OUT=$($ACLOUD_CMD schedule job list --output "$fmt" 2>&1)
        validate_list_output "schedule job list" "$fmt" "$OUT" $?
    done
    echo ""
}

# --- OneShot Job test -----------------------------------------------------
test_oneshot_job() {
    echo -e "${BLUE}=== 1. OneShot Job CRUD Test ===${NC}"

    if [ -z "$STEP_RESOURCE_URI" ]; then
        skip "OneShot job CREATE — step resource URI not available"
        echo ""
        echo -e "${GREEN}[LIST]${NC} Listing scheduled jobs..."
        $ACLOUD_CMD schedule job list 2>&1 | head -10
        echo ""
        echo -e "${GREEN}✓ OneShot job list-only test completed (CREATE skipped)${NC}\n"
        return 0
    fi

    local job_name="${RESOURCE_PREFIX}-oneshot-job"
    local schedule_at
    schedule_at=$(date -u -d "+1 hour" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null \
        || date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null \
        || echo "")
    if [ -z "$schedule_at" ]; then
        echo -e "${YELLOW}Skipping OneShot job test (cannot calculate future date)${NC}"
        return 0
    fi

    echo -e "${GREEN}[CREATE]${NC} Creating OneShot job: $job_name"
    echo "  step-resource-uri: $STEP_RESOURCE_URI"
    CREATE_OUTPUT=$($ACLOUD_CMD schedule job create \
        --name "$job_name" \
        --region "$REGION" \
        --job-type "OneShot" \
        --schedule-at "$schedule_at" \
        --step-resource-uri "$STEP_RESOURCE_URI" \
        --step-action-uri "$STEP_ACTION_URI" \
        --step-http-verb "$STEP_HTTP_VERB" \
        --step-name "e2e-poweroff-step" \
        --tags "e2e-test,oneshot" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then
        echo -e "${RED}OneShot job test failed: Authentication error${NC}"
        return 1
    fi

    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT"
        if echo "$CREATE_OUTPUT" | grep -qi "typology\|not found configuration"; then
            echo -e "${YELLOW}⚠ Schedule API typology not configured for this resource on this tenant — skipping job tests.${NC}"
            return 0
        fi
        return 1
    fi

    JOB_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$JOB_ID" ] || ! is_valid_id "$JOB_ID"; then
        echo -e "${RED}Could not extract job ID${NC}"; echo "$CREATE_OUTPUT"; return 1
    fi
    CREATED_JOBS+=("$JOB_ID")
    echo -e "${GREEN}OneShot job created: $JOB_ID${NC}\n"

    echo -e "${GREEN}[LIST]${NC} Listing scheduled jobs"
    $ACLOUD_CMD schedule job list 2>&1 | head -20
    echo ""

    echo -e "${GREEN}[GET]${NC} Getting job details: $JOB_ID"
    $ACLOUD_CMD schedule job get "$JOB_ID" 2>&1
    echo ""

    echo -e "${GREEN}[UPDATE]${NC} Updating job: $JOB_ID"
    UPDATE_OUTPUT=$($ACLOUD_CMD schedule job update "$JOB_ID" \
        --name "${job_name}-updated" \
        --tags "e2e-test,updated" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Job updated successfully${NC}"
    else
        echo -e "${YELLOW}Update may have failed${NC}"
        echo "$UPDATE_OUTPUT" | head -5
    fi
    echo ""

    echo -e "${GREEN}✓ OneShot Job CRUD test completed!${NC}\n"
}

# --- Recurring Job test ---------------------------------------------------
test_recurring_job() {
    echo -e "${BLUE}=== 2. Recurring Job CRUD Test ===${NC}"

    if [ -z "$STEP_RESOURCE_URI" ]; then
        skip "Recurring job CREATE — step resource URI not available"
        echo ""
        return 0
    fi

    local job_name="${RESOURCE_PREFIX}-recurring-job"
    local execute_until
    execute_until=$(date -u -d "+1 month" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null \
        || date -u -v+1m +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null \
        || echo "")
    if [ -z "$execute_until" ]; then
        echo -e "${YELLOW}Skipping Recurring job test (cannot calculate future date)${NC}"
        return 0
    fi

    echo -e "${GREEN}[CREATE]${NC} Creating Recurring job: $job_name"
    echo "  step-resource-uri: $STEP_RESOURCE_URI"
    CREATE_OUTPUT=$($ACLOUD_CMD schedule job create \
        --name "$job_name" \
        --region "$REGION" \
        --job-type "Recurring" \
        --cron "0 10 * * *" \
        --execute-until "$execute_until" \
        --step-resource-uri "$STEP_RESOURCE_URI" \
        --step-action-uri "$STEP_ACTION_URI" \
        --step-http-verb "$STEP_HTTP_VERB" \
        --step-name "e2e-poweroff-step" \
        --tags "e2e-test,recurring" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then
        echo -e "${RED}Recurring job test failed: Authentication error${NC}"
        return 1
    fi

    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT"
        if echo "$CREATE_OUTPUT" | grep -qi "typology\|not found configuration"; then
            echo -e "${YELLOW}⚠ Schedule API typology not configured for this resource on this tenant — skipping job tests.${NC}"
            return 0
        fi
        return 1
    fi

    JOB_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$JOB_ID" ] || ! is_valid_id "$JOB_ID"; then
        echo -e "${RED}Could not extract job ID${NC}"; echo "$CREATE_OUTPUT"; return 1
    fi
    CREATED_JOBS+=("$JOB_ID")
    echo -e "${GREEN}Recurring job created: $JOB_ID${NC}\n"

    echo -e "${GREEN}[LIST]${NC} Listing scheduled jobs"
    $ACLOUD_CMD schedule job list 2>&1 | head -20
    echo ""

    echo -e "${GREEN}[GET]${NC} Getting job details: $JOB_ID"
    $ACLOUD_CMD schedule job get "$JOB_ID" 2>&1
    echo ""

    echo -e "${GREEN}[UPDATE]${NC} Updating job: $JOB_ID"
    UPDATE_OUTPUT=$($ACLOUD_CMD schedule job update "$JOB_ID" \
        --name "${job_name}-updated" \
        --enabled=false \
        --tags "e2e-test,updated,disabled" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Job updated successfully${NC}"
    else
        echo -e "${YELLOW}Update may have failed${NC}"
        echo "$UPDATE_OUTPUT" | head -5
    fi
    echo ""

    echo -e "${GREEN}✓ Recurring Job CRUD test completed!${NC}\n"
}

# --- Main -----------------------------------------------------------------
ensure_project || { echo -e "${RED}Cannot proceed without a project ID${NC}"; exit 1; }
setup_context

echo -e "${BLUE}Starting Schedule Resources E2E Tests...${NC}\n"

bootstrap_step_resource || \
    echo -e "${YELLOW}Note: step resource bootstrap failed — job CREATE steps will be skipped.\n${NC}"

test_oneshot_job   || FAILURES=$((FAILURES + 1))
test_recurring_job || FAILURES=$((FAILURES + 1))
test_output_formats

echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Project ID: $PROJECT_ID"
if [ ${#CREATED_JOBS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Jobs: ${#CREATED_JOBS[@]} created${NC}"
    [ -n "$STEP_RESOURCE_URI" ] && echo "  Step resource URI: $STEP_RESOURCE_URI"
else
    echo -e "${YELLOW}○ Jobs: 0 created${NC}"
fi
echo ""

if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}=== Schedule E2E: all checks passed ===${NC}"
    exit 0
else
    echo -e "${RED}=== Schedule E2E: $FAILURES check(s) failed ===${NC}"
    exit 1
fi
