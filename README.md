# acloud — The official CLI for Aruba Cloud

[![Build](https://github.com/Arubacloud/acloud-cli/actions/workflows/build.yml/badge.svg)](https://github.com/Arubacloud/acloud-cli/actions/workflows/build.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Arubacloud/acloud-cli)](https://github.com/Arubacloud/acloud-cli/blob/main/go.mod)
[![Release](https://img.shields.io/github/v/release/Arubacloud/acloud-cli)](https://github.com/Arubacloud/acloud-cli/releases)
[![Codecov](https://codecov.io/gh/Arubacloud/acloud-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/Arubacloud/acloud-cli)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

A single, unified command-line interface for managing the full range of Aruba Cloud infrastructure resources directly from your terminal.

---

## Features

- **Full resource coverage** — compute, networking, storage, databases, containers, security, and scheduling
- **Multiple output formats** — human-readable table, `table-json`, `table-yaml`, `json`, `yaml`
- **Shell completion** — Bash, Zsh, Fish, and PowerShell
- **Named contexts** — switch between projects without repeating `--project-id`
- **Safe delete** — `--dry-run` to validate before deleting; `--yes` to skip prompts
- **Pagination** — `--limit` and `--offset` on all list commands
- **Debug mode** — full HTTP request/response logging via `--debug`
- **CI/CD ready** — scriptable, non-interactive, JSON/YAML output for pipeline integration

---

## Resource Coverage

| Category | Resources |
|----------|-----------|
| **Management** | Projects |
| **Compute** | Cloud Servers, Key Pairs |
| **Storage** | Block Storage, Snapshots, Backups, Restores |
| **Network** | VPC, Subnet, Security Group, Security Rule, Elastic IP, Load Balancer, VPC Peering, VPN Tunnel, VPN Route |
| **Container** | Kubernetes as a Service (KaaS), Container Registry |
| **Database** | DBaaS instances, Databases, Users, Grants, Backups |
| **Security** | KMS Keys |
| **Schedule** | Jobs (OneShot and Recurring) |

---

## Installation

### macOS — Homebrew

```bash
brew tap Arubacloud/tap
brew install acloud
```

### Linux — apt (Debian / Ubuntu)

```bash
curl -fsSL https://arubacloud.github.io/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/arubacloud.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/arubacloud.gpg] https://arubacloud.github.io/apt stable main" | \
  sudo tee /etc/apt/sources.list.d/arubacloud.list
sudo apt update && sudo apt install acloud
```

### Linux — rpm (RHEL / Fedora / Amazon Linux)

```bash
sudo rpm -i https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud_linux_amd64.rpm
```

### Windows — Scoop

```powershell
scoop bucket add arubacloud https://github.com/Arubacloud/scoop-bucket
scoop install acloud
```

### Manual binary

Precompiled static binaries (no runtime dependencies) are on the [releases page](https://github.com/Arubacloud/acloud-cli/releases/latest).

```bash
# Linux (amd64)
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64
chmod +x acloud-linux-amd64 && sudo mv acloud-linux-amd64 /usr/local/bin/acloud

# macOS (Apple Silicon)
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-arm64
chmod +x acloud-darwin-arm64 && sudo mv acloud-darwin-arm64 /usr/local/bin/acloud
```

```powershell
# Windows (PowerShell)
Invoke-WebRequest `
  -Uri "https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-windows-amd64.exe" `
  -OutFile "acloud.exe"
```

Move `acloud.exe` to a directory on your `PATH`.

### Verify

```bash
acloud --version
```

---

## Quick Start

### 1. Configure credentials

```bash
# Prompts for the secret with echo disabled (recommended)
acloud config set --client-id YOUR_CLIENT_ID

# Or pass both flags (suitable for CI/CD — prefer environment variables instead)
acloud config set --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET
```

> **Security note**: Avoid passing `--client-secret` in interactive sessions — it will appear in your shell history.

Credentials are stored in `~/.acloud.yaml`.

### 2. Set a context

Contexts bind a name to a project ID so you never need to pass `--project-id` again:

```bash
acloud context set prod --project-id "YOUR_PROJECT_ID"
acloud context use prod
```

### 3. Start managing resources

```bash
acloud management project list
acloud compute cloudserver list
acloud network vpc list
acloud storage blockstorage list
acloud database dbaas list
```

---

## Shell Completion

```bash
# Bash
echo 'source <(acloud completion bash)' >> ~/.bashrc

# Zsh
echo 'source <(acloud completion zsh)' >> ~/.zshrc

# Fish
acloud completion fish | source

# PowerShell
acloud completion powershell | Out-String | Invoke-Expression
```

---

## Output Formats

All `list` and `get` commands accept a global `--output` (`-o`) flag:

| Value | Description |
|-------|-------------|
| `table` | Human-readable fixed-width table (default) |
| `table-json` | Flat JSON array — one object per row, easy to pipe to `jq` |
| `table-yaml` | Flat YAML sequence — one mapping per row |
| `json` | Full SDK response as indented JSON |
| `yaml` | Full SDK response as YAML |

```bash
acloud network vpc list                   # table (default)
acloud network vpc list -o table-json     # flat JSON — pipe to jq
acloud network vpc list -o json           # full SDK envelope

# Example: extract all VPC IDs with jq
acloud network vpc list -o table-json | jq -r '.[].id'
```

---

## Context Management

```bash
acloud context set prod    --project-id "prod-project-id"
acloud context set dev     --project-id "dev-project-id"
acloud context set staging --project-id "staging-project-id"

acloud context use prod
acloud context current
acloud context list
acloud context delete staging
```

---

## Safe Delete

Every delete command supports `--dry-run` to confirm the resource exists without removing it:

```bash
acloud storage blockstorage delete <volume-id> --dry-run
# [dry-run] Would delete block storage '<volume-id>'. Resource exists and is accessible.
```

Pass `--yes` (or `-y`) to skip the interactive confirmation prompt.

---

## Pagination

All list commands support `--limit` and `--offset`:

```bash
acloud storage blockstorage list --limit 10             # first 10 results
acloud storage blockstorage list --limit 10 --offset 10 # second page
```

---

## Debug Mode

```bash
acloud --debug network vpc list
acloud -d network vpc list       # short form
```

Enables full HTTP request/response logging including headers, query parameters, and JSON bodies. Output goes to stderr and does not interfere with command output.

> **Security warning**: Debug output may include credentials and tokens. Do not use in shared terminal sessions or paste its output publicly.

---

## Documentation

Full documentation: **https://arubacloud.github.io/acloud-cli/**

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development guidelines.

## License

See [LICENSE](LICENSE) for licensing details.
