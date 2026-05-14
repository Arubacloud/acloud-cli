# CONVENTIONS.md — Code Conventions & Standards

## File Naming

- Category parent command: `cmd/<category>.go` (e.g., `cmd/storage.go`)
- Resource command: `cmd/<category>.<resource>.go` (e.g., `cmd/storage.blockstorage.go`)
- Tests: `cmd/<file>_test.go` and `cmd/<file>_test_enhanced.go` (extended fixtures)

---

## Flag Naming

All flags use **kebab-case** (not camelCase or snake_case).

**Standard flags reused across commands:**

| Flag | Short | Type | Scope | Purpose |
|------|-------|------|-------|---------|
| `--project-id` | — | string | all | Target project (always optional; context is fallback) |
| `--name` | — | string | create/update | Resource name (marked required on create) |
| `--region` | — | string | create | Region code (marked required on create) |
| `--tags` | — | string slice | create/update | Comma-separated tags |
| `--output` | `-o` | string | all | Output format: `table` (default), `table-json`, `table-yaml`, `json`, `yaml` |
| `--limit` | — | int | list | Maximum number of results to return |
| `--offset` | — | int | list | Number of results to skip (for pagination) |
| `--yes` | `-y` | bool | delete | Skip interactive confirmation prompt |
| `--dry-run` | — | bool | delete | Validate existence without deleting; prints `[dry-run] Would delete …` |

Flag descriptions follow this style:
- `"Project ID (uses context if not specified)"`
- `"Name for the block storage (required)"`
- `"Skip confirmation prompt"`

---

## Cobra Command Struct Fields

**Always set:** `Use`, `Short`, `RunE`

**Set when needed:**
- `Args` — use `cobra.ExactArgs(N)` or `cobra.NoArgs` for validation
- `Long` — set on parent/category commands and all create commands
- `Example` — set on all create commands (multi-line `backtick` string)
- `ValidArgsFunction` — set on get/update/delete commands that accept a resource ID

**Never set:** `Aliases`, `Deprecated`, `Hidden`, `PreRun`, `PostRun`

```go
var blockstorageGetCmd = &cobra.Command{
    Use:   "get [volume-id]",
    Short: "Get block storage details",
    Args:  cobra.ExactArgs(1),
    Run:   func(cmd *cobra.Command, args []string) { ... },
}

var blockstorageCmd = &cobra.Command{
    Use:  "blockstorage",
    Short: "Manage block storage",
    Long: `Perform CRUD operations on block storage in Aruba Cloud.`,
}
```

`Short` is imperative, verb-first: `"Create a new VPC"`, `"Get block storage details"`, `"Delete a VPC"`.
`Long` follows the pattern: `"Perform CRUD operations on <resource> in Aruba Cloud."`.

---

## Import Organization

```go
import (
    "context"          // stdlib: concurrency first
    "encoding/json"    // stdlib: alphabetical
    "fmt"
    "os"
    "strings"

    "github.com/Arubacloud/sdk-go/pkg/types"  // external: alphabetical
    "github.com/spf13/cobra"
)
```

Two groups: stdlib, then external. Each group is alphabetically ordered.

---

## Variable Naming in Handlers

```go
Run: func(cmd *cobra.Command, args []string) {
    // Always these names:
    projectID, err := GetProjectID(cmd)
    client, err    := GetArubaClient()
    ctx            := context.Background()

    // Flag values — match the flag name (kebab → camelCase):
    name, _         := cmd.Flags().GetString("name")
    region, _       := cmd.Flags().GetString("region")
    tags, _         := cmd.Flags().GetStringSlice("tags")
    confirm, _      := cmd.Flags().GetBool("yes")

    // API response:
    response, err := client.From...().Resource().Op(ctx, projectID, ...)

    // Table rows:
    var rows [][]string
}
```

Single resource extracted from response: use singular noun matching the resource (`volume`, `vpc`, `server`, `dbaas`), not pluralized.

---

## Argument Validation

- Use `cobra.ExactArgs(N)` in the command struct; do not re-validate inside `Run`.
- Validate flags **before** calling `GetArubaClient()` so the SDK is never initialized needlessly.
- Required flags that cannot be enforced with `MarkFlagRequired` (e.g., conditional) are checked manually with an early return:
  ```go
  if name == "" {
      fmt.Println("Error: --name is required")
      return
  }
  ```

---

## Pointer Dereferencing

SDK response fields are pointers. Always nil-check before use:

```go
name := ""
if resource.Metadata.Name != nil {
    name = *resource.Metadata.Name
}
```

Never dereference a response pointer without a nil guard.

---

## Adding a New Resource

1. Create `cmd/<category>.<resource>.go`. Define all subcommand vars at package level.
2. Register in `init()`:
   ```go
   func init() {
       parentCmd.AddCommand(resourceCmd)
       resourceCmd.AddCommand(resourceCreateCmd)
       resourceCmd.AddCommand(resourceGetCmd)
       resourceCmd.AddCommand(resourceUpdateCmd)
       resourceCmd.AddCommand(resourceDeleteCmd)
       resourceCmd.AddCommand(resourceListCmd)

       // Flags
       resourceCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
       resourceCreateCmd.Flags().String("name", "", "Name (required)")
       resourceCreateCmd.Flags().String("region", "", "Region code (required)")
       resourceCreateCmd.MarkFlagRequired("name")
       resourceCreateCmd.MarkFlagRequired("region")

       resourceDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

       // Completion
       resourceGetCmd.ValidArgsFunction    = completeResourceID
       resourceUpdateCmd.ValidArgsFunction = completeResourceID
       resourceDeleteCmd.ValidArgsFunction = completeResourceID
   }
   ```
3. Use `GetArubaClient()` and `GetProjectID(cmd)` from `cmd/root.go` — never read flags or initialize the SDK directly.
4. Implement `completeResourceID` following the shell completion pattern in `ARCHITECTURE.md`.
5. Register the parent category command in `cmd/<category>.go`'s `init()` if it doesn't already exist.

---

## Standard Command Bodies

### list
```go
projectID, err := GetProjectID(cmd)
if err != nil { fmt.Printf("Error: %v\n", err); return }

client, err := GetArubaClient()
if err != nil { fmt.Printf("Error initializing client: %v\n", err); return }

ctx := context.Background()
response, err := client.From<Svc>().<Resource>().List(ctx, projectID, nil)
if err != nil { fmt.Printf("Error listing <resources>: %v\n", err); return }

if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
    headers := []TableColumn{
        {Header: "NAME", Width: 30},
        {Header: "ID",   Width: 26},
        // ...
    }
    var rows [][]string
    for _, r := range response.Data.Values {
        rows = append(rows, []string{safePtrStr(r.Metadata.Name), safePtrStr(r.Metadata.ID), ...})
    }
    PrintOutput(response.Data, headers, rows)
} else {
    fmt.Println("No <resources> found")
}
```

### get
```go
resourceID := args[0]
// GetProjectID, GetArubaClient, context.Background() ...
resp, err := client.From<Svc>().<Resource>().Get(ctx, projectID, resourceID, nil)
if err != nil { ... return }
// check resp.IsError() ...

// Honour -o json / -o yaml before the human-formatted block
format := resolveOutputFormat()
if format == OutputFormatJSON || format == OutputFormatYAML {
    PrintOutput(resp.Data, nil, nil)
    return nil
}

fmt.Println("\n<Resource> Details:")
fmt.Println("===================")
if resp.Data.Metadata.ID != nil { fmt.Printf("ID:   %s\n", *resp.Data.Metadata.ID) }
// ...
```

### create
```go
// 1. GetProjectID
// 2. Extract flags; validate required ones early
// 3. GetArubaClient
// 4. Build types.<Resource>Request{} (nested struct)
// 5. Call .Create(ctx, projectID, request, nil)
// 6. Check err, then response.IsError()
// 7. PrintOutput with single-row result (pass response.Data as first arg)
PrintOutput(response.Data, headers, [][]string{row})
```

### update
```go
// 1. Get current resource via .Get() to preserve unmodified fields
// 2. Build request from current values
// 3. Overwrite only flags that were explicitly Changed:
if name != "" { updateReq.Metadata.Name = name }
if cmd.Flags().Changed("tags") { updateReq.Metadata.Tags = tags }
// 4. Call .Update(ctx, projectID, id, request, nil)
// 5. PrintOutput with single-row result
PrintOutput(response.Data, headers, [][]string{row})
```

### delete
```go
// 1. --dry-run: call Get to validate existence, print msgDryRun, return nil
dryRun, _ := cmd.Flags().GetBool("dry-run")
if dryRun {
    // GetProjectID, GetArubaClient, call .Get() to validate
    if err != nil { return fmt.Errorf("getting <resource>: %w", err) }
    fmt.Println(msgDryRun("<resource type>", id))
    return nil
}

// 2. Confirmation (before GetArubaClient in the non-dry-run path):
confirmed, err := confirmDelete("<resource type>", id)
if err != nil { return err }
if !confirmed { return nil }

// 3. GetProjectID, GetArubaClient, .Delete(ctx, projectID, id, nil)
// 4. Success message:
fmt.Println(msgDeleted("<resource type>", id))
```

`confirmDelete(resourceType, id string) (bool, error)` is a helper in `cmd/root.go` that detects non-interactive stdin and skips the prompt when `--yes` is set or when stdin is not a terminal. Use it — do not inline the prompt.

---

## Test Conventions

- Tests live in `package cmd` (same package as the code).
- Use `t.TempDir()` for isolated file paths; override `HOME` (or `USERPROFILE` on Windows) to redirect config/context files.
- Use `defer cleanup()` to restore environment variables.
- Skip live-API tests with `ACLOUD_TEST_SKIP_CLIENT=true`.
- Table-driven tests are preferred for multiple input/output cases.

### Test client — httptest harness (v0.2.0+)

Since SDK v0.2.0, wrapper types (`aruba.CloudServer`, `aruba.VPC`, …) carry unexported internal state that can only be populated by the SDK's own adapters. Hand-built fake structs cannot produce a hydrated wrapper with real `.ID()`/`.State()`/`List[T]` values.

Use the `arubaTestServer` harness in `cmd/mock_test.go` instead:

```go
func TestMyCommand(t *testing.T) {
    srv := newArubaTestServer(t)          // real aruba.Client pointed at httptest.Server
    srv.OnGet("/projects/p1/cloudServers", jsonResponse(200, types.CloudServerList{
        // ... fixture fields ...
    }))

    err := runCmd(srv.Client(), []string{"cloud-server", "list", "--project-id", "p1"})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

- Register routes with `OnGet`/`OnPost`/`OnPut`/`OnDelete`/`OnPatch` before calling `runCmd`.
- `jsonResponse(status, body)` marshals `body` (a `types.*Response` or `types.*List`) to JSON.
- `errorResponse(status, title, detail)` emits a `types.ErrorResponse` body so the SDK surfaces an `*aruba.HTTPError`.
- Unregistered routes cause `t.Errorf` (not a silent 404) — mis-keyed paths fail loudly.
- `runCmd`/`runCmdCapture`/`resetCmdFlags`/`strPtr` helpers are unchanged; pass `srv.Client()` where the old code passed a `newMockClient(...)`.
- Clear the client cache after each test (handled automatically by `runCmd` via `defer resetClientState()`).
