# Container Registry

Container Registry provides a private Docker container registry for storing and managing container images in Aruba Cloud.

## Command Syntax

```bash
acloud container containerregistry <command> [flags] [arguments]
```

## Available Commands

### `create`

Create a new container registry.

**Syntax:**
```bash
acloud container containerregistry create [flags]
```

**Required Flags:**
- `--name <string>` - Name for the container registry
- `--region <string>` - Region code (e.g., `ITBG-Bergamo`)
- `--public-ip-id <string>` - Elastic IP ID
- `--vpc-id <string>` - VPC ID
- `--subnet-id <string>` - Subnet ID
- `--security-group-id <string>` - Security group ID
- `--block-storage-id <string>` - Block storage ID

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--tags <stringSlice>` - Tags (comma-separated)
- `--billing-period <string>` - Billing period: Hour, Month, Year (default: Hour)
- `--admin-username <string>` - Administrator username
- `--concurrent-users <string>` - Concurrent users plan (e.g., `Small`, or a numeric value like `20`)

**Example:**
```bash
acloud container containerregistry create \
  --name "my-registry" \
  --region "ITBG-Bergamo" \
  --public-ip-id "<eip-id>" \
  --vpc-id "<vpc-id>" \
  --subnet-id "<subnet-id>" \
  --security-group-id "<sg-id>" \
  --block-storage-id "<vol-id>" \
  --admin-username "adminuser" \
  --concurrent-users "Small" \
  --billing-period "Hour" \
  --tags "production,registry"
```

**Note:** The security group must allow TCP/443 (HTTPS) ingress and allow egress traffic for the registry to function correctly.

### `list`

List all container registries in the project.

**Syntax:**
```bash
acloud container containerregistry list [flags]
```

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**Example:**
```bash
acloud container containerregistry list
```

**Output:**
The command displays a table with the following columns:
- NAME
- ID
- REGION
- STATUS

### `get`

Get detailed information about a specific container registry.

**Syntax:**
```bash
acloud container containerregistry get <registry-id> [flags]
```

**Arguments:**
- `<registry-id>` - The ID of the container registry

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--verbose` - Show detailed JSON output

**Example:**
```bash
acloud container containerregistry get 69495ef64d0cdc87949b71ec
```

**Output:**
The command displays detailed information including:
- ID and URI
- Name and region
- Public IP, VPC, Subnet, Security Group, Block Storage
- Billing plan
- Admin user (if configured)
- Concurrent users
- Status
- Creation date and creator
- Tags

### `update`

Update a container registry's properties (name, tags, billing period, concurrent users).

**Syntax:**
```bash
acloud container containerregistry update <registry-id> [flags]
```

**Arguments:**
- `<registry-id>` - The ID of the container registry

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--name <string>` - New name for the container registry
- `--tags <stringSlice>` - New tags (comma-separated)
- `--billing-period <string>` - Billing period: Hour, Month, Year
- `--concurrent-users <string>` - Number of concurrent users

**Example:**
```bash
acloud container containerregistry update 69495ef64d0cdc87949b71ec \
  --name "my-registry-updated" \
  --tags "production,registry,updated" \
  --billing-period "Year" \
  --concurrent-users 10
```

**Note:** At least one of `--name`, `--tags`, `--billing-period`, or `--concurrent-users` must be provided.

### `delete`

Delete a container registry.

**Syntax:**
```bash
acloud container containerregistry delete <registry-id> [flags]
```

**Arguments:**
- `<registry-id>` - The ID of the container registry

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt

**Example:**
```bash
acloud container containerregistry delete 69495ef64d0cdc87949b71ec --yes
```

## Auto-completion

The CLI provides auto-completion for container registry IDs:

```bash
acloud container containerregistry get <TAB>
acloud container containerregistry update <TAB>
acloud container containerregistry delete <TAB>
```

## Common Workflows

### Creating a Container Registry

1. **Ensure required resources exist:**
   - Public IP (Elastic IP)
   - VPC
   - Subnet
   - Security Group
   - Block Storage

2. **Create the container registry:**
   ```bash
   acloud container containerregistry create \
     --name "my-registry" \
     --region "ITBG-Bergamo" \
     --public-ip-id "<eip-id>" \
     --vpc-id "<vpc-id>" \
     --subnet-id "<subnet-id>" \
     --security-group-id "<sg-id>" \
     --block-storage-id "<vol-id>" \
     --admin-username "adminuser" \
     --concurrent-users "Small" \
     --billing-period "Hour"
   ```

3. **Wait for the registry to be ready** and check status:
   ```bash
   acloud container containerregistry get <registry-id>
   ```

### Updating Registry Metadata

```bash
# Update registry name and tags
acloud container containerregistry update <registry-id> \
  --name "new-name" \
  --tags "production,updated"

# Update billing period
acloud container containerregistry update <registry-id> \
  --billing-period "Year"

# Update concurrent users
acloud container containerregistry update <registry-id> \
  --concurrent-users 20
```

### Listing and Filtering Registries

```bash
# List all registries
acloud container containerregistry list

# Use grep to filter by name or tags
acloud container containerregistry list | grep "production"
```

## Best Practices

- **Naming**: Use descriptive names that indicate the registry's purpose (e.g., `prod-registry`, `dev-registry`)
- **Tags**: Use tags to organize registries by environment, project, or team
- **Billing Period**: Choose appropriate billing periods based on expected usage
- **Concurrent Users**: Set concurrent users based on your team size and usage patterns
- **Security**: Ensure security groups are properly configured to restrict access
- **Storage**: Monitor block storage usage and expand as needed
- **Cleanup**: Delete unused registries to avoid unnecessary costs

## Related Resources

- [KaaS](./kaas.md) - Kubernetes clusters for running containerized applications
- [Network Resources](../network.md) - Configure networking, VPCs, and security groups
- [Storage Resources](../storage.md) - Block storage for registry data

