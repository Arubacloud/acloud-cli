# Database Resources

The `database` category provides commands for managing database resources in Aruba Cloud, including DBaaS (Database as a Service) instances, databases, users, and backups.

## Available Resources

### [DBaaS](database/dbaas.md)

DBaaS (Database as a Service) instances provide managed database services in Aruba Cloud.

**Quick Commands:**
```bash
# List all DBaaS instances
acloud database dbaas list

# Get DBaaS instance details
acloud database dbaas get <dbaas-id>

# Create a DBaaS instance
acloud database dbaas create --name "my-dbaas" --region "ITBG-Bergamo" --engine-id <engine-id> --flavor <flavor>

# Update a DBaaS instance
acloud database dbaas update <dbaas-id> --tags "production,updated"

# Delete a DBaaS instance
acloud database dbaas delete <dbaas-id>
```

### [DBaaS Databases](database/dbaas.database.md)

Databases within DBaaS instances that store your data.

**Quick Commands:**
```bash
# List all databases in a DBaaS instance
acloud database dbaas database list <dbaas-id>

# Get database details
acloud database dbaas database get <dbaas-id> <database-name>

# Create a database
acloud database dbaas database create <dbaas-id> --name "my-database"

# Update a database (rename)
acloud database dbaas database update <dbaas-id> <database-name> --name "new-name"

# Delete a database
acloud database dbaas database delete <dbaas-id> <database-name>
```

### [DBaaS Users](database/dbaas.user.md)

Users that can access databases within DBaaS instances.

**Quick Commands:**
```bash
# List all users in a DBaaS instance
acloud database dbaas user list <dbaas-id>

# Get user details
acloud database dbaas user get <dbaas-id> <username>

# Create a user
acloud database dbaas user create <dbaas-id> --username "my-user" --password "secure-password"

# Update a user (change password)
acloud database dbaas user update <dbaas-id> <username> --password "new-password"

# Delete a user
acloud database dbaas user delete <dbaas-id> <username>
```

### [Database Backups](database/backup.md)

Backups of databases for disaster recovery and point-in-time restoration.

**Quick Commands:**
```bash
# List all database backups
acloud database backup list

# Get backup details
acloud database backup get <backup-id>

# Create a backup
acloud database backup create --name "my-backup" --region "ITBG-Bergamo" --dbaas-id <dbaas-id> --database-name <database-name>

# Delete a backup
acloud database backup delete <backup-id>
```

## Command Structure

All database commands follow this structure:

```
acloud database <resource> <action> [arguments] [flags]
```

Where:
- `<resource>`: The type of resource (e.g., `dbaas`, `dbaas database`, `dbaas user`, `backup`)
- `<action>`: The operation to perform (e.g., `list`, `get`, `create`, `update`, `delete`)
- `[arguments]`: Required arguments (e.g., resource IDs, names)
- `[flags]`: Optional flags (e.g., `--name`, `--region`, `--tags`)

## Common Patterns

### Listing Resources

```bash
acloud database <resource> list
```

### Getting Resource Details

```bash
acloud database <resource> get <resource-id>
```

### Creating Resources

```bash
acloud database <resource> create [required-args] [flags]
```

### Updating Resources

```bash
acloud database <resource> update <resource-id> [flags]
```

### Deleting Resources

```bash
acloud database <resource> delete <resource-id> [--yes]
```

## Project Context

Database resources are scoped to a project. You can either:

1. **Use the `--project-id` flag:**
   ```bash
   acloud database dbaas list --project-id <project-id>
   ```

2. **Set a context:**
   ```bash
   acloud context set my-prod --project-id <project-id>
   acloud database dbaas list  # Uses context project ID
   ```

See [Installation - Context Management](../installation.md#context-management) for more information.

## Next Steps

- Explore [Management Resources](management.md) for organization-level resources
- Check [Storage Resources](storage.md) for storage operations
- Review [Network Resources](network.md) for networking capabilities

