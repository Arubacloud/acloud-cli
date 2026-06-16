#!/bin/bash

# E2E Test Script for Database Resources
# Tests CRUD operations for DBaaS, DBaaS Databases, DBaaS Users, and Database Backups.
# Fully self-contained: bootstraps VPC/Subnet/SG/ElasticIP when env-var URIs are absent
# and tears them down in reverse order on exit.

# Don't exit on error - we want to continue and show summary
# set -e  # Exit on error

# shellcheck source=../common.sh
source "$(dirname "${BASH_SOURCE[0]}")/../common.sh"

# --- State tracking -------------------------------------------------------
CREATED_DBAAS=()
CREATED_DATABASES=()
CREATED_USERS=()
CREATED_GRANTS=0   # count only; API does not return grant IDs in responses
CREATED_BACKUPS=()
DBAAS_ID=""

# Bootstrapped dep IDs (empty = pre-supplied; do not delete on exit)
BOOTSTRAP_PROJECT_ID=""
BOOTSTRAP_VPC_ID=""
BOOTSTRAP_SUBNET_ID=""
BOOTSTRAP_SG_ID=""
BOOTSTRAP_ELASTIC_IP_ID=""

# Resolved dep IDs (populated by ensure_* functions)
VPC_ID="${ACLOUD_VPC_ID:-}"
SUBNET_ID="${ACLOUD_SUBNET_ID:-}"
SG_ID="${ACLOUD_SECURITY_GROUP_ID:-}"
ELASTIC_IP_ID="${ACLOUD_ELASTIC_IP_ID:-}"

# Engine/flavor/zone/storage: prefer explicit env vars, then hardcoded defaults.
ENGINE_ID="${ACLOUD_ENGINE_ID:-mysql-8.0}"
FLAVOR="${ACLOUD_FLAVOR:-DBO4A8}"
ZONE="${ACLOUD_ZONE:-ITBG-1}"
STORAGE_SIZE="${ACLOUD_STORAGE_SIZE:-50}"

print_banner "Database"

# --- Cleanup --------------------------------------------------------------
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test resources...${NC}"

    # Delete database backups
    for backup_id in "${CREATED_BACKUPS[@]}"; do
        if is_valid_id "$backup_id"; then
            echo "Deleting backup: $backup_id"
            $ACLOUD_CMD database backup delete "$backup_id" --yes 2>&1 || true
        fi
    done

    # When grants were created, individual user/database deletes will always 409
    # ("user has granted access" / "users have access to this database").
    # The DBaaS cascade delete handles all sub-resources atomically, so skip
    # the individual deletes when grants exist to avoid confusing error noise.
    if [ "$CREATED_GRANTS" -gt 0 ]; then
        echo "Skipping individual user/database deletes — grants exist; DBaaS cascade delete will remove them"
    else
        # Delete DBaaS users
        if [ -n "$DBAAS_ID" ] && is_valid_id "$DBAAS_ID"; then
            for user in "${CREATED_USERS[@]}"; do
                echo "Deleting user: $user"
                $ACLOUD_CMD database dbaas user delete "$DBAAS_ID" "$user" --yes 2>&1 || true
            done
        fi

        # Delete DBaaS databases
        if [ -n "$DBAAS_ID" ] && is_valid_id "$DBAAS_ID"; then
            for db in "${CREATED_DATABASES[@]}"; do
                echo "Deleting database: $db"
                $ACLOUD_CMD database dbaas database delete "$DBAAS_ID" "$db" --yes 2>&1 || true
            done
        fi
    fi

    # Delete DBaaS instances — wait for Active before delete to avoid 400
    for dbaas_id in "${CREATED_DBAAS[@]}"; do
        if is_valid_id "$dbaas_id"; then
            echo "Waiting for DBaaS $dbaas_id before delete..."
            wait_for_status "$ACLOUD_CMD database dbaas get $dbaas_id" '^(Active|Ready)$' 300 2>/dev/null || true
            echo "Deleting DBaaS instance: $dbaas_id"
            $ACLOUD_CMD database dbaas delete "$dbaas_id" --yes 2>&1 || true
            # Poll until the DBaaS is fully gone before removing networking deps
            local wait_del=0
            echo "  Waiting for DBaaS $dbaas_id to be fully removed..."
            while [ "$wait_del" -lt 300 ]; do
                $ACLOUD_CMD database dbaas get "$dbaas_id" >/dev/null 2>&1 || { echo "  → DBaaS $dbaas_id removed"; break; }
                sleep 10
                wait_del=$((wait_del + 10))
            done
        fi
    done

    # Bootstrapped infra in reverse dep order
    if [ -n "$BOOTSTRAP_ELASTIC_IP_ID" ]; then
        echo "Deleting bootstrapped elastic IP: $BOOTSTRAP_ELASTIC_IP_ID"
        $ACLOUD_CMD network elasticip delete "$BOOTSTRAP_ELASTIC_IP_ID" --yes 2>&1 || true
    fi
    if [ -n "$BOOTSTRAP_SG_ID" ] && [ -n "$BOOTSTRAP_VPC_ID" ]; then
        echo "Deleting bootstrapped security group: $BOOTSTRAP_SG_ID"
        $ACLOUD_CMD network securitygroup delete "$BOOTSTRAP_VPC_ID" "$BOOTSTRAP_SG_ID" --yes 2>&1 || true
    fi
    if [ -n "$BOOTSTRAP_SUBNET_ID" ] && [ -n "$BOOTSTRAP_VPC_ID" ]; then
        echo "Deleting bootstrapped subnet: $BOOTSTRAP_SUBNET_ID"
        wait_for_status "$ACLOUD_CMD network subnet get $BOOTSTRAP_VPC_ID $BOOTSTRAP_SUBNET_ID" '^(Active|Ready)$' 60 2>/dev/null || true
        $ACLOUD_CMD network subnet delete "$BOOTSTRAP_VPC_ID" "$BOOTSTRAP_SUBNET_ID" --yes 2>&1 || true
        wait_for_removal "$ACLOUD_CMD network subnet get $BOOTSTRAP_VPC_ID $BOOTSTRAP_SUBNET_ID" 120 2>/dev/null || true
    fi
    if [ -n "$BOOTSTRAP_VPC_ID" ]; then
        echo "Deleting bootstrapped VPC: $BOOTSTRAP_VPC_ID"
        local vpc_del_elapsed=0
        while [ "$vpc_del_elapsed" -lt 120 ]; do
            vpc_out=$($ACLOUD_CMD network vpc delete "$BOOTSTRAP_VPC_ID" --yes 2>&1)
            if [ $? -eq 0 ]; then echo "$vpc_out"; break; fi
            # 404 means already gone — stop retrying immediately
            if echo "$vpc_out" | grep -qi "404\|Not Found"; then echo "$vpc_out"; break; fi
            echo "$vpc_out"
            sleep 15
            vpc_del_elapsed=$((vpc_del_elapsed + 15))
        done
        wait_for_removal "$ACLOUD_CMD network vpc get $BOOTSTRAP_VPC_ID" 120 2>/dev/null || true
    fi

    # Sweep for DBaaS instances that exist in the project but weren't tracked —
    # e.g. a failed create call that returned 400 but still persisted the resource.
    # This must run before the project delete or the project will stay blocked.
    if [ -n "$BOOTSTRAP_PROJECT_ID" ] || [ "$PROJECT_ID" != "your-project-id" ]; then
        local sweep_id="${BOOTSTRAP_PROJECT_ID:-$PROJECT_ID}"
        local leftover_ids
        leftover_ids=$($ACLOUD_CMD database dbaas list --output table-json 2>/dev/null \
            | python3 -c "
import json,sys
try:
    d=json.load(sys.stdin)
    for item in d:
        v=item.get('id','')
        if v: print(v)
except: pass
" 2>/dev/null || true)
        for lid in $leftover_ids; do
            if is_valid_id "$lid"; then
                echo "  Sweeping untracked DBaaS instance: $lid"
                wait_for_status "$ACLOUD_CMD database dbaas get $lid" '^(Active|Ready|Failed|Error)$' 120 2>/dev/null || true
                $ACLOUD_CMD database dbaas delete "$lid" --yes 2>&1 || true
                wait_for_removal "$ACLOUD_CMD database dbaas get $lid" 300 2>/dev/null || true
            fi
        done
    fi

    # Delete bootstrapped project last (retry — DBaaS and VPC deletions are async)
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
    if [ -n "$VPC_ID" ]; then
        echo "  → using pre-supplied VPC: $VPC_ID"
        return 0
    fi
    echo "Bootstrapping VPC for database suite..."
    local out
    out=$($ACLOUD_CMD network vpc create \
        --name "${RESOURCE_PREFIX}-db-vpc" \
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
    echo "Bootstrapping Subnet for database suite..."
    local _ts="${RESOURCE_PREFIX##*-}"
    local cidr="10.$(( (_ts % 200) + 10 )).1.0/24"
    local out
    out=$($ACLOUD_CMD network subnet create "$VPC_ID" \
        --name "${RESOURCE_PREFIX}-db-subnet" \
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
    echo "Bootstrapping Security Group for database suite..."
    local out
    out=$($ACLOUD_CMD network securitygroup create "$VPC_ID" \
        --name "${RESOURCE_PREFIX}-db-sg" \
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

ensure_elastic_ip() {
    if [ -n "$ELASTIC_IP_ID" ]; then
        echo "  → using pre-supplied Elastic IP: $ELASTIC_IP_ID"
        return 0
    fi
    echo "Bootstrapping Elastic IP for database suite..."
    local out
    out=$($ACLOUD_CMD network elasticip create \
        --name "${RESOURCE_PREFIX}-db-eip" \
        --region "$REGION" \
        --billing-period Hour \
        --tags "e2e-test" 2>&1) || { echo -e "${RED}Elastic IP create failed: $out${NC}"; return 1; }
    local eip_id
    eip_id=$(extract_id "$out")
    if [ -z "$eip_id" ] || ! is_valid_id "$eip_id"; then
        echo -e "${RED}Could not extract Elastic IP ID: $out${NC}"; return 1
    fi
    BOOTSTRAP_ELASTIC_IP_ID="$eip_id"
    echo "  → waiting for Elastic IP $eip_id to be Active..."
    wait_for_status "$ACLOUD_CMD network elasticip get $eip_id" '^(Active|NotUsed|Ready)$' 180 || true
    ELASTIC_IP_ID="$eip_id"
    echo "  → Elastic IP $eip_id ready"
}

# --- Test --output flag ---------------------------------------------------
# dbaas list always tested; dbaas database list only when DBAAS_ID is set.
test_output_formats() {
    echo -e "${BLUE}--- Testing database list --output flag ---${NC}"

    local label="database dbaas list"
    local cmd="$ACLOUD_CMD database dbaas list"
    for fmt in json yaml; do
        echo -e "${YELLOW}Testing $label --output $fmt...${NC}"
        OUT=$(eval "$cmd --output $fmt" 2>&1)
        validate_list_output "$label" "$fmt" "$OUT" $?
    done

    if [ -n "$DBAAS_ID" ] && is_valid_id "$DBAAS_ID"; then
        local db_label="database dbaas database list"
        local db_cmd="$ACLOUD_CMD database dbaas database list $DBAAS_ID"
        for fmt in json yaml; do
            echo -e "${YELLOW}Testing $db_label --output $fmt...${NC}"
            OUT=$(eval "$db_cmd --output $fmt" 2>&1)
            validate_list_output "$db_label" "$fmt" "$OUT" $?
        done
    else
        echo -e "${YELLOW}⚠ Skipping 'dbaas database list' format test — no DBaaS instance available${NC}"
    fi
    echo ""
}

# --- Test DBaaS -----------------------------------------------------------
test_dbaas() {
    echo -e "${BLUE}=== 1. DBaaS CRUD Test ===${NC}"

    echo "Resolving database networking dependencies..."
    ensure_vpc            || { echo -e "${RED}VPC bootstrap failed — skipping DBaaS${NC}"; return 1; }
    ensure_subnet         || return 1
    ensure_security_group || return 1
    ensure_elastic_ip     || return 1

    local dbaas_name="${RESOURCE_PREFIX}-dbaas"

    echo -e "${GREEN}[CREATE]${NC} Creating DBaaS instance: $dbaas_name"
    echo "  (zone=$ZONE, flavor=$FLAVOR, engine=$ENGINE_ID, vpc=$VPC_ID)"
    CREATE_OUTPUT=$($ACLOUD_CMD database dbaas create \
        --name "$dbaas_name" \
        --region "$REGION" \
        --zone "$ZONE" \
        --engine-id "$ENGINE_ID" \
        --flavor "$FLAVOR" \
        --storage-size "$STORAGE_SIZE" \
        --vpc-id "$VPC_ID" \
        --subnet-id "$SUBNET_ID" \
        --security-group-id "$SG_ID" \
        --elastic-ip-id "$ELASTIC_IP_ID" \
        --tags "e2e-test,created-by-script" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then
        echo -e "${RED}DBaaS test failed: Authentication error${NC}"
        return 1
    fi

    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        return 1
    fi

    DBAAS_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${RED}Could not extract DBaaS ID from create output${NC}"
        echo "CREATE_OUTPUT: $CREATE_OUTPUT" >&2
        return 1
    fi

    CREATED_DBAAS+=("$DBAAS_ID")
    echo -e "${GREEN}DBaaS created: $DBAAS_ID${NC}"

    echo -e "${GREEN}[LIST]${NC} Listing DBaaS instances"
    $ACLOUD_CMD database dbaas list 2>&1 | head -20

    echo -e "${GREEN}[GET]${NC} Getting DBaaS details: $DBAAS_ID"
    $ACLOUD_CMD database dbaas get "$DBAAS_ID" 2>&1

    echo "Waiting for DBaaS $DBAAS_ID to be Active..."
    local dbaas_ready=0
    wait_for_status "$ACLOUD_CMD database dbaas get $DBAAS_ID" '^(Active|Ready)$' 600 && dbaas_ready=1

    if [ "$dbaas_ready" -eq 1 ]; then
        echo -e "${GREEN}[UPDATE]${NC} Updating DBaaS: $DBAAS_ID"
        UPDATE_OUTPUT=$($ACLOUD_CMD database dbaas update "$DBAAS_ID" \
            --tags "e2e-test,updated" \
            --zone "$ZONE" 2>&1)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}DBaaS updated successfully${NC}"
        else
            echo -e "${YELLOW}Update may have failed${NC}"
            echo "$UPDATE_OUTPUT" | head -5
        fi
    else
        echo -e "${YELLOW}[UPDATE]${NC} Skipping — DBaaS did not reach Active after 600s."
    fi

    echo -e "${GREEN}✓ DBaaS CRUD test completed!${NC}\n"
    return 0
}

# --- Test DBaaS Database --------------------------------------------------
test_dbaas_database() {
    echo -e "${BLUE}=== 2. DBaaS Database CRUD Test ===${NC}"

    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${YELLOW}Skipping DBaaS database test (no DBaaS instance available)${NC}\n"
        return 0
    fi

    # MySQL database names: letters, digits, underscore only (no hyphens)
    local _ts="${RESOURCE_PREFIX##*-}"
    local db_name="e2edb${_ts}"

    echo "Waiting for DBaaS $DBAAS_ID to be Active before database operations..."
    wait_for_status "$ACLOUD_CMD database dbaas get $DBAAS_ID" '^(Active|Ready)$' 180 2>/dev/null || true

    echo -e "${GREEN}[CREATE]${NC} Creating database: $db_name"
    CREATE_OUTPUT=$($ACLOUD_CMD database dbaas database create "$DBAAS_ID" \
        --name "$db_name" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then return 1; fi
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        return 1
    fi

    CREATED_DATABASES+=("$db_name")
    echo -e "${GREEN}Database created: $db_name${NC}"

    echo -e "${GREEN}[LIST]${NC} Listing databases in DBaaS: $DBAAS_ID"
    $ACLOUD_CMD database dbaas database list "$DBAAS_ID" 2>&1

    echo -e "${GREEN}[GET]${NC} Getting database details: $db_name"
    $ACLOUD_CMD database dbaas database get "$DBAAS_ID" "$db_name" 2>&1

    # Database rename (UPDATE) is not supported by the API (returns 405)
    echo -e "${YELLOW}[UPDATE]${NC} Skipping database rename — API does not support PUT on databases (405)"

    echo -e "${GREEN}✓ DBaaS Database CRUD test completed!${NC}\n"
    return 0
}

# --- Test DBaaS User ------------------------------------------------------
test_dbaas_user() {
    echo -e "${BLUE}=== 3. DBaaS User CRUD Test ===${NC}"

    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${YELLOW}Skipping DBaaS user test (no DBaaS instance available)${NC}\n"
        return 0
    fi

    # Username scoped to this DBaaS instance — use the same short name as the
    # Terraform provider example (restapi) which is verified to pass API validation.
    local username="restapi"
    local password="Prova123456789AC@"

    echo "Waiting for DBaaS $DBAAS_ID to be Active before user operations..."
    wait_for_status "$ACLOUD_CMD database dbaas get $DBAAS_ID" '^(Active|Ready)$' 180 2>/dev/null || true

    echo -e "${GREEN}[CREATE]${NC} Creating user: $username"
    CREATE_OUTPUT=$($ACLOUD_CMD database dbaas user create "$DBAAS_ID" \
        --username "$username" \
        --password "$password" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then return 1; fi
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        return 1
    fi

    CREATED_USERS+=("$username")
    echo -e "${GREEN}User created: $username${NC}"

    echo -e "${GREEN}[LIST]${NC} Listing users in DBaaS: $DBAAS_ID"
    $ACLOUD_CMD database dbaas user list "$DBAAS_ID" 2>&1

    echo -e "${GREEN}[GET]${NC} Getting user details: $username"
    $ACLOUD_CMD database dbaas user get "$DBAAS_ID" "$username" 2>&1

    # User password update is not supported by the API (returns 405)
    echo -e "${YELLOW}[UPDATE]${NC} Skipping user password update — API does not support PUT on users (405)"

    echo -e "${GREEN}✓ DBaaS User CRUD test completed!${NC}\n"
    return 0
}

# --- Test DBaaS Grant -----------------------------------------------------
test_dbaas_grant() {
    echo -e "${BLUE}=== 4. DBaaS Grant CRUD Test ===${NC}"

    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${YELLOW}Skipping DBaaS grant test (no DBaaS instance available)${NC}\n"
        return 0
    fi
    if [ ${#CREATED_USERS[@]} -eq 0 ] || [ ${#CREATED_DATABASES[@]} -eq 0 ]; then
        echo -e "${YELLOW}Skipping DBaaS grant test (no user or database available)${NC}\n"
        return 0
    fi

    local username="${CREATED_USERS[0]}"
    local db_name="${CREATED_DATABASES[0]}"
    local role="liteadmin"

    echo "Waiting for DBaaS $DBAAS_ID to be Active before grant operations..."
    wait_for_status "$ACLOUD_CMD database dbaas get $DBAAS_ID" '^(Active|Ready)$' 180 2>/dev/null || true

    echo -e "${GREEN}[CREATE]${NC} Creating grant: $username → $db_name ($role)"
    CREATE_OUTPUT=$($ACLOUD_CMD database dbaas grant create "$DBAAS_ID" "$db_name" \
        --username "$username" \
        --role "$role" 2>&1)
    exit_code=$?

    if ! check_auth_error "$CREATE_OUTPUT"; then return 1; fi
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        return 1
    fi
    CREATED_GRANTS=$((CREATED_GRANTS + 1))
    echo -e "${GREEN}Grant created: $username → $db_name${NC}"

    echo -e "${GREEN}[LIST]${NC} Listing grants on database: $db_name"
    GRANT_LIST_OUTPUT=$($ACLOUD_CMD database dbaas grant list "$DBAAS_ID" "$db_name" 2>&1)
    echo "$GRANT_LIST_OUTPUT"

    # The grant list/create responses do not include an ID field (API limitation),
    # so individual grant DELETE is not possible. DBaaS cascade delete handles cleanup.
    local grant_id
    grant_id=$(echo "$GRANT_LIST_OUTPUT" | awk 'NR>1 && NF>0 {print $1; exit}')
    if [ -n "$grant_id" ] && is_valid_id "$grant_id"; then
        echo -e "${GREEN}[GET]${NC} Getting grant details: $grant_id"
        $ACLOUD_CMD database dbaas grant get "$DBAAS_ID" "$db_name" "$grant_id" 2>&1
    else
        echo -e "${YELLOW}Could not extract grant ID for cleanup (cleanup via user delete)${NC}"
    fi

    # Grant update is not supported by the API
    echo -e "${YELLOW}[UPDATE]${NC} Skipping grant update — API does not support PUT on grants"

    echo -e "${GREEN}✓ DBaaS Grant CRUD test completed!${NC}\n"
    return 0
}

# --- Test Database Backup -------------------------------------------------
test_backup() {
    echo -e "${BLUE}=== 5. Database Backup CRUD Test ===${NC}"

    if [ -z "$DBAAS_ID" ] || ! is_valid_id "$DBAAS_ID"; then
        echo -e "${YELLOW}Skipping backup test (no DBaaS instance available)${NC}\n"
        return 0
    fi

    if [ ${#CREATED_DATABASES[@]} -eq 0 ]; then
        echo -e "${YELLOW}Skipping backup test (no database available)${NC}\n"
        return 0
    fi

    # Backup name: alphanumeric only (no hyphens, same constraint as database/user names)
    local _ts="${RESOURCE_PREFIX##*-}"
    local backup_name="e2ebackup${_ts}"
    local database_name="${CREATED_DATABASES[0]}"

    # The backup service has its own database registry that may lag behind the DBaaS
    # database API. Retry up to 3 times with a 15-second delay on "database not found"
    # errors to handle the propagation window.
    echo -e "${GREEN}[CREATE]${NC} Creating backup: $backup_name"
    local attempt=0 exit_code=1
    while [ $attempt -lt 3 ]; do
        CREATE_OUTPUT=$($ACLOUD_CMD database backup create \
            --name "$backup_name" \
            --region "$REGION" \
            --zone "$ZONE" \
            --dbaas-id "$DBAAS_ID" \
            --database-name "$database_name" \
            --billing-period "Hour" \
            --tags "e2e-test" 2>&1)
        exit_code=$?
        if [ $exit_code -eq 0 ]; then break; fi
        if echo "$CREATE_OUTPUT" | grep -qi "database.*not found\|not found.*database"; then
            attempt=$((attempt + 1))
            if [ $attempt -lt 3 ]; then
                echo -e "${YELLOW}  Backup service hasn't synced database yet — retrying in 15s (attempt $((attempt + 1))/3)...${NC}"
                sleep 15
                continue
            fi
        fi
        break
    done

    if ! check_auth_error "$CREATE_OUTPUT"; then return 1; fi
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}CREATE failed:${NC}"
        echo "$CREATE_OUTPUT" >&2
        return 1
    fi

    BACKUP_ID=$(extract_id "$CREATE_OUTPUT")
    if [ -z "$BACKUP_ID" ] || ! is_valid_id "$BACKUP_ID"; then
        echo -e "${RED}Could not extract backup ID from create output${NC}"
        echo "CREATE_OUTPUT: $CREATE_OUTPUT" >&2
        return 1
    fi

    CREATED_BACKUPS+=("$BACKUP_ID")
    echo -e "${GREEN}Backup created: $BACKUP_ID${NC}"

    echo -e "${GREEN}[LIST]${NC} Listing database backups"
    $ACLOUD_CMD database backup list 2>&1 | head -20

    echo -e "${GREEN}[GET]${NC} Getting backup details: $BACKUP_ID"
    $ACLOUD_CMD database backup get "$BACKUP_ID" 2>&1

    echo -e "${GREEN}✓ Database Backup CRUD test completed!${NC}\n"
    return 0
}

# --- Main -----------------------------------------------------------------
ensure_project || { echo -e "${RED}Cannot proceed without a project ID${NC}"; exit 1; }
setup_context

echo -e "${BLUE}Starting Database Resources E2E Tests...${NC}\n"

test_dbaas             || FAILURES=$((FAILURES + 1))
test_dbaas_database    || FAILURES=$((FAILURES + 1))
test_dbaas_user        || FAILURES=$((FAILURES + 1))
test_dbaas_grant       || FAILURES=$((FAILURES + 1))
test_backup            || FAILURES=$((FAILURES + 1))
test_output_formats

# Test summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Project ID: $PROJECT_ID"
if [ -n "$DBAAS_ID" ]; then
    echo "DBaaS ID: $DBAAS_ID"
fi
if [ ${#CREATED_DBAAS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ DBaaS Instances: ${#CREATED_DBAAS[@]} created${NC}"
else
    echo -e "${YELLOW}○ DBaaS Instances: 0 created${NC}"
fi
if [ ${#CREATED_DATABASES[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Databases: ${#CREATED_DATABASES[@]} created${NC}"
else
    echo -e "${YELLOW}○ Databases: 0 created${NC}"
fi
if [ ${#CREATED_USERS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Users: ${#CREATED_USERS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Users: 0 created${NC}"
fi
if [ "$CREATED_GRANTS" -gt 0 ]; then
    echo -e "${GREEN}✓ Grants: $CREATED_GRANTS created${NC}"
else
    echo -e "${YELLOW}○ Grants: 0 created${NC}"
fi
if [ ${#CREATED_BACKUPS[@]} -gt 0 ]; then
    echo -e "${GREEN}✓ Backups: ${#CREATED_BACKUPS[@]} created${NC}"
else
    echo -e "${YELLOW}○ Backups: 0 created${NC}"
fi
echo ""

if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}=== Database E2E: all checks passed ===${NC}"
    exit 0
else
    echo -e "${RED}=== Database E2E: $FAILURES check(s) failed ===${NC}"
    exit 1
fi
