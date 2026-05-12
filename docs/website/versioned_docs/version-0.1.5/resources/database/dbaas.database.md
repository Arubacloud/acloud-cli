# DBaaS Database Management

Databases within DBaaS instances store your data. Each DBaaS instance can contain multiple databases.

## Available Commands

- `acloud database dbaas database create` - Create a new database in a DBaaS instance
- `acloud database dbaas database list` - List all databases in a DBaaS instance
- `acloud database dbaas database get` - Get details of a specific database
- `acloud database dbaas database update` - Update database name
- `acloud database dbaas database delete` - Delete a database

## Create Database

Create a new database within a DBaaS instance.

### Usage

```bash
acloud database dbaas database create <dbaas-id> --name <name> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance

### Required Flags

- `--name` - Name for the database

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas database create 69455aa70d0972656501d45d \
  --name "my-database"
```

## List Databases

List all databases in a DBaaS instance.

### Usage

```bash
acloud database dbaas database list <dbaas-id> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas database list 69455aa70d0972656501d45d
```

## Get Database Details

Retrieve detailed information about a specific database.

### Usage

```bash
acloud database dbaas database get <dbaas-id> <database-name> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance
- `database-name` (required): The name of the database

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas database get 69455aa70d0972656501d45d "my-database"
```

## Update Database

Rename a database.

### Usage

```bash
acloud database dbaas database update <dbaas-id> <database-name> --name <new-name> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance
- `database-name` (required): The current name of the database

### Required Flags

- `--name` - New name for the database

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas database update 69455aa70d0972656501d45d "my-database" \
  --name "renamed-database"
```

## Delete Database

Delete a database from a DBaaS instance.

### Usage

```bash
acloud database dbaas database delete <dbaas-id> <database-name> [--yes] [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance
- `database-name` (required): The name of the database to delete

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt

### Example

```bash
acloud database dbaas database delete 69455aa70d0972656501d45d "my-database" --yes
```

## Related Resources

- [DBaaS](dbaas.md) - Manage DBaaS instances
- [DBaaS Users](dbaas.user.md) - Manage users for databases
- [Database Backups](backup.md) - Create backups of databases

