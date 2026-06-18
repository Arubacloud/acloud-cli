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

### [Cryptographic Keys](security/key.md)

Cryptographic keys are nested inside a KMS instance and provide AES or RSA encryption material.

**Quick Commands:**
```bash
# List all keys in a KMS instance
acloud security key list --kms-id <kms-id>

# Get key details
acloud security key get <key-id> --kms-id <kms-id>

# Create a key
acloud security key create --kms-id <kms-id> --name "my-key" --algorithm "Aes"

# Delete a key
acloud security key delete <key-id> --kms-id <kms-id>
```

### [KMIP Services](security/kmip.md)

KMIP services are nested inside a KMS instance and expose a KMIP endpoint for certificate-based client authentication.

**Quick Commands:**
```bash
# List all KMIP services in a KMS instance
acloud security kmip list --kms-id <kms-id>

# Get KMIP service details
acloud security kmip get <kmip-id> --kms-id <kms-id>

# Create a KMIP service (and wait for certificate)
acloud security kmip create --kms-id <kms-id> --name "my-kmip" --wait

# Download certificate and private key (PEM)
acloud security kmip download <kmip-id> --kms-id <kms-id>

# Delete a KMIP service
acloud security kmip delete <kmip-id> --kms-id <kms-id>
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

