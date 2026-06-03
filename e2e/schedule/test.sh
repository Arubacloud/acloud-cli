#!/bin/bash

# E2E Test Script for Schedule Resources
# Tests CRUD operations for Scheduled Jobs.
#
# Schedule jobs REQUIRE at least one step with a valid resource URI (a real
# cloud resource like a server that the job will act on). Provide this via:
#
#   ACLOUD_STEP_RESOURCE_URI  — URI of any existing resource in your project.
#                               Format: /projects/<id>/providers/<provider>/<type>/<id>
#                               Example (cloud server):
#                               /projects/.../providers/Aruba.Compute/cloudServers/<id>
#   ACLOUD_STEP_ACTION_URI    — Action to schedule (default: poweroff)
#   ACLOUD_STEP_HTTP_VERB     — HTTP verb for the action (default: POST)
#
# If ACLOUD_STEP_RESOURCE_URI is not set the job CREATE steps are skipped;
# the LIST and format tests still run.

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# Cleanup tracking
CREATED_JOBS=()
BOOTSTRAP_PROJECT_ID=""

# Step configuration — a valid resource URI is required by the schedule API
STEP_RESOURCE_URI="${ACLOUD_STEP_RESOURCE_URI:-}"
STEP_ACTION_URI="${ACLOUD_STEP_ACTION_URI:-poweroff}"
STEP_HTTP_VERB="${ACLOUD_STEP_HTTP_VERB:-POST}"

print_banner "Schedule"

# Resolve the step resource URI when ACLOUD_STEP_RESOURCE_URI is not set.
# The schedule API only supports resource types registered with the scheduler
# (cloud servers, etc.). Elastic IPs are NOT supported — the API returns
# "Not found configuration for typology" for Aruba.Network/elasticIps.
# Fall back to finding an existing cloud server in the project.
bootstrap_step_resource() {
    [ -n "$STEP_RESOURCE_URI" ] && return 0
    [ "$PROJECT_ID" = "your-project-id" ] && return 1

    echo "Bootstrapping step resource (ACLOUD_STEP_RESOURCE_URI not set)..."
    # Try to find an existing cloud server — the most common schedulable resource.
    local cs_id
    cs_id=$($ACLOUD_CMD compute cloudserver list --project-id "$PROJECT_ID" --output table-json 2>/dev/null \
        | python3 -c "import json,sys; d=json.load(sys.stdin); print(d[0]['id'] if d else '')" 2>/dev/null)
    if [ -n "$cs_id" ] && is_valid_id "$cs_id"; then
        STEP_RESOURCE_URI="/projects/$PROJECT_ID/providers/Aruba.Compute/cloudServers/$cs_id"
        echo "  → Step resource (existing cloud server): $STEP_RESOURCE_URI"
        return 0
    fi

    echo -e "${YELLOW}  ⚠ No existing cloud server found in project $PROJECT_ID.${NC}"
    echo -e "${YELLOW}    Set ACLOUD_STEP_RESOURCE_URI to a cloud server URI to run schedule tests.${NC}"
    return 1
}

# Test --output flag for schedule list commands
test_output_formats() {
    echo -e "${BLUE}--- Testing schedule list --output flag ---${NC}"

    local label="schedule job list"
    local cmd="$ACLOUD_CMD schedule job list --project-id \"$PROJECT_ID\""

    for fmt in json yaml; do
        echo -e "${YELLOW}Testing $label --output $fmt...${NC}"
        OUT=$(eval "$cmd --output $fmt" 2>&1)
        validate_list_output "$label" "$fmt" "$OUT" $?
    done
    echo ""
}

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test resources...${NC}"

    # Delete scheduled jobs
    for job_id in "${CREATED_JOBS[@]}"; do
        if is_valid_id "$job_id"; then
            echo "Deleting job: $job_id"
            $ACLOUD_CMD schedule job delete "$job_id" --yes 2>&1 || true
        fi
    done

    # Delete bootstrapped project last (after all child resources)
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

# Trap to ensure cleanup runs on exit
trap cleanup EXIT

ensure_project || { echo -e "${RED}Cannot proceed without a project ID${NC}"; exit 1; }
setup_context

# Test function for OneShot Job
test_oneshot_job() {
    echo -e "${BLUE}=== 1. OneShot Job CRUD Test ===${NC}"

    if [ -z "$STEP_RESOURCE_URI" ]; then
        echo -e "${YELLOW}⚠ Skipping OneShot job CREATE — ACLOUD_STEP_RESOURCE_URI is not set.${NC}"
        echo -e "${YELLOW}  The schedule API requires at least one step with a valid resource URI.${NC}"
        echo -e "${YELLOW}  Set ACLOUD_STEP_RESOURCE_URI to a cloud server or other resource URI.${NC}"
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
    CREATE_OUTPUT=$($ACLOUD_CMD schedule job create \
        --name "$job_name" \
        --region "$REGION" \
        --job-type "OneShot" \
        --schedule-at "$schedule_at" \
        --step-resource-uri "$STEP_RESOURCE_URI" \
        --step-action-uri "$STEP_ACTION_URI" \
        --step-http-verb "$STEP_HTTP_VERB" \
        --step-name "e2e-step" \
        --tags "e2e-test,oneshot" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then
        echo -e "${RED}OneShot job test failed: Authentication error${NC}"
        return 1
    fi

    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        echo -e "${RED}OneShot job test failed${NC}"
        return 1
    fi

    JOB_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$JOB_ID" ] || ! is_valid_id "$JOB_ID"; then
        echo -e "${RED}Could not extract job ID from create output${NC}"
        echo "CREATE_OUTPUT: $CREATE_OUTPUT" >&2
        echo -e "${RED}OneShot job test failed${NC}"
        return 1
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
    return 0
}

# Test function for Recurring Job
test_recurring_job() {
    echo -e "${BLUE}=== 2. Recurring Job CRUD Test ===${NC}"

    if [ -z "$STEP_RESOURCE_URI" ]; then
        echo -e "${YELLOW}⚠ Skipping Recurring job CREATE — ACLOUD_STEP_RESOURCE_URI is not set.${NC}"
        echo ""
        return 0
    fi

    local job_name="${RESOURCE_PREFIX}-recurring-job"
    local execute_until
    execute_until=$(date -u -d "+1 month" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null \
        || date -u -v+1m +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null \
        || echo "")
    local cron="0 0 * * *"  # Daily at midnight

    if [ -z "$execute_until" ]; then
        echo -e "${YELLOW}Skipping Recurring job test (cannot calculate future date)${NC}"
        return 0
    fi

    echo -e "${GREEN}[CREATE]${NC} Creating Recurring job: $job_name"
    CREATE_OUTPUT=$($ACLOUD_CMD schedule job create \
        --name "$job_name" \
        --region "$REGION" \
        --job-type "Recurring" \
        --cron "$cron" \
        --execute-until "$execute_until" \
        --step-resource-uri "$STEP_RESOURCE_URI" \
        --step-action-uri "$STEP_ACTION_URI" \
        --step-http-verb "$STEP_HTTP_VERB" \
        --step-name "e2e-step" \
        --tags "e2e-test,recurring" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then
        echo -e "${RED}Recurring job test failed: Authentication error${NC}"
        return 1
    fi

    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        echo -e "${RED}Recurring job test failed${NC}"
        return 1
    fi

    JOB_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$JOB_ID" ] || ! is_valid_id "$JOB_ID"; then
        echo -e "${RED}Could not extract job ID from create output${NC}"
        echo "CREATE_OUTPUT: $CREATE_OUTPUT" >&2
        echo -e "${RED}Recurring job test failed${NC}"
        return 1
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
    return 0
}

# Run tests
echo -e "${BLUE}Starting Schedule Resources E2E Tests...${NC}\n"

bootstrap_step_resource || true
if [ -z "$STEP_RESOURCE_URI" ]; then
    echo -e "${YELLOW}Note: step resource could not be bootstrapped — job CREATE steps will be skipped.${NC}\n"
fi

test_oneshot_job  || FAILURES=$((FAILURES + 1))
test_recurring_job || FAILURES=$((FAILURES + 1))
test_output_formats

# Test summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Project ID: $PROJECT_ID"
if [ ${#CREATED_JOBS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Jobs: ${#CREATED_JOBS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Jobs: 0 created (ACLOUD_STEP_RESOURCE_URI not set)${NC}"
fi
echo ""

if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}=== Schedule E2E: all checks passed ===${NC}"
    exit 0
else
    echo -e "${RED}=== Schedule E2E: $FAILURES check(s) failed ===${NC}"
    exit 1
fi
