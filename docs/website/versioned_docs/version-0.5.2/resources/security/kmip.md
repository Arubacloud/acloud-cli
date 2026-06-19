# KMIP Services

KMIP (Key Management Interoperability Protocol) services are nested inside a KMS instance and expose a KMIP endpoint for client certificate-based authentication. After creation, a certificate and private key pair can be downloaded once the service reaches `CertificateAvailable` status.

## Available Commands

- `acloud security kmip create` - Create a new KMIP service inside a KMS instance
- `acloud security kmip list` - List all KMIP services in a KMS instance
- `acloud security kmip get` - Get details of a specific KMIP service
- `acloud security kmip delete` - Delete a KMIP service
- `acloud security kmip download` - Download the KMIP certificate and private key (PEM)

## Create KMIP Service

Create a new KMIP service inside an existing KMS instance.

### Usage

```bash
acloud security kmip create --kms-id <kms-id> --name <name> [flags]
```

### Required Flags

- `--kms-id` - ID of the parent KMS instance
- `--name` - Name for the KMIP service

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--wait` - Block until the certificate becomes available (respects `--timeout`)

### Example

```bash
acloud security kmip create \
  --kms-id "69455aa70d0972656501d45d" \
  --name "my-kmip-service" \
  --wait
```

## List KMIP Services

List all KMIP services inside a KMS instance.

### Usage

```bash
acloud security kmip list --kms-id <kms-id> [flags]
```

### Required Flags

- `--kms-id` - ID of the parent KMS instance

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--limit` - Maximum number of results to return
- `--offset` - Number of results to skip

### Example

```bash
acloud security kmip list --kms-id "69455aa70d0972656501d45d"
```

## Get KMIP Service Details

Retrieve detailed information about a specific KMIP service.

### Usage

```bash
acloud security kmip get <kmip-id> --kms-id <kms-id> [flags]
```

### Arguments

- `kmip-id` (required): The unique ID of the KMIP service

### Required Flags

- `--kms-id` - ID of the parent KMS instance

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud security kmip get abc123 --kms-id "69455aa70d0972656501d45d"
```

## Delete KMIP Service

Delete a KMIP service.

### Usage

```bash
acloud security kmip delete <kmip-id> --kms-id <kms-id> [--yes] [flags]
```

### Arguments

- `kmip-id` (required): The unique ID of the KMIP service

### Required Flags

- `--kms-id` - ID of the parent KMS instance

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt
- `--dry-run` - Validate resource exists without deleting

### Example

```bash
acloud security kmip delete abc123 --kms-id "69455aa70d0972656501d45d" --yes
```

## Download Certificate

Download the PEM-encoded certificate and private key for a KMIP service. The service must have reached `CertificateAvailable` status before the download is available.

### Usage

```bash
acloud security kmip download <kmip-id> --kms-id <kms-id> [flags]
```

### Arguments

- `kmip-id` (required): The unique ID of the KMIP service

### Required Flags

- `--kms-id` - ID of the parent KMS instance

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud security kmip download abc123 --kms-id "69455aa70d0972656501d45d"
```

The output contains both the certificate and private key in PEM format, which you can redirect to files:

```bash
acloud security kmip download abc123 --kms-id "69455aa70d0972656501d45d" > kmip-cert.pem
```

## Related Resources

- [KMS Key Management](kms.md) - Manage the KMS instances that contain KMIP services
- [Cryptographic Keys](key.md) - Manage cryptographic keys nested inside KMS instances
