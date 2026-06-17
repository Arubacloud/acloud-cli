# Restore Management

Restore operations use the Aruba.Storage/restore resource to restore block storage volumes from backups. Unlike snapshot restores which create new volumes, backup restores write data back to existing volumes.

## Available Commands

- `acloud storage restore` - Create a restore operation from a backup
- `acloud storage restore list` - List restore operations for a backup
- `acloud storage restore get` - Get details of a specific restore operation
- `acloud storage restore update` - Update restore operation name and tags
- `acloud storage restore delete` - Delete a restore operation

## Restore vs Snapshot Restore

| Feature | Backup Restore | Snapshot Restore |
|---------|---------------|------------------|
| API Resource | Aruba.Storage/restore | Creates new volume |
| Target | Existing volume | New volume |
| Hierarchical | Nested under backup | Standalone |
| Use Case | Restore to same volume | Clone to new volume |

## Create Restore Operation

Create a restore operation to restore data from a backup to an existing volume.

**Important:** The restore operation writes data TO an existing volume. The target volume must exist and will be overwritten.

### Usage

```bash
acloud storage restore <backup-id> <volume-id> --name <name> [flags]
```

### Arguments

- `backup-id` - The ID of the backup to restore from
- `volume-id` - The ID of the target volume (will be overwritten)

### Required Flags

- `--name` - Name for the restore operation
- `--region` - Region code

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--tags` - Tags (comma-separated)

### Example

```bash
# Restore backup to the same volume it was created from
acloud storage restore 694594818f4a09c12b5e0c19 69455aa70d0972656501d45d \
  --name "restore-after-failure" \
  --region "ITBG-Bergamo" \
  --tags "recovery,production"

# Restore to a different volume
acloud storage restore 694594818f4a09c12b5e0c19 69455bb80d0972656501d45e \
  --name "clone-to-test" \
  --region "ITBG-Bergamo"
```

### Output

```
Creating restore operation with the following parameters:
  Name:       restore-after-failure
  Region:     ITBG-Bergamo
  Backup ID:  694594818f4a09c12b5e0c19
  Backup URI: /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/backups/694594818f4a09c12b5e0c19
  Volume ID:  69455aa70d0972656501d45d
  Volume URI: /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
  Tags:       [recovery production]

Restore operation created successfully!
ID:              6945953a8f4a09c12b5e0c71
Name:            restore-after-failure
Creation Date:   19-12-2025 18:11:06
Status:          InCreation
```

## List Restore Operations

List all restore operations for a specific backup.

**Note:** Restore operations are hierarchical and nested under backups, so you must specify the backup ID.

### Usage

```bash
acloud storage restore list <backup-id> [flags]
```

### Arguments

- `backup-id` - The ID of the backup (supports auto-completion)

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud storage restore list 694594818f4a09c12b5e0c19
```

### Output

```
NAME                    ID                          STATUS
restore-after-failure   6945953a8f4a09c12b5e0c71   Active
test-restore            6945954b8f4a09c12b5e0c72   Active
```

## Get Restore Details

Get detailed information about a specific restore operation.

### Usage

```bash
acloud storage restore get <backup-id> <restore-id> [flags]
```

### Arguments

- `backup-id` - The ID of the backup (supports auto-completion)
- `restore-id` - The ID of the restore operation (auto-completes based on selected backup)

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Auto-completion

Restore commands support hierarchical auto-completion:

1. First, press TAB to see available backup IDs
2. After entering a backup ID, press TAB again to see restore IDs for that backup

```bash
acloud storage restore get <TAB>
# Shows backup IDs:
# 67649dac8c7bb1c5d7c80631    MyBackup
# ...

acloud storage restore get 67649dac8c7bb1c5d7c80631 <TAB>
# Shows restore IDs for that backup:
# 67664dde0aca19a92c2c48bb    RestoreOperation1
# ...
```

### Example

```bash
acloud storage restore get 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71
```

### Output

```
Restore Operation Details:
==========================
ID:              6945953a8f4a09c12b5e0c71
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/backups/694594818f4a09c12b5e0c19/restores/6945953a8f4a09c12b5e0c71
Name:            restore-after-failure
Target Volume:   /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
Region:          ITBG-Bergamo
Status:          Active
Creation Date:   19-12-2025 18:11:06
Created By:      aru-297647
Tags:            [recovery production]
```

## Update Restore Operation

Update the name and/or tags of a restore operation.

**Important:** Restore operations can only be updated when in "Active" status (not "InCreation").

### Usage

```bash
acloud storage restore update <backup-id> <restore-id> [flags]
```

### Arguments

- `backup-id` - The ID of the backup (supports auto-completion)
- `restore-id` - The ID of the restore operation (auto-completes based on selected backup)

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--name` - New name for the restore operation
- `--tags` - New tags (comma-separated)

**Note:** At least one of `--name` or `--tags` must be provided.

### Example

```bash
# Update name only
acloud storage restore update 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71 \
  --name "restore-renamed"

# Update tags only
acloud storage restore update 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71 \
  --tags "prod,final"

# Update both
acloud storage restore update 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71 \
  --name "production-restore" \
  --tags "prod,final"
```

### Output

```
Restore operation updated successfully!
ID:              6945953a8f4a09c12b5e0c71
Name:            production-restore
Tags:            [prod final]
```

## Delete Restore Operation

Delete a restore operation record. This does not undo the restore; it only deletes the operation metadata.

### Usage

```bash
acloud storage restore delete <backup-id> <restore-id> [flags]
```

### Arguments

- `backup-id` - The ID of the backup (supports auto-completion)
- `restore-id` - The ID of the restore operation (auto-completes based on selected backup)

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `-y, --yes` - Skip confirmation prompt

### Example

```bash
# With confirmation prompt
acloud storage restore delete 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71

# Skip confirmation
acloud storage restore delete 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71 --yes
```

### Output

```
Restore operation 6945953a8f4a09c12b5e0c71 deleted successfully!
```

## Restore Workflow

### 1. Identify Backup
```bash
# List available backups
acloud storage backup list

# Get backup details to verify it's the right one
acloud storage backup get <backup-id>
```

### 2. Prepare Target Volume
```bash
# Ensure target volume exists and is ready
acloud storage blockstorage get <volume-id>

# IMPORTANT: Detach volume from VM if attached
# The restore will overwrite all data on the volume
```

### 3. Create Restore Operation
```bash
acloud storage restore <backup-id> <volume-id> \
  --name "restore-$(date +%Y%m%d)" \
  --region "ITBG-Bergamo"
```

### 4. Monitor Restore
```bash
# Check restore status
acloud storage restore get <backup-id> <restore-id>

# Wait until status is "Active"
```

### 5. Verify and Cleanup
```bash
# After successful restore, delete the operation record
acloud storage restore delete <backup-id> <restore-id> --yes
```

## Restore Status

- **InCreation**: Restore operation is in progress
- **Active**: Restore completed successfully
- **Failed**: Restore encountered an error

## Important Considerations

### Data Overwrite
- Restore operations **overwrite** the target volume completely
- All existing data on the target volume will be lost
- Always verify you're restoring to the correct volume
- Consider creating a snapshot of the target volume before restore

### Volume Requirements
- Target volume must exist before creating restore
- Target volume should be detached from VMs
- Target volume must be in the same region as the backup
- Target volume size should match or exceed backup size

### Backup Availability
- Backup must be in "Active" status
- Cannot update backup while restore is in progress
- Multiple restores can be created from the same backup

### Hierarchical Structure
- Restore operations are nested under backups
- All restore commands require backup-id parameter
- Use `restore list <backup-id>` to see all restores for a backup

## Best Practices

1. **Test in non-production first**: Always test restore procedures in a test environment
2. **Verify backup before restore**: Check backup details and status before restoring
3. **Detach volumes**: Ensure target volumes are detached from VMs
4. **Create pre-restore snapshot**: Snapshot the target volume before restore as a safety measure
5. **Monitor restore status**: Check restore status until it's "Active"
6. **Validate after restore**: Verify data integrity after restore completes
7. **Clean up records**: Delete restore operation records after verification
8. **Document procedures**: Keep runbooks for emergency restore scenarios

## Troubleshooting

### Restore Fails to Create
- Verify backup is in "Active" status
- Check target volume exists and is accessible
- Ensure region matches between backup and volume
- Verify no other restore is running on the same volume

### Cannot Update Backup
- Error: "Backup can't be deleted or modified because there is a running restore operation"
- Solution: Wait for restore to complete or delete the restore operation

### Restore Stuck in "InCreation"
- Wait a few minutes as restore operations can take time
- Check backup and volume status
- Contact support if status doesn't change after extended period

## Notes

- Restore operations are asynchronous and start in "InCreation" status
- Restores write TO existing volumes (different from snapshot restore)
- You can track restore history by keeping operation records
- Restore operations use the Aruba.Storage/restore API resource
- Tags can help organize and track restore operations by purpose or date
