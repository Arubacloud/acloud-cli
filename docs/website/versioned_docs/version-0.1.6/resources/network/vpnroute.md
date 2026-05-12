# VPN Tunnel Route

VPN Tunnel Routes define routing rules for traffic flowing through VPN tunnels in Aruba Cloud. These routes control how network traffic is directed between your on-premises network and your VPC via a VPN tunnel by specifying cloud subnet and on-premises subnet CIDRs.

## Commands

### List VPN Tunnel Routes

List all routes for a specific VPN tunnel.

```bash
acloud network vpnroute list <vpn-tunnel-id> [flags]
```

**Arguments:**
- `vpn-tunnel-id` - The ID of the VPN tunnel

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network vpnroute list 1234567890abcdef
```

**Output:**
```
NAME            ID                        CLOUD SUBNET      ONPREM SUBNET     STATUS
route-1         1234567890abcdef123456   10.0.1.0/24       192.168.1.0/24    Active
route-2         1234567890abcdef123457   10.0.2.0/24       192.168.2.0/24    Active
```

### Get VPN Tunnel Route Details

Get detailed information about a specific VPN tunnel route.

```bash
acloud network vpnroute get <vpn-tunnel-id> <route-id> [flags]
```

**Arguments:**
- `vpn-tunnel-id` - The ID of the VPN tunnel
- `route-id` - The ID of the route (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network vpnroute get 1234567890abcdef 1234567890abcdef123456
```

**Output:**
```
VPN Route Details:
==================
ID:              1234567890abcdef123456
URI:             /projects/.../vpnroutes/1234567890abcdef123456
Name:            route-1
Region:          ITBG-Bergamo
Cloud Subnet:    10.0.1.0/24
OnPrem Subnet:   192.168.1.0/24
Creation Date:   06-08-2025 07:44:52
Created By:      aru-297647
Tags:            [vpn,route,production]
Status:          Active
```

### Create VPN Tunnel Route

Create a new route for a VPN tunnel.

```bash
acloud network vpnroute create <vpn-tunnel-id> [flags]
```

**Required Flags:**
- `--name string` - VPN Route name
- `--region string` - Region code (e.g., ITBG-Bergamo)
- `--cloud-subnet string` - CIDR of the cloud subnet
- `--onprem-subnet string` - CIDR of the on-prem subnet

**Optional Flags:**
- `--tags strings` - Tags for the VPN route (comma-separated)
- `--project-id string` - Project ID (uses context if not specified)
- `-v, --verbose` - Show detailed debug information

**Examples:**
```bash
# Create a basic VPN route
acloud network vpnroute create 1234567890abcdef \
  --name "route-1" \
  --region ITBG-Bergamo \
  --cloud-subnet "10.0.1.0/24" \
  --onprem-subnet "192.168.1.0/24"

# Create VPN route with tags
acloud network vpnroute create 1234567890abcdef \
  --name "production-route" \
  --region ITBG-Bergamo \
  --cloud-subnet "10.0.2.0/24" \
  --onprem-subnet "192.168.2.0/24" \
  --tags "vpn,production,network"
```

**Output:**
```
NAME            ID                        CLOUD SUBNET      ONPREM SUBNET     STATUS
route-1         1234567890abcdef123456   10.0.1.0/24       192.168.1.0/24    Active
```

**Notes:**
- The VPN route will be in **InCreation** state initially
- Use `acloud network vpnroute get` to check when it becomes **Active**

### Update VPN Tunnel Route

Update an existing VPN tunnel route's properties.

```bash
acloud network vpnroute update <vpn-tunnel-id> <route-id> [flags]
```

**Arguments:**
- `vpn-tunnel-id` - The ID of the VPN tunnel
- `route-id` - The ID of the route (supports auto-completion)

**Flags:**
- `--name string` - New name for the VPN route
- `--tags strings` - New tags for the VPN route (comma-separated)
- `--cloud-subnet string` - CIDR of the cloud subnet
- `--onprem-subnet string` - CIDR of the on-prem subnet
- `--project-id string` - Project ID (uses context if not specified)

**Note:** At least one field must be provided for update.

**Examples:**
```bash
# Update VPN route name
acloud network vpnroute update 1234567890abcdef 1234567890abcdef123456 \
  --name "updated-route-1"

# Update cloud subnet
acloud network vpnroute update 1234567890abcdef 1234567890abcdef123456 \
  --cloud-subnet "10.0.3.0/24"

# Update multiple fields
acloud network vpnroute update 1234567890abcdef 1234567890abcdef123456 \
  --name "production-route" \
  --cloud-subnet "10.0.2.0/24" \
  --onprem-subnet "192.168.2.0/24" \
  --tags "vpn,production,updated"
```

**Output:**
```
NAME            ID                        CLOUD SUBNET      ONPREM SUBNET     STATUS
production-route 1234567890abcdef123456   10.0.2.0/24       192.168.2.0/24    Active
```

**Restrictions:**
- Cannot update VPN routes in **InCreation** state
- Wait for the VPN route to reach **Active** state before updating

### Delete VPN Tunnel Route

Delete a VPN tunnel route.

```bash
acloud network vpnroute delete <vpn-tunnel-id> <route-id> [flags]
```

**Arguments:**
- `vpn-tunnel-id` - The ID of the VPN tunnel
- `route-id` - The ID of the route (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)
- `-y, --yes` - Skip confirmation prompt

**Examples:**
```bash
# Delete with confirmation prompt
acloud network vpnroute delete 1234567890abcdef 1234567890abcdef123456

# Delete without confirmation
acloud network vpnroute delete 1234567890abcdef 1234567890abcdef123456 --yes
```

**Confirmation Prompt:**
```
Are you sure you want to delete VPN route 1234567890abcdef123456? This action cannot be undone.
Type 'yes' to confirm: yes
```

**Output:**
```
ID                              STATUS
1234567890abcdef123456         deleted
```

**Notes:**
- Deletion cannot be undone
- Ensure the VPN tunnel is not dependent on the route before deletion

## Shell Auto-completion

The VPN Tunnel Route commands support intelligent auto-completion for route IDs:

```bash
# Enable completion (bash)
source <(acloud completion bash)

# Type command and press TAB to see available route IDs
acloud network vpnroute get <vpn-tunnel-id> <TAB>
acloud network vpnroute update <vpn-tunnel-id> <TAB>
acloud network vpnroute delete <vpn-tunnel-id> <TAB>
```

Auto-completion shows route IDs with their names:
```
1234567890abcdef123456    route-1
1234567890abcdef123457    route-2
```

## VPN Route Properties

### Cloud Subnet

The cloud subnet CIDR represents the network range in your VPC that should be accessible through the VPN tunnel.

**Examples:**
- `10.0.1.0/24` - Specific subnet in VPC
- `10.0.0.0/16` - Entire VPC network range

### On-Premises Subnet

The on-premises subnet CIDR represents the network range in your on-premises infrastructure that should be accessible through the VPN tunnel.

**Examples:**
- `192.168.1.0/24` - Specific on-premises subnet
- `192.168.0.0/16` - Entire on-premises network range

## VPN Route States

VPN routes can be in the following states:

| State | Description | Can Update? | Can Delete? |
|-------|-------------|-------------|-------------|
| InCreation | VPN route is being created | ❌ No | ❌ No |
| Active | VPN route is ready to use | ✅ Yes | ✅ Yes |

## Common Workflows

### Setting Up VPN Routes

```bash
# 1. Create VPN tunnel (if not exists)
VPN_TUNNEL_ID=$(acloud network vpntunnel create \
  --name "prod-vpn-tunnel" \
  --region ITBG-Bergamo | grep "ID:" | awk '{print $2}')

# 2. Wait for tunnel to be Active
while true; do
  STATUS=$(acloud network vpntunnel get $VPN_TUNNEL_ID | grep "Status:" | awk '{print $2}')
  if [ "$STATUS" = "Active" ]; then
    break
  fi
  echo "Waiting for VPN tunnel to become Active... (current: $STATUS)"
  sleep 5
done

# 3. Create routes for different subnets
acloud network vpnroute create $VPN_TUNNEL_ID \
  --name "vpc-subnet-1" \
  --region ITBG-Bergamo \
  --cloud-subnet "10.0.1.0/24" \
  --onprem-subnet "192.168.1.0/24"

acloud network vpnroute create $VPN_TUNNEL_ID \
  --name "vpc-subnet-2" \
  --region ITBG-Bergamo \
  --cloud-subnet "10.0.2.0/24" \
  --onprem-subnet "192.168.2.0/24"

# 4. List all routes
acloud network vpnroute list $VPN_TUNNEL_ID
```

### Updating VPN Routes

```bash
VPN_TUNNEL_ID="1234567890abcdef"
ROUTE_ID="1234567890abcdef123456"

# Update cloud subnet
acloud network vpnroute update $VPN_TUNNEL_ID $ROUTE_ID \
  --cloud-subnet "10.0.3.0/24"

# Update on-premises subnet
acloud network vpnroute update $VPN_TUNNEL_ID $ROUTE_ID \
  --onprem-subnet "192.168.3.0/24"

# Update name and tags
acloud network vpnroute update $VPN_TUNNEL_ID $ROUTE_ID \
  --name "updated-route" \
  --tags "vpn,production,updated"
```

## Best Practices

1. **Use Descriptive Names**
   ```bash
   --name "vpc-subnet-1-to-onprem"
   --name "production-vpn-route"
   ```

2. **Tag Your Routes**
   ```bash
   --tags "vpn,production,network"
   --tags "vpn,development,test"
   ```

3. **Plan Subnet Mappings**
   - Ensure cloud and on-premises subnets don't overlap
   - Use clear naming conventions for route identification

4. **Wait for Active State**
   ```bash
   # Check status before updating
   acloud network vpnroute get <vpn-tunnel-id> <route-id>
   # Ensure Status is "Active"
   acloud network vpnroute update <vpn-tunnel-id> <route-id> --name "new-name"
   ```

## Troubleshooting

### "Cannot update VPN route while in InCreation state"

**Problem:** Trying to update a VPN route that hasn't finished creating.

**Solution:**
```bash
# Check current status
acloud network vpnroute get <vpn-tunnel-id> <route-id>

# Wait for Status to become "Active"
# Then retry the update
acloud network vpnroute update <vpn-tunnel-id> <route-id> --name "new-name"
```

### "Error: at least one field must be provided for update"

**Problem:** Update command called without any changes.

**Solution:**
```bash
# Provide at least one field to update
acloud network vpnroute update <vpn-tunnel-id> <route-id> --name "new-name"
# or
acloud network vpnroute update <vpn-tunnel-id> <route-id> --tags tag1,tag2
```

## Related Commands

- [VPN Tunnel](vpntunnel.md) - Manage VPN tunnels
- [VPC](vpc.md) - Manage VPCs
