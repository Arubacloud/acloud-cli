# Security Resources

The `security` category provides commands for managing security resources in Aruba Cloud, including Key Management System (KMS) keys for encryption and security.

## Available Resources

### [KMS Keys](security/kms.md)

KMS (Key Management Service) keys provide encryption key management for securing your data and resources.

**Quick Commands:**
```bash
# List all KMS keys
acloud security kms list

# Get KMS key details
acloud security kms get <kms-id>

# Create a KMS key
acloud security kms create --name "my-kms-key" --region "ITBG-Bergamo" --billing-period "Hour"

# Update a KMS key
acloud security kms update <kms-id> --name "updated-name" --tags "production"

# Delete a KMS key
acloud security kms delete <kms-id>
```

## Command Structure

All security commands follow this structure:

```
acloud security <resource> <action> [arguments] [flags]
```

Where:
- `<resource>`: The type of resource (e.g., `kms`)
- `<action>`: The operation to perform (e.g., `list`, `get`, `create`, `update`, `delete`)
- `[arguments]`: Required arguments (e.g., resource IDs)
- `[flags]`: Optional flags (e.g., `--name`, `--region`, `--tags`)

## Common Patterns

### Listing Resources

```bash
acloud security <resource> list
```

### Getting Resource Details

```bash
acloud security <resource> get <resource-id>
```

### Creating Resources

```bash
acloud security <resource> create [required-args] [flags]
```

### Updating Resources

```bash
acloud security <resource> update <resource-id> [flags]
```

### Deleting Resources

```bash
acloud security <resource> delete <resource-id> [--yes]
```

## Project Context

Security resources are scoped to a project. You can either:

1. **Use the `--project-id` flag:**
   ```bash
   acloud security kms list --project-id <project-id>
   ```

2. **Set a context:**
   ```bash
   acloud context set my-prod --project-id <project-id>
   acloud security kms list  # Uses context project ID
   ```

See [Installation - Context Management](../installation.md#context-management) for more information.

## Next Steps

- Explore [Management Resources](management.md) for organization-level resources
- Check [Storage Resources](storage.md) for storage operations
- Review [Network Resources](network.md) for networking capabilities
- See [Database Resources](database.md) for database management
- Review [Schedule Resources](schedule.md) for job scheduling

