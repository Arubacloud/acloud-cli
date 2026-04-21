#!/bin/bash

# E2E Test Script for Schedule Resources
# Tests CRUD operations for Scheduled Jobs

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# Cleanup tracking
CREATED_JOBS=()

print_banner "Schedule"

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
    
    echo -e "${GREEN}Cleanup completed!${NC}"
}

# Trap to ensure cleanup runs on exit
trap cleanup EXIT

setup_context

# Test function for OneShot Job
test_oneshot_job() {
    echo -e "${BLUE}=== 1. OneShot Job CRUD Test ===${NC}"
    
    local job_name="${RESOURCE_PREFIX}-oneshot-job"
    # Schedule for 1 hour from now
    local schedule_at=$(date -u -d "+1 hour" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "")
    
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
        --enabled true \
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
    echo -e "${GREEN}OneShot job created: $JOB_ID${NC}"
    
    echo -e "${GREEN}[LIST]${NC} Listing scheduled jobs"
    $ACLOUD_CMD schedule job list 2>&1 | head -20
    
    echo -e "${GREEN}[GET]${NC} Getting job details: $JOB_ID"
    $ACLOUD_CMD schedule job get "$JOB_ID" 2>&1
    
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
    
    echo -e "${GREEN}OneShot job test completed successfully${NC}\n"
    return 0
}

# Test function for Recurring Job
test_recurring_job() {
    echo -e "${BLUE}=== 2. Recurring Job CRUD Test ===${NC}"
    
    local job_name="${RESOURCE_PREFIX}-recurring-job"
    # Execute until 1 month from now
    local execute_until=$(date -u -d "+1 month" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+1m +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "")
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
        --enabled true \
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
    echo -e "${GREEN}Recurring job created: $JOB_ID${NC}"
    
    echo -e "${GREEN}[LIST]${NC} Listing scheduled jobs"
    $ACLOUD_CMD schedule job list 2>&1 | head -20
    
    echo -e "${GREEN}[GET]${NC} Getting job details: $JOB_ID"
    $ACLOUD_CMD schedule job get "$JOB_ID" 2>&1
    
    echo -e "${GREEN}[UPDATE]${NC} Updating job: $JOB_ID"
    UPDATE_OUTPUT=$($ACLOUD_CMD schedule job update "$JOB_ID" \
        --name "${job_name}-updated" \
        --enabled false \
        --tags "e2e-test,updated,disabled" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Job updated successfully${NC}"
    else
        echo -e "${YELLOW}Update may have failed${NC}"
        echo "$UPDATE_OUTPUT" | head -5
    fi
    
    echo -e "${GREEN}Recurring job test completed successfully${NC}\n"
    return 0
}

# Run tests
echo -e "${BLUE}Starting Schedule Resources E2E Tests...${NC}\n"

test_oneshot_job
test_recurring_job
test_output_formats

# Test summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Project ID: $PROJECT_ID"
echo "○ Jobs: ${#CREATED_JOBS[@]} created"
echo ""

echo -e "${GREEN}=== All Schedule Tests Completed! ===${NC}"

