# VPC Peering

VPC Peering allows you to connect two Virtual Private Clouds (VPCs) in Aruba Cloud, enabling private network traffic between them. Peering connections are useful for sharing resources or enabling communication between different environments or projects.

## Commands

### List VPC Peerings
List all VPC peering connections for a VPC.

```bash
acloud network vpcpeering list <vpc-id>
```

**Arguments:**
- `vpc-id` - The ID of the VPC

**Example:**
```bash
acloud network vpcpeering list 689307f4745108d3c6343b5a
```

**Output:**
```
NAME         ID                        PEER VPC                  REGION        STATUS
prod-peer    6949666e4d0cdc87949b7204  /.../vpcs/69485a584d0cdc87949b6ff8  ITBG-Bergamo  Active
```

### Get VPC Peering Details
Get details about a specific VPC peering connection.

```bash
acloud network vpcpeering get <vpc-id> <peering-id>
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `peering-id` - The ID of the VPC peering connection

**Example:**
```bash
acloud network vpcpeering get 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204
```

**Output:**
```
VPC Peering Details:
====================
ID:              6949666e4d0cdc87949b7204
Name:            prod-peer
Peer VPC:        /.../vpcs/69485a584d0cdc87949b6ff8
Region:          ITBG-Bergamo
Creation Date:   06-08-2025 07:44:52
Created By:      aru-297647
Tags:            [production peering]
Status:          Active
```

### Create VPC Peering
Create a new VPC peering connection.

```bash
acloud network vpcpeering create <vpc-id> --peer-vpc-id <peer-vpc-id> --name <name> --region <region>
```

**Required Flags:**
- `--peer-vpc-id string` - The URI of the peer VPC
- `--name string` - Name for the peering connection
- `--region string` - Region code (e.g., ITBG-Bergamo)

**Optional Flags:**
- `--tags strings` - Tags for the peering (comma-separated)

**Example:**
```bash
acloud network vpcpeering create 689307f4745108d3c6343b5a --peer-vpc-id /projects/.../vpcs/69485a584d0cdc87949b6ff8 --name prod-peer --region ITBG-Bergamo
```

**Output:**
```
VPC Peering created successfully!
ID:      6949666e4d0cdc87949b7204
Name:    prod-peer
Peer VPC: /.../vpcs/69485a584d0cdc87949b6ff8
Region:  ITBG-Bergamo
```

### Update VPC Peering
Update an existing VPC peering connection.

```bash
acloud network vpcpeering update <vpc-id> <peering-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `peering-id` - The ID of the VPC peering connection

**Flags:**
- `--name string` - New name for the peering
- `--tags strings` - New tags for the peering (comma-separated)

**Example:**
```bash
acloud network vpcpeering update 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 --name "new-peer-name"
```

**Output:**
```
VPC Peering updated successfully!
ID:      6949666e4d0cdc87949b7204
Name:    new-peer-name
```

### Delete VPC Peering
Delete a VPC peering connection.

```bash
acloud network vpcpeering delete <vpc-id> <peering-id>
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `peering-id` - The ID of the VPC peering connection

**Example:**
```bash
acloud network vpcpeering delete 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204
```

**Output:**
```
VPC Peering 6949666e4d0cdc87949b7204 deleted successfully!
```

## Shell Auto-completion

The VPC Peering commands support auto-completion for VPC IDs and peering IDs.

## Best Practices
- Use descriptive names for peering connections.
- Tag peerings by environment or purpose.

## Troubleshooting
- Ensure both VPCs are in the same region and project if required.
- Check peering status before updating or deleting.

## Related Commands
- [VPC Peering Route](vpcpeeringroute.md)
- [VPC](vpc.md)
