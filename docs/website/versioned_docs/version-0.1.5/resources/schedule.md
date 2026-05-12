# Schedule Resources

The `schedule` category provides commands for managing scheduled jobs in Aruba Cloud. Scheduled jobs allow you to automate tasks that run at specific times or on recurring schedules.

## Available Resources

### [Jobs](schedule/job.md)

Scheduled jobs that execute actions at specified times or on recurring schedules using CRON expressions.

**Quick Commands:**
```bash
# List all scheduled jobs
acloud schedule job list

# Get job details
acloud schedule job get <job-id>

# Create a OneShot job (runs once at a specific time)
acloud schedule job create --name "my-oneshot-job" --region "ITBG-Bergamo" --job-type "OneShot" --schedule-at "2024-12-31T23:59:59Z"

# Create a Recurring job (runs on a schedule)
acloud schedule job create --name "my-recurring-job" --region "ITBG-Bergamo" --job-type "Recurring" --cron "0 0 * * *" --execute-until "2025-12-31T23:59:59Z"

# Update a job
acloud schedule job update <job-id> --name "updated-name" --enabled false

# Delete a job
acloud schedule job delete <job-id>
```

## Command Structure

All schedule commands follow this structure:

```
acloud schedule <resource> <action> [arguments] [flags]
```

Where:
- `<resource>`: The type of resource (e.g., `job`)
- `<action>`: The operation to perform (e.g., `list`, `get`, `create`, `update`, `delete`)
- `[arguments]`: Required arguments (e.g., resource IDs)
- `[flags]`: Optional flags (e.g., `--name`, `--job-type`, `--cron`)

## Job Types

### OneShot Jobs

OneShot jobs execute once at a specified date and time. They are useful for:
- One-time maintenance tasks
- Scheduled deployments
- Time-based automation

**Required flags:**
- `--schedule-at`: Date and time when the job should run (ISO 8601 format)

### Recurring Jobs

Recurring jobs execute on a schedule defined by a CRON expression. They are useful for:
- Daily backups
- Weekly reports
- Periodic maintenance

**Required flags:**
- `--cron`: CRON expression defining the schedule
- `--execute-until`: End date until which the job can run

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

**Examples:**
- `0 0 * * *` - Daily at midnight
- `0 */6 * * *` - Every 6 hours
- `0 0 1 * *` - First day of every month at midnight
- `0 0 * * 0` - Every Sunday at midnight

## Common Patterns

### Listing Jobs

```bash
acloud schedule job list
```

### Getting Job Details

```bash
acloud schedule job get <job-id>
```

### Creating a OneShot Job

```bash
acloud schedule job create \
  --name "backup-job" \
  --region "ITBG-Bergamo" \
  --job-type "OneShot" \
  --schedule-at "2024-12-31T23:59:59Z" \
  --enabled true \
  --tags "backup,automation"
```

### Creating a Recurring Job

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

### Updating a Job

```bash
acloud schedule job update <job-id> \
  --name "updated-name" \
  --enabled false \
  --tags "updated,disabled"
```

### Deleting a Job

```bash
acloud schedule job delete <job-id> [--yes]
```

## Project Context

Scheduled jobs are scoped to a project. You can either:

1. **Use the `--project-id` flag:**
   ```bash
   acloud schedule job list --project-id <project-id>
   ```

2. **Set a context:**
   ```bash
   acloud context set my-prod --project-id <project-id>
   acloud schedule job list  # Uses context project ID
   ```

See [Installation - Context Management](../installation.md#context-management) for more information.

## Next Steps

- Explore [Management Resources](management.md) for organization-level resources
- Check [Storage Resources](storage.md) for storage operations
- Review [Network Resources](network.md) for networking capabilities
- See [Database Resources](database.md) for database management

