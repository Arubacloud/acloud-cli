# Network Resources

The `network` category provides commands for managing network resources in Aruba Cloud.

## Available Resources

### [VPC](network/vpc.md)

Virtual Private Clouds (VPCs) provide isolated network environments for your resources.

**Quick Commands:**
```bash
# List all VPCs
acloud network vpc list

# Get VPC details
acloud network vpc get <vpc-id>

# Create a VPC
acloud network vpc create --name "my-vpc" --region ITBG-Bergamo

# Update a VPC
acloud network vpc update <vpc-id> --name "new-name" --tags tag1,tag2

# Delete a VPC
acloud network vpc delete <vpc-id>
```

### [Elastic IP](network/elasticip.md)

Elastic IPs are static public IP addresses that can be assigned to your resources.

**Quick Commands:**
```bash
# List all Elastic IPs
acloud network elasticip list

# Get Elastic IP details
acloud network elasticip get <eip-id>

# Create an Elastic IP
acloud network elasticip create --name "my-eip" --region ITBG-Bergamo --billing-period Hour

# Update an Elastic IP
acloud network elasticip update <eip-id> --name "new-name" --tags tag1,tag2

# Delete an Elastic IP
acloud network elasticip delete <eip-id>
```

### [Security Rule](network/securityrule.md)

Security Rules define the firewall rules for Security Groups. They specify direction (Ingress/Egress), protocol, ports, and target (IP addresses or Security Groups).

**Quick Commands:**
```bash
# List all security rules for a security group
acloud network securityrule list <vpc-id> <securitygroup-id>

# Get security rule details
acloud network securityrule get <vpc-id> <securitygroup-id> <securityrule-id>

# Create a security rule
acloud network securityrule create <vpc-id> <securitygroup-id> \
  --name "allow-http" --region ITBG-Bergamo \
  --direction Ingress --protocol TCP --port 80 \
  --target-kind Ip --target-value "0.0.0.0/0"

# Update a security rule
acloud network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> \
  --port 443

# Delete a security rule
acloud network securityrule delete <vpc-id> <securitygroup-id> <securityrule-id>
```

### [Load Balancer](network/loadbalancer.md)

Load Balancers distribute traffic across multiple resources. Note: Load Balancers are read-only via the CLI.

**Quick Commands:**
```bash
# List all Load Balancers
acloud network loadbalancer list

# Get Load Balancer details
acloud network loadbalancer get <lb-id>
```

## Common Patterns

### Using Project Context

All network commands support project context. Set a context to avoid specifying `--project-id` every time:

```bash
# Set current context
acloud context use my-project

# Now you can run commands without --project-id
acloud network vpc list
acloud network elasticip list
```

### Shell Auto-completion

The CLI provides intelligent auto-completion for resource IDs:

```bash
# Type the command and press TAB to see available VPC IDs
acloud network vpc get <TAB>

# Type the command and press TAB to see available Elastic IP IDs
acloud network elasticip update <TAB>
```

### Tagging Resources

Use tags to organize and categorize your network resources:

```bash
# Create with tags
acloud network vpc create --name "prod-vpc" --region ITBG-Bergamo --tags production,critical

# Update tags
acloud network vpc update <vpc-id> --tags production,updated,network
acloud network elasticip update <eip-id> --tags production,public
```

### Regional Resources

Network resources are regional. Supported regions:
- `ITBG-Bergamo` - Italy, Bergamo

Specify the region when creating resources:

```bash
acloud network vpc create --name "my-vpc" --region ITBG-Bergamo
acloud network elasticip create --name "my-eip" --region ITBG-Bergamo --billing-period Hour
```

## Resource Lifecycle

### VPC Lifecycle
1. **InCreation** - VPC is being created
2. **Active** - VPC is ready to use
3. **Deleting** - VPC is being deleted

### Elastic IP Lifecycle
1. **InCreation** - Elastic IP is being created
2. **NotUsed** - Elastic IP is created but not attached
3. **InUse** - Elastic IP is attached to a resource

### Load Balancer Lifecycle
1. **InCreation** - Load Balancer is being created
2. **Active** - Load Balancer is ready to use

### State Restrictions

- Resources cannot be updated while in **InCreation** state
- Resources in **Deleting** state will be removed automatically

## Best Practices

1. **Use Descriptive Names**: Give your resources meaningful names for easier management
   ```bash
   acloud network vpc create --name "production-vpc" --region ITBG-Bergamo
   ```

2. **Tag Your Resources**: Use tags to organize resources by environment, team, or purpose
   ```bash
   acloud network vpc update <vpc-id> --tags production,team-devops,critical
   ```

3. **Use Contexts**: Set up project contexts to streamline your workflow
   ```bash
   acloud context create prod-env --project-id <project-id>
   acloud context use prod-env
   ```

4. **Check Status Before Updates**: Ensure resources are not in transitional states
   ```bash
   acloud network vpc get <vpc-id>
   # Check that Status is "Active" before updating
   acloud network vpc update <vpc-id> --name "new-name"
   ```

5. **Use Confirmation Flags**: Skip confirmation prompts in scripts with `--yes`
   ```bash
   acloud network vpc delete <vpc-id> --yes
   ```

## Examples

### Setting Up a New Network Environment

```bash
# Create a VPC
acloud network vpc create \
  --name "production-vpc" \
  --region ITBG-Bergamo \
  --tags production,network

# Create an Elastic IP for external access
acloud network elasticip create \
  --name "prod-api-ip" \
  --region ITBG-Bergamo \
  --billing-period Month \
  --tags production,api

# List all resources
acloud network vpc list
acloud network elasticip list
```

### Updating Network Resources

```bash
# Update VPC tags
acloud network vpc update <vpc-id> --tags production,updated,critical

# Rename Elastic IP
acloud network elasticip update <eip-id> --name "prod-api-gateway-ip"

# Update both name and tags
acloud network elasticip update <eip-id> \
  --name "new-name" \
  --tags production,critical,frontend
```

### Cleaning Up Resources

```bash
# Delete Elastic IP (with confirmation)
acloud network elasticip delete <eip-id>

# Delete VPC without confirmation prompt
acloud network vpc delete <vpc-id> --yes

# Check remaining resources
acloud network vpc list
acloud network elasticip list
```

## Troubleshooting

### "Cannot update resource while in InCreation state"

Wait for the resource to finish creation before attempting updates:

```bash
# Check current status
acloud network vpc get <vpc-id>

# Wait for Status to show "Active", then retry
acloud network vpc update <vpc-id> --name "new-name"
```

### "Failed to create VPC - Status: 400"

Ensure you're using the correct region format:

```bash
# Correct format
acloud network vpc create --name "my-vpc" --region ITBG-Bergamo

# Incorrect format (will fail)
acloud network vpc create --name "my-vpc" --region eu-west-1
```

### "Error: required flag(s) 'project-id' not set"

Either set a project context or provide the project ID explicitly:

```bash
# Option 1: Use context
acloud context use my-project
acloud network vpc list

# Option 2: Provide project ID
acloud network vpc list --project-id <project-id>
```

## Related Documentation

- [VPC Documentation](network/vpc.md)
- [Subnet Documentation](network/subnet.md)
- [Security Group Documentation](network/securitygroup.md)
- [Security Rule Documentation](network/securityrule.md)
- [Elastic IP Documentation](network/elasticip.md)
- [Load Balancer Documentation](network/loadbalancer.md)
- [VPC Peering Documentation](network/vpcpeering.md)
- [VPC Peering Route Documentation](network/vpcpeeringroute.md)
- [VPN Tunnel Documentation](network/vpntunnel.md)
- [VPN Tunnel Route Documentation](network/vpnroute.md)
- [Context Management](../installation.md#context-management)
