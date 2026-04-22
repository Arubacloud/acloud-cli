# Elastic IP

Elastic IPs are static public IP addresses that can be assigned to your cloud resources.

## Commands

### List Elastic IPs

List all Elastic IPs in your project.

```bash
acloud network elasticip list [flags]
```

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
# List Elastic IPs using context
acloud network elasticip list

# List with explicit project ID
acloud network elasticip list --project-id 68398923fb2cb026400d4d31
```

**Output:**
```
NAME                    ID                        ADDRESS          STATUS
prod-api-ip             68ffa0ddce76e7da20465721  209.227.232.229  InUse
staging-web-ip          69007f71ce76e7da20465a52  95.110.142.229   InUse
dev-test-ip             6908820cf974c5deb5decd6c  209.227.232.182  NotUsed
```

### Get Elastic IP Details

Get detailed information about a specific Elastic IP.

```bash
acloud network elasticip get <elastic-ip-id> [flags]
```

**Arguments:**
- `elastic-ip-id` - The ID of the Elastic IP (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network elasticip get 68ffa0ddce76e7da20465721
```

**Output:**
```
Elastic IP Details:
===================
ID:              68ffa0ddce76e7da20465721
URI:             /projects/.../elasticIps/68ffa0ddce76e7da20465721
Name:            prod-api-ip
Address:         209.227.232.229
Billing Period:  Hour
Linked Resources: 2
Creation Date:   27-10-2025 16:42:05
Created By:      aru-297647
Tags:            [production api public]
Status:          InUse
```

### Create Elastic IP

Create a new Elastic IP.

```bash
acloud network elasticip create [flags]
```

**Required Flags:**
- `--name string` - Name for the Elastic IP
- `--region string` - Region code (e.g., ITBG-Bergamo)
- `--billing-period string` - Billing period: Hour, Month, or Year

**Optional Flags:**
- `--tags strings` - Tags for the Elastic IP (comma-separated)
- `--project-id string` - Project ID (uses context if not specified)

**Examples:**
```bash
# Create hourly-billed Elastic IP
acloud network elasticip create \
  --name "my-elastic-ip" \
  --region ITBG-Bergamo \
  --billing-period Hour

# Create monthly-billed Elastic IP with tags
acloud network elasticip create \
  --name "prod-api-ip" \
  --region ITBG-Bergamo \
  --billing-period Month \
  --tags production,api,public

# Create yearly-billed Elastic IP
acloud network elasticip create \
  --name "long-term-ip" \
  --region ITBG-Bergamo \
  --billing-period Year \
  --tags production,critical
```

**Output:**
```
Elastic IP created successfully!
ID:      69485a704d0cdc87949b6ffe
Name:    my-elastic-ip
```

**Notes:**
- Elastic IPs are in **InCreation** state initially
- The IP address is assigned automatically
- Billing starts once the IP is created

**Billing Period Options:**
- `Hour` - Pay per hour (flexible, higher unit cost)
- `Month` - Monthly commitment (cost-effective for steady usage)
- `Year` - Annual commitment (most cost-effective for long-term use)

### Update Elastic IP

Update an existing Elastic IP's name and/or tags.

```bash
acloud network elasticip update <elastic-ip-id> [flags]
```

**Arguments:**
- `elastic-ip-id` - The ID of the Elastic IP (supports auto-completion)

**Flags:**
- `--name string` - New name for the Elastic IP
- `--tags strings` - New tags for the Elastic IP (comma-separated)
- `--project-id string` - Project ID (uses context if not specified)

**Note:** At least one of `--name` or `--tags` must be provided.

**Examples:**
```bash
# Update Elastic IP name
acloud network elasticip update 68ffa0ddce76e7da20465721 --name "prod-api-gateway"

# Update Elastic IP tags
acloud network elasticip update 68ffa0ddce76e7da20465721 --tags production,api,critical

# Update both name and tags
acloud network elasticip update 68ffa0ddce76e7da20465721 \
  --name "prod-frontend-ip" \
  --tags production,frontend,load-balanced
```

**Output:**
```
Elastic IP updated successfully!
ID:      68ffa0ddce76e7da20465721
Name:    prod-frontend-ip
Tags:    [production frontend load-balanced]
```

**Restrictions:**
- Cannot update Elastic IPs in **InCreation** state
- Cannot change the IP address or billing period after creation
- Wait for **NotUsed** or **InUse** state before updating

### Delete Elastic IP

Delete an Elastic IP.

```bash
acloud network elasticip delete <elastic-ip-id> [flags]
```

**Arguments:**
- `elastic-ip-id` - The ID of the Elastic IP (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)
- `-y, --yes` - Skip confirmation prompt

**Examples:**
```bash
# Delete with confirmation prompt
acloud network elasticip delete 68ffa0ddce76e7da20465721

# Delete without confirmation
acloud network elasticip delete 68ffa0ddce76e7da20465721 --yes

# Delete with explicit project ID
acloud network elasticip delete 68ffa0ddce76e7da20465721 \
  --project-id 68398923fb2cb026400d4d31 \
  --yes
```

**Confirmation Prompt:**
```
Are you sure you want to delete Elastic IP 68ffa0ddce76e7da20465721? This action cannot be undone.
Type 'yes' to confirm: yes

Elastic IP 68ffa0ddce76e7da20465721 deleted successfully!
```

**Notes:**
- Detach Elastic IP from resources before deletion
- Deletion is immediate and cannot be undone
- Billing stops after deletion

## Shell Auto-completion

The Elastic IP commands support intelligent auto-completion for Elastic IP IDs:

```bash
# Enable completion (bash)
source <(acloud completion bash)

# Type command and press TAB to see available Elastic IP IDs
acloud network elasticip get <TAB>
acloud network elasticip update <TAB>
acloud network elasticip delete <TAB>
```

Auto-completion shows Elastic IP IDs with their names:
```
68ffa0ddce76e7da20465721    prod-api-ip
69007f71ce76e7da20465a52    staging-web-ip
6908820cf974c5deb5decd6c    dev-test-ip
```

## Elastic IP States

Elastic IPs can be in the following states:

| State | Description | Can Update? | Can Delete? |
|-------|-------------|-------------|-------------|
| InCreation | IP is being created | ❌ No | ❌ No |
| NotUsed | IP created but not attached | ✅ Yes | ✅ Yes |
| InUse | IP is attached to a resource | ✅ Yes | ⚠️ Detach first |

## Elastic IP Properties

### IP Address

The IP address is assigned automatically by Aruba Cloud:
- Cannot be specified during creation
- Cannot be changed after creation
- Remains constant for the Elastic IP's lifetime
- Released when the Elastic IP is deleted

### Billing Period

The billing period is set during creation and cannot be changed:
- **Hour**: Most flexible, billed hourly
- **Month**: Monthly commitment, better value
- **Year**: Annual commitment, best value

To change billing period, you must delete and recreate the Elastic IP.

### Linked Resources

Elastic IPs can be attached to resources:
- Virtual machines
- Load balancers
- Network interfaces

The `Linked Resources` count shows how many resources are using this IP.

## Common Workflows

### Creating and Configuring an Elastic IP

```bash
# 1. Create the Elastic IP
EIP_ID=$(acloud network elasticip create \
  --name "prod-api-ip" \
  --region ITBG-Bergamo \
  --billing-period Month \
  --tags production | grep "ID:" | awk '{print $2}')

# 2. Wait for creation to complete
while true; do
  STATUS=$(acloud network elasticip get $EIP_ID | grep "Status:" | awk '{print $2}')
  if [ "$STATUS" = "NotUsed" ] || [ "$STATUS" = "InUse" ]; then
    break
  fi
  echo "Waiting for Elastic IP to be ready... (current: $STATUS)"
  sleep 5
done

# 3. Get the assigned IP address
IP_ADDR=$(acloud network elasticip get $EIP_ID | grep "Address:" | awk '{print $2}')
echo "Elastic IP ready: $IP_ADDR"

# 4. Update with additional tags
acloud network elasticip update $EIP_ID --tags production,api,public

# 5. Verify configuration
acloud network elasticip get $EIP_ID
```

### Managing Multiple Elastic IPs

```bash
# List all Elastic IPs
acloud network elasticip list

# Tag IPs by purpose
acloud network elasticip update <eip-id-1> --tags production,api
acloud network elasticip update <eip-id-2> --tags staging,web
acloud network elasticip update <eip-id-3> --tags development,testing

# Get details of all Elastic IPs
for eip_id in $(acloud network elasticip list | tail -n +2 | awk '{print $2}'); do
  echo "=== Elastic IP: $eip_id ==="
  acloud network elasticip get $eip_id
  echo ""
done
```

### Cleaning Up Unused Elastic IPs

```bash
# List all Elastic IPs
acloud network elasticip list

# Identify unused IPs (Status: NotUsed)
acloud network elasticip list | grep "NotUsed"

# Delete unused Elastic IPs
acloud network elasticip delete <unused-eip-id> --yes

# Verify deletion
acloud network elasticip list
```

## Best Practices

1. **Choose Appropriate Billing Period**
   ```bash
   # Short-term testing
   acloud network elasticip create --name "test-ip" --region ITBG-Bergamo --billing-period Hour
   
   # Long-term production
   acloud network elasticip create --name "prod-ip" --region ITBG-Bergamo --billing-period Year
   ```

2. **Use Descriptive Names**
   ```bash
   acloud network elasticip create \
     --name "prod-api-gateway-ip" \
     --region ITBG-Bergamo \
     --billing-period Month
   ```

3. **Tag by Environment and Purpose**
   ```bash
   acloud network elasticip update <eip-id> --tags production,frontend,load-balanced
   ```

4. **Monitor Usage**
   ```bash
   # Check which IPs are in use
   acloud network elasticip list | grep "InUse"
   
   # Check which IPs are not being used
   acloud network elasticip list | grep "NotUsed"
   ```

5. **Clean Up Unused IPs**
   ```bash
   # Unused IPs still incur charges
   # Delete IPs you're not using
   acloud network elasticip delete <unused-eip-id> --yes
   ```

## Cost Optimization

### Billing Period Selection

Choose the right billing period based on expected usage:

| Duration | Recommended Period | Reason |
|----------|-------------------|--------|
| < 1 month | Hour | Flexibility, no commitment |
| 1-12 months | Month | Balance of flexibility and cost |
| > 12 months | Year | Maximum savings |

### Identifying Unused IPs

```bash
# Find all NotUsed IPs
acloud network elasticip list | grep "NotUsed"

# Get details to determine if they're needed
acloud network elasticip get <eip-id>

# Delete if no longer needed
acloud network elasticip delete <eip-id> --yes
```

## Troubleshooting

### "Cannot update Elastic IP while in InCreation state"

**Problem:** Trying to update an Elastic IP that hasn't finished creating.

**Solution:**
```bash
# Check current status
acloud network elasticip get <eip-id>

# Wait for Status to become "NotUsed" or "InUse"
# Then retry the update
acloud network elasticip update <eip-id> --name "new-name"
```

### "Error: --billing-period must be Hour, Month, or Year"

**Problem:** Invalid billing period specified.

**Solution:**
```bash
# Use one of the valid values (case-sensitive)
acloud network elasticip create \
  --name "my-ip" \
  --region ITBG-Bergamo \
  --billing-period Month  # Not "monthly" or "month"
```

### IP Address Not Showing After Creation

**Problem:** Elastic IP created but address field is empty.

**Solution:**
```bash
# Wait a few seconds for IP assignment
sleep 5

# Check again
acloud network elasticip get <eip-id>

# The address should now be populated
```

### Cannot Delete IP Attached to Resource

**Problem:** Elastic IP is in use and cannot be deleted.

**Solution:**
```bash
# Check which resources are using the IP
acloud network elasticip get <eip-id>

# Detach the IP from resources first (via resource management)
# Then delete the Elastic IP
acloud network elasticip delete <eip-id> --yes
```

## Related Commands

- [VPC](vpc.md) - Network isolation for Elastic IPs
- [Load Balancer](loadbalancer.md) - Use Elastic IPs with load balancers
- [Context Management](../../installation.md#context-management) - Manage project contexts
