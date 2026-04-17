# Scheduled Job Management

Scheduled jobs allow you to automate tasks that run at specific times or on recurring schedules using CRON expressions.

## Available Commands

- `acloud schedule job create` - Create a new scheduled job
- `acloud schedule job list` - List all scheduled jobs
- `acloud schedule job get` - Get details of a specific job
- `acloud schedule job update` - Update job properties
- `acloud schedule job delete` - Delete a scheduled job

## Job Types

### OneShot Jobs

OneShot jobs execute once at a specified date and time. They are useful for:
- One-time maintenance tasks
- Scheduled deployments
- Time-based automation

### Recurring Jobs

Recurring jobs execute on a schedule defined by a CRON expression. They are useful for:
- Daily backups
- Weekly reports
- Periodic maintenance

## Create OneShot Job

Create a job that runs once at a specific time.

### Usage

```bash
acloud schedule job create --name <name> --region <region> --job-type "OneShot" --schedule-at <datetime> [flags]
```

### Required Flags

- `--name` - Name for the job
- `--region` - Region code (e.g., "ITBG-Bergamo")
- `--job-type` - Must be "OneShot"
- `--schedule-at` - Date and time when the job should run (ISO 8601 format, e.g., "2024-12-31T23:59:59Z")

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--enabled` - Enable the job (default: true)
- `--tags` - Tags (comma-separated)

### Example

```bash
acloud schedule job create \
  --name "deploy-release" \
  --region "ITBG-Bergamo" \
  --job-type "OneShot" \
  --schedule-at "2024-12-31T23:59:59Z" \
  --enabled true \
  --tags "deployment,production"
```

## Create Recurring Job

Create a job that runs on a recurring schedule.

### Usage

```bash
acloud schedule job create --name <name> --region <region> --job-type "Recurring" --cron <cron-expression> --execute-until <datetime> [flags]
```

### Required Flags

- `--name` - Name for the job
- `--region` - Region code (e.g., "ITBG-Bergamo")
- `--job-type` - Must be "Recurring"
- `--cron` - CRON expression defining the schedule
- `--execute-until` - End date until which the job can run (ISO 8601 format)

### Optional Flags

- `--project-id` - Project ID (uses context if not specified)
- `--enabled` - Enable the job (default: true)
- `--tags` - Tags (comma-separated)

### Example

```bash
acloud schedule job create \
  --name "daily-backup" \
  --region "ITBG-Bergamo" \
  --job-type "Recurring" \
  --cron "0 2 * * *" \
  --execute-until "2025-12-31T23:59:59Z" \
  --enabled true \
  --tags "backup,daily"
```

## CRON Expression Format

CRON expressions follow the standard format:
```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
* * * * *
```

### Common CRON Examples

- `0 0 * * *` - Daily at midnight
- `0 */6 * * *` - Every 6 hours
- `0 0 1 * *` - First day of every month at midnight
- `0 0 * * 0` - Every Sunday at midnight
- `0 0 1 1 *` - January 1st at midnight (yearly)
- `*/15 * * * *` - Every 15 minutes

## List Jobs

List all scheduled jobs in your project.

### Usage

```bash
acloud schedule job list [flags]
```

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud schedule job list
```

## Get Job Details

Retrieve detailed information about a specific job.

### Usage

```bash
acloud schedule job get <job-id> [flags]
```

### Arguments

- `job-id` (required): The unique ID of the job

### Flags

- `--project-id` - Project ID (uses context if not specified)

### Example

```bash
acloud schedule job get 69455aa70d0972656501d45d
```

## Update Job

Update job properties such as name, enabled status, and tags.

### Usage

```bash
acloud schedule job update <job-id> [flags]
```

### Arguments

- `job-id` (required): The unique ID of the job

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--name` - New name for the job
- `--enabled` - Enable or disable the job
- `--tags` - New tags (comma-separated)

### Example

```bash
acloud schedule job update 69455aa70d0972656501d45d \
  --name "updated-job-name" \
  --enabled false \
  --tags "updated,disabled"
```

## Delete Job

Delete a scheduled job.

### Usage

```bash
acloud schedule job delete <job-id> [--yes] [flags]
```

### Arguments

- `job-id` (required): The unique ID of the job

### Flags

- `--project-id` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt

### Example

```bash
acloud schedule job delete 69455aa70d0972656501d45d --yes
```

## Best Practices

- Use descriptive names for jobs
- Set appropriate `execute-until` dates for recurring jobs
- Test CRON expressions before creating jobs
- Monitor job execution status
- Use tags to organize and filter jobs
- Disable jobs instead of deleting when temporarily not needed

## Related Resources

- [Database Backups](../database/backup.md) - Automate backup creation
- [Storage Backups](../storage/backup.md) - Schedule storage backups

