# DBaaS User Management

Users that can access databases within DBaaS instances. Each user has a username and password for authentication.

## Available Commands

- `acloud database dbaas user create` - Create a new user in a DBaaS instance
- `acloud database dbaas user list` - List all users in a DBaaS instance
- `acloud database dbaas user get` - Get details of a specific user
- `acloud database dbaas user update` - Update user password
- `acloud database dbaas user delete` - Delete a user

## Create User

Create a new user within a DBaaS instance.

### Usage

```bash
acloud database dbaas user create <dbaas-id> --username <username> --password <password> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance

### Required Flags

- `--username` - Username for the user
- `--password` - Password for the user

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas user create 69455aa70d0972656501d45d \
  --username "app-user" \
  --password "SecurePassword123!"
```

## List Users

List all users in a DBaaS instance.

### Usage

```bash
acloud database dbaas user list <dbaas-id> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas user list 69455aa70d0972656501d45d
```

## Get User Details

Retrieve detailed information about a specific user.

### Usage

```bash
acloud database dbaas user get <dbaas-id> <username> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance
- `username` (required): The username

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas user get 69455aa70d0972656501d45d "app-user"
```

## Update User

Change a user's password.

### Usage

```bash
acloud database dbaas user update <dbaas-id> <username> --password <new-password> [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance
- `username` (required): The username

### Required Flags

- `--password` - New password for the user

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud database dbaas user update 69455aa70d0972656501d45d "app-user" \
  --password "NewSecurePassword456!"
```

## Delete User

Delete a user from a DBaaS instance.

### Usage

```bash
acloud database dbaas user delete <dbaas-id> <username> [--yes] [flags]
```

### Arguments

- `dbaas-id` (required): The unique ID of the DBaaS instance
- `username` (required): The username to delete

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt

### Example

```bash
acloud database dbaas user delete 69455aa70d0972656501d45d "app-user" --yes
```

## Security Best Practices

- Use strong passwords (minimum 12 characters, mix of letters, numbers, and symbols)
- Rotate passwords regularly
- Use different passwords for different users
- Never share passwords or commit them to version control
- Consider using a password manager

## Related Resources

- [DBaaS](dbaas.md) - Manage DBaaS instances
- [DBaaS Databases](dbaas.database.md) - Manage databases
- [Database Backups](backup.md) - Create backups of databases

