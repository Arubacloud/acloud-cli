# acloud-cli

[![GitHub release](https://img.shields.io/github/tag/Arubacloud/acloud-cli.svg?label=release)](https://github.com/Arubacloud/acloud-cli/releases/latest)

**acloud-cli** is the official Command Line Interface (CLI) for the **Aruba Cloud Management Platform**.  
It allows developers, DevOps engineers, and platform operators to interact with Aruba Cloud APIs directly from the terminal for automation, scripting, and infrastructure management.

> ⚠️ **Development Status**  
> This CLI is under active development and is **not production-ready**.  
> Commands, APIs, and behavior may change between releases.

---

## Features and Capabilities

The Aruba Cloud CLI provides programmatic access to the following platform services:

- Project and organization management
- Block storage volumes, snapshots, backups, and restores
- Network resources such as VPCs, subnets, security groups, and VPNs
- Kubernetes as a Service (KaaS) cluster management
- Infrastructure lifecycle operations and automation workflows

This tool is designed for:
- Infrastructure as Code (IaC) workflows
- CI/CD pipelines
- Automation and scripting
- Advanced terminal-based cloud management

---

## Installation

Precompiled binaries are available for Windows, Linux, and macOS.  
No additional runtime dependencies are required.

### Windows

```powershell
Invoke-WebRequest `
  -Uri "https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-windows-amd64.exe" `
  -OutFile "acloud.exe"
```

acloud.exe --help

Optionally move acloud.exe to a directory included in your PATH.

## Linux

### Ubuntu 22.04+ and most modern distributions
```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64
chmod +x acloud-linux-amd64
sudo mv acloud-linux-amd64 /usr/local/bin/acloud
```


### Ubuntu 20.04 or older WSL distributions (GLIBC 2.31 compatible)
```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64-ubuntu20
chmod +x acloud-linux-amd64-ubuntu20
sudo mv acloud-linux-amd64-ubuntu20 /usr/local/bin/acloud
```
If you encounter GLIBC errors such as GLIBC_2.34 not found, use the -ubuntu20 binary.

## MacOS
```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-amd64
chmod +x acloud-darwin-amd64
sudo mv acloud-darwin-amd64 /usr/local/bin/acloud
```
---

## Configuration
Before using the CLI, you must configure your Aruba Cloud API credentials.

### Set Credentials
```bash
# Recommended: pass --client-id on the command line; the secret is prompted securely (echo disabled)
acloud config set --client-id YOUR_CLIENT_ID
# Enter client secret: (hidden input)

# Alternative for CI/automation (both flags on the command line)
acloud config set --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET
```

> **Security note**: Avoid passing `--client-secret` interactively — it will appear in your shell history. Omitting the flag causes the CLI to prompt for it with echo disabled.

Credentials are stored securely in:
```bash
~/.acloud.yaml
```

### View Configuration
```bash
acloud config show
```
---

## Quick Start
### 1. Configure Credentials
```bash
acloud config set --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET
```

### 2. Create and Use a Context (Recommended)
Contexts allow you to work with a specific project without repeatedly passing --project-id
```bash
acloud context set my-prod --project-id "YOUR_PROJECT_ID"
acloud context use my-prod
```

### 3. Explore Resources
```bash
# List projects
acloud management project list

# List block storage volumes
acloud storage blockstorage list

# List snapshots
acloud storage snapshot list
```

## Context Management
Manage multiple project contexts to simplify multi-environment workflows:
```bash
acloud context set prod --project-id "prod-project-id"
acloud context set dev --project-id "dev-project-id"
acloud context set staging --project-id "staging-project-id"

acloud context use prod
acloud context use dev

acloud context current
acloud context list
acloud context delete staging
```bash

## Usage
```bash
acloud --help
acloud config --help
```

## Debug Mode
```bash
acloud --debug network vpc list
# Short form
acloud -d network vpc list
```

Debug mode enables:
- HTTP request and response logging
- Detailed JSON payloads
- Full error response details

> **Security Warning**: Debug output may include credentials and tokens from HTTP headers. Do not use in shared terminal sessions or paste its output publicly.

Debug output is sent to stderr and does not interfere with command output.

## Output Format

All list and get commands accept a global `--output` (`-o`) flag:

| Value | Description |
|-------|-------------|
| `table` | Human-readable fixed-width table (default) |
| `table-json` | JSON array of flat snake_case objects (one per row) |
| `table-yaml` | YAML sequence of flat snake_case mappings (one per row) |
| `json` | Full SDK response object as indented JSON |
| `yaml` | Full SDK response object as YAML |

```bash
acloud network vpc list                  # table (default)
acloud network vpc list -o table-json    # flat JSON array — easy to pipe to jq
acloud network vpc list -o json          # full SDK response envelope
```

Use `table-json` for scripting with tools like `jq`:
```bash
acloud storage blockstorage list -o table-json | jq '.[].name'
```

## Safe Delete (Dry Run)

Every delete command supports `--dry-run` to validate existence without deleting:

```bash
acloud storage blockstorage delete <volume-id> --dry-run
# [dry-run] Would delete block storage '<volume-id>'. Resource exists and is accessible.
```

Pass `--yes` (or `-y`) to skip the interactive confirmation prompt.

## Pagination

All list commands accept `--limit` and `--offset` flags:

```bash
acloud storage blockstorage list --limit 10           # first 10 results
acloud storage blockstorage list --limit 10 --offset 10  # second page
```

## Documentation
📚 Full documentation is available at:
https://arubacloud.github.io/acloud-cli/

The documentation website includes:
- Getting started guides
- Authentication and configuration references
- Complete command and resource documentation
- Examples and tutorials
- Versioned documentation for each CLI release

Local source files are available in the docs/ directory.

## Testing
End-to-end (E2E) tests validate CRUD operations across all resource categories.

### Required Environment Variables
```bash
export ACLOUD_PROJECT_ID="your-project-id"
export ACLOUD_REGION="ITBG-Bergamo"
```

## Run E2E Tests
```bash
./e2e/management/test.sh
./e2e/storage/test.sh
./e2e/network/test.sh
./e2e/container/test.sh
```

Container (KaaS) tests require additional environment variables.
See [e2e/README.md](e2e/README.md) for full instructions and prerequisites.

## Contributing
Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for development guidelines.

## License
See the [LICENSE](LICENSE) file for licensing details.
