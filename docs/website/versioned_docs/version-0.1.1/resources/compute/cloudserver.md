# Cloud Servers

Cloud servers are virtual machine instances that run your applications and workloads in Aruba Cloud.

## Command Syntax

```bash
acloud compute cloudserver <command> [flags] [arguments]
```

## Available Commands

### `create`

Create a new cloud server instance.

**Syntax:**
```bash
acloud compute cloudserver create [flags]
```


**Required Flags:**
- `--name <string>` - Name for the cloud server
- `--region <string>` - Region code (e.g., `ITBG-Bergamo`)
- `--flavor <string>` - Flavor name (e.g., `small`, `medium`, `large`)
- `--image <string>` - Image ID or name
- `--vpc-uri <string>` - VPC URI (e.g., `/projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id}`)
- `--subnet-uri <stringSlice>` - Subnet URI(s) (comma-separated)
- `--security-group-uri <stringSlice>` - Security Group URI(s) (comma-separated)

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--keypair <string>` - Key pair name for SSH access
- `--tags <stringSlice>` - Tags (comma-separated)
- `--user-data-file <string>` - Path to cloud-init YAML file (will be base64 encoded)
- `--elasticip-uri <string>` - Elastic IP URI (optional)
- `--billing-period <string>` - Billing period: Hour, Month, Year (optional, default: Hour)

**Example:**
```bash
acloud compute cloudserver create \
  --name "web-server" \
  --region "ITBG-Bergamo" \
  --flavor "small" \
  --image "ubuntu-22.04" \
  --keypair "my-keypair" \
  --tags "production,web" \
  --user-data-file "/path/to/cloud-init.yaml"
```

**Note:** The `--user-data-file` flag accepts a path to a cloud-init YAML file. The file content will be automatically base64 encoded and included in the cloud server creation request. This allows you to configure the server during initialization using cloud-init scripts.

### `list`

List all cloud servers in the project.

**Syntax:**
```bash
acloud compute cloudserver list [flags]
```

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**Example:**
```bash
acloud compute cloudserver list
```

**Output:**
The command displays a table with the following columns:
- NAME
- ID
- LOCATION
- FLAVOR
- STATUS

### `get`

Get detailed information about a specific cloud server.

**Syntax:**
```bash
acloud compute cloudserver get <server-id> [flags]
```

**Arguments:**
- `<server-id>` - The ID of the cloud server

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--verbose` - Show detailed JSON output

**Example:**
```bash
acloud compute cloudserver get 69495ef64d0cdc87949b71ec
```

**Output:**
The command displays detailed information including:
- ID and URI
- Name and region
- Flavor details (CPU, RAM, HD)
- Image information
- Key pair (if configured)
- Status
- Creation date and creator
- Tags

### `update`

Update a cloud server's properties (name, tags).

**Syntax:**
```bash
acloud compute cloudserver update <server-id> [flags]
```

**Arguments:**
- `<server-id>` - The ID of the cloud server

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--name <string>` - New name for the cloud server
- `--tags <stringSlice>` - New tags (comma-separated)

**Example:**
```bash
acloud compute cloudserver update 69495ef64d0cdc87949b71ec \
  --name "web-server-updated" \
  --tags "production,web,updated"
```

**Note:** At least one of `--name` or `--tags` must be provided.

### `delete`

Delete a cloud server instance.

**Syntax:**
```bash
acloud compute cloudserver delete <server-id> [flags]
```

**Arguments:**
- `<server-id>` - The ID of the cloud server

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt

**Example:**
```bash
acloud compute cloudserver delete 69495ef64d0cdc87949b71ec --yes
```

### `power-on`

Power on a cloud server.

**Syntax:**
```bash
acloud compute cloudserver power-on <server-id> [flags]
```

**Arguments:**
- `<server-id>` - The ID of the cloud server

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**Example:**
```bash
acloud compute cloudserver power-on 69495ef64d0cdc87949b71ec
```

**Output:**
```
Cloud server powered on successfully!
Server: web-server
Status: Active
```

### `power-off`

Power off a cloud server.

**Syntax:**
```bash
acloud compute cloudserver power-off <server-id> [flags]
```

**Arguments:**
- `<server-id>` - The ID of the cloud server

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**Example:**
```bash
acloud compute cloudserver power-off 69495ef64d0cdc87949b71ec
```

**Output:**
```
Cloud server powered off successfully!
Server: web-server
Status: Stopped
```

### `set-password`

Set or change the password for a cloud server.

**Syntax:**
```bash
acloud compute cloudserver set-password <server-id> [flags]
```

**Arguments:**
- `<server-id>` - The ID of the cloud server

**Required Flags:**
- `--password <string>` - New password for the cloud server

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**Example:**
```bash
acloud compute cloudserver set-password 69495ef64d0cdc87949b71ec --password "MySecurePassword123!"
```

**Output:**
```
Cloud server password set successfully!
Server ID: 69495ef64d0cdc87949b71ec
```

**Security Note:** Passwords provided via command line flags may be visible in process lists and shell history. Consider using environment variables or secure password management tools.

### `connect`

Get SSH connection information for a cloud server with an Elastic IP.

**Syntax:**
```bash
acloud compute cloudserver connect <server-id> [flags]
```

**Arguments:**
- `<server-id>` - The ID of the cloud server

**Required Flags:**
- `--user <string>` - SSH username (required - see below for image-specific users)

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**SSH Username by Image Type:**
The SSH username depends on the image/template used when creating the cloud server:
- **Ubuntu/Debian images**: Use `ubuntu`
- **CentOS/RHEL images**: Use `centos` or `root`
- **Other Linux distributions**: Typically `root`, but check the image documentation
- **Windows images**: Not applicable (use RDP instead)

For detailed information about accessing Cloud Servers and default users, see the [Aruba Cloud Knowledge Base](https://kb.arubacloud.com/cmp/en/computing/cloud-server.aspx).

**Example:**
```bash
# For Ubuntu/Debian images
acloud compute cloudserver connect 69495ef64d0cdc87949b71ec --user ubuntu

# For CentOS/RHEL images
acloud compute cloudserver connect 69495ef64d0cdc87949b71ec --user centos
```

**Output:**
The command will:
1. Get the cloud server details
2. Check for an Elastic IP in linked resources
3. Retrieve the Elastic IP address
4. Print the SSH connection command

```
Connect by running: ssh ubuntu@203.0.113.42
```

**Note:** 
- The cloud server must have an Elastic IP linked to use this command. If no Elastic IP is found, the command will display an error message.
- The `--user` flag is required. If not provided or set to `<user>`, the command will display an error with guidance on common SSH users.

## Auto-completion

The CLI provides auto-completion for cloud server IDs:

```bash
acloud compute cloudserver get <TAB>
acloud compute cloudserver update <TAB>
acloud compute cloudserver delete <TAB>
acloud compute cloudserver power-on <TAB>
acloud compute cloudserver power-off <TAB>
acloud compute cloudserver set-password <TAB>
acloud compute cloudserver connect <TAB>
```

## Common Workflows

### Launching a New Server

1. **Create a key pair** (if needed):
   ```bash
   acloud compute keypair create --name "my-keypair" --public-key "$(cat ~/.ssh/id_rsa.pub)"
   ```

2. **Create the cloud server**:
   ```bash
   acloud compute cloudserver create \
     --name "app-server" \
     --region "ITBG-Bergamo" \
     --flavor "medium" \
     --image "your-image-id" \
     --keypair "my-keypair" \
     --user-data-file "/path/to/cloud-init.yaml"
   ```

3. **Wait for the server to be ready** and check status:
   ```bash
   acloud compute cloudserver get <server-id>
   ```

### Updating Server Metadata

```bash
# Update server name and tags
acloud compute cloudserver update <server-id> \
  --name "new-name" \
  --tags "production,updated"
```

### Listing and Filtering Servers

```bash
# List all servers
acloud compute cloudserver list

# Use grep to filter by name or tags
acloud compute cloudserver list | grep "production"
```

### Managing Server Power State

```bash
# Power off a server
acloud compute cloudserver power-off <server-id>

# Power on a server
acloud compute cloudserver power-on <server-id>

# Check server status
acloud compute cloudserver get <server-id>
```

### Setting Server Password

```bash
# Set or change server password
acloud compute cloudserver set-password <server-id> --password "NewPassword123!"

# Using environment variable for better security
acloud compute cloudserver set-password <server-id> --password "$SERVER_PASSWORD"
```

### Connecting to a Server via SSH

```bash
# Get SSH connection command (user is required)
# For Ubuntu/Debian images
acloud compute cloudserver connect <server-id> --user ubuntu

# For CentOS/RHEL images
acloud compute cloudserver connect <server-id> --user centos

# The command will output: "Connect by running: ssh user@ip-address"
```

**Important:** The SSH username depends on the image/template used. Common defaults:
- Ubuntu/Debian: `ubuntu`
- CentOS/RHEL: `centos` or `root`
- Other Linux: `root` (check image documentation)

See the [Aruba Cloud Knowledge Base](https://kb.arubacloud.com/cmp/en/computing/cloud-server.aspx) for detailed information about accessing Cloud Servers.

## Best Practices

- **Naming**: Use descriptive names that indicate the server's purpose (e.g., `web-server-prod`, `db-server-staging`)
- **Tags**: Use tags to organize servers by environment, project, or team
- **Flavors**: Choose appropriate flavors based on your workload requirements
- **Key Pairs**: Always use key pairs for SSH access instead of passwords
- **Monitoring**: Check server status before performing operations
- **Cleanup**: Delete unused servers to avoid unnecessary costs

## Related Resources

- [Key Pairs](./keypair.md) - Manage SSH key pairs for server access
- [Network Resources](../network.md) - Configure networking and security groups
- [Storage Resources](../storage.md) - Attach block storage volumes

