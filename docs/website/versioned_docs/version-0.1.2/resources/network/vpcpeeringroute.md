# VPC Peering Route

VPC Peering Routes define routing rules for traffic between peered VPCs in Aruba Cloud. These routes control how network traffic is directed between VPCs connected via a VPC Peering connection by specifying local and remote network addresses.

## Commands

### List VPC Peering Routes

List all routes for a specific VPC peering connection.

```bash
acloud network vpcpeeringroute list <vpc-id> <peering-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `peering-id` - The ID of the VPC peering connection

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network vpcpeeringroute list 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204
```

**Output:**
```
NAME            ID                        LOCAL NETWORK     REMOTE NETWORK    STATUS
route-1         1234567890abcdef123456   10.0.1.0/24       10.1.1.0/24       Active
route-2         1234567890abcdef123457   10.0.2.0/24       10.1.2.0/24       Active
```

### Get VPC Peering Route Details

Get detailed information about a specific VPC peering route.

```bash
acloud network vpcpeeringroute get <vpc-id> <peering-id> <route-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `peering-id` - The ID of the VPC peering connection
- `route-id` - The ID of the route (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network vpcpeeringroute get 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456
```

**Output:**
```
VPC Peering Route Details:
==========================
ID:              1234567890abcdef123456
URI:             /projects/.../vpcpeeringroutes/1234567890abcdef123456
Name:            route-1
Local Network:   10.0.1.0/24
Remote Network:  10.1.1.0/24
Billing Period:  Hour
Creation Date:   06-08-2025 07:44:52
Created By:      aru-297647
Tags:            [vpc,peering,production]
Status:          Active
```

### Create VPC Peering Route

Create a new route for a VPC peering connection.

```bash
acloud network vpcpeeringroute create <vpc-id> <peering-id> [flags]
```

**Required Flags:**
- `--name string` - VPC Peering Route name
- `--local-network string` - Local network address in CIDR notation
- `--remote-network string` - Remote network address in CIDR notation

**Optional Flags:**
- `--billing-period string` - Billing period: Hour, Month, Year (default: Hour)
- `--tags strings` - Tags for the VPC peering route (comma-separated)
- `--project-id string` - Project ID (uses context if not specified)
- `-v, --verbose` - Show detailed debug information

**Examples:**
```bash
# Create a basic VPC peering route
acloud network vpcpeeringroute create 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 \
  --name "route-1" \
  --local-network "10.0.1.0/24" \
  --remote-network "10.1.1.0/24"

# Create VPC peering route with billing period and tags
acloud network vpcpeeringroute create 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 \
  --name "production-route" \
  --local-network "10.0.2.0/24" \
  --remote-network "10.1.2.0/24" \
  --billing-period Month \
  --tags "vpc,peering,production"
```

**Output:**
```
NAME            ID                        LOCAL NETWORK     REMOTE NETWORK    STATUS
route-1         1234567890abcdef123456   10.0.1.0/24       10.1.1.0/24       Active
```

**Notes:**
- The VPC peering route will be in **InCreation** state initially
- Use `acloud network vpcpeeringroute get` to check when it becomes **Active**

### Update VPC Peering Route

Update an existing VPC peering route's properties.

```bash
acloud network vpcpeeringroute update <vpc-id> <peering-id> <route-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `peering-id` - The ID of the VPC peering connection
- `route-id` - The ID of the route (supports auto-completion)

**Flags:**
- `--name string` - New name for the VPC peering route
- `--tags strings` - New tags for the VPC peering route (comma-separated)
- `--local-network string` - Local network address in CIDR notation
- `--remote-network string` - Remote network address in CIDR notation
- `--billing-period string` - Billing period: Hour, Month, Year
- `--project-id string` - Project ID (uses context if not specified)

**Note:** At least one field must be provided for update.

**Examples:**
```bash
# Update VPC peering route name
acloud network vpcpeeringroute update 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456 \
  --name "updated-route-1"

# Update local network
acloud network vpcpeeringroute update 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456 \
  --local-network "10.0.3.0/24"

# Update billing period
acloud network vpcpeeringroute update 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456 \
  --billing-period Month

# Update multiple fields
acloud network vpcpeeringroute update 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456 \
  --name "production-route" \
  --local-network "10.0.2.0/24" \
  --remote-network "10.1.2.0/24" \
  --billing-period Month \
  --tags "vpc,peering,production,updated"
```

**Output:**
```
NAME            ID                        LOCAL NETWORK     REMOTE NETWORK    STATUS
production-route 1234567890abcdef123456   10.0.2.0/24       10.1.2.0/24       Active
```

**Restrictions:**
- Cannot update VPC peering routes in **InCreation** state
- Wait for the VPC peering route to reach **Active** state before updating

### Delete VPC Peering Route

Delete a VPC peering route.

```bash
acloud network vpcpeeringroute delete <vpc-id> <peering-id> <route-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `peering-id` - The ID of the VPC peering connection
- `route-id` - The ID of the route (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)
- `-y, --yes` - Skip confirmation prompt

**Examples:**
```bash
# Delete with confirmation prompt
acloud network vpcpeeringroute delete 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456

# Delete without confirmation
acloud network vpcpeeringroute delete 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456 --yes
```

**Confirmation Prompt:**
```
Are you sure you want to delete VPC peering route 1234567890abcdef123456? This action cannot be undone.
Type 'yes' to confirm: yes
```

**Output:**
```
ID                              STATUS
1234567890abcdef123456         deleted
```

**Notes:**
- Deletion cannot be undone
- Ensure the VPC peering connection is not dependent on the route before deletion

## Shell Auto-completion

The VPC Peering Route commands support intelligent auto-completion for route IDs:

```bash
# Enable completion (bash)
source <(acloud completion bash)

# Type command and press TAB to see available route IDs
acloud network vpcpeeringroute get <vpc-id> <peering-id> <TAB>
acloud network vpcpeeringroute update <vpc-id> <peering-id> <TAB>
acloud network vpcpeeringroute delete <vpc-id> <peering-id> <TAB>
```

Auto-completion shows route IDs with their names:
```
1234567890abcdef123456    route-1
1234567890abcdef123457    route-2
```

## VPC Peering Route Properties

### Local Network Address

The local network address (CIDR) represents the network range in the local VPC that should be accessible through the peering connection.

**Examples:**
- `10.0.1.0/24` - Specific subnet in local VPC
- `10.0.0.0/16` - Entire local VPC network range

### Remote Network Address

The remote network address (CIDR) represents the network range in the remote VPC that should be accessible through the peering connection.

**Examples:**
- `10.1.1.0/24` - Specific subnet in remote VPC
- `10.1.0.0/16` - Entire remote VPC network range

### Billing Period

The billing period determines how the VPC peering route is billed:

- **Hour**: Pay-per-hour billing (default)
- **Month**: Monthly billing
- **Year**: Annual billing (best cost savings)

## VPC Peering Route States

VPC peering routes can be in the following states:

| State | Description | Can Update? | Can Delete? |
|-------|-------------|-------------|-------------|
| InCreation | VPC peering route is being created | ❌ No | ❌ No |
| Active | VPC peering route is ready to use | ✅ Yes | ✅ Yes |

## Common Workflows

### Setting Up VPC Peering Routes

```bash
# 1. Create VPC peering (if not exists)
VPC_ID="689307f4745108d3c6343b5a"
PEERING_ID=$(acloud network vpcpeering create $VPC_ID \
  --name "prod-peering" \
  --region ITBG-Bergamo | grep "ID:" | awk '{print $2}')

# 2. Wait for peering to be Active
while true; do
  STATUS=$(acloud network vpcpeering get $VPC_ID $PEERING_ID | grep "Status:" | awk '{print $2}')
  if [ "$STATUS" = "Active" ]; then
    break
  fi
  echo "Waiting for VPC peering to become Active... (current: $STATUS)"
  sleep 5
done

# 3. Create routes for different subnets
acloud network vpcpeeringroute create $VPC_ID $PEERING_ID \
  --name "subnet-1-route" \
  --local-network "10.0.1.0/24" \
  --remote-network "10.1.1.0/24"

acloud network vpcpeeringroute create $VPC_ID $PEERING_ID \
  --name "subnet-2-route" \
  --local-network "10.0.2.0/24" \
  --remote-network "10.1.2.0/24" \
  --billing-period Month

# 4. List all routes
acloud network vpcpeeringroute list $VPC_ID $PEERING_ID
```

### Updating VPC Peering Routes

```bash
VPC_ID="689307f4745108d3c6343b5a"
PEERING_ID="6949666e4d0cdc87949b7204"
ROUTE_ID="1234567890abcdef123456"

# Update local network
acloud network vpcpeeringroute update $VPC_ID $PEERING_ID $ROUTE_ID \
  --local-network "10.0.3.0/24"

# Update remote network
acloud network vpcpeeringroute update $VPC_ID $PEERING_ID $ROUTE_ID \
  --remote-network "10.1.3.0/24"

# Update billing period
acloud network vpcpeeringroute update $VPC_ID $PEERING_ID $ROUTE_ID \
  --billing-period Year

# Update name and tags
acloud network vpcpeeringroute update $VPC_ID $PEERING_ID $ROUTE_ID \
  --name "updated-route" \
  --tags "vpc,peering,production,updated"
```

## Best Practices

1. **Use Descriptive Names**
   ```bash
   --name "vpc1-subnet1-to-vpc2-subnet1"
   --name "production-peering-route"
   ```

2. **Tag Your Routes**
   ```bash
   --tags "vpc,peering,production"
   --tags "vpc,peering,development"
   ```

3. **Plan Network Mappings**
   - Ensure local and remote networks don't overlap
   - Use clear naming conventions for route identification
   - Document network mappings for future reference

4. **Choose Appropriate Billing Period**
   - Use **Hour** for temporary or testing routes
   - Use **Month** for production routes with variable usage
   - Use **Year** for stable, long-term routes (best cost savings)

5. **Wait for Active State**
   ```bash
   # Check status before updating
   acloud network vpcpeeringroute get <vpc-id> <peering-id> <route-id>
   # Ensure Status is "Active"
   acloud network vpcpeeringroute update <vpc-id> <peering-id> <route-id> --name "new-name"
   ```

## Troubleshooting

### "Cannot update VPC peering route while in InCreation state"

**Problem:** Trying to update a VPC peering route that hasn't finished creating.

**Solution:**
```bash
# Check current status
acloud network vpcpeeringroute get <vpc-id> <peering-id> <route-id>

# Wait for Status to become "Active"
# Then retry the update
acloud network vpcpeeringroute update <vpc-id> <peering-id> <route-id> --name "new-name"
```

### "Error: at least one field must be provided for update"

**Problem:** Update command called without any changes.

**Solution:**
```bash
# Provide at least one field to update
acloud network vpcpeeringroute update <vpc-id> <peering-id> <route-id> --name "new-name"
# or
acloud network vpcpeeringroute update <vpc-id> <peering-id> <route-id> --tags tag1,tag2
```

## Related Commands

- [VPC Peering](vpcpeering.md) - Manage VPC peering connections
- [VPC](vpc.md) - Manage VPCs
