# KMS Key Management

KMS (Key Management System) keys provide encryption key management for securing your data and resources in Aruba Cloud.

## Available Commands

- `acloud security kms create` - Create a new KMS key
- `acloud security kms list` - List all KMS keys
- `acloud security kms get` - Get details of a specific KMS key
- `acloud security kms update` - Update KMS key name and tags
- `acloud security kms delete` - Delete a KMS key

## Create KMS Key

Create a new KMS key in your project.

### Usage

```bash
acloud security kms create --name <name> --region <region> [flags]
```

### Required Flags

- `--name` - Name for the KMS key
- `--region` - Region code (e.g., "ITBG-Bergamo")

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--billing-period` - Billing period: Hour, Month, Year (default: "Hour")
- `--tags` - Tags (comma-separated)

### Example

```bash
acloud security kms create \
  --name "my-encryption-key" \
  --region "ITBG-Bergamo" \
  --billing-period "Hour" \
  --tags "production,encryption"
```

## List KMS Keys

List all KMS keys in your project.

### Usage

```bash
acloud security kms list [flags]
```

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud security kms list
```

## Get KMS Key Details

Retrieve detailed information about a specific KMS key.

### Usage

```bash
acloud security kms get <kms-id> [flags]
```

### Arguments

- `kms-id` (required): The unique ID of the KMS key

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud security kms get 69455aa70d0972656501d45d
```

## Update KMS Key

Update KMS key name and tags.

### Usage

```bash
acloud security kms update <kms-id> [flags]
```

### Arguments

- `kms-id` (required): The unique ID of the KMS key

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--name` - New name for the KMS key
- `--tags` - New tags (comma-separated)

### Example

```bash
acloud security kms update 69455aa70d0972656501d45d \
  --name "updated-key-name" \
  --tags "production,updated"
```

## Delete KMS Key

Delete a KMS key.

### Usage

```bash
acloud security kms delete <kms-id> [--yes] [flags]
```

### Arguments

- `kms-id` (required): The unique ID of the KMS key

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt

### Example

```bash
acloud security kms delete 69455aa70d0972656501d45d --yes
```

## Security Best Practices

- Use descriptive names for KMS keys
- Organize keys using tags
- Rotate keys regularly according to your security policy
- Monitor key usage and access
- Use different keys for different environments (dev, staging, production)
- Never share or expose key material

## Related Resources

- [Database Resources](../database/dbaas.md) - Use KMS keys with databases
- [Storage Resources](../storage/blockstorage.md) - Encrypt storage volumes

