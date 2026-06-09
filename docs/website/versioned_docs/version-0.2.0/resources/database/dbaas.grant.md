# DBaaS Grant Management

Grants define which users have access to which databases within a DBaaS instance, and at what privilege level (e.g., `liteadmin`).

## Available Commands

- `acloud database dbaas grant create` - Grant a user access to a database
- `acloud database dbaas grant list` - List all grants on a database
- `acloud database dbaas grant get` - Get details of a specific grant
- `acloud database dbaas grant delete` - Revoke a grant

## Create Grant

Grant a user access to a specific database within a DBaaS instance.

### Usage

```bash
acloud database dbaas grant create <dbaas-id> <database-name> --username <username> --role <role> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance
- `database-name` (required): The name of the database to grant access to

### Required Flags

- `--username` - The username to grant access to
- `--role` - The privilege role (e.g., `liteadmin`)

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas grant create 69455aa70d0972656501d45d "my-database" \
  --username "restapi" \
  --role "liteadmin"
```

> **Note:** The DBaaS instance must be in `Active` state before creating grants. The grant update (PUT) operation is not supported by the API — revoke and re-create to change a grant's role.

## List Grants

List all grants on a specific database within a DBaaS instance.

### Usage

```bash
acloud database dbaas grant list <dbaas-id> <database-name> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance
- `database-name` (required): The name of the database

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas grant list 69455aa70d0972656501d45d "my-database"
```

## Get Grant Details

Retrieve details of a specific grant.

### Usage

```bash
acloud database dbaas grant get <dbaas-id> <database-name> <grant-id> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance
- `database-name` (required): The name of the database
- `grant-id` (required): The unique ID of the grant

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas grant get 69455aa70d0972656501d45d "my-database" 69455aa70d0972656501d4ab
```

## Delete Grant

Revoke a user's access to a database.

### Usage

```bash
acloud database dbaas grant delete <dbaas-id> <database-name> <grant-id> [--yes] [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance
- `database-name` (required): The name of the database
- `grant-id` (required): The unique ID of the grant to revoke

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt

### Example

```bash
acloud database dbaas grant delete 69455aa70d0972656501d45d "my-database" 69455aa70d0972656501d4ab --yes
```

> **Note:** Grants are also removed when the parent database or DBaaS instance is deleted.

## Typical Workflow

```bash
# 1. Create a DBaaS instance and wait for Active status
acloud database dbaas create --name "my-dbaas" --region "ITBG-Bergamo" --zone "ITBG-1" \
  --engine-id "mysql-8.0" --flavor "DBO4A8" --storage-size 50 \
  --vpc-id "<vpc-id>" --subnet-id "<subnet-id>" --security-group-id "<sg-id>" \
  --elastic-ip-id "<eip-id>"

# 2. Create a database
acloud database dbaas database create <dbaas-id> --name "appdb"

# 3. Create a user
acloud database dbaas user create <dbaas-id> --username "restapi" --password "SecurePass123!"

# 4. Grant the user access
acloud database dbaas grant create <dbaas-id> "appdb" --username "restapi" --role "liteadmin"

# 5. Verify the grant
acloud database dbaas grant list <dbaas-id> "appdb"
```

## Related Resources

- [DBaaS](dbaas.md) - Manage DBaaS instances
- [DBaaS Databases](dbaas.database.md) - Manage databases within DBaaS instances
- [DBaaS Users](dbaas.user.md) - Manage users for DBaaS instances
- [Database Backups](backup.md) - Create and manage database backups
