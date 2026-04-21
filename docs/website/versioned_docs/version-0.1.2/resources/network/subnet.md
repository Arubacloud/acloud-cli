# Subnet

Subnets allow you to segment your VPC network into smaller, isolated sections. Each subnet can have its own CIDR block and can be used to organize resources, control routing, and apply security policies within a VPC.

## Subnet Types

Subnets can be created in two types:

- **Basic Subnet**: Automatically assigned CIDR block by the system. No DHCP configuration required.
- **Advanced Subnet**: Custom CIDR block specified by the user. Requires DHCP configuration with `--dhcp-enabled` flag. Optionally supports custom routes and DNS servers.

## Commands

### List Subnets
List all subnets in a VPC.

```bash
acloud network subnet list <vpc-id>
```

**Arguments:**
- `vpc-id` - The ID of the VPC

**Example:**
```bash
acloud network subnet list 689307f4745108d3c6343b5a
```

**Output:**
```
NAME         ID                        CIDR           STATUS
subnet-1     1234567890abcdef          10.0.1.0/24    Active
subnet-2     0987654321fedcba          10.0.2.0/24    Active
```

### Get Subnet Details
Get details about a specific subnet.

```bash
acloud network subnet get <vpc-id> <subnet-id>
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `subnet-id` - The ID of the subnet

**Example:**
```bash
acloud network subnet get 689307f4745108d3c6343b5a 1234567890abcdef
```

**Output:**
```
Subnet Details:
===============
ID:            1234567890abcdef
Name:          subnet-1
Type:          Advanced
CIDR:          10.0.1.0/24
DHCP Enabled:  true
DHCP Routes:
  - 0.0.0.0/0 -> 10.0.0.1
DHCP DNS:      [8.8.8.8 8.8.4.4]
Status:        Active
```

### Create Subnet
Create a new subnet in a VPC.

```bash
acloud network subnet create <vpc-id> --name <name> --region <region> [flags]
```

**Required Flags:**
- `--name string` - Name for the subnet
- `--region string` - Region for the subnet

**Optional Flags:**
- `--cidr string` - CIDR block for the subnet (if provided, creates Advanced subnet type)
- `--tags stringSlice` - Subnet tags (comma-separated or multiple flags)
- `--dhcp-enabled` - Enable DHCP for Advanced subnet type (required when `--cidr` is provided)
- `--dhcp-routes stringSlice` - DHCP routes for Advanced subnet type (format: `destination:gateway`, e.g., `0.0.0.0/0:10.0.0.1`)
- `--dhcp-dns stringSlice` - DHCP DNS servers for Advanced subnet type (e.g., `8.8.8.8`, `8.8.4.4`)

**Create Basic Subnet (Auto-assigned CIDR):**
```bash
acloud network subnet create 689307f4745108d3c6343b5a --name subnet-1 --region "ITBG-Bergamo"
```

**Create Advanced Subnet (Custom CIDR with DHCP):**
```bash
acloud network subnet create 689307f4745108d3c6343b5a \
  --name subnet-1 \
  --region "ITBG-Bergamo" \
  --cidr 10.0.1.0/24 \
  --dhcp-enabled \
  --dhcp-routes "0.0.0.0/0:10.0.0.1" \
  --dhcp-dns "8.8.8.8" "8.8.4.4"
```

**Output:**
```
NAME         ID                        REGION          CIDR           STATUS
subnet-1     1234567890abcdef          ITBG-Bergamo    10.0.1.0/24    Active
```

### Update Subnet
Update an existing subnet's name, CIDR, tags, or DHCP configuration.

```bash
acloud network subnet update <vpc-id> <subnet-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `subnet-id` - The ID of the subnet

**Flags:**
- `--name string` - New name for the subnet
- `--cidr string` - New CIDR block
- `--tags stringSlice` - Subnet tags (comma-separated or multiple flags)
- `--dhcp-enabled` - Enable/disable DHCP for Advanced subnet type
- `--dhcp-routes stringSlice` - DHCP routes for Advanced subnet type (format: `destination:gateway`)
- `--dhcp-dns stringSlice` - DHCP DNS servers for Advanced subnet type

**Examples:**

Update subnet name:
```bash
acloud network subnet update 689307f4745108d3c6343b5a 1234567890abcdef --name new-subnet-name
```

Update DHCP routes for Advanced subnet:
```bash
acloud network subnet update 689307f4745108d3c6343b5a 1234567890abcdef \
  --dhcp-routes "192.168.1.0/24:10.0.0.1" "0.0.0.0/0:10.0.0.1"
```

Update DHCP DNS servers:
```bash
acloud network subnet update 689307f4745108d3c6343b5a 1234567890abcdef \
  --dhcp-dns "1.1.1.1" "1.0.0.1"
```

**Output:**
```
NAME              ID                        CIDR           STATUS
new-subnet-name   1234567890abcdef          10.0.1.0/24    Active
```

### Delete Subnet
Delete a subnet from a VPC.

```bash
acloud network subnet delete <vpc-id> <subnet-id>
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `subnet-id` - The ID of the subnet

**Example:**
```bash
acloud network subnet delete 689307f4745108d3c6343b5a 1234567890abcdef
```

**Output:**
```
Subnet 1234567890abcdef deleted successfully!
```

## Shell Auto-completion

The subnet commands support auto-completion for VPC IDs and subnet IDs.

## Best Practices
- Use descriptive names for subnets based on their purpose.
- Avoid overlapping CIDR blocks between subnets.
- For Advanced subnets, always enable DHCP with `--dhcp-enabled` when providing a custom CIDR.
- Configure appropriate DHCP routes and DNS servers for Advanced subnets to ensure proper network connectivity.
- Use Basic subnets when you don't need custom CIDR configuration.

## Troubleshooting
- Ensure the VPC is **Active** before creating subnets.
- Check for CIDR conflicts when adding or updating subnets.

## Related Commands
- [VPC](vpc.md)
- [Security Group](securitygroup.md)
