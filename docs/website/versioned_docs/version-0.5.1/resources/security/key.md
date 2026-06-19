# Cryptographic Keys

Cryptographic keys are nested inside a KMS instance and provide the actual encryption material. Each key has an algorithm (AES symmetric or RSA asymmetric) and a lifecycle status.

## Available Commands

- `acloud security key create` - Create a new cryptographic key inside a KMS instance
- `acloud security key list` - List all keys in a KMS instance
- `acloud security key get` - Get details of a specific key
- `acloud security key delete` - Delete a key

## Create Key

Create a new cryptographic key inside an existing KMS instance.

### Usage

```bash
acloud security key create --kms-id <kms-id> --name <name> --algorithm <algorithm> [flags]
```

### Required Flags

- `--kms-id` - ID of the parent KMS instance
- `--name` - Name for the key
- `--algorithm` - Cryptographic algorithm: `Aes` (symmetric) or `Rsa` (asymmetric)

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud security key create \
  --kms-id "69455aa70d0972656501d45d" \
  --name "my-aes-key" \
  --algorithm "Aes"
```

## List Keys

List all keys inside a KMS instance.

### Usage

```bash
acloud security key list --kms-id <kms-id> [flags]
```

### Required Flags

- `--kms-id` - ID of the parent KMS instance

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--limit` - Maximum number of results to return
- `--offset` - Number of results to skip

### Example

```bash
acloud security key list --kms-id "69455aa70d0972656501d45d"
```

## Get Key Details

Retrieve detailed information about a specific key.

### Usage

```bash
acloud security key get <key-id> --kms-id <kms-id> [flags]
```

### Arguments

- `key-id` (required): The unique ID of the key

### Required Flags

- `--kms-id` - ID of the parent KMS instance

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud security key get abc123 --kms-id "69455aa70d0972656501d45d"
```

## Delete Key

Delete a cryptographic key. This action is irreversible.

### Usage

```bash
acloud security key delete <key-id> --kms-id <kms-id> [--yes] [flags]
```

### Arguments

- `key-id` (required): The unique ID of the key

### Required Flags

- `--kms-id` - ID of the parent KMS instance

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt
- `--dry-run` - Validate resource exists without deleting

### Example

```bash
acloud security key delete abc123 --kms-id "69455aa70d0972656501d45d" --yes
```

## Related Resources

- [KMS Key Management](kms.md) - Manage the KMS instances that contain keys
- [KMIP Services](kmip.md) - Manage KMIP services nested inside KMS instances
