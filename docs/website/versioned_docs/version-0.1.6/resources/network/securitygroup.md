# Security Group

Security Groups act as virtual firewalls for your resources within a VPC. They control inbound and outbound traffic at the instance level, allowing you to define rules based on protocols, ports, and source/destination IP ranges.

## Commands

### List Security Groups
List all security groups in a VPC.

```bash
acloud network securitygroup list <vpc-id>
```

**Arguments:**
- `vpc-id` - The ID of the VPC

**Example:**
```bash
acloud network securitygroup list 689307f4745108d3c6343b5a
```

**Output:**
```
NAME         ID                        DESCRIPTION         STATUS
web-sg       1234567890abcdef          Allow web traffic   Active
```

### Get Security Group Details
Get details about a specific security group.

```bash
acloud network securitygroup get <vpc-id> <security-group-id>
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `security-group-id` - The ID of the security group

**Example:**
```bash
acloud network securitygroup get 689307f4745108d3c6343b5a 1234567890abcdef
```

**Output:**
```
Security Group Details:
======================
ID:            1234567890abcdef
Name:          web-sg
Description:   Allow web traffic
Status:        Active
Rules:         3
```

### Create Security Group
Create a new security group in a VPC.

```bash
acloud network securitygroup create <vpc-id> --name <name> --description <desc>
```

**Required Flags:**
- `--name string` - Name for the security group
- `--description string` - Description of the security group

**Example:**
```bash
acloud network securitygroup create 689307f4745108d3c6343b5a --name web-sg --description "Allow web traffic"
```

**Output:**
```
Security Group created successfully!
ID:          1234567890abcdef
Name:        web-sg
Description: Allow web traffic
```

### Update Security Group
Update an existing security group's name or description.

```bash
acloud network securitygroup update <vpc-id> <security-group-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `security-group-id` - The ID of the security group

**Flags:**
- `--name string` - New name for the security group
- `--description string` - New description

**Example:**
```bash
acloud network securitygroup update 689307f4745108d3c6343b5a 1234567890abcdef --name "new-sg-name"
```

**Output:**
```
Security Group updated successfully!
ID:          1234567890abcdef
Name:        new-sg-name
```

### Delete Security Group
Delete a security group from a VPC.

```bash
acloud network securitygroup delete <vpc-id> <security-group-id>
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `security-group-id` - The ID of the security group

**Example:**
```bash
acloud network securitygroup delete 689307f4745108d3c6343b5a 1234567890abcdef
```

**Output:**
```
Security Group 1234567890abcdef deleted successfully!
```

## Shell Auto-completion

The security group commands support auto-completion for VPC IDs and security group IDs.

## Best Practices
- Use descriptive names and descriptions for security groups.
- Regularly review and update security group rules.

## Troubleshooting
- Ensure the VPC is **Active** before creating security groups.
- Check for conflicting or overly permissive rules.

## Related Commands
- [Subnet](subnet.md)
- [VPC](vpc.md)
