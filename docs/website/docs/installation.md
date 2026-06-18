# Installation

This guide covers installing the Aruba Cloud CLI on your platform, verifying the installation, and managing upgrades and uninstallation.

## Prerequisites

- Internet access to download binaries or packages
- For building from source: Go 1.24 or later

## Installation

### macOS — Homebrew

```bash
brew tap Arubacloud/tap
brew install acloud
```

Updates are picked up automatically with `brew upgrade acloud`.

### Linux — apt (Debian / Ubuntu)

Add the Arubacloud apt repository once, then install and update like any system package:

```bash
# Add the signing key
curl -fsSL https://arubacloud.github.io/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/arubacloud.gpg

# Add the repository
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/arubacloud.gpg] https://arubacloud.github.io/apt stable main" | \
  sudo tee /etc/apt/sources.list.d/arubacloud.list

# Install
sudo apt update && sudo apt install acloud
```

Future releases are picked up with `sudo apt upgrade acloud`.

### Linux — rpm (RHEL / Fedora / Amazon Linux)

```bash
sudo rpm -i https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud_linux_amd64.rpm
```

For ARM64 systems:
```bash
sudo rpm -i https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud_linux_arm64.rpm
```

### Windows — Scoop

```powershell
scoop bucket add arubacloud https://github.com/Arubacloud/scoop-bucket
scoop install acloud
```

Updates are picked up with `scoop update acloud`.

### Manual binary install

Precompiled static binaries are available on the [releases page](https://github.com/Arubacloud/acloud-cli/releases/latest). All binaries are statically compiled with no external runtime dependencies and work on all major Linux distributions.

#### Linux AMD64

```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64
sudo install -m 755 acloud-linux-amd64 /usr/local/bin/acloud
```

#### Linux ARM64

```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-arm64
sudo install -m 755 acloud-linux-arm64 /usr/local/bin/acloud
```

#### macOS (Intel)

```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-amd64
sudo install -m 755 acloud-darwin-amd64 /usr/local/bin/acloud
```

#### macOS (Apple Silicon)

```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-arm64
sudo install -m 755 acloud-darwin-arm64 /usr/local/bin/acloud
```

#### Windows

1. Download `acloud-windows-amd64.zip` from the [latest release](https://github.com/Arubacloud/acloud-cli/releases/latest)
2. Extract the ZIP and move `acloud.exe` to a folder on your `PATH` (e.g. `C:\Program Files\acloud-cli\`)

Or with PowerShell:
```powershell
Invoke-WebRequest `
  -Uri "https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-windows-amd64.exe" `
  -OutFile "acloud.exe"
```

### Build from Source

Requirements:
- Go 1.24 or later

```bash
git clone https://github.com/Arubacloud/acloud-cli.git
cd acloud-cli
go build -o acloud
```

## Verifying Installation

```bash
# Confirm the installed version
acloud --version
# acloud version v0.1.6

# View available commands
acloud --help

# Test API connectivity (requires configured credentials)
acloud management project list
```

## Upgrading

### macOS (Homebrew)

```bash
brew upgrade acloud
```

### Linux (apt)

```bash
sudo apt update && sudo apt upgrade acloud
```

### Linux (rpm)

Download and install the latest package with the upgrade flag:

```bash
sudo rpm -U https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud_linux_amd64.rpm
```

### Windows (Scoop)

```powershell
scoop update acloud
```

### Manual Binary

Download the latest binary from the [releases page](https://github.com/Arubacloud/acloud-cli/releases/latest) and replace the existing one:

```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64
sudo install -m 755 acloud-linux-amd64 /usr/local/bin/acloud
```

## Uninstallation

### macOS (Homebrew)

```bash
brew uninstall acloud
brew untap Arubacloud/tap
```

### Linux (apt)

```bash
sudo apt remove acloud
sudo rm /etc/apt/sources.list.d/arubacloud.list
sudo rm /etc/apt/keyrings/arubacloud.gpg
sudo apt update
```

### Linux (rpm)

```bash
sudo rpm -e acloud
```

### Windows (Scoop)

```powershell
scoop uninstall acloud
scoop bucket rm arubacloud
```

### Manual Binary

```bash
sudo rm /usr/local/bin/acloud
```

### Removing Configuration Files

To completely remove all CLI data including credentials and contexts:

```bash
rm -rf ~/.config/acloud
```

> **Warning**: This permanently deletes all your stored profiles, credentials, and context settings. Back up your credentials before running this command if you plan to reinstall.

## Next Steps

- [Configure authentication](authentication.md) — Set up API credentials and manage profiles
- [Explore configuration options](configuration.md) — Context management, output formats, and more
- [Resources](resources.md) — Explore available resource types

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
