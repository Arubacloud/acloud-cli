# VPC Network (Virtual Private Cloud Network)

Virtual Private Clouds provide isolated network environments for your cloud resources.

## Commands

### List VPCs

List all VPCs in your project.

```bash
acloud network vpc list [flags]
```

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
# List VPCs using context
acloud network vpc list

# List VPCs with explicit project ID
acloud network vpc list --project-id 68398923fb2cb026400d4d31
```

**Output:**
```
NAME            ID                        SUBNETS    STATUS
production-vpc  689307f4745108d3c6343b5a  4          Active
test-vpc        69485a584d0cdc87949b6ff8  0          InCreation
```

### Get VPC Details

Get detailed information about a specific VPC.

```bash
acloud network vpc get <vpc-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network vpc get 689307f4745108d3c6343b5a
```

**Output:**
```
VPC Details:
============
ID:              689307f4745108d3c6343b5a
URI:             /projects/.../vpcs/689307f4745108d3c6343b5a
Name:            production-vpc
Default:         false
Linked Resources: 4
Creation Date:   06-08-2025 07:44:52
Created By:      aru-297647
Tags:            [production network critical]
Status:          Active
```

### Create VPC

Create a new VPC.

```bash
acloud network vpc create [flags]
```

**Required Flags:**
- `--name string` - Name for the VPC
- `--region string` - Region code (e.g., ITBG-Bergamo)

**Optional Flags:**
- `--tags strings` - Tags for the VPC (comma-separated)
- `--project-id string` - Project ID (uses context if not specified)

**Examples:**
```bash
# Create a basic VPC
acloud network vpc create --name "my-vpc" --region ITBG-Bergamo

# Create VPC with tags
acloud network vpc create \
  --name "production-vpc" \
  --region ITBG-Bergamo \
  --tags production,network,critical

# Create VPC with explicit project ID
acloud network vpc create \
  --name "my-vpc" \
  --region ITBG-Bergamo \
  --project-id 68398923fb2cb026400d4d31
```

**Output:**
```
VPC created successfully!
ID:      69485a584d0cdc87949b6ff8
Name:    my-vpc
Default: false
```

**Notes:**
- VPCs are created with `Default: false` and `Preset: false` automatically
- The VPC will be in **InCreation** state initially
- Use `acloud network vpc get <vpc-id>` to check when it becomes **Active**

### Update VPC

Update an existing VPC's name and/or tags.

```bash
acloud network vpc update <vpc-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC (supports auto-completion)

**Flags:**
- `--name string` - New name for the VPC
- `--tags strings` - New tags for the VPC (comma-separated)
- `--project-id string` - Project ID (uses context if not specified)

**Note:** At least one of `--name` or `--tags` must be provided.

**Examples:**
```bash
# Update VPC name
acloud network vpc update 689307f4745108d3c6343b5a --name "new-vpc-name"

# Update VPC tags
acloud network vpc update 689307f4745108d3c6343b5a --tags production,updated,network

# Update both name and tags
acloud network vpc update 689307f4745108d3c6343b5a \
  --name "production-vpc" \
  --tags production,critical,frontend
```

**Output:**
```
VPC updated successfully!
ID:      689307f4745108d3c6343b5a
Name:    production-vpc
Tags:    [production critical frontend]
```

**Restrictions:**
- Cannot update VPCs in **InCreation** state
- Wait for the VPC to reach **Active** state before updating

### Delete VPC

Delete a VPC.

```bash
acloud network vpc delete <vpc-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)
- `-y, --yes` - Skip confirmation prompt

**Examples:**
```bash
# Delete with confirmation prompt
acloud network vpc delete 689307f4745108d3c6343b5a

# Delete without confirmation
acloud network vpc delete 689307f4745108d3c6343b5a --yes

# Delete with explicit project ID
acloud network vpc delete 689307f4745108d3c6343b5a \
  --project-id 68398923fb2cb026400d4d31 \
  --yes
```

**Confirmation Prompt:**
```
Are you sure you want to delete VPC 689307f4745108d3c6343b5a? This action cannot be undone.
Type 'yes' to confirm: yes

VPC 689307f4745108d3c6343b5a deleted successfully!
```

**Notes:**
- Deleted VPCs will show **Deleting** status before being removed
- Ensure no resources are using the VPC before deletion
- Deletion cannot be undone

## Shell Auto-completion

The VPC commands support intelligent auto-completion for VPC IDs:

```bash
# Enable completion (bash)
source <(acloud completion bash)

# Type command and press TAB to see available VPC IDs
acloud network vpc get <TAB>
acloud network vpc update <TAB>
acloud network vpc delete <TAB>
```

Auto-completion shows VPC IDs with their names:
```
689307f4745108d3c6343b5a    production-vpc
69485a584d0cdc87949b6ff8    test-vpc
```

## VPC States

VPCs can be in the following states:

| State | Description | Can Update? | Can Delete? |
|-------|-------------|-------------|-------------|
| InCreation | VPC is being created | ❌ No | ❌ No |
| Active | VPC is ready to use | ✅ Yes | ✅ Yes |
| Deleting | VPC is being deleted | ❌ No | ❌ No |

## VPC Properties

### Default Property

The `Default` property indicates whether a VPC is the default VPC for the project:
- Managed automatically by Aruba Cloud
- Cannot be set by users via CLI
- Only one VPC per project can be default

### Preset Property

The `Preset` property indicates whether a VPC uses preset configurations:
- Always set to `false` for user-created VPCs
- Cannot be modified via CLI

### Linked Resources

VPCs can have linked resources (subnets, network interfaces, etc.):
- Shown as count in list view
- Details visible in get command
- Must be removed before VPC deletion

## Common Workflows

### Creating and Configuring a VPC

```bash
# 1. Create the VPC
VPC_ID=$(acloud network vpc create \
  --name "production-vpc" \
  --region ITBG-Bergamo \
  --tags production | grep "ID:" | awk '{print $2}')

# 2. Wait for creation to complete
while true; do
  STATUS=$(acloud network vpc get $VPC_ID | grep "Status:" | awk '{print $2}')
  if [ "$STATUS" = "Active" ]; then
    break
  fi
  echo "Waiting for VPC to become Active... (current: $STATUS)"
  sleep 5
done

# 3. Update with additional tags
acloud network vpc update $VPC_ID --tags production,critical,network

# 4. Verify configuration
acloud network vpc get $VPC_ID
```

### Managing Multiple VPCs

```bash
# List all VPCs
acloud network vpc list

# Tag VPCs by environment
acloud network vpc update <vpc-id-1> --tags production,backend
acloud network vpc update <vpc-id-2> --tags staging,frontend
acloud network vpc update <vpc-id-3> --tags development,testing

# Get details of all VPCs
for vpc_id in $(acloud network vpc list | tail -n +2 | awk '{print $2}'); do
  echo "=== VPC: $vpc_id ==="
  acloud network vpc get $vpc_id
  echo ""
done
```

### Cleaning Up Test VPCs

```bash
# List all VPCs
acloud network vpc list

# Delete test VPCs (skip confirmation with --yes)
acloud network vpc delete <test-vpc-id> --yes

# Verify deletion
acloud network vpc list
```

## Best Practices

1. **Use Descriptive Names**
   ```bash
   acloud network vpc create --name "prod-backend-vpc" --region ITBG-Bergamo
   ```

2. **Tag by Environment and Purpose**
   ```bash
   acloud network vpc update <vpc-id> --tags production,backend,critical
   ```

3. **Wait for Active State Before Configuration**
   ```bash
   # Check status before updating
   acloud network vpc get <vpc-id>
   # Ensure Status is "Active"
   acloud network vpc update <vpc-id> --name "new-name"
   ```

4. **Use Project Contexts**
   ```bash
   acloud context use prod-project
   acloud network vpc list  # No need for --project-id
   ```

5. **Document VPC Purpose in Tags**
   ```bash
   acloud network vpc create \
     --name "api-vpc" \
     --region ITBG-Bergamo \
     --tags api,public-facing,load-balanced
   ```

## Troubleshooting

### "Cannot update VPC while in InCreation state"

**Problem:** Trying to update a VPC that hasn't finished creating.

**Solution:**
```bash
# Check current status
acloud network vpc get <vpc-id>

# Wait for Status to become "Active"
# Then retry the update
acloud network vpc update <vpc-id> --name "new-name"
```

### "Failed to create VPC - Status: 400"

**Problem:** Invalid region format or missing required fields.

**Solution:**
```bash
# Use correct region format
acloud network vpc create --name "my-vpc" --region ITBG-Bergamo

# Not: --region eu-west-1 (wrong format)
```

### "Error: at least one of --name or --tags must be provided"

**Problem:** Update command called without any changes.

**Solution:**
```bash
# Provide at least one field to update
acloud network vpc update <vpc-id> --name "new-name"
# or
acloud network vpc update <vpc-id> --tags tag1,tag2
```

### VPC Shows as Default but Not Created by Me

**Explanation:** Aruba Cloud automatically creates a default VPC for each project. This is normal and managed by the platform.

## Related Commands

- [Elastic IP](elasticip.md) - Assign public IPs within VPCs
- [Load Balancer](loadbalancer.md) - Distribute traffic within VPCs
- [Context Management](../../installation.md#context-management) - Manage project contexts
