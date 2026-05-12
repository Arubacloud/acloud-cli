# Security Rule

Security Rules define the firewall rules for Security Groups within a VPC. They control inbound and outbound traffic by specifying direction, protocol, ports, and target (IP addresses or other Security Groups).

## Commands

### List Security Rules

List all security rules for a specific security group.

```bash
acloud network securityrule list <vpc-id> <securitygroup-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `securitygroup-id` - The ID of the security group

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network securityrule list 689307f4745108d3c6343b5a 1234567890abcdef
```

**Output:**
```
NAME            ID                        DIRECTION    PROTOCOL    PORT    TARGET                    STATUS
allow-http      1234567890abcdef123456   Ingress      TCP         80      Ip:0.0.0.0/0             Active
allow-https     1234567890abcdef123457   Ingress      TCP         443     Ip:0.0.0.0/0             Active
allow-ssh       1234567890abcdef123458   Ingress      TCP         22      Ip:10.0.0.0/8            Active
```

### Get Security Rule Details

Get detailed information about a specific security rule.

```bash
acloud network securityrule get <vpc-id> <securitygroup-id> <securityrule-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `securitygroup-id` - The ID of the security group
- `securityrule-id` - The ID of the security rule (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network securityrule get 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456
```

**Output:**
```
Security Rule Details:
=====================
ID:              1234567890abcdef123456
URI:             /projects/.../securityrules/1234567890abcdef123456
Name:            allow-http
Region:          ITBG-Bergamo
Direction:       Ingress
Protocol:        TCP
Port:            80
Target Kind:     Ip
Target Value:    0.0.0.0/0
Creation Date:   06-08-2025 07:44:52
Created By:      aru-297647
Tags:            [web http public]
Status:          Active
```

### Create Security Rule

Create a new security rule for a security group.

```bash
acloud network securityrule create <vpc-id> <securitygroup-id> [flags]
```

**Required Flags:**
- `--name string` - Security Rule name
- `--region string` - Region code (e.g., ITBG-Bergamo)
- `--direction string` - Direction: Ingress or Egress
- `--protocol string` - Protocol: ANY, TCP, UDP, ICMP
- `--target-kind string` - Target Kind: Ip or SecurityGroup
- `--target-value string` - Target Value: If kind = Ip, the value must be a valid network address in CIDR notation (included 0.0.0.0/0). If kind = SecurityGroup, the value must be a valid URI of any security group within the same VPC

**Optional Flags:**
- `--port string` - Port: a single numeric port, a port range or * (required for TCP/UDP)
- `--tags strings` - Tags for the security rule (comma-separated)
- `--project-id string` - Project ID (uses context if not specified)
- `-v, --verbose` - Show detailed debug information

**Examples:**
```bash
# Create a basic ingress rule allowing HTTP traffic
acloud network securityrule create 689307f4745108d3c6343b5a 1234567890abcdef \
  --name "allow-http" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 80 \
  --target-kind Ip \
  --target-value "0.0.0.0/0"

# Create a rule allowing SSH from specific network
acloud network securityrule create 689307f4745108d3c6343b5a 1234567890abcdef \
  --name "allow-ssh" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 22 \
  --target-kind Ip \
  --target-value "10.0.0.0/8" \
  --tags "ssh,admin,secure"

# Create a rule allowing all traffic from another security group
acloud network securityrule create 689307f4745108d3c6343b5a 1234567890abcdef \
  --name "allow-from-app-sg" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol ANY \
  --target-kind SecurityGroup \
  --target-value "/projects/.../securitygroups/9876543210fedcba"

# Create an egress rule allowing all outbound traffic
acloud network securityrule create 689307f4745108d3c6343b5a 1234567890abcdef \
  --name "allow-all-outbound" \
  --region ITBG-Bergamo \
  --direction Egress \
  --protocol ANY \
  --target-kind Ip \
  --target-value "0.0.0.0/0"
```

**Output:**
```
NAME            ID                        DIRECTION    PROTOCOL    PORT    STATUS
allow-http      1234567890abcdef123456   Ingress      TCP         80      Active
```

**Notes:**
- Port is required for TCP and UDP protocols
- Port can be omitted for ANY and ICMP protocols
- Port can be a single number (e.g., "80"), a range (e.g., "8000-9000"), or "*" for all ports
- The security rule will be in **InCreation** state initially
- Use `acloud network securityrule get` to check when it becomes **Active**

### Update Security Rule

Update an existing security rule's name or tags.

**Important:** Security rules can only be updated by changing the name or tags. Properties such as direction, protocol, port, and target cannot be modified. To change these properties, you must delete and recreate the security rule.

```bash
acloud network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `securitygroup-id` - The ID of the security group
- `securityrule-id` - The ID of the security rule (supports auto-completion)

**Flags:**
- `--name string` - New name for the security rule
- `--tags strings` - New tags for the security rule (comma-separated)
- `--project-id string` - Project ID (uses context if not specified)

**Note:** At least one field (--name or --tags) must be provided for update.

**Debug Mode:** Use the global `--debug` flag to see detailed HTTP request/response information when troubleshooting update issues:

```bash
acloud --debug network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --tags test
```

This will show the request payload and full error details if the update fails.

**Examples:**
```bash
# Update security rule name
acloud network securityrule update 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456 \
  --name "allow-http-updated"

# Update tags
acloud network securityrule update 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456 \
  --tags "web,https,secure"

# Update both name and tags
acloud network securityrule update 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456 \
  --name "allow-https" \
  --tags "web,https,secure"
```

**Output:**
```
NAME            ID                        DIRECTION    PROTOCOL    PORT    STATUS
allow-https     1234567890abcdef123456   Ingress      TCP         443     Active
```

**Restrictions:**
- Cannot update security rules in **InCreation** state
- Wait for the security rule to reach **Active** state before updating
- Only name and tags can be updated; properties (direction, protocol, port, target) cannot be changed

### Delete Security Rule

Delete a security rule.

```bash
acloud network securityrule delete <vpc-id> <securitygroup-id> <securityrule-id> [flags]
```

**Arguments:**
- `vpc-id` - The ID of the VPC
- `securitygroup-id` - The ID of the security group
- `securityrule-id` - The ID of the security rule (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)
- `-y, --yes` - Skip confirmation prompt

**Examples:**
```bash
# Delete with confirmation prompt
acloud network securityrule delete 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456

# Delete without confirmation
acloud network securityrule delete 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456 --yes
```

**Confirmation Prompt:**
```
Are you sure you want to delete security rule 1234567890abcdef123456? This action cannot be undone.
Type 'yes' to confirm: yes
```

**Output:**
```
ID                              STATUS
1234567890abcdef123456         deleted
```

**Notes:**
- Deletion cannot be undone
- Ensure no resources depend on the security rule before deletion

## Shell Auto-completion

The security rule commands support intelligent auto-completion for security rule IDs:

```bash
# Enable completion (bash)
source <(acloud completion bash)

# Type command and press TAB to see available security rule IDs
acloud network securityrule get <vpc-id> <securitygroup-id> <TAB>
acloud network securityrule update <vpc-id> <securitygroup-id> <TAB>
acloud network securityrule delete <vpc-id> <securitygroup-id> <TAB>
```

Auto-completion shows security rule IDs with their names:
```
1234567890abcdef123456    allow-http
1234567890abcdef123457    allow-https
1234567890abcdef123458    allow-ssh
```

## Security Rule Properties

### Direction

- **Ingress**: Inbound traffic (traffic coming into the security group)
- **Egress**: Outbound traffic (traffic going out from the security group)

### Protocol

- **ANY**: All protocols
- **TCP**: Transmission Control Protocol
- **UDP**: User Datagram Protocol
- **ICMP**: Internet Control Message Protocol

### Port

- **Single port**: A numeric value (e.g., "80", "443", "22")
- **Port range**: A range of ports (e.g., "8000-9000")
- **All ports**: Use "*" to allow all ports
- **Note**: Port is required for TCP and UDP, optional for ANY and ICMP

### Target

The target specifies the source (for Ingress) or destination (for Egress) of the traffic:

- **IP Target**: Use CIDR notation (e.g., "0.0.0.0/0" for all IPs, "10.0.0.0/8" for private network)
- **Security Group Target**: Use the URI of another security group within the same VPC

## Security Rule States

Security rules can be in the following states:

| State | Description | Can Update? | Can Delete? |
|-------|-------------|-------------|-------------|
| InCreation | Security rule is being created | ❌ No | ❌ No |
| Active | Security rule is ready to use | ✅ Yes | ✅ Yes |

## Common Workflows

### Setting Up Web Server Security Rules

```bash
# 1. Create security group (if not exists)
VPC_ID="689307f4745108d3c6343b5a"
SG_ID=$(acloud network securitygroup create $VPC_ID \
  --name "web-server-sg" \
  --region ITBG-Bergamo | grep "ID:" | awk '{print $2}')

# 2. Allow HTTP traffic from anywhere
acloud network securityrule create $VPC_ID $SG_ID \
  --name "allow-http" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 80 \
  --target-kind Ip \
  --target-value "0.0.0.0/0"

# 3. Allow HTTPS traffic from anywhere
acloud network securityrule create $VPC_ID $SG_ID \
  --name "allow-https" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 443 \
  --target-kind Ip \
  --target-value "0.0.0.0/0"

# 4. Allow SSH from admin network only
acloud network securityrule create $VPC_ID $SG_ID \
  --name "allow-ssh-admin" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 22 \
  --target-kind Ip \
  --target-value "10.0.0.0/8" \
  --tags "ssh,admin,secure"

# 5. Allow all outbound traffic
acloud network securityrule create $VPC_ID $SG_ID \
  --name "allow-all-outbound" \
  --region ITBG-Bergamo \
  --direction Egress \
  --protocol ANY \
  --target-kind Ip \
  --target-value "0.0.0.0/0"
```

### Allowing Communication Between Security Groups

```bash
VPC_ID="689307f4745108d3c6343b5a"
APP_SG_ID="1234567890abcdef"
DB_SG_ID="9876543210fedcba"

# Get the URI of the database security group
DB_SG_URI=$(acloud network securitygroup get $VPC_ID $DB_SG_ID | grep "URI:" | awk '{print $2}')

# Allow app security group to access database security group
acloud network securityrule create $VPC_ID $DB_SG_ID \
  --name "allow-from-app" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 5432 \
  --target-kind SecurityGroup \
  --target-value "$DB_SG_URI"
```

### Updating Security Rules

```bash
VPC_ID="689307f4745108d3c6343b5a"
SG_ID="1234567890abcdef"
RULE_ID="1234567890abcdef123456"

# Update port range
acloud network securityrule update $VPC_ID $SG_ID $RULE_ID \
  --port "8080-8090"

# Update target to restrict access
acloud network securityrule update $VPC_ID $SG_ID $RULE_ID \
  --target-value "192.168.1.0/24"

# Update name and tags
acloud network securityrule update $VPC_ID $SG_ID $RULE_ID \
  --name "allow-http-restricted" \
  --tags "web,http,restricted"
```

## Best Practices

1. **Principle of Least Privilege**
   ```bash
   # Good: Allow specific port from specific network
   --port 22 --target-value "10.0.0.0/8"
   
   # Avoid: Allow all ports from anywhere
   --port "*" --target-value "0.0.0.0/0"
   ```

2. **Use Descriptive Names**
   ```bash
   --name "allow-http-public"
   --name "allow-ssh-admin-only"
   --name "allow-db-from-app"
   ```

3. **Tag Your Rules**
   ```bash
   --tags "web,public,http"
   --tags "admin,ssh,secure"
   --tags "database,internal"
   ```

4. **Use Security Group Targets for Internal Communication**
   ```bash
   # Better: Reference security group
   --target-kind SecurityGroup --target-value "/projects/.../securitygroups/..."
   
   # Less secure: Use IP ranges
   --target-kind Ip --target-value "10.0.0.0/8"
   ```

5. **Review Rules Regularly**
   ```bash
   # List all rules for a security group
   acloud network securityrule list <vpc-id> <securitygroup-id>
   
   # Review each rule
   acloud network securityrule get <vpc-id> <securitygroup-id> <rule-id>
   ```

6. **Document Rules with Tags**
   ```bash
   --tags "purpose=web-server,environment=production,managed-by=devops"
   ```

## Troubleshooting

### "Cannot update security rule while in InCreation state"

**Problem:** Trying to update a security rule that hasn't finished creating.

**Solution:**
```bash
# Check current status
acloud network securityrule get <vpc-id> <securitygroup-id> <securityrule-id>

# Wait for Status to become "Active"
# Then retry the update
acloud network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --name "new-name"
```

### "Error: --port is required"

**Problem:** Port is required for TCP and UDP protocols but was not provided.

**Solution:**
```bash
# For TCP/UDP, always provide port
--protocol TCP --port 80
--protocol UDP --port 53

# For ANY and ICMP, port can be omitted
--protocol ANY
--protocol ICMP
```

### "Error: at least one field must be provided for update"

**Problem:** Update command called without any changes.

**Solution:**
```bash
# Provide at least one field to update
acloud network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --name "new-name"
# or
acloud network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --tags tag1,tag2
```

### Invalid Target Value

**Problem:** Target value format is incorrect.

**Solution:**
```bash
# For IP targets, use CIDR notation
--target-kind Ip --target-value "0.0.0.0/0"        # All IPs
--target-kind Ip --target-value "10.0.0.0/8"       # Private network
--target-kind Ip --target-value "192.168.1.0/24"   # Specific subnet

# For Security Group targets, use full URI
--target-kind SecurityGroup --target-value "/projects/.../securitygroups/..."
```

## Related Commands

- [Security Group](securitygroup.md) - Manage security groups
- [VPC](vpc.md) - Manage VPCs
- [Subnet](subnet.md) - Manage subnets

