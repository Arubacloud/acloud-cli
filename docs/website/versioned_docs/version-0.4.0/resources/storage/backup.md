# Backup Management

Backups are managed copies of block storage volumes using the Aruba.Storage/backup resource. Unlike snapshots, backups provide advanced features like incremental backups, retention policies, and scheduled operations.

## Available Commands

- `acloud storage backup` - Create a new backup from a volume
- `acloud storage backup list` - List all backups
- `acloud storage backup get` - Get details of a specific backup
- `acloud storage backup update` - Update backup name and tags
- `acloud storage backup delete` - Delete a backup

## Backup vs Snapshot

| Feature | Backup | Snapshot |
|---------|--------|----------|
| API Resource | Aruba.Storage/backup | Aruba.Storage/snapshot |
| Backup Type | Full or Incremental | Full only |
| Retention Policy | Configurable days | Manual management |
| Billing Period | Hour, Month, Year | Default billing |
| Use Case | Long-term data protection | Quick point-in-time copies |

## Create Backup

Create a backup from a block storage volume with advanced options.

### Usage

```bash
acloud storage backup <volume-id> --name <name> [flags]
```

### Arguments

- `volume-id` - The ID of the block storage volume to backup

### Required Flags

- `--name` - Name for the backup

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--region` - Region code (default: "ITBG-Bergamo")
- `--type` - Backup type: Full or Incremental (default: "Full")
- `--retention-days` - Number of days to retain the backup
- `--billing-period` - Billing period: Hour, Month, Year
- `--tags` - Tags (comma-separated, max 20 chars per tag)

### Example

```bash
# Create a full backup with 7-day retention
acloud storage backup 69455aa70d0972656501d45d \
  --name "weekly-backup" \
  --region "ITBG-Bergamo" \
  --type "Full" \
  --retention-days 7 \
  --tags "weekly,production"

# Create an incremental backup
acloud storage backup 69455aa70d0972656501d45d \
  --name "daily-incremental" \
  --type "Incremental" \
  --retention-days 1
```

### Output

```
Creating storage backup with the following parameters:
  Name:           weekly-backup
  Type:           Full
  Region:         ITBG-Bergamo
  Volume ID:      69455aa70d0972656501d45d
  Volume URI:     /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
  Retention Days: 7
  Tags:           [weekly production]

Storage backup created successfully!
ID:              694594818f4a09c12b5e0c19
Name:            weekly-backup
Type:            Full
Creation Date:   19-12-2025 18:08:01
```

## List Backups

List all backups in your project.

### Usage

```bash
acloud storage backup list [flags]
```

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud storage backup list
```

### Output

```
NAME              ID                          TYPE          STATUS
weekly-backup     694594818f4a09c12b5e0c19   Full          Active
daily-backup      694595918f4a09c12b5e0c20   Incremental   Active
monthly-archive   694596a18f4a09c12b5e0c21   Full          Active
```

## Get Backup Details

Get detailed information about a specific backup.

### Usage

```bash
acloud storage backup get <backup-id> [flags]
```

### Arguments

- `backup-id` - The ID of the backup (supports auto-completion)

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud storage backup get 694594818f4a09c12b5e0c19
```

### Output

```
Storage Backup Details:
=======================
ID:              694594818f4a09c12b5e0c19
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/backups/694594818f4a09c12b5e0c19
Name:            weekly-backup
Type:            Full
Source Volume:   /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
Retention Days:  7
Region:          ITBG-Bergamo
Status:          Active
Creation Date:   19-12-2025 18:08:01
Created By:      aru-297647
Tags:            [weekly production]
```

## Update Backup

Update the name and/or tags of a backup.

**Important:** Backup cannot be updated while a restore operation is running on the associated volume. The backup must be in "Active" status (not "InCreation" or "Deleting").

### Usage

```bash
acloud storage backup update <backup-id> [flags]
```

### Arguments

- `backup-id` - The ID of the backup (supports auto-completion)

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--name` - New name for the backup
- `--tags` - New tags (comma-separated)

**Note:** At least one of `--name` or `--tags` must be provided.

### Example

```bash
# Update name only
acloud storage backup update 694594818f4a09c12b5e0c19 --name "backup-renamed"

# Update tags only
acloud storage backup update 694594818f4a09c12b5e0c19 --tags "prod,critical"

# Update both
acloud storage backup update 694594818f4a09c12b5e0c19 \
  --name "production-backup" \
  --tags "prod,critical"
```

### Output

```
Backup updated successfully!
ID:              694594818f4a09c12b5e0c19
Name:            production-backup
Tags:            [prod critical]
```

## Delete Backup

Delete a backup. This action cannot be undone.

**Important:** Cannot delete a backup if there are active restore operations.

### Usage

```bash
acloud storage backup delete <backup-id> [flags]
```

### Arguments

- `backup-id` - The ID of the backup (supports auto-completion)

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `-y, --yes` - Skip confirmation prompt

### Example

```bash
# With confirmation prompt
acloud storage backup delete 694594818f4a09c12b5e0c19

# Skip confirmation
acloud storage backup delete 694594818f4a09c12b5e0c19 --yes
```

### Output

```
Backup 694594818f4a09c12b5e0c19 deleted successfully!
```

## Backup Types

### Full Backup
- Complete copy of the volume
- Can be used independently for restore
- Takes more time and storage space
- Recommended for: Weekly/monthly archives, critical data protection

### Incremental Backup
- Only stores changes since last backup
- Faster and uses less storage
- Requires previous backups for restore
- Recommended for: Daily backups, frequent snapshots

## Retention Policy

- Backups can have a retention period in days
- After retention period expires, backups may be automatically deleted
- Set retention based on your compliance and recovery requirements
- Common retention periods: 7 days (weekly), 30 days (monthly), 365 days (yearly)

## Best Practices

1. **Use Full backups for baseline**: Create weekly full backups as recovery points
2. **Incremental for daily**: Use incremental backups for daily protection
3. **Tag appropriately**: Use tags to identify backup purpose and schedule
4. **Monitor retention**: Track backup ages and adjust retention as needed
5. **Test restores**: Periodically verify backups can be restored successfully
6. **Delete old backups**: Clean up backups after retention period

## Notes

- Backups are created asynchronously and start in "InCreation" status
- Once "Active", backups can be used for restore operations
- Tags must be maximum 20 characters each
- The volume ID can be found using `acloud storage blockstorage list`
- Backups use the Aruba.Storage/backup API resource (different from snapshots)
