# Contributing to acloud-cli

Thank you for your interest in contributing to acloud-cli! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Adding New Commands](#adding-new-commands)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Development Setup

### Prerequisites

- **Go 1.24.2 or higher** - [Download Go](https://golang.org/dl/)
- **Git** - For version control
- **Aruba Cloud API credentials** - For testing (Client ID and Client Secret)

### Getting Started

1. **Fork and clone the repository**

```bash
git clone https://github.com/Arubacloud/acloud-cli.git
cd acloud-cli
```

2. **Install dependencies**

```bash
go mod download
```

3. **Configure your credentials**

```bash
# Build the CLI first
go build -o acloud

# Configure credentials
./acloud config set --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET
```

4. **Build and test**

```bash
go build -o acloud
./acloud --help
```

## Project Structure

```
acloud-cli/
├── cmd/                    # Command implementations
│   ├── root.go            # Root command and shared utilities
│   ├── config.go          # Configuration management
│   ├── context.go         # Context management
│   ├── management/         # Management resources (projects)
│   ├── storage/           # Storage resources (block storage, snapshots, etc.)
│   ├── network/           # Network resources (VPC, subnets, etc.)
│   └── ...                # Other resource categories
├── docs/                   # Documentation
│   ├── README.md          # Documentation index
│   ├── getting-started.md # Getting started guide
│   └── resources/         # Resource-specific documentation
├── e2e/                    # End-to-end tests
│   ├── README.md          # E2E testing guide
│   ├── management/        # Management resource tests
│   ├── storage/           # Storage resource tests
│   └── network/           # Network resource tests
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── LICENSE                 # Apache 2.0 License
├── README.md               # User documentation
└── CONTRIBUTING.md         # This file
```

## Adding New Commands

### Command Organization

Commands are organized by resource category:
- **Management** (`cmd/management.*.go`) - Organization-level resources
- **Storage** (`cmd/storage.*.go`) - Storage resources
- **Network** (`cmd/network.*.go`) - Network resources
- **Compute** (`cmd/compute.go`) - Compute resources
- **Other categories** - As needed

### Adding a New Resource Command

1. **Create a new file** in the appropriate category (e.g., `cmd/network.newresource.go`)

2. **Define the command structure** following the CRUD pattern:

```go
package cmd

import (
    "context"
    "fmt"
    "strings"

    "github.com/Arubacloud/sdk-go/pkg/types"
    "github.com/spf13/cobra"
)

func init() {
    // Register commands
    networkCmd.AddCommand(newresourceCmd)
    newresourceCmd.AddCommand(newresourceCreateCmd)
    newresourceCmd.AddCommand(newresourceGetCmd)
    newresourceCmd.AddCommand(newresourceListCmd)
    newresourceCmd.AddCommand(newresourceUpdateCmd)
    newresourceCmd.AddCommand(newresourceDeleteCmd)

    // Define flags
    newresourceCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
    newresourceCreateCmd.Flags().String("name", "", "Resource name (required)")
    newresourceCreateCmd.Flags().String("region", "", "Region code (required)")
    newresourceCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
}

var newresourceCmd = &cobra.Command{
    Use:   "newresource",
    Short: "Manage new resources",
    Long:  `Perform CRUD operations on new resources in Aruba Cloud.`,
}

var newresourceCreateCmd = &cobra.Command{
    Use:   "create [vpc-id]",
    Short: "Create a new resource",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        vpcID := args[0]
        name, _ := cmd.Flags().GetString("name")
        region, _ := cmd.Flags().GetString("region")
        tags, _ := cmd.Flags().GetStringSlice("tags")

        // Validate required fields
        if name == "" {
            fmt.Println("Error: --name is required")
            return
        }

        // Get project ID and client
        projectID, err := GetProjectID(cmd)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }

        client, err := GetArubaClient()
        if err != nil {
            fmt.Printf("Error initializing client: %v\n", err)
            return
        }

        // Build request
        req := types.NewResourceRequest{
            Metadata: types.RegionalResourceMetadataRequest{
                ResourceMetadataRequest: types.ResourceMetadataRequest{
                    Name: name,
                    Tags: tags,
                },
                Location: types.LocationRequest{
                    Value: region,
                },
            },
            // ... properties
        }

        // Make API call
        ctx := context.Background()
        resp, err := client.FromNetwork().NewResources().Create(ctx, projectID, vpcID, req, nil)
        if err != nil {
            fmt.Printf("Error creating resource: %v\n", err)
            return
        }

        // Handle response
        if resp != nil && resp.IsError() && resp.Error != nil {
            fmt.Printf("Failed to create resource - Status: %d\n", resp.StatusCode)
            if resp.Error.Title != nil {
                fmt.Printf("Error: %s\n", *resp.Error.Title)
            }
            if resp.Error.Detail != nil {
                fmt.Printf("Detail: %s\n", *resp.Error.Detail)
            }
            return
        }

        // Display result
        if resp != nil && resp.Data != nil && resp.Data.Metadata.ID != nil {
            headers := []TableColumn{
                {Header: "NAME", Width: 30},
                {Header: "ID", Width: 26},
                {Header: "STATUS", Width: 15},
            }
            row := []string{
                name,
                *resp.Data.Metadata.ID,
                func() string {
                    if resp.Data.Status.State != nil {
                        return *resp.Data.Status.State
                    }
                    return ""
                }(),
            }
            PrintTable(headers, [][]string{row})
        }
    },
}
```

3. **Follow the standard patterns:**
   - Use `GetProjectID(cmd)` for project ID resolution
   - Use `GetArubaClient()` for SDK client (with caching)
   - Use `PrintTable()` for consistent table output
   - Handle errors consistently with `response.IsError() && response.Error != nil`
   - Always display `Tags` field in `get` commands (even if empty)
   - Use `LocationResponse.Value` (not `.Code`) for region display

4. **Add auto-completion** (if applicable):

```go
newresourceGetCmd.ValidArgsFunction = completeNewResourceID
newresourceUpdateCmd.ValidArgsFunction = completeNewResourceID
newresourceDeleteCmd.ValidArgsFunction = completeNewResourceID

func completeNewResourceID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    // Implementation for ID completion
}
```

5. **Update documentation** in `docs/resources/` directory

## Code Style

### General Guidelines

- Follow standard Go conventions and idioms
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and concise
- Handle errors explicitly (don't ignore them)

### Formatting

```bash
# Format all code
go fmt ./...

# Check for common mistakes
go vet ./...

# Run linter (if configured)
golangci-lint run
```

### Error Handling

Always use the standardized error handling pattern:

```go
if resp != nil && resp.IsError() && resp.Error != nil {
    fmt.Printf("Failed to create resource - Status: %d\n", resp.StatusCode)
    if resp.Error.Title != nil {
        fmt.Printf("Error: %s\n", *resp.Error.Title)
    }
    if resp.Error.Detail != nil {
        fmt.Printf("Detail: %s\n", *resp.Error.Detail)
    }
    return
}
```

### Region Handling

- Always use `LocationResponse.Value` (not `.Code`) for display
- Use `LocationResponse.Value` directly in API requests (no normalization needed)
- Handle nil `LocationResponse` gracefully

### Output Formatting

- Use `PrintTable()` for list commands
- Display detailed information for `get` commands
- Always show `Tags` field (even if empty: `Tags: []`)

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./cmd/...
```

### End-to-End (E2E) Tests

E2E tests validate CRUD operations across all resource categories:

```bash
# Set required environment variables
export ACLOUD_PROJECT_ID="your-project-id"
export ACLOUD_REGION="ITBG-Bergamo"

# Run E2E tests
./e2e/management/test.sh
./e2e/storage/test.sh
./e2e/network/test.sh
```

See [e2e/README.md](e2e/README.md) for detailed E2E testing documentation.

### Testing with Debug Mode

Use the `--debug` flag to see HTTP requests/responses:

```bash
./acloud --debug network vpc create --name test-vpc --region ITBG-Bergamo
```

### Test Checklist

Before submitting changes, ensure:
- [ ] All CRUD operations work (Create, Read, Update, Delete)
- [ ] Error handling is consistent
- [ ] Output formatting is correct
- [ ] Auto-completion works (if applicable)
- [ ] Documentation is updated
- [ ] E2E tests pass (if applicable)

## Documentation

### Documentation Structure

Documentation is organized in the `docs/` directory:

- `docs/getting-started.md` - Installation and setup guide
- `docs/resources/` - Resource-specific documentation
  - `management/` - Management resources
  - `storage/` - Storage resources
  - `network/` - Network resources

### Adding Documentation

1. **Create resource documentation** in `docs/resources/<category>/<resource>.md`

2. **Include:**
   - Resource overview
   - Command examples (CREATE, LIST, GET, UPDATE, DELETE)
   - Required and optional flags
   - Example outputs
   - Best practices
   - Troubleshooting tips

3. **Update category index** (e.g., `docs/resources/network.md`)

4. **Update main README** if adding new resource categories

### Documentation Style

- Use clear, concise language
- Include practical examples
- Show both command syntax and expected output
- Document error scenarios
- Keep examples up-to-date with actual CLI behavior

## Submitting Changes

### Before Submitting

1. **Ensure code is formatted:**
   ```bash
   go fmt ./...
   go vet ./...
   ```

2. **Run tests:**
   ```bash
   go test ./...
   ```

3. **Update documentation** for any new features or changes

4. **Test your changes** with real API calls

### Pull Request Process

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** and commit:
   ```bash
   git add .
   git commit -m "Add: description of your changes"
   ```

3. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```

4. **Create a Pull Request** on GitHub with:
   - Clear description of changes
   - Reference to related issues (if any)
   - Screenshots or examples (if applicable)
   - Confirmation that tests pass

### Commit Message Guidelines

Use clear, descriptive commit messages:

- **Format:** `<type>: <description>`
- **Types:**
  - `Add:` - New features
  - `Fix:` - Bug fixes
  - `Update:` - Updates to existing features
  - `Refactor:` - Code refactoring
  - `Docs:` - Documentation changes
  - `Test:` - Test additions or updates

**Examples:**
```
Add: VPC peering CRUD operations
Fix: Region value display in get commands
Update: Standardize error handling across all commands
Docs: Add E2E testing documentation
```

## Building

### Build for Current Platform

```bash
go build -o acloud
```

### Build for All Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o acloud-linux-amd64

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o acloud-darwin-amd64

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o acloud-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o acloud-windows-amd64.exe
```

## Release Process

Releases are automated via GitHub Actions:

1. **Create and push a new tag:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **GitHub Actions automatically:**
   - Builds binaries for Linux, macOS (Intel & ARM), and Windows
   - Creates a GitHub release
   - Uploads binaries as release assets

## SDK Integration

The CLI uses the official Aruba Cloud Go SDK (`github.com/Arubacloud/sdk-go`).

### Client Initialization

The client is cached for performance. Use `GetArubaClient()` which:
- Loads credentials from `~/.acloud.yaml`
- Checks for `--debug` flag to enable logging
- Caches the client for reuse within the same execution

### SDK Methods

Access different resource types via:
- `client.FromNetwork()` - Network resources
- `client.FromStorage()` - Storage resources
- `client.FromManagement()` - Management resources
- `client.FromCompute()` - Compute resources
- And more...

See [SDK_INTEGRATION.md](SDK_INTEGRATION.md) for details.

## Common Patterns

### Standard CRUD Implementation

1. **Create:** Validate inputs → Build request → Call SDK → Display result
2. **List:** Call SDK → Format as table → Display
3. **Get:** Call SDK → Display detailed information (including Tags)
4. **Update:** Fetch current → Merge changes → Call SDK → Display result
5. **Delete:** Confirm (if needed) → Call SDK → Display confirmation

### Flag Definitions

```go
// Required flags
resourceCreateCmd.Flags().String("name", "", "Resource name (required)")
resourceCreateCmd.MarkFlagRequired("name")

// Optional flags
resourceCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")

// Global flags
rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")
```

### Table Output

```go
headers := []TableColumn{
    {Header: "NAME", Width: 30},
    {Header: "ID", Width: 26},
    {Header: "STATUS", Width: 15},
}
rows := [][]string{
    {name, id, status},
}
PrintTable(headers, rows)
```

## Getting Help

- **Open an issue** for bug reports or feature requests
- **Check existing issues** before creating a new one
- **Provide detailed information:**
  - CLI version
  - Operating system
  - Steps to reproduce
  - Expected vs actual behavior
  - Error messages (with `--debug` output if applicable)

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Follow the project's coding standards

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.

Thank you for contributing to acloud-cli! 🎉
