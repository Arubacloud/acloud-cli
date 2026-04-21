# ARCHITECTURE.md — Code Architecture & Design Patterns

## Execution Flow

`main.go` → `cmd.Execute()` → root Cobra command → subcommand handler.

Global flags registered on `rootCmd` (available on every command):
- `--debug, -d` — enables HTTP request/response logging to stderr (microsecond-precision, stderr output)
- `--project-id` — target project (falls back to active context if omitted)

Commands use `Run` (not `RunE`). Errors are printed with `fmt.Printf` and the handler returns early — no exit codes are used in resource commands.

---

## Client Initialization & Caching

`GetArubaClient()` in `cmd/root.go` returns a cached `aruba.Client`. The cache is package-level state protected by `sync.Mutex`.

**Cache invalidation** — the cached instance is reused only when ALL of the following match the prior call:
```
cachedClientID == config.ClientID
cachedSecret   == config.ClientSecret
cachedDebug    == debugEnabled
cachedBaseURL  == baseURL
cachedTokenIssuer == tokenIssuerURL
```

**Client construction** (when cache misses):
```go
options := aruba.DefaultOptions()
// WithNativeLogger() added if --debug is set
aruba.NewClient(options)
```

**Defaults applied inside `GetArubaClient()`** (not in `LoadConfig`):
- `BaseURL` → `https://api.arubacloud.com` if empty
- `TokenIssuerURL` → predefined Aruba identity URL if empty

If `LoadConfig()` fails (missing `~/.acloud.yaml`), the error is wrapped:
> `"failed to load configuration: %w. Please run 'acloud config set' to configure credentials"`

---

## SDK Call Pattern

All resource operations follow the builder pattern through the client:

```go
client.FromStorage().Volumes().Create(ctx, projectID, request, nil)
client.FromCompute().CloudServers().Get(ctx, projectID, id, nil)
client.FromNetwork().VPCs().List(ctx, projectID, nil)
```

- The 4th argument (`options`) is always `nil` in current commands.
- `ctx` is always `context.Background()`, declared inline in the handler.
- The response carries `.IsError()`, `.StatusCode`, `.Error.Title`, `.Error.Detail`, and `.Data`.

**Response error check pattern:**
```go
if response != nil && response.IsError() && response.Error != nil {
    fmt.Printf("Failed - Status: %d\n", response.StatusCode)
    if response.Error.Title != nil {
        fmt.Printf("Error: %s\n", *response.Error.Title)
    }
    if response.Error.Detail != nil {
        fmt.Printf("Detail: %s\n", *response.Error.Detail)
    }
    return
}
```

---

## Project ID Resolution

`GetProjectID(cmd)` in `cmd/root.go` resolves in order:
1. `--project-id` flag value (if non-empty)
2. `GetCurrentProjectID()` → reads `CurrentContext` from `~/.acloud-context.yaml`, returns its `ProjectID`
3. Returns error: `"project ID not specified. Use --project-id flag or set a context with 'acloud context use <name>'"`

---

## Config Subsystem

**File:** `~/.acloud.yaml` (permissions `0600` on write)

**Struct:**
```go
type Config struct {
    ClientID       string `yaml:"clientId"`
    ClientSecret   string `yaml:"clientSecret"`
    BaseURL        string `yaml:"baseUrl,omitempty"`
    TokenIssuerURL string `yaml:"tokenIssuerUrl,omitempty"`
}
```

- Missing file → `LoadConfig()` returns an error (no graceful degradation to empty config).
- Partial config → zero values for missing fields; defaults are applied later in `GetArubaClient()`.
- `SaveConfig()` marshals to YAML and writes with `os.WriteFile(..., 0600)`.

---

## Context Subsystem

**File:** `~/.acloud-context.yaml`

**Struct:**
```go
type Context struct {
    CurrentContext string             `yaml:"current-context"`
    Contexts       map[string]CtxInfo `yaml:"contexts"`
}
type CtxInfo struct {
    ProjectID string `yaml:"project-id"`
    Name      string `yaml:"name,omitempty"`
}
```

**Command behaviours:**
- `context set <name> --project-id <id>` — creates/updates a named context but does **not** switch to it automatically.
- `context use <name>` — validates the name exists, then sets `CurrentContext`.
- `context delete <name>` — removes the context; clears `CurrentContext` if it was the active one.
- `context list` — prints all contexts with `*` marking the current one.

---

## Command Registration

Commands are registered in `init()` functions inside each file. Resource files register with the parent defined in the category base file:

```go
// cmd/storage.go
func init() { rootCmd.AddCommand(storageCmd) }

// cmd/storage.blockstorage.go
func init() {
    storageCmd.AddCommand(blockstorageCmd)
    blockstorageCmd.AddCommand(blockstorageListCmd)
    blockstorageCmd.AddCommand(blockstorageGetCmd)
    // ...
}
```

No `PreRun`, `PostRun`, or middleware hooks exist anywhere in the codebase.

---

## Output Patterns

The global `--output / -o` flag (declared once on `rootCmd`, inherited by every command) controls the output format. Five canonical modes are supported:

| Mode | Aliases | Shape |
|------|---------|-------|
| `table` | `std`, `standard` | Fixed-width text table (default) |
| `table-json` | `std-json`, `standard-json` | JSON array of flat snake_case row objects |
| `table-yaml` | `std-yaml`, `standard-yaml` | YAML sequence of flat snake_case mappings |
| `json` | — | Full SDK response object as indented JSON |
| `yaml` | — | Full SDK response object as YAML |

### Unified output (`PrintOutput`)

`PrintOutput(obj any, headers []TableColumn, rows [][]string)` in `cmd/root.go`:
- Pure dispatcher: calls `resolveOutputFormat()` and delegates to one of five private functions.
- `table` / `table-json` / `table-yaml` branches use `headers` + `rows` (flat, pre-formatted strings).
- `json` / `yaml` branches use `obj` (the full SDK response); `obj=nil` emits `{}`.

| Private function | Format | Notes |
|---|---|---|
| `printJSON(obj)` | `json` | `json.MarshalIndent`; nil → `{}` |
| `printYAML(obj)` | `yaml` | JSON → `interface{}` → `yaml.Encoder` round-trip (SDK structs have `json` tags but no `yaml` tags; keeps camelCase keys) |
| `printTableJSON(headers, rows)` | `table-json` | Hand-built ordered JSON array — `json.Marshal` of a map loses column order |
| `printTableYAML(headers, rows)` | `table-yaml` | `yaml.Node` sequence — preserves column order |
| `printTable(headers, rows)` | `table` (default) | Fixed-width `%-Ns` printf; values longer than Width truncated with `"..."` |

`PrintTable(headers, rows)` is a thin shim around `PrintOutput(nil, headers, rows)` kept for backward compatibility; it will be removed in a follow-up.

```go
headers := []TableColumn{
    {Header: "NAME",    Width: 30},
    {Header: "ID",      Width: 26},
    {Header: "STATUS",  Width: 15},
}
// list: pass full response so -o json / -o yaml emit the SDK envelope
PrintOutput(response.Data, headers, rows)
```

### `get` command output

For `table` mode, detail views use `fmt.Printf` with labeled fields:
```
Resource Details:
=================
ID:    <value>
Name:  <value>
```

For `json` / `yaml` modes, get commands have an early-return that emits the full SDK data object:
```go
format := resolveOutputFormat()
if format == OutputFormatJSON || format == OutputFormatYAML {
    PrintOutput(resp.Data, nil, nil)
    return nil
}
// … fmt.Printf labeled-field block for table modes …
```

---

## Destructive Operation Pattern (Delete)

Every delete command supports two safety mechanisms:

1. **`--dry-run`** — calls `Get` to validate existence, prints `msgDryRun(kind, id)`, and returns without deleting.
2. **`--yes` / `-y`** — skips the interactive confirmation prompt (also skipped when stdin is not a terminal).

```go
// --dry-run: validate, report, return
dryRun, _ := cmd.Flags().GetBool("dry-run")
if dryRun {
    // GetProjectID, GetArubaClient, call .Get() ...
    fmt.Println(msgDryRun("<resource type>", id))
    return nil
}

// Confirmation (uses confirmDelete helper from root.go):
confirmed, err := confirmDelete("<resource type>", id)
if err != nil { return err }
if !confirmed { return nil }

// proceed with SDK delete call
```

Flags registered in `init()`:
```go
resourceDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
resourceDeleteCmd.Flags().Bool("dry-run", false, "Validate existence without deleting")
```

`confirmDelete(resourceType, id string) (bool, error)` in `cmd/root.go` detects non-interactive stdin, respects `--yes`, and prompts when appropriate — never inline the prompt.

---

## Shell Completion

Completion functions live in the same file as the command they complete. Pattern:

```go
func completeBlockStorageID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    projectID, err := GetProjectID(cmd)
    if err != nil { return nil, cobra.ShellCompDirectiveNoFileComp }

    client, err := GetArubaClient()
    if err != nil { return nil, cobra.ShellCompDirectiveNoFileComp }

    ctx := context.Background()
    response, err := client.FromStorage().Volumes().List(ctx, projectID, nil)
    if err != nil { return nil, cobra.ShellCompDirectiveNoFileComp }

    var completions []string
    for _, v := range response.Data.Values {
        if v.Metadata.ID != nil && strings.HasPrefix(*v.Metadata.ID, toComplete) {
            completions = append(completions, fmt.Sprintf("%s\t%s", *v.Metadata.ID, *v.Metadata.Name))
        }
    }
    return completions, cobra.ShellCompDirectiveNoFileComp
}

// Registered in init():
blockstorageGetCmd.ValidArgsFunction = completeBlockStorageID
```

Always returns `cobra.ShellCompDirectiveNoFileComp`. On any error, returns `nil, cobra.ShellCompDirectiveNoFileComp` (fail silently).

---

## Error Handling Rules

- Resource commands use `Run` (not `RunE`). Errors are printed and the function returns.
- `os.Exit` is called only in `cmd/root.go` (if `Execute()` fails) and `cmd/config.go` (hard validation errors during initial setup). Never in resource commands.
- SDK call errors from `err != nil` and API-level errors from `response.IsError()` are handled separately (see SDK Call Pattern above).

---

## Request Building

SDK requests use nested struct composition with pointer-valued optional fields:

```go
types.BlockStorageRequest{
    Metadata: types.RegionalResourceMetadataRequest{
        ResourceMetadataRequest: types.ResourceMetadataRequest{
            Name: name,
            Tags: tags,
        },
        Location: types.LocationRequest{Value: region},
    },
    Properties: types.BlockStoragePropertiesRequest{
        SizeGB:        size,
        BillingPeriod: billingPeriod,
        Type:          types.BlockStorageType(volumeType),
    },
}
```

**Update pattern** — fetch the current resource first to preserve values not being updated, then overwrite changed fields:
```go
getResp, _ := client.From...().Resource().Get(ctx, projectID, id, nil)
current := getResp.Data
updateReq := buildRequestFrom(current)    // preserve current values
if name != "" { updateReq.Metadata.Name = name }
if cmd.Flags().Changed("tags") { updateReq.Metadata.Tags = tags }
```
