# Snapshot Management

Snapshots are point-in-time copies of block storage volumes. They capture the entire state of a volume at a specific moment.

## Available Commands

- `acloud storage snapshot create` - Create a new snapshot from a volume
- `acloud storage snapshot list` - List snapshots for a volume
- `acloud storage snapshot get` - Get details of a specific snapshot
- `acloud storage snapshot update` - Update snapshot name and tags
- `acloud storage snapshot delete` - Delete a snapshot

## Create Snapshot

Create a snapshot from an existing block storage volume.

### Usage

```bash
acloud storage snapshot create --name <name> --region <region> --volume-uri <volume-uri> [flags]
```

### Required Flags

- `--name` - Name for the snapshot
- `--region` - Region code
- `--volume-uri` - Source volume URI

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--tags` - Tags (comma-separated)

### Example

```bash
acloud storage snapshot create \
  --name "backup-before-upgrade" \
  --region "ITBG-Bergamo" \
  --volume-uri "/projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d" \
  --tags "backup,pre-upgrade"
```

### Output

```
Creating snapshot with the following parameters:
  Name:       backup-before-upgrade
  Region:     ITBG-Bergamo
  Volume URI: /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
  Tags:       [backup pre-upgrade]

Snapshot created successfully!
ID:              6944fd760d0972656501d431
Name:            backup-before-upgrade
Creation Date:   18-12-2025 17:23:50
```

## List Snapshots

List all snapshots for a specific block storage volume.

### Usage

```bash
acloud storage snapshot list --volume-uri <volume-uri> [flags]
```

### Required Flags

- `--volume-uri` - Block storage volume URI

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `-v, --verbose` - Show detailed debug information

### Example

```bash
acloud storage snapshot list --volume-uri "/projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d"
```

### Output

```
NAME                      ID                          SIZE(GB)  STATUS
backup-before-upgrade     6944fd760d0972656501d431   50        Available
daily-backup-20251218     6944fe870d0972656501d432   50        Available
```

## Get Snapshot Details

Get detailed information about a specific snapshot.

### Usage

```bash
acloud storage snapshot get <snapshot-id> [flags]
```

### Arguments

- `snapshot-id` - The ID of the snapshot (supports auto-completion)

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud storage snapshot get 6944fd760d0972656501d431
```

### Output

```
Snapshot Details:
=================
ID:              6944fd760d0972656501d431
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/snapshots/6944fd760d0972656501d431
Name:            backup-before-upgrade
Size (GB):       50
Source Volume:   /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
Region:          ITBG-Bergamo
Status:          Available
Creation Date:   18-12-2025 17:23:50
Created By:      aru-297647
Tags:            [backup pre-upgrade]
```

## Update Snapshot

Update the name and/or tags of a snapshot.

### Usage

```bash
acloud storage snapshot update <snapshot-id> [flags]
```

### Arguments

- `snapshot-id` - The ID of the snapshot (supports auto-completion)

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--name` - New name for the snapshot
- `--tags` - New tags (comma-separated)

**Note:** At least one of `--name` or `--tags` must be provided.

### Example

```bash
# Update name only
acloud storage snapshot update 6944fd760d0972656501d431 --name "pre-upgrade-snapshot"

# Update tags only
acloud storage snapshot update 6944fd760d0972656501d431 --tags "backup,important"

# Update both
acloud storage snapshot update 6944fd760d0972656501d431 \
  --name "pre-upgrade-snapshot" \
  --tags "backup,important"
```

### Output

```
Snapshot updated successfully!
ID:              6944fd760d0972656501d431
Name:            pre-upgrade-snapshot
Tags:            [backup important]
```

## Delete Snapshot

Delete a snapshot. This action cannot be undone.

### Usage

```bash
acloud storage snapshot delete <snapshot-id> [flags]
```

### Arguments

- `snapshot-id` - The ID of the snapshot (supports auto-completion)

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `-y, --yes` - Skip confirmation prompt

### Example

```bash
# With confirmation prompt
acloud storage snapshot delete 6944fd760d0972656501d431

# Skip confirmation
acloud storage snapshot delete 6944fd760d0972656501d431 --yes
```

### Output

```
Snapshot 6944fd760d0972656501d431 deleted successfully!
```

## Notes

- Snapshots are point-in-time copies and do not include changes made after creation
- Snapshots are stored in the same region as the source volume
- You can create multiple snapshots from the same volume
- Snapshots can be used to restore volumes to a previous state
- The volume URI can be found using `acloud storage blockstorage get <volume-id>`
