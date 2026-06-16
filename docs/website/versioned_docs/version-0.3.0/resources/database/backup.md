# Database Backup Management

Database backups provide point-in-time copies of your databases for disaster recovery and restoration.

## Available Commands

- `acloud database backup create` - Create a new database backup
- `acloud database backup list` - List all database backups
- `acloud database backup get` - Get details of a specific backup
- `acloud database backup delete` - Delete a database backup

**Note:** Database backups do not support update operations.

## Create Backup

Create a new backup of a database in a DBaaS instance.

### Usage

```bash
acloud database backup create --name <name> --region <region> --dbaas-id <dbaas-id> --database-name <database-name> [flags]
```

### Required Flags

- `--name` - Name for the backup
- `--region` - Region code (e.g., "ITBG-Bergamo")
- `--dbaas-id` - DBaaS instance ID
- `--database-name` - Name of the database to backup

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--billing-period` - Billing period: Hour, Month, Year (default: "Hour")
- `--tags` - Tags (comma-separated)

### Example

```bash
acloud database backup create \
  --name "daily-backup" \
  --region "ITBG-Bergamo" \
  --dbaas-id "69455aa70d0972656501d45d" \
  --database-name "my-database" \
  --billing-period "Hour" \
  --tags "daily,automated"
```

## List Backups

List all database backups in your project.

### Usage

```bash
acloud database backup list [flags]
```

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database backup list
```

## Get Backup Details

Retrieve detailed information about a specific backup.

### Usage

```bash
acloud database backup get <backup-id> [flags]
```

### Arguments

- `backup-id` (required): The unique ID of the backup

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database backup get 69455aa70d0972656501d45d
```

## Delete Backup

Delete a database backup.

### Usage

```bash
acloud database backup delete <backup-id> [--yes] [flags]
```

### Arguments

- `backup-id` (required): The unique ID of the backup

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt

### Example

```bash
acloud database backup delete 69455aa70d0972656501d45d --yes
```

## Backup Best Practices

- Create regular backups of critical databases
- Store backups in different regions for disaster recovery
- Test restore procedures regularly
- Keep multiple backup versions
- Automate backup creation using scheduled jobs

## Related Resources

- [DBaaS](dbaas.md) - Manage DBaaS instances
- [DBaaS Databases](dbaas.database.md) - Manage databases
- [Schedule Jobs](../schedule/job.md) - Automate backup creation

