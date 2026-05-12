# Load Balancer

Load Balancers distribute incoming network traffic across multiple resources to ensure high availability and reliability.

**Note:** Load Balancers are currently read-only via the CLI. You can list and view details, but cannot create, update, or delete them through the CLI.

## Commands

### List Load Balancers

List all Load Balancers in your project.

```bash
acloud network loadbalancer list [flags]
```

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
# List Load Balancers using context
acloud network loadbalancer list

# List with explicit project ID
acloud network loadbalancer list --project-id 68398923fb2cb026400d4d31
```

**Output:**
```
NAME                                     ID                        ADDRESS          STATUS
ingress-nginx-controller                 68ffa1797912602cb16794dc  209.227.232.229  Active
api-gateway-lb                           69485b8a4d0cdc87949b7012  95.110.142.230   Active
```

### Get Load Balancer Details

Get detailed information about a specific Load Balancer.

```bash
acloud network loadbalancer get <load-balancer-id> [flags]
```

**Arguments:**
- `load-balancer-id` - The ID of the Load Balancer (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network loadbalancer get 68ffa1797912602cb16794dc
```

**Output:**
```
Load Balancer Details:
======================
ID:              68ffa1797912602cb16794dc
URI:             /projects/.../loadBalancers/68ffa1797912602cb16794dc
Name:            ingress-nginx-controller
Address:         209.227.232.229
Linked Resources: 2
VPC:             689307f4745108d3c6343b5a
Creation Date:   27-10-2025 16:44:41
Created By:      aru-297647
Tags:            [kubernetes ingress production]
Status:          Active
```

## Shell Auto-completion

The Load Balancer commands support intelligent auto-completion for Load Balancer IDs:

```bash
# Enable completion (bash)
source <(acloud completion bash)

# Type command and press TAB to see available Load Balancer IDs
acloud network loadbalancer get <TAB>
```

Auto-completion shows Load Balancer IDs with their names:
```
68ffa1797912602cb16794dc    ingress-nginx-controller
69485b8a4d0cdc87949b7012    api-gateway-lb
```

## Load Balancer States

Load Balancers can be in the following states:

| State | Description |
|-------|-------------|
| InCreation | Load Balancer is being created |
| Active | Load Balancer is ready and distributing traffic |
| Updating | Load Balancer configuration is being updated |
| Deleting | Load Balancer is being deleted |

## Load Balancer Properties

### Address

The public IP address of the Load Balancer:
- Assigned automatically
- Used to route incoming traffic
- Typically an Elastic IP

### VPC Association

Load Balancers are associated with a VPC:
- Shown as VPC ID in the details
- Determines network isolation
- Affects routing and security rules

### Linked Resources

Load Balancers distribute traffic to linked resources:
- Backend servers
- Target groups
- Health check endpoints

The `Linked Resources` count shows how many backends are configured.

### Tags

Load Balancers support tags for organization:
- Set when the Load Balancer is created
- Visible via the get command
- Cannot be modified via CLI (read-only)

## Common Workflows

### Viewing Load Balancer Information

```bash
# List all Load Balancers
acloud network loadbalancer list

# Get details of a specific Load Balancer
acloud network loadbalancer get <lb-id>

# View Load Balancer with specific tags
acloud network loadbalancer list | grep "production"
```

### Monitoring Load Balancers

```bash
# Check status of all Load Balancers
acloud network loadbalancer list

# Get detailed info including linked resources
for lb_id in $(acloud network loadbalancer list | tail -n +2 | awk '{print $2}'); do
  echo "=== Load Balancer: $lb_id ==="
  acloud network loadbalancer get $lb_id
  echo ""
done
```

### Finding Load Balancers by Address

```bash
# List all Load Balancers with their addresses
acloud network loadbalancer list

# Filter by specific address
acloud network loadbalancer list | grep "209.227.232.229"
```

## Integration with Other Resources

### With Elastic IPs

Load Balancers often use Elastic IPs:
```bash
# Get Load Balancer address
LB_IP=$(acloud network loadbalancer get <lb-id> | grep "Address:" | awk '{print $2}')

# Find the corresponding Elastic IP
acloud network elasticip list | grep "$LB_IP"
```

### With VPCs

Load Balancers are associated with VPCs:
```bash
# Get Load Balancer VPC
VPC_ID=$(acloud network loadbalancer get <lb-id> | grep "VPC:" | awk '{print $2}')

# Get VPC details
acloud network vpc get $VPC_ID
```

## Best Practices

1. **Regular Monitoring**
   ```bash
   # Check Load Balancer status regularly
   acloud network loadbalancer list
   ```

2. **Document Configuration**
   ```bash
   # Save Load Balancer details for reference
   acloud network loadbalancer get <lb-id> > lb-config-backup.txt
   ```

3. **Track Resource Associations**
   ```bash
   # Note which VPC and resources are associated
   acloud network loadbalancer get <lb-id>
   ```

4. **Use with Project Contexts**
   ```bash
   acloud context use prod-project
   acloud network loadbalancer list  # No need for --project-id
   ```

## Limitations

### Read-Only Access

Load Balancers are read-only via the CLI:
- ❌ Cannot create Load Balancers
- ❌ Cannot update Load Balancer configuration
- ❌ Cannot delete Load Balancers
- ❌ Cannot modify tags
- ✅ Can list Load Balancers
- ✅ Can view Load Balancer details

### Managing Load Balancers

To create, update, or delete Load Balancers, use:
- Aruba Cloud Web Console
- Aruba Cloud API directly
- Infrastructure as Code tools (Terraform, etc.)

The CLI provides read-only access for monitoring and reference.

## Use Cases

### Monitoring Traffic Distribution

```bash
# View Load Balancer status
acloud network loadbalancer list

# Check linked resources count
acloud network loadbalancer get <lb-id> | grep "Linked Resources"
```

### Infrastructure Auditing

```bash
# List all Load Balancers with their configurations
for lb_id in $(acloud network loadbalancer list | tail -n +2 | awk '{print $2}'); do
  echo "Load Balancer ID: $lb_id"
  acloud network loadbalancer get $lb_id
  echo "---"
done > load-balancers-audit.txt
```

### Integration with Scripts

```bash
#!/bin/bash
# Check if a Load Balancer is active

LB_ID="68ffa1797912602cb16794dc"
STATUS=$(acloud network loadbalancer get $LB_ID | grep "Status:" | awk '{print $2}')

if [ "$STATUS" = "Active" ]; then
  echo "Load Balancer is healthy"
  exit 0
else
  echo "Load Balancer is not active: $STATUS"
  exit 1
fi
```

### Finding Load Balancers by VPC

```bash
#!/bin/bash
# Find all Load Balancers in a specific VPC

VPC_ID="689307f4745108d3c6343b5a"

echo "Load Balancers in VPC: $VPC_ID"
for lb_id in $(acloud network loadbalancer list | tail -n +2 | awk '{print $2}'); do
  LB_VPC=$(acloud network loadbalancer get $lb_id | grep "VPC:" | awk '{print $2}')
  if [ "$LB_VPC" = "$VPC_ID" ]; then
    LB_NAME=$(acloud network loadbalancer get $lb_id | grep "Name:" | awk '{print $2}')
    echo "  - $LB_NAME ($lb_id)"
  fi
done
```

## Troubleshooting

### "Error: required flag(s) 'project-id' not set"

**Problem:** No project context set and project ID not provided.

**Solution:**
```bash
# Option 1: Use context
acloud context use my-project
acloud network loadbalancer list

# Option 2: Provide project ID
acloud network loadbalancer list --project-id <project-id>
```

### Load Balancer Not Showing in List

**Problem:** Expected Load Balancer is not visible.

**Solution:**
```bash
# Check you're using the correct project
acloud context current

# Switch to the correct project if needed
acloud context use correct-project
acloud network loadbalancer list
```

### Cannot Find Load Balancer Address

**Problem:** Need to find which Load Balancer uses a specific IP.

**Solution:**
```bash
# List all Load Balancers with addresses
acloud network loadbalancer list

# Or search for specific IP
acloud network loadbalancer list | grep "209.227.232.229"
```

## Output Format

### List Output

The list command shows a table with:
- **NAME**: Load Balancer name (may be truncated)
- **ID**: Unique identifier
- **ADDRESS**: Public IP address
- **STATUS**: Current state (Active, InCreation, etc.)

### Get Output

The get command shows detailed information:
- Basic metadata (ID, URI, Name)
- Network configuration (Address, VPC)
- Resource associations (Linked Resources count)
- Timestamps (Creation Date, Created By)
- Tags (if configured)
- Current status

## Related Commands

- [VPC](vpc.md) - View VPC associated with Load Balancers
- [Elastic IP](elasticip.md) - View Elastic IPs used by Load Balancers
- [Context Management](../../installation.md#context-management) - Manage project contexts

## Future Enhancements

The following features may be added in future versions:
- Create Load Balancer
- Update Load Balancer configuration
- Delete Load Balancer
- Modify Load Balancer tags
- Manage backend pools
- Configure health checks
- View traffic statistics

For now, use the Aruba Cloud Web Console or API for write operations.
