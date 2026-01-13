# Aruba Cloud CLI Documentation

Welcome to the Aruba Cloud CLI (`acloud`) documentation. This CLI provides a powerful command-line interface for managing your Aruba Cloud resources.

## Table of Contents

- [Getting Started](getting-started.md)
  - [Installation](getting-started.md#installation)
  - [Authentication](getting-started.md#authentication)
  - [Context Management](getting-started.md#context-management)
  - [Auto-completion](getting-started.md#auto-completion)
- [Resources](https://arubacloud.github.io/acloud-cli/docs/resources)
  - Full documentation is available on the [documentation website](https://arubacloud.github.io/acloud-cli/)
  - For local development, see `docs/website/docs/resources/`


## Quick Start

```bash
# Configure credentials
acloud config set

# Set up a context (optional but recommended)
acloud context set my-prod --project-id "your-project-id"
acloud context use my-prod

# List projects
acloud management project list

# List storage resources (uses context)
acloud storage blockstorage list
acloud storage snapshot list

# List network resources
acloud network vpc list
acloud network elasticip list
acloud network loadbalancer list
```

## Getting Help

For help with any command, use the `--help` flag:

```bash
acloud --help
acloud management --help
acloud management project --help
acloud management project create --help
```

## Debug Mode

The CLI provides a global `--debug` (or `-d`) flag for troubleshooting. When enabled, it shows HTTP request/response details, request payloads, and full error information. See [Getting Started - Debug Mode](getting-started.md#debug-mode) for more details.

## Additional Resources

- [GitHub Repository](https://github.com/Arubacloud/acloud-cli)
- [Aruba Cloud Documentation](https://kb.arubacloud.com/en/home.aspx)
- [SDK Documentation](https://github.com/Arubacloud/sdk-go)
