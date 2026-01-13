# Installation

This guide covers installing the Aruba Cloud CLI on your platform and initial configuration.

## Installation

### Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/Arubacloud/acloud-cli/releases).

#### Linux AMD64

**For Ubuntu 22.04+ or newer distributions:**
```bash
# Download and extract
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64.tar.gz
tar -xzf acloud-linux-amd64.tar.gz

# Move to PATH
sudo mv acloud-linux-amd64 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

**For Ubuntu 20.04 or older WSL distributions (GLIBC 2.31 compatible):**

If you encounter GLIBC version errors (e.g., `GLIBC_2.34 not found`), use the Ubuntu 20.04 compatible binary:
```bash
# Download and extract Ubuntu 20.04 compatible binary
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64-ubuntu20.tar.gz
tar -xzf acloud-linux-amd64-ubuntu20.tar.gz

# Move to PATH
sudo mv acloud-linux-amd64-ubuntu20 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

> **Note:** The Ubuntu 20.04 compatible binary works on Ubuntu 20.04, 22.04, 24.04, and newer versions. Use this version if you're running older WSL distributions or encounter GLIBC compatibility issues.

#### Linux ARM64

**For Ubuntu 22.04+ or newer distributions:**
```bash
# Download and extract
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-arm64.tar.gz
tar -xzf acloud-linux-arm64.tar.gz

# Move to PATH
sudo mv acloud-linux-arm64 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

**For Ubuntu 20.04 or older distributions (GLIBC 2.31 compatible):**

If you encounter GLIBC version errors, use the Ubuntu 20.04 compatible binary:
```bash
# Download and extract Ubuntu 20.04 compatible binary
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-arm64-ubuntu20.tar.gz
tar -xzf acloud-linux-arm64-ubuntu20.tar.gz

# Move to PATH
sudo mv acloud-linux-arm64-ubuntu20 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

#### macOS (Intel)
```bash
# Download and extract
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-amd64.tar.gz
tar -xzf acloud-darwin-amd64.tar.gz

# Move to PATH
sudo mv acloud-darwin-amd64 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

#### macOS (Apple Silicon)
```bash
# Download and extract
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-arm64.tar.gz
tar -xzf acloud-darwin-arm64.tar.gz

# Move to PATH
sudo mv acloud-darwin-arm64 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

#### Windows

**Using PowerShell:**
```powershell
# Download
Invoke-WebRequest -Uri "https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-windows-amd64.zip" -OutFile "acloud-windows-amd64.zip"

# Extract
Expand-Archive -Path acloud-windows-amd64.zip -DestinationPath .

# Add to PATH (current session)
$env:Path += ";$PWD"

# Or move to a permanent location and add to PATH
# Move acloud-windows-amd64.exe to C:\Program Files\acloud-cli\
# Then add C:\Program Files\acloud-cli\ to your system PATH
```

**Using Command Prompt:**
1. Download `acloud-windows-amd64.zip` from the [releases page](https://github.com/Arubacloud/acloud-cli/releases)
2. Extract the ZIP file
3. Move `acloud-windows-amd64.exe` to a folder (e.g., `C:\Program Files\acloud-cli\`)
4. Add that folder to your system PATH environment variable
5. Rename `acloud-windows-amd64.exe` to `acloud.exe` for convenience

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

2. **Configure the CLI**:
   ```bash
   acloud config set
   ```

3. **Enter your credentials** when prompted:
   - Client ID
   - Client Secret

4. **Verify configuration**:
   ```bash
   acloud config show
   ```

### Configuration File

Credentials are stored in `~/.acloud.yaml`:

```yaml
clientId: your-client-id
clientSecret: your-client-secret
```

**Security Note**: Keep your credentials secure. The configuration file contains sensitive information.

### Environment Variables

You can also set credentials via environment variables:

```bash
export ACLOUD_CLIENT_ID="your-client-id"
export ACLOUD_CLIENT_SECRET="your-client-secret"
```

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
- Explore the [Resource Documentation](resources/management.md)

## Debug Mode

The CLI provides a global `--debug` (or `-d`) flag that enables verbose logging to help troubleshoot issues. When enabled, it shows:

- **HTTP Request/Response details**: All HTTP requests and responses made by the SDK
- **Request payloads**: JSON-formatted request bodies being sent to the API
- **Error details**: Full error response bodies when requests fail

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

### GLIBC Version Errors

If you see errors like:
```
acloud: /lib/x86_64-linux-gnu/libc.so.6: version 'GLIBC_2.34' not found
```

This means your Linux distribution has an older GLIBC version than required. **Solution:** Use the Ubuntu 20.04 compatible binary:

```bash
# Download Ubuntu 20.04 compatible binary instead
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64-ubuntu20.tar.gz
tar -xzf acloud-linux-amd64-ubuntu20.tar.gz
sudo mv acloud-linux-amd64-ubuntu20 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

The Ubuntu 20.04 compatible binaries work on Ubuntu 20.04, 22.04, 24.04, and newer versions.

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
