# Block Storage Management

Block storage volumes are persistent storage devices that can be attached to virtual machines in Aruba Cloud.

## Available Commands

- `acloud storage blockstorage create` - Create a new block storage volume
- `acloud storage blockstorage list` - List all block storage volumes
- `acloud storage blockstorage get` - Get details of a specific volume
- `acloud storage blockstorage update` - Update volume name and tags
- `acloud storage blockstorage delete` - Delete a block storage volume

## Create Block Storage

Create a new block storage volume in your project.

### Usage

```bash
acloud storage blockstorage create --name <name> --region <region> --size <size-gb> [flags]
```

### Required Flags

- `--name` - Name for the block storage volume
- `--size` - Size in GB (must be greater than 0)

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--region` - Region code (default: "ITBG-Bergamo")
- `--zone` - Zone/datacenter (optional)
- `--type` - Volume type: Standard or Performance (default: "Standard")
- `--billing-period` - Billing period: Hour, Month, Year (default: "Hour")
- `--tags` - Tags (comma-separated)

### Example

```bash
# Create a 50GB standard block storage
acloud storage blockstorage create \
  --name "my-data-volume" \
  --region "ITBG-Bergamo" \
  --size 50 \
  --type "Standard" \
  --billing-period "Hour" \
  --tags "env,production"
```

### Output

```
Creating block storage with:
  Name: my-data-volume
  Region: ITBG-Bergamo
  Size: 50 GB
  Type: Standard
  Billing Period: Hour
  Project ID: 68398923fb2cb026400d4d31

Block storage created successfully!
ID:              69455aa70d0972656501d45d
Name:            my-data-volume
Size (GB):       50
Type:            Standard
Zone:            DC-BG-IT-1
Region:          ITBG-Bergamo
Status:          NotUsed
Creation Date:   18-12-2025 18:49:06
```

## Create a Bootable Block Storage (for Custom Images)

To create a bootable block storage volume (for example, to use a custom image), use the following command. The `--region` flag is required, and the `--zone` flag is optional:

```bash
acloud storage blockstorage create \
  --name boot-ubuntu \
  --region ITBG-Bergamo \
  --set-bootable \
  --billing-period Hour \
  --size 20 \
  --tags boot \
  --type Performance \
  --image LU22-001
```

Example output:
```
Block storage created successfully!
ID:              697b389bce7dfeef91532563
Name:            boot-ubuntu
Size (GB):       20
Type:            Performance
Zone:            
Region:          ITBG-Bergamo
Status:          InCreation
Creation Date:   29-01-2026 10:38:19
```

> **Note:**
> - The `--region` flag is required. The `--zone` flag is optional and only needed for zonal block storage.
> - Use `--set-bootable` to ensure the volume is created as bootable (required when using the `--image` flag).
> - Replace `LU22-001` with the desired image code.
> - Adjust other parameters as needed for your use case.

### List of Available Images for Bootable Block Storage

Below are some of the available image codes you can use with the `--image` flag when creating a bootable block storage. For the full and up-to-date list, see the [official ArubaCloud API documentation](https://api.arubacloud.com/docs/metadata/#cloud-server-bootvolume).

| Image Code         | Description           | OS Flavor        |
|--------------------|----------------------|------------------|
| alma8              | AlmaLinux 8 64bit    | Linux            |
| alma9              | AlmaLinux 9 64bit    | Linux            |
| DE11-001           | Debian 11 64bit      | Linux            |
| DE12-001           | Debian 12 64bit      | Linux            |
| LU20-001           | Ubuntu 20.04 64bit   | Linux            |
| LU22-001           | Ubuntu 22.04 64bit   | Linux            |
| LU24-001           | Ubuntu 24.04 64bit   | Linux            |
| osuse15_2_x64_1_0  | openSUSE 15 64bit    | Linux            |
| WS19-001_W2K19_1_0 | Windows Server 2019  | Windows          |
| WS22-001_W2K22_1_0 | Windows Server 2022  | Windows          |

> **Note:** Use the value in the "Image Code" column with the `--image` flag. For example: `--image LU22-001` for Ubuntu 22.04.

## List Block Storage

List all block storage volumes in your project.

### Usage

```bash
acloud storage blockstorage list [flags]
```

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `-v, --verbose` - Show detailed debug information

### Example

```bash
acloud storage blockstorage list
```

### Output

```
NAME              ID                          SIZE(GB)  REGION  ZONE         TYPE       STATUS
my-data-volume    69455aa70d0972656501d45d   50        ITBG-Bergamo   DC-BG-IT-1   Standard   NotUsed
app-volume        69455bb80d0972656501d45e   100       ITBG-Bergamo   DC-BG-IT-1   Standard   Used
```

## Get Block Storage Details

Get detailed information about a specific block storage volume.

### Usage

```bash
acloud storage blockstorage get <volume-id> [flags]
```

### Arguments

- `volume-id` - The ID of the block storage volume

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Auto-completion

Volume IDs support auto-completion. Press TAB after typing the command to see available volumes:

```bash
acloud storage blockstorage get <TAB>
# Shows:
# 6965a6c3ffc0fd1ef8ba5612    MyVolume
# 6965a6c3ffc0fd1ef8ba5613    DataVolume
```

### Example

```bash
acloud storage blockstorage get 69455aa70d0972656501d45d
```

### Output

```
Block Storage Details:
======================
ID:              69455aa70d0972656501d45d
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
Name:            my-data-volume
Size (GB):       50
Type:            Standard
Zone:            DC-BG-IT-1
Region:          ITBG-Bergamo
Bootable:        false
Status:          NotUsed
Creation Date:   18-12-2025 18:49:06
Created By:      aru-297647
Tags:            [env production]
```

## Update Block Storage

Update the name and/or tags of a block storage volume.

**Note:** Size updates are not currently supported by the API. Volume must be in "Used" or "NotUsed" status to be updated.

### Usage

```bash
acloud storage blockstorage update <volume-id> [flags]
```

### Arguments

- `volume-id` - The ID of the block storage volume (supports auto-completion)

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--name` - New name for the volume
- `--tags` - New tags (comma-separated)

**Note:** At least one of `--name` or `--tags` must be provided.

### Example

```bash
# Update name only
acloud storage blockstorage update 69455aa70d0972656501d45d --name "prod-data-volume"

# Update tags only
acloud storage blockstorage update 69455aa70d0972656501d45d --tags "production,critical"

# Update both
acloud storage blockstorage update 69455aa70d0972656501d45d \
  --name "prod-data-volume" \
  --tags "production,critical"
```

### Output

```
Block storage updated successfully!
ID:              69455aa70d0972656501d45d
Name:            prod-data-volume
Tags:            [production critical]
Size (GB):       50
Type:            Standard
```

## Delete Block Storage

Delete a block storage volume. This action cannot be undone.

### Usage

```bash
acloud storage blockstorage delete <volume-id> [flags]
```

### Arguments

- `volume-id` - The ID of the block storage volume (supports auto-completion)

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `-y, --yes` - Skip confirmation prompt

### Example

```bash
# With confirmation prompt
acloud storage blockstorage delete 69455aa70d0972656501d45d

# Skip confirmation
acloud storage blockstorage delete 69455aa70d0972656501d45d --yes
```

### Output

```
Block storage 69455aa70d0972656501d45d deleted successfully!
```

## Notes

- Block storage volumes can be created with different types (Standard or Performance)
- Volumes must be detached from VMs before deletion
- The zone is automatically assigned if not specified
- Billing periods can be Hour, Month, or Year
- Tags are useful for organizing and filtering resources
