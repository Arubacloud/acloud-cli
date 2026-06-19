---
id: database
title: Database (DBaaS)
sidebar_label: Database
description: End-to-end workflow for provisioning a MySQL DBaaS instance, creating a database, managing users and permissions, and connecting.
---

# Database (DBaaS) Example

This guide covers the complete end-to-end workflow for provisioning a MySQL DBaaS instance using the Aruba Cloud CLI: from network setup to database creation, user management, and connecting.

## Prerequisites

Before starting, ensure you have:
- The CLI configured with valid credentials (`acloud config set`)
- An active project (use `acloud management project list` to verify)
- An active VPC, or follow Step 0 to prepare networking resources

---

## Step 0: Prepare Networking Resources

### List Available VPCs

```bash
acloud network vpc list
```

Example output:
```
NAME       ID                        REGION         SUBNETS    STATUS
prod-vpc   69495ef64d0cdc87949b71ec  ITBG-Bergamo   3          Active
```

Note the VPC `ID`. Ensure `STATUS` is `Active` before proceeding.

---

### List or Create a Subnet

```bash
acloud network subnet list 69495ef64d0cdc87949b71ec
```

Example output:
```
NAME        ID                         REGION         CIDR              STATUS
db-subnet   694ba1737712ac0032dbe50a   ITBG-Bergamo   192.168.10.0/24   Active
```

Note the subnet `ID`.

---

### List or Create a Security Group

The security group must allow **inbound TCP on port 3306** for MySQL connectivity.

```bash
acloud network securitygroup list 69495ef64d0cdc87949b71ec
```

If you need to create a new security group with a MySQL inbound rule:

```bash
# Create security group
acloud network securitygroup create \
  --name "db-security-group" \
  --vpc-id "69495ef64d0cdc87949b71ec" \
  --region "ITBG-Bergamo"

# Add inbound rule for MySQL (port 3306)
acloud network securityrule create 69495ef64d0cdc87949b71ec <securitygroup-id> \
  --direction Inbound \
  --protocol TCP \
  --port-range-min 3306 \
  --port-range-max 3306 \
  --remote-ip-prefix "0.0.0.0/0"
```

Note the security group `ID`.

---

### Get an Elastic IP for Public Access

```bash
acloud network elasticip list
```

Note the Elastic IP `ID`. An Elastic IP is required to connect to the DBaaS instance from outside the VPC.

---

## Step 1: Create the DBaaS Instance

Create a new MySQL 8.0 DBaaS instance with all required networking flags:

```bash
acloud database dbaas create \
  --name "prod-mysql" \
  --region "ITBG-Bergamo" \
  --zone "ITBG-1" \
  --engine-id "mysql-8.0" \
  --flavor "DBO4A8" \
  --storage-size 50 \
  --vpc-id "69495ef64d0cdc87949b71ec" \
  --subnet-id "694ba1737712ac0032dbe50a" \
  --security-group-id "694b05ac4d0cdc87949b75f9" \
  --elastic-ip-id "694bb7897712ac0032dbe60c" \
  --tags "production,mysql"
```

Example output:
```
DBaaS instance created successfully!
ID:              69455aa70d0972656501d45d
Name:            prod-mysql
Engine:          mysql-8.0
Flavor:          DBO4A8
Storage (GB):    50
Region:          ITBG-Bergamo
Status:          InCreation
Creation Date:   18-06-2026 10:00:00
```

> **Note**: DBaaS provisioning typically takes several minutes. Use `--wait` with an extended timeout to block until the instance is `Active`:
> ```bash
> acloud --timeout 15m database dbaas create ... --wait
> ```

---

## Step 2: Wait for the DBaaS Instance to Become Active

Monitor the provisioning status:

```bash
acloud database dbaas list
```

Wait until `STATUS` shows `Active`:

```
NAME         ID                        ENGINE     FLAVOR    STORAGE  REGION         STATUS
prod-mysql   69455aa70d0972656501d45d  mysql-8.0  DBO4A8    50 GB    ITBG-Bergamo   Active
```

Get full details including the connection endpoint:

```bash
acloud database dbaas get 69455aa70d0972656501d45d
```

---

## Step 3: Create a Database

Once the DBaaS instance is `Active`, create a database within it:

```bash
acloud database dbaas database create 69455aa70d0972656501d45d \
  --name "appdb"
```

Example output:
```
Database created successfully!
DBaaS ID:   69455aa70d0972656501d45d
Name:       appdb
```

Verify the database was created:

```bash
acloud database dbaas database list 69455aa70d0972656501d45d
```

---

## Step 4: Create a Database User

Create a user that will connect to the database:

```bash
acloud database dbaas user create 69455aa70d0972656501d45d \
  --username "app-user" \
  --password "SecurePassword123!"
```

Example output:
```
User created successfully!
DBaaS ID:   69455aa70d0972656501d45d
Username:   app-user
```

> **Security best practice**: Use a strong password (minimum 12 characters, mix of uppercase, lowercase, numbers, and symbols). Never commit passwords to version control.

Verify the user was created:

```bash
acloud database dbaas user list 69455aa70d0972656501d45d
```

---

## Step 5: Grant User Access to the Database

Assign the user access to the database with the `liteadmin` role:

```bash
acloud database dbaas grant create 69455aa70d0972656501d45d "appdb" \
  --username "app-user" \
  --role "liteadmin"
```

Example output:
```
Grant created successfully!
DBaaS ID:      69455aa70d0972656501d45d
Database:      appdb
Username:      app-user
Role:          liteadmin
```

Verify the grant:

```bash
acloud database dbaas grant list 69455aa70d0972656501d45d "appdb"
```

---

## Step 6: Retrieve Connection Details

Get the full DBaaS instance details to find the connection endpoint:

```bash
acloud database dbaas get 69455aa70d0972656501d45d
```

Note the public IP address from the Elastic IP assigned to your DBaaS instance.

---

## Step 7: Connect to the Database

Use any standard MySQL client to connect:

```bash
mysql -h <elastic-ip-address> -P 3306 -u app-user -p appdb
# Enter password: (hidden input)
```

Or with a connection string:

```bash
mysql --host=<elastic-ip-address> --port=3306 --user=app-user --password --database=appdb
```

> **Note**: Ensure your local machine's IP address is allowed by the security group inbound rule on port 3306.

---

## Step 8: Verify the Setup

Once connected, verify the database and user are correctly configured:

```sql
-- Show available databases
SHOW DATABASES;

-- Select the database
USE appdb;

-- Show current user
SELECT USER();

-- Verify permissions
SHOW GRANTS FOR 'app-user'@'%';
```

---

## Best Practices

- **Backups**: Schedule regular backups using `acloud database dbaas backup create`
- **Security groups**: Restrict port 3306 to specific IP ranges instead of `0.0.0.0/0`
- **User management**: Create separate users for each application; avoid using the admin user directly
- **Password rotation**: Rotate passwords regularly by deleting and re-creating users
- **Tagging**: Use tags to track environment and cost allocation (e.g., `production,billing-team-a`)
- **Wait flag**: Use `--wait --timeout 15m` when creating DBaaS instances to catch provisioning failures early

---

## Step 9: Cleanup

To delete the DBaaS instance and all its resources:

```bash
# Delete the DBaaS instance (users, databases, and grants are removed automatically)
acloud database dbaas delete 69455aa70d0972656501d45d --yes
```

> **Warning**: Deleting a DBaaS instance permanently deletes all databases, users, grants, and data within it. Ensure you have backups before proceeding.

## Related Resources

- [DBaaS](../resources/database/dbaas.md) - Full DBaaS command reference
- [DBaaS Databases](../resources/database/dbaas.database.md) - Database management commands
- [DBaaS Users](../resources/database/dbaas.user.md) - User management commands
- [DBaaS Grants](../resources/database/dbaas.grant.md) - Grant management commands
- [Database Backups](../resources/database/backup.md) - Backup and restore commands
