# DBaaS Management

DBaaS (Database as a Service) instances provide managed database services in Aruba Cloud.

## Available Commands

- `acloud database dbaas create` - Create a new DBaaS instance
- `acloud database dbaas list` - List all DBaaS instances
- `acloud database dbaas get` - Get details of a specific DBaaS instance
- `acloud database dbaas update` - Update DBaaS instance tags
- `acloud database dbaas delete` - Delete a DBaaS instance

## Create DBaaS Instance

Create a new DBaaS instance in your project.

### Usage

```bash
acloud database dbaas create --name <name> --region <region> --zone <zone> --engine-id <engine-id> --flavor <flavor> --storage-size <gb> [flags]
```

### Required Flags

- `--name` - Name for the DBaaS instance
- `--region` - Region code (e.g., "ITBG-Bergamo")
- `--zone` - Availability zone (e.g., "ITBG-1")
- `--engine-id` - Database engine identifier (e.g., `mysql-8.0`)
- `--flavor` - Database flavor/plan code (e.g., `DBO4A8`)
- `--storage-size` - Storage size in GB (e.g., `50`)

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--vpc-id` - VPC ID (required when the project has a VPC)
- `--subnet-id` - Subnet ID (required when the project has a VPC)
- `--security-group-id` - Security group ID (required when the project has a VPC)
- `--elastic-ip-id` - Elastic IP ID for public access
- `--tags` - Tags (comma-separated)

### Example

```bash
acloud database dbaas create \
  --name "my-database" \
  --region "ITBG-Bergamo" \
  --zone "ITBG-1" \
  --engine-id "mysql-8.0" \
  --flavor "DBO4A8" \
  --storage-size 50 \
  --vpc-id "<vpc-id>" \
  --subnet-id "<subnet-id>" \
  --security-group-id "<sg-id>" \
  --elastic-ip-id "<eip-id>" \
  --tags "production,mysql"
```

**Note:** Database and user `update` (PUT) operations are not supported by the API and will return HTTP 405. Use `delete` and `create` to replace a database or user.

## List DBaaS Instances

List all DBaaS instances in your project.

### Usage

```bash
acloud database dbaas list [flags]
```

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas list
```

## Get DBaaS Instance Details

Retrieve detailed information about a specific DBaaS instance.

### Usage

```bash
acloud database dbaas get <dbaas-id> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas get 69455aa70d0972656501d45d
```

## Update DBaaS Instance

Update tags for a DBaaS instance.

### Usage

```bash
acloud database dbaas update <dbaas-id> --tags <tags> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--tags` - New tags (comma-separated)

### Example

```bash
acloud database dbaas update 69455aa70d0972656501d45d --tags "production,updated"
```

## Delete DBaaS Instance

Delete a DBaaS instance.

### Usage

```bash
acloud database dbaas delete <dbaas-id> [--yes] [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt

### Example

```bash
acloud database dbaas delete 69455aa70d0972656501d45d --yes
```

## Related Resources

- [DBaaS Databases](dbaas.database.md) - Manage databases within DBaaS instances
- [DBaaS Users](dbaas.user.md) - Manage users for DBaaS instances
- [Database Backups](backup.md) - Create and manage database backups

