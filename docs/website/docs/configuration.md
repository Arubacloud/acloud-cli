---
id: configuration
title: Configuration
sidebar_label: Configuration
description: Configure context management, shell auto-completion, output formats, and advanced CLI options.
---

# Configuration

This page covers CLI configuration options beyond credentials: context management, shell auto-completion, output formats, and advanced flags.

## Context Management

The CLI provides context management to avoid passing `--project-id` repeatedly. Contexts allow you to save project IDs and switch between them easily.

### Setting up a Context

Create a context with a project ID:

```bash
acloud context set my-prod --project-id "66a10244f62b99c686572a9f"
```

> **Single-context shortcut**: When only one context is configured, the CLI uses it automatically — you do not need to run `acloud context use` first. If you add a second context later, you will need to explicitly switch with `acloud context use <name>`.

### Using a Context

Switch to a saved context when you have more than one:

```bash
acloud context use my-prod
```

> **Switching profiles**: Contexts store project IDs that belong to a specific tenant. When you switch credential profiles with `acloud config profile use`, the CLI will ask whether to clear all contexts (since the saved project IDs are likely invalid for the new profile). Pass `--clear-contexts` to clear them without being prompted:
> ```bash
> acloud config profile use staging --clear-contexts
> ```

Once a context is active, you can run commands without specifying `--project-id`:

```bash
# Works without --project-id
acloud storage blockstorage list
acloud storage snapshot list
acloud management project get <project-id>
```

### Managing Contexts

**List all contexts:**
```bash
acloud context list
```

Output shows all contexts with the current one marked with `*`:
```
Contexts:
=========
my-prod              Project ID: 66a10244f62b99c686572a9f *
my-dev               Project ID: 66a10244f62b99c686572a9e
my-staging           Project ID: 66a10244f62b99c686572a9d

* = current context
```

**Show current context:**
```bash
acloud context current
```

**Delete a context:**
```bash
acloud context delete my-dev
```

### Context File

Contexts are stored in `~/.config/acloud/context.yaml`:

```yaml
current-context: my-prod
contexts:
  my-prod:
    project-id: 66a10244f62b99c686572a9f
  my-dev:
    project-id: 66a10244f62b99c686572a9e
```

> **Legacy path**: Earlier versions stored this file at `~/.acloud-context.yaml`. The CLI migrates it automatically on first run.

### Overriding Context

You can always override the context by explicitly passing `--project-id`:

```bash
# Uses context project ID
acloud storage blockstorage list

# Overrides with specific project ID
acloud storage blockstorage list --project-id "different-project-id"
```

## Auto-completion

The CLI supports shell auto-completion for commands, flags, and resource IDs.

### Bash

#### Current Session
```bash
source <(acloud completion bash)
```

#### Permanent Installation

**Linux:**
```bash
acloud completion bash | sudo tee /etc/bash_completion.d/acloud
```

**macOS:**
```bash
acloud completion bash > $(brew --prefix)/etc/bash_completion.d/acloud
```

After installation, restart your shell or run:
```bash
source ~/.bashrc  # or ~/.bash_profile on macOS
```

### Zsh

Add to `~/.zshrc`:

```bash
# Enable completion
autoload -Uz compinit
compinit

# Load acloud completion
source <(acloud completion zsh)
```

Or for permanent installation:
```bash
acloud completion zsh > "${fpath[1]}/_acloud"
```

### Fish

```bash
acloud completion fish | source
```

Or for permanent installation:
```bash
acloud completion fish > ~/.config/fish/completions/acloud.fish
```

### PowerShell

Add to your PowerShell profile:

```powershell
acloud completion powershell | Out-String | Invoke-Expression
```

## Features of Auto-completion

The auto-completion system provides:

1. **Command completion**: Tab-complete commands and subcommands
   ```bash
   acloud man<TAB>  # completes to "management"
   ```

2. **Flag completion**: Tab-complete available flags
   ```bash
   acloud config set --<TAB>  # shows available flags
   ```

3. **Resource ID completion**: Tab-complete resource IDs with descriptions

   **Management Resources:**
   ```bash
   acloud management project get <TAB>
   # Shows:
   # 655b2822af30f667f826994e    defaultproject
   # 66a10244f62b99c686572a9f    develop
   # ...
   ```

   **Storage Resources:**
   ```bash
   # Block Storage
   acloud storage blockstorage get <TAB>
   # Shows:
   # 6965a6c3ffc0fd1ef8ba5612    MyVolume
   # 6965a6c3ffc0fd1ef8ba5613    DataVolume
   # ...

   # Snapshots
   acloud storage snapshot get <TAB>
   # Shows:
   # 696c9edce63c1af07d60d0c7    MySnapshot
   # 696c9edce63c1af07d60d0c8    BackupSnapshot
   # ...

   # Backups
   acloud storage backup get <TAB>
   # Shows:
   # 67649dac8c7bb1c5d7c80631    MyBackup
   # 67649dac8c7bb1c5d7c80632    DailyBackup
   # ...

   # Restores (hierarchical: backup-id then restore-id)
   acloud storage restore get <TAB>
   # First shows backup IDs:
   # 67649dac8c7bb1c5d7c80631    MyBackup
   # ...
   acloud storage restore get 67649dac8c7bb1c5d7c80631 <TAB>
   # Then shows restore IDs for that backup:
   # 67664dde0aca19a92c2c48bb    RestoreOperation1
   # ...
   ```

   Auto-completion works with `get`, `update`, and `delete` commands for all resources.

## Output Format

All list and get commands support a global `--output` (or `-o`) flag that controls the output format.

| Value | Description |
|-------|-------------|
| `table` | Human-readable fixed-width table (default) |
| `table-json` | JSON array of flat snake_case objects (one per row) |
| `table-yaml` | YAML sequence of flat snake_case mappings (one per row) |
| `json` | Full SDK response object as indented JSON |
| `yaml` | Full SDK response object as YAML |

```bash
# Default table output
acloud network vpc list

# Flat JSON array — easy to pipe to jq
acloud network vpc list -o table-json

# Full SDK response envelope
acloud network vpc list -o json
```

`table-json` is the best choice for scripting with tools like `jq`:
```bash
acloud storage blockstorage list -o table-json | jq '.[].name'
```

## Synchronous Provisioning (--wait)

By default, `create` and `update` commands return as soon as the API accepts the request — the resource may still be provisioning in the background. Pass `--wait` to block until the resource reaches an operational state (`Active`, `Running`, etc.) or fails.

```bash
# Block until the VPC is Active
acloud network vpc create --name "prod-vpc" --region "IT-BG" --wait

# Block until the DBaaS is Active; extend the timeout to 10 minutes
acloud --timeout 10m database dbaas create --name "prod-db" --region "IT-BG" \
  --engine-id "mysql-8.0" --flavor "DBO4A8" --storage-size 50 \
  --zone "ITBG-1" --wait
```

Progress is printed to stderr (so it doesn't mix with `-o json` output):
```
Waiting for VPC 'prod-vpc' to become Active... [12s]
VPC 'prod-vpc' is Active after 15s.
```

**Supported resources:** `cloudserver`, `vpc`, `subnet`, `securitygroup`, `vpntunnel`, `elasticip`, `kaas`, `containerregistry`, `dbaas`, `kms`, `blockstorage` — on both create and update commands.

**Timeout:** `--wait` respects the global `--timeout` flag (default `30s`). For resources that take longer to provision (KaaS clusters, DBaaS instances), set a longer timeout:
```bash
acloud --timeout 15m container kaas create ... --wait
```

**Exit code:** Returns non-zero if the resource reaches a failure state (`Error`, `Failed`) or the timeout expires before the resource becomes operational.

## Safe Delete (Dry Run)

Every delete command supports `--dry-run` to validate that a resource exists without deleting it:

```bash
acloud storage blockstorage delete <volume-id> --dry-run
# [dry-run] Would delete block storage '<volume-id>'. Resource exists and is accessible.
```

Pass `--yes` (or `-y`) to skip the interactive confirmation prompt.

## Pagination

All list commands support `--limit` and `--offset` flags to paginate large result sets.

| Flag | Description |
|------|-------------|
| `--limit N` | Return at most N results |
| `--offset N` | Skip the first N results |

```bash
# First page of 10 results
acloud storage blockstorage list --limit 10

# Second page
acloud storage blockstorage list --limit 10 --offset 10
```

When neither flag is passed the API returns its default result set.

## Debug Mode

The CLI provides a global `--debug` (or `-d`) flag that enables verbose logging to help troubleshoot issues. When enabled, it shows:

- **HTTP Request/Response details**: All HTTP requests and responses made by the SDK
- **Request payloads**: JSON-formatted request bodies being sent to the API
- **Error details**: Full error response bodies when requests fail

> **Security Warning**: Debug output may include credentials and tokens from HTTP headers. Do not use `--debug` in shared terminal sessions or paste its output publicly.

### Usage

Add the `--debug` flag to any command:

```bash
# Enable debug logging for a command
acloud --debug network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --tags test

# Short form
acloud -d network vpc list
```

### Example Output

When debug mode is enabled, you'll see additional output like:

```
[ArubaSDK] 2025-01-15 10:30:45.123456 HTTP Request: PUT https://api.arubacloud.com/...
[ArubaSDK] 2025-01-15 10:30:45.234567 Request Headers: ...
[ArubaSDK] 2025-01-15 10:30:45.345678 Request Body: {...}

=== DEBUG: Security Rule Update Request ===
VPC ID: 69495ef64d0cdc87949b71ec
Security Group ID: 694b05ac4d0cdc87949b75f9
Security Rule ID: 694b06564d0cdc87949b7608
Request Payload:
{
  "metadata": {
    "name": "my-rule",
    "tags": ["test"],
    ...
  },
  ...
}
==========================================

[ArubaSDK] 2025-01-15 10:30:46.456789 HTTP Response: 200 OK
[ArubaSDK] 2025-01-15 10:30:46.567890 Response Body: {...}
```

**Note**: Debug output is sent to `stderr`, so it won't interfere with normal command output and can be redirected separately if needed.

## Troubleshooting

### Debugging API Errors

If you encounter API errors (e.g., 500 Internal Server Error), use the `--debug` flag to see the full request and response:

```bash
acloud --debug network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --tags test
```

This will show:
- The exact request payload being sent
- The full HTTP response (including error details)
- Any SDK-level logging

### Auto-completion not working

1. Ensure bash-completion is installed:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install bash-completion

   # macOS
   brew install bash-completion
   ```

2. Reload your shell configuration:
   ```bash
   source ~/.bashrc
   ```
