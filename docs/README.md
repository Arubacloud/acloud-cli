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

## Output Format

All list and get commands accept a global `--output` (or `-o`) flag that controls the output format:

| Value | Description |
|-------|-------------|
| `table` | Human-readable fixed-width table (default) |
| `table-json` | JSON array of flat snake_case objects (one per row) |
| `table-yaml` | YAML sequence of flat snake_case mappings (one per row) |
| `json` | Full SDK response object as indented JSON |
| `yaml` | Full SDK response object as YAML |

See [Getting Started - Output Format](getting-started.md#output-format) for examples and scripting tips.

## Pagination

All list commands accept `--limit` and `--offset` flags for page-based navigation through large result sets. See [Getting Started - Pagination](getting-started.md#pagination) for details.

## Debug Mode

The CLI provides a global `--debug` (or `-d`) flag for troubleshooting. When enabled, it shows HTTP request/response details, request payloads, and full error information.

> **Security Warning**: Debug output may include credentials and tokens from HTTP headers. Do not use in shared terminal sessions or paste its output publicly.

See [Getting Started - Debug Mode](getting-started.md#debug-mode) for more details.

## Additional Resources

- [GitHub Repository](https://github.com/Arubacloud/acloud-cli)
- [Aruba Cloud Documentation](https://kb.arubacloud.com/en/home.aspx)
- [SDK Documentation](https://github.com/Arubacloud/sdk-go)
