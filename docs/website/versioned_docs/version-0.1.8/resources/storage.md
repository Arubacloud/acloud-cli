# Storage Resources



The `storage` category provides commands for managing storage resources in Aruba Cloud.



## Available Resources



### [Block Storage](storage/blockstorage.md)



Block storage volumes are persistent storage devices that can be attached to virtual machines.



**Quick Commands:**

```bash
# List all block storage volumes
acloud storage blockstorage list

# Get volume details
acloud storage blockstorage get <volume-id>

# Create a volume
acloud storage blockstorage create --name "my-volume" --size 50

# Update a volume
acloud storage blockstorage update <volume-id> --name "new-name"

# Delete a volume
acloud storage blockstorage delete <volume-id>
```



### [Snapshots](storage/snapshot.md)



Snapshots are point-in-time copies of block storage volumes for quick backups and cloning.



**Quick Commands:**

```bash
# List snapshots for a volume
acloud storage snapshot list --volume-uri <volume-uri>

# Get snapshot details
acloud storage snapshot get <snapshot-id>

# Create a snapshot
acloud storage snapshot create --name "backup" --region "ITBG-Bergamo" --volume-uri <uri>

# Update a snapshot
acloud storage snapshot update <snapshot-id> --tags "important"

# Delete a snapshot
acloud storage snapshot delete <snapshot-id>
```



### [Backups](storage/backup.md)



Backups provide advanced data protection with full/incremental backup types and retention policies.



**Quick Commands:**

```bash
# List all backups
acloud storage backup list

# Get backup details
acloud storage backup get <backup-id>

# Create a backup
acloud storage backup <volume-id> --name "weekly-backup" --type "Full" --retention-days 7

# Update a backup
acloud storage backup update <backup-id> --tags "production"

# Delete a backup
acloud storage backup delete <backup-id>
```



### [Restore Operations](storage/restore.md)



Restore operations allow you to restore block storage volumes from backups.



**Quick Commands:**

```bash
# List restore operations for a backup
acloud storage restore list <backup-id>

# Get restore details
acloud storage restore get <backup-id> <restore-id>

# Create a restore operation
acloud storage restore <backup-id> <volume-id> --name "restore-op" --region "ITBG-Bergamo"

# Update a restore operation
acloud storage restore update <backup-id> <restore-id> --name "new-name"

# Delete a restore operation
acloud storage restore delete <backup-id> <restore-id>
```



## Command Structure



All storage commands follow this structure:



```

acloud storage <resource> <action> [arguments] [flags]

```



Where:

- `<resource>`: The type of resource (e.g., `blockstorage`, `snapshot`, `backup`, `restore`)

- `<action>`: The operation to perform (e.g., `list`, `get`, `create`, `update`, `delete`)

- `[arguments]`: Required arguments (e.g., resource IDs)

- `[flags]`: Optional flags (e.g., `--name`, `--size`, `--type`)



## Common Patterns



### Listing Resources



```bash
acloud storage <resource> list [arguments]
```



Lists all resources of the specified type with key information displayed in a table format.



### Getting Resource Details



```bash
acloud storage <resource> get <resource-id> [arguments]
```



Displays detailed information about a specific resource.



### Creating Resources



```bash
acloud storage <resource> <arguments> --flag1 value1 --flag2 value2
```



Creates a new resource with the specified properties.



### Updating Resources



```bash
acloud storage <resource> update <resource-id> [arguments] --flag1 value1
```



Updates an existing resource. Only provided fields are modified.



### Deleting Resources



```bash
acloud storage <resource> delete <resource-id> [arguments]
```



Deletes the specified resource. May prompt for confirmation.



## Storage Architecture



Storage resources are organized hierarchically:



```

Project

├── Block Storage Volumes

│   ├── Snapshots (point-in-time copies)

│   └── Backups (with retention policies)

│       └── Restore Operations (nested under backups)

```



## Resource Relationships



- **Block Storage** → **Snapshots**: One-to-many (a volume can have multiple snapshots)

- **Block Storage** → **Backups**: One-to-many (a volume can have multiple backups)

- **Backups** → **Restore Operations**: One-to-many (a backup can have multiple restore operations)



## Next Steps



- [Block Storage Management Guide](storage/blockstorage.md)

- [Snapshot Management Guide](storage/snapshot.md)

- [Backup Management Guide](storage/backup.md)

- [Restore Operations Guide](storage/restore.md)


*Example:*
``bash
acloud storage blockstorage get 6942fe38f4a09c12b5ded74
``

*Output:*
``
Block Storage Details:
===========
ID:       6942fe38f4a09c12b5ded74
URI:       /projects/68398923fb2cb02640d4d31/providers/Aruba.Storage/blockStorages/6942fe38f4a09c12b5ded74
Name:      my-volume
Size (GB):    10
Type:      Standard
Zone:      ITBG-3
Region:     ITBG-Bergamo
Status:     NotUsed
Creation Date:  18-12-2025 10:30:0
Created By:   aru-123456
Tags:      [environment=production owner=devops]
``

*Note:* The URI field shows the ful resource path neded for creating snapshots.

## Create Block Storage

``bash
acloud storage blockstorage create \
 -name "my-volume" \
 -size 10
``

*With zone:*
``bash
acloud storage blockstorage create \
 -name "my-volume" \
 -zone "ITBG-3" \
 -size 10
``

*With al options:*
``bash
acloud storage blockstorage create \
 -name "my-volume" \
 -region "ITBG-Bergamo" \
 -zone "ITBG-3" \
 -size 10 \
 -type "Standard" \
 -biling-period "Hour" \
 -tags "environment=production,owner=devops"
``

*Required Flags:*
- `-name`: Name for the block storage volume
- `-size`: Size in GB

*Optional Flags:*
- `-zone`: Zone/datacenter (e.g., `ITBG-3`) - Optional, can be omited
- `-region`: Region code (default: `ITBG-Bergamo`)
- `-type`: Volume type (`Standard` or `Performance`, default: `Standard`)
- `-biling-period`: Biling period (`Hour`, `Month`, `Year`, default: `Hour`)
- `-tags`: Coma-separated tags
- `-project-id`: Project ID (uses context if not specified)

*Example Output:*
``
Creating block storage with:
 Name: my-volume
 Region: ITBG-Bergamo
 Zone: ITBG-3
 Size: 10 GB
 Type: Standard
 Biling Period: Hour
 Project ID: 68398923fb2cb02640d4d31

Block storage created sucesfuly!
ID:       vol-987654321
Name:      my-volume
Size (GB):    10
Type:      Standard
Zone:      ITBG-3
Region:     ITBG-Bergamo
Status:     InCreation
Creation Date:  18-12-2025 16:46:27
``

*Comon Region Codes:*
- `ITBG-Bergamo` - Italy, Bergamo (default)
- `ITMIL-Milano` - Italy, Milan
- `CZPRG-Prague` - Czech Republic, Prague

## Update Block Storage

*Important:* Block storage can only be updated when its status is `Used` or `NotUsed`. Updates are not alowed during provisioning states like `InCreation`.

You can update the name and/or tags of a block storage volume:

``bash
# Update name
acloud storage blockstorage update 6942fe38f4a09c12b5ded74 -name "new-volume-name"

# Update tags
acloud storage blockstorage update 6942fe38f4a09c12b5ded74 -tags "env=prod,team=backend"

# Update both
acloud storage blockstorage update 6942fe38f4a09c12b5ded74 \
 -name "new-name" \
 -tags "env=prod,team=backend"
``

*Example Output:*
``bash
$ acloud storage blockstorage update 69457430d09726501d43c -name "my-updated-volume"

Block storage updated sucesfuly!
ID:       69457430d09726501d43c
Name:      my-updated-volume
Size (GB):    5
Type:      Standard
``

*Status Validation:*
``bash
$ acloud storage blockstorage update 6945a70d09726501d45d -name "should-fail"
Eror: Canot update block storage with status 'InCreation'
Block storage can only be updated when status is 'Used' or 'NotUsed'
``

*Note:* Size and type updates are not suported by the API at this time. Only name and tags can be modified.

## Delete Block Storage

``bash
# With confirmation prompt
acloud storage blockstorage delete 6942fe38f4a09c12b5ded74

# Skip confirmation
acloud storage blockstorage delete 6942fe38f4a09c12b5ded74 -yes
``

*Example with confirmation:*
``bash
$ acloud storage blockstorage delete 6942fe38f4a09c12b5ded74
Are you sure you want to delete block storage 6942fe38f4a09c12b5ded74? (yes/no): yes

Block storage 6942fe38f4a09c12b5ded74 deleted sucesfuly!
``

*Flags:*
- `-y, -yes` - Skip confirmation prompt (useful for scripts)

*Warning:* Deleting a block storage volume is permanent and canot be undone.

# Snapshots

Snapshots provide point-in-time copies of block storage volumes for backup and recovery purposes.

## List Snapshots

List snapshots for a specific block storage volume:

``bash
# List snapshots for a volume (requires volume URI)
acloud storage snapshot list -volume-uri "/projects/68398923fb2cb02640d4d31/providers/Aruba.Storage/blockStorages/6942fe38f4a09c12b5ded74"

# With explicit project ID
acloud storage snapshot list \
 -volume-uri "/projects/68398923fb2cb02640d4d31/providers/Aruba.Storage/blockStorages/6942fe38f4a09c12b5ded74" \
 -project-id "6a1024f62b9c686572a9f"
``

*Output:*
``
NAME              ID             SIZE(GB)   STATUS     
backup-snapshot-01       6942abc8f4a09c12b5def12  10     Active    
daily-backup          6942def8f4a09c12b5def34  250     Active    
updated-test-snapshot     6945d620d09726501d47  824634892352 Active
``

*Required Flags:*
- `-volume-uri` - Ful URI of the source block storage volume (get this from `blockstorage get` comand)

*Optional Flags:*
- `-project-id` - Project ID (uses context if not specified)

*Note:* The `-volume-uri` flag is required to filter snapshots for a specific volume. Use `blockstorage get <volume-id>` to get the volume URI.

## Get Snapshot Details

``bash
acloud storage snapshot get <snapshot-id>
``

*Example:*
``bash
acloud storage snapshot get 6942abc8f4a09c12b5def12
``

*Output:*
``
Snapshot Details:
=========
ID:       6942abc8f4a09c12b5def12
URI:       /projects/68398923fb2cb02640d4d31/providers/Aruba.Storage/snapshots/6942abc8f4a09c12b5def12
Name:      daily-backup
Size (GB):    10
Source Volume:  /projects/68398923fb2cb02640d4d31/providers/Aruba.Storage/blockStorages/6942fe38f4a09c12b5ded74
Region:     ITBG-Bergamo
Status:     Active
Creation Date:  19-12-2025 14:12:50
Created By:   aru-297647
``

*Note:* The Region field shows the region value (e.g., `ITBG-Bergamo`) as returned by the API.
## Create Snapshot

To create a snapshot, you ned the ful URI of the source block storage volume. You can get this from the `blockstorage get` comand.

``bash
# First, get the volume URI
acloud storage blockstorage get 6942fe38f4a09c12b5ded74

# Then create the snapshot using the URI
acloud storage snapshot create \
 -name "my-snapshot" \
 -region "ITBG-Bergamo" \
 -volume-uri "/projects/68398923fb2cb02640d4d31/providers/Aruba.Storage/blockStorages/6942fe38f4a09c12b5ded74"
``

*With tags:*
``bash
acloud storage snapshot create \
 -name "daily-backup" \
 -region "ITBG-Bergamo" \
 -volume-uri "/projects/68398923fb2cb02640d4d31/providers/Aruba.Storage/blockStorages/6942fe38f4a09c12b5ded74" \
 -tags "type=backup,schedule=daily"
``

*Required Flags:*
- `-name`: Name for the snapshot
- `-region`: Region code (e.g., `ITBG-Bergamo`)
- `-volume-uri`: Ful URI of the source volume (get this from `blockstorage get` comand)

*Optional Flags:*
- `-tags`: Coma-separated tags
- `-project-id`: Project ID (uses context if not specified)

*Example Output:*
``
Creating snapshot with the folowing parameters:
 Name:    daily-backup
 Region:   ITBG-Bergamo
 Volume URI: /projects/68398923fb2cb02640d4d31/providers/Aruba.Storage/blockStorages/6942fe38f4a09c12b5ded74
 Tags:    [type=backup schedule=daily]

Snapshot created sucesfuly!
ID:       6942abc8f4a09c12b5def12
Name:      daily-backup
Creation Date:  19-12-2025 14:12:50
``

*Important Notes:*
- The snapshot wil initialy be in `InCreation` status and wil transition to `Active` when ready
- SDK v0.1.2+ properly handles block storage state validation during snapshot creation
- Creation may take several minutes depending on volume size
Name:      daily-backup
Size (GB):    10
Creation Date:  19-12-2025 14:30:0
``

## Update Snapshot

You can update the name and/or tags of a snapshot:

``bash
# Update name
acloud storage snapshot update 6942abc8f4a09c12b5def12 -name "new-snapshot-name"

# Update tags
acloud storage snapshot update 6942abc8f4a09c12b5def12 -tags "type=backup,retention=30days"

# Update both
acloud storage snapshot update 6942abc8f4a09c12b5def12 \
 -name "new-name" \
 -tags "type=backup,retention=30days"
``

*Example:*
``bash
acloud storage snapshot update 6945d620d09726501d47 -name "updated-test-snapshot"
``

*Output:*
``
Snapshot updated sucesfuly!
``

*Important Notes:*
- Only name and tags can be updated
- The CLI uses the region value directly from the API (e.g., `ITBG-Bergamo`)
- No region conversion is neded - the region value is used as-is during updates

## Delete Snapshot

``bash
# With confirmation prompt
acloud storage snapshot delete 6942abc8f4a09c12b5def12

# Skip confirmation (useful for automation)
acloud storage snapshot delete 6942abc8f4a09c12b5def12 -yes
``

*Example with confirmation:*
``bash
$ acloud storage snapshot delete 6945d620d09726501d47
Are you sure you want to delete snapshot 6945d620d09726501d47? (yes/no): yes

Snapshot 6945d620d09726501d47 deleted sucesfuly!
``

*Flags:*
- `-y, -yes` - Skip confirmation prompt (useful for scripts and automation)

*Warning:* Deleting a snapshot is permanent and canot be undone. Always verify the snapshot ID before deletion.

# Best Practices

## Block Storage

1. *Naming Convention*: Use descriptive names that indicate purpose and environment
  ``bash
  -name "prod-db-data-volume"
  -name "dev-ap-logs"
  ``

2. *Taging Strategy*: Use consistent tags for organization and cost tracking
  ``bash
  -tags "environment=production,aplication=database,cost-center=enginering"
  ``

3. *Size Planing*: Chose apropriate size based on growth projections
  - Start with minimum required size
  - Monitor usage regularly
  - Note: Size canot be increased after creation via update comand
  - Consider costs vs. performance

4. *Type Selection*:
  - *Standard*: General-purpose workloads, cost-efective
  - *Performance*: High I/O workloads, databases, critical aplications

5. *Zone Selection*:
  - Zone parameter is *optional* when creating block storage
  - Can be omited for region-wide deployment
  - Specify zone (e.g., `ITBG-3`) for specific datacenter placement

6. *Regional Placement*: Create volumes in the same region as your compute resources
  - Default region: `ITBG-Bergamo`
  - Use proper region format: `ITBG-Bergamo`

7. *Biling Period*: Chose based on usage duration
  - *Hour*: Short-term testing, development (default)
  - *Month*: Production workloads
  - *Year*: Long-term stable infrastructure (best cost savings)

8. *Update Constraints*: Block storage can only be updated when status is `Used` or `NotUsed`
  - Canot update during provisioning (`InCreation` state)
  - Only name and tags can be modified
  - Size and type changes not suported

## Snapshots

1. *Geting Volume URI*: Always use `blockstorage get` to get the ful URI for snapshot operations
  ``bash
  # Get the URI from the block storage details
  acloud storage blockstorage get 6942fe38f4a09c12b5ded74
  # Copy the URI field for snapshot creation and listing
  ``

2. *Listing Snapshots*: The volume URI is required to list snapshots
  ``bash
  acloud storage snapshot list -volume-uri "/projects/PROJECT_ID/providers/Aruba.Storage/blockStorages/VOLUME_ID"
  ``

3. *Regular Backups*: Create snapshots on a regular schedule
  ``bash
  -tags "schedule=daily,retention=7days"
  -tags "schedule=wekly,retention=4weks"
  ``

4. *Pre-Change Snapshots*: Always create a snapshot before major changes
  ``bash
  acloud storage snapshot create \
   -name "pre-upgrade-$(date +%Y%m%d)" \
   -region "ITBG-Bergamo" \
   -volume-uri "/projects/PROJECT_ID/providers/Aruba.Storage/blockStorages/VOLUME_ID"
  ``

5. *Naming Convention*: Include date and purpose in snapshot names
  ``bash
  -name "prod-db-daily-20251218"
  -name "pre-migration-backup-20251218"
  ``

6. *Retention Policy*: Implement a clear retention policy
  - Daily snapshots: 7 days
  - Wekly snapshots: 4 weks
  - Monthly snapshots: 12 months

7. *Cleanup*: Delete old snapshots to manage costs
  ``bash
  # List snapshots for a volume
  acloud storage snapshot list -volume-uri "/projects/PROJECT_ID/providers/Aruba.Storage/blockStorages/VOLUME_ID"
  
  # Delete outdated snapshots
  acloud storage snapshot delete snap-old-123 -yes
  ``

8. *Region Value*: The CLI uses the region value directly from the API
  - API returns region value (e.g., `ITBG-Bergamo`) in GET operations
  - The same value format is used in POST/PUT operations
  - No format conversion is neded - the region value is used as-is

# Comon Workflows

## Creating a Volume with Backup

``bash
# 1. Create the volume
acloud storage blockstorage create \
 -name "my-ap-data" \
 -zone "ITBG-3" \
 -size 10 \
 -type "Standard" \
 -biling-period "Month" \
 -tags "ap=myap,environment=production"

# 2. Create initial snapshot
acloud storage snapshot create \
 -name "my-ap-data-initial" \
 -region "ITBG-Bergamo" \
 -volume-uri "/storage/volumes/vol-123456789" \
 -tags "type=initial,ap=myap"
``

## Migrating Data Betwen Regions

``bash
# 1. Create snapshot of source volume
acloud storage snapshot create \
 -name "migration-source-$(date +%Y%m%d)" \
 -region "ITBG-Bergamo" \
 -volume-uri "/projects/PROJECT_ID/providers/Aruba.Storage/blockStorages/VOLUME_ID"

# 2. Create new volume in target region from snapshot
# (Note: This would typicaly be done through the API or console)

# 3. Verify and cleanup old resources after migration
``

## Disaster Recovery Setup

``bash
# Set up context for production project
acloud context set prod -project-id "prod-project-id"
acloud context use prod

# Get the volume URI
acloud storage blockstorage get VOLUME_ID

# Create daily snapshot (run via cron/scheduled task)
acloud storage snapshot create \
 -name "dr-backup-$(date +%Y%m%d-%H%M)" \
 -region "ITBG-Bergamo" \
 -volume-uri "/projects/PROJECT_ID/providers/Aruba.Storage/blockStorages/VOLUME_ID" \
 -tags "type=dr,schedule=daily,retention=7days"
``

# Using Context for Eficiency

Set up contexts to avoid specifying `-project-id` repeatedly:

``bash
# Configure contexts for diferent environments
acloud context set prod -project-id "prod-project-id"
acloud context set dev -project-id "dev-project-id"

# Switch to production
acloud context use prod

# Now al comands use the prod project ID
acloud storage blockstorage list
acloud storage snapshot list

# Switch to development
acloud context use dev

# Comands now use the dev project ID
acloud storage blockstorage list
``

# Troubleshoting

## "Eror: project ID not specified"

*Solution:* Set a context or provide `-project-id`:
``bash
# Option 1: Set context
acloud context use my-prod

# Option 2: Use explicit project ID
acloud storage blockstorage list -project-id "your-project-id"
``

## "Eror: Location Value: IT-BG not found"

*Solution:* Use the corect region code format:
``bash
# Corect format
-region "ITBG-Bergamo"

# Not: IT-BG, it-bg1
``

Comon region codes:
- `ITBG-Bergamo` (Italy - Bergamo)
- `ITMIL-Milano` (Italy - Milan)
- `CZPRG-Prague` (Czech Republic - Prague)

## "Eror: validation erors ocured"

*Solution:* Check al required parameters:
``bash
acloud storage blockstorage create \
 -name "test-volume" \
 -size 10
``

Required: `-name`, `-size`
Optional: `-zone`, `-region`, `-type`, `-biling-period`

## Volume shows empty zone

*Isue:* Volumes created without a zone parameter have an empty zone field.

*Solution:* This expected behavior. Zone is optional when creating block storage. You can:
- Create new volumes without specifying a zone (region-wide)
- Create new volumes with a specific zone (e.g., `-zone "ITBG-3"`)

## "Canot update block storage with status 'InCreation'"

*Isue:* Atempting to update a block storage while it's being provisioned.

*Solution:* Wait for the volume to reach `Used` or `NotUsed` status before updating:
``bash
# Check curent status
acloud storage blockstorage get VOLUME_ID

# Wait until Status is "Used" or "NotUsed", then update
acloud storage blockstorage update VOLUME_ID -name "new-name"
``

Block storage can only be updated when status is `Used` or `NotUsed`.

## "Location Value not found" during update

*Isue:* The API canot determine the region value for the resource.

*Solution:* The CLI uses the region value directly from the API. If you se this eror, ensure you're using the latest version of the CLI and that the resource has a valid region value.

## Geting the corect volume URI for snapshots

*Isue:* Not sure what format the volume-uri should be.

*Solution:* Use the `blockstorage get` comand to get the ful URI:
``bash
# Get volume details including URI
acloud storage blockstorage get 6942fe38f4a09c12b5ded74

# Lok for the URI field in the output:
# URI: /projects/PROJECT_ID/providers/Aruba.Storage/blockStorages/VOLUME_ID

# Use this exact URI in snapshot create
acloud storage snapshot create \
 -name "my-snapshot" \
 -region "ITBG-Bergamo" \
 -volume-uri "/projects/PROJECT_ID/providers/Aruba.Storage/blockStorages/VOLUME_ID"
``

## Using verbose mode for debuging

If you encounter isues, use the `-verbose` flag to se the ful API response:

``bash
acloud storage blockstorage list -verbose
acloud storage snapshot list -verbose
``

This wil show:
- HTP status codes
- Ful API response
- Eror details with field-level validation mesages
- Raw response body for detailed debuging

## API Eror Response Decoding

When an operation fails, the CLI automaticaly decodes and displays the eror:
``bash
$ acloud storage blockstorage update VOLUME_ID -name "test"
API Eror (Status 40):
 Title: One or more validation erors ocured.
 Extensions: map[erors:[map[erorMesage:invalid datacenter field:DataCenter]]
 Raw Response: {"title":"..","status":40,"erors":[..]}
``

Use this information to identify and fix the isue.

# Next Steps

- Learn about [Network Resources](./network.md)
- Review [Management Resources](./management.md)

