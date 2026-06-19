# Getting Started with Aruba Cloud CLI

This guide will help you install, configure, and start using the Aruba Cloud CLI.

## Installation

### Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/Arubacloud/acloud-cli/releases).

#### Linux

You can install the CLI using either the direct binary or the tarball:

**Option 1: Download the direct binary**
```bash
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64
sudo mv acloud-linux-amd64 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

**Option 2: Download and extract the tarball**
```bash
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64.tar.gz
tar -xzf acloud-linux-amd64.tar.gz
sudo mv acloud /usr/local/bin/
sudo chmod +x /usr/local/bin/acloud
```


#### macOS (Intel)

You can install the CLI using either the direct binary or the tarball:

**Option 1: Download the direct binary**
```bash
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-amd64
sudo mv acloud-darwin-amd64 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

**Option 2: Download and extract the tarball**
```bash
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-amd64.tar.gz
tar -xzf acloud-darwin-amd64.tar.gz
sudo mv acloud /usr/local/bin/
sudo chmod +x /usr/local/bin/acloud
```

#### macOS (Apple Silicon)

You can install the CLI using either the direct binary or the tarball:

**Option 1: Download the direct binary**
```bash
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-arm64
sudo mv acloud-darwin-arm64 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

**Option 2: Download and extract the tarball**
```bash
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-arm64.tar.gz
tar -xzf acloud-darwin-arm64.tar.gz
sudo mv acloud /usr/local/bin/
sudo chmod +x /usr/local/bin/acloud
```


#### Windows

You can install the CLI using either the direct binary or the zip archive:

**Option 1: Download the direct binary**
1. Download `acloud-windows-amd64.exe` from the [latest release](https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-windows-amd64.exe)
2. Move it to a folder (e.g., `C:\\Program Files\\acloud-cli\\`)
3. Add that folder to your system PATH
4. Optionally, rename to `acloud.exe` for convenience

**Option 2: Download and extract the zip archive**
1. Download `acloud-windows-amd64.zip` from the [latest release](https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-windows-amd64.zip)
2. Extract the ZIP file
3. Move `acloud-windows-amd64.exe` to a folder (e.g., `C:\\Program Files\\acloud-cli\\`)
4. Add that folder to your system PATH
5. Optionally, rename to `acloud.exe` for convenience

### Build from Source

Requirements:
- Go 1.24 or later

```bash
git clone https://github.com/Arubacloud/acloud-cli.git
cd acloud-cli
go build -o acloud
```

## Authentication

The Aruba Cloud CLI requires API credentials to authenticate with Aruba Cloud services.

### Setting up Credentials

1. **Obtain API Credentials**: Get your Client ID and Client Secret from the Aruba Cloud console.

2. **Configure the CLI** — pass `--client-id` on the command line; the secret is read securely with echo disabled:
   ```bash
   acloud config set --client-id YOUR_CLIENT_ID
   # Enter client secret: (hidden input, does not appear in shell history)
   ```

   For CI/automation, provide the secret via environment variable instead of the prompt:
   ```bash
   ACLOUD_CLIENT_SECRET=YOUR_CLIENT_SECRET acloud config set --client-id YOUR_CLIENT_ID
   ```

   > **Security note**: `--client-secret` is intentionally not a supported flag — passing secrets on the command line exposes them in process lists and shell history. Use the interactive prompt or `ACLOUD_CLIENT_SECRET`.

3. **Verify configuration**:
   ```bash
   acloud config show
   ```

### Configuration File

Credentials are stored in `~/.config/acloud/config.yaml` (XDG Base Directory, file permissions `0600`):

```yaml
profiles:
  default:
    clientId: your-client-id
    clientSecret: your-client-secret
```

**Security Note**: Keep your credentials secure. The configuration file contains sensitive information.

## Context Management

The CLI provides context management to avoid passing `--project-id` repeatedly. Contexts allow you to save project IDs and switch between them easily.

### Setting up a Context

Create a context with a project ID:

```bash
acloud context set my-prod --project-id "66a10244f62b99c686572a9f"
```

### Using a Context

Switch to a saved context:

```bash
acloud context use my-prod
```

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

Contexts are stored in `~/.acloud-context.yaml`:

```yaml
current-context: my-prod
contexts:
  my-prod:
    project-id: 66a10244f62b99c686572a9f
  my-dev:
    project-id: 66a10244f62b99c686572a9e
```

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

   # Backup create — completes volume ID
   acloud storage backup <TAB>
   # Shows:
   # 6965a6c3ffc0fd1ef8ba5612    MyVolume
   # ...

   # Backup get/update/delete — completes backup ID
   acloud storage backup get <TAB>
   # Shows:
   # 67649dac8c7bb1c5d7c80631    MyBackup
   # 67649dac8c7bb1c5d7c80632    DailyBackup
   # ...

   # Restore create — hierarchical: backup-id then volume-id
   acloud storage restore <TAB>
   # First shows backup IDs:
   # 67649dac8c7bb1c5d7c80631    MyBackup
   # ...
   acloud storage restore 67649dac8c7bb1c5d7c80631 <TAB>
   # Then shows volume IDs:
   # 6965a6c3ffc0fd1ef8ba5612    MyVolume
   # ...

   # Restore get/update/delete — hierarchical: backup-id then restore-id
   acloud storage restore get <TAB>
   # First shows backup IDs:
   # 67649dac8c7bb1c5d7c80631    MyBackup
   # ...
   acloud storage restore get 67649dac8c7bb1c5d7c80631 <TAB>
   # Then shows restore IDs for that backup:
   # 67664dde0aca19a92c2c48bb    RestoreOperation1
   # ...
   ```

   **Database Resources:**
   ```bash
   # DBaaS database commands — hierarchical: dbaas-id then database-name
   acloud database dbaas database list <TAB>
   # Shows DBaaS instance IDs:
   # 69455aa70d0972656501d45d    my-dbaas
   # ...
   acloud database dbaas database get 69455aa70d0972656501d45d <TAB>
   # Shows database names in that instance:
   # appdb
   # reporting
   # ...

   # DBaaS user commands — hierarchical: dbaas-id then username
   acloud database dbaas user list <TAB>
   # Shows DBaaS instance IDs
   acloud database dbaas user get 69455aa70d0972656501d45d <TAB>
   # Shows usernames in that instance

   # DBaaS grant commands — hierarchical: dbaas-id, database-name, grant-id
   acloud database dbaas grant list <TAB>
   # Shows DBaaS instance IDs
   acloud database dbaas grant list 69455aa70d0972656501d45d <TAB>
   # Shows database names
   acloud database dbaas grant get 69455aa70d0972656501d45d appdb <TAB>
   # Shows grant IDs for that database
   ```

   Auto-completion works with `get`, `update`, `delete`, `list`, and `create` commands for all resources.

## Verifying Installation

Test your installation:

```bash
# Check version
acloud --version

# View available commands
acloud --help

# Test API connectivity
acloud management project list
```

## Next Steps

- Learn about [Project Management](resources/management/project.md)
- Explore [Management Resources](resources/management.md)
- Explore [Storage Resources](resources/storage.md)
- Read the [Command Reference](command-reference.md)

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

## Safe Delete (Dry Run)

Every delete command supports `--dry-run` to validate that a resource exists and is accessible without actually deleting it:

```bash
acloud storage blockstorage delete <volume-id> --dry-run
# [dry-run] Would delete block storage '<volume-id>'. Resource exists and is accessible.
```

Pass `--yes` (or `-y`) to skip the interactive confirmation prompt in non-interactive scripts.

## Pagination

All list commands support `--limit` and `--offset` flags to paginate large result sets.

| Flag | Description |
|------|-------------|
| `--limit N` | Return at most N results |
| `--offset N` | Skip the first N results (for page-based navigation) |

```bash
# First page of 10 results
acloud storage blockstorage list --limit 10

# Second page of 10 results
acloud storage blockstorage list --limit 10 --offset 10

# Third page
acloud storage blockstorage list --limit 10 --offset 20
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

### "Error initializing client"

This usually means credentials are not configured. Run:
```bash
acloud config set
```

### "No projects found"

Ensure your credentials have the correct permissions and you have projects in your account.

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
