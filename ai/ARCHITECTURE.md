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

SDK v0.3.0 uses a fluent **wrapper layer**. `client.From<Svc>()` returns a typed
sub-client; each CRUD method returns a hydrated wrapper type or `*aruba.List[T]`
rather than raw `types.*Response` structs.

```go
// Top-level resources (e.g. project) — addressed by aruba.Ref:
client.FromProject().Get(ctx, projectRef(id))      // → (*aruba.Project, error)
client.FromProject().List(ctx, listOpts(cmd)...)   // → (*aruba.List[*aruba.Project], error)
client.FromProject().Create(ctx, proj)             // → (*aruba.Project, error)
client.FromProject().Update(ctx, proj)             // → (*aruba.Project, error)
client.FromProject().Delete(ctx, projectRef(id))   // → error

// Project-scoped resources — wrapper built with InProject(projectRef):
client.FromCompute().CloudServers().Get(ctx, ref)  // → (*aruba.CloudServer, error)
client.FromStorage().Volumes().List(ctx, ...)      // → (*aruba.List[*aruba.BlockStorage], error)

// Regional resources carry .InRegion(region) on the wrapper builder.
// Zonal resources additionally carry .InZone(zone).
// Resources that need WaitUntilActive call it on the returned wrapper.
```

**Ref addressing** — resources are addressed by `aruba.Ref` (an interface with `URI()
string`). Use `aruba.URI("/projects/"+id)` (wrapped by `projectRef` in `cmd/root.go`)
for top-level refs and chain `InProject(proj)` / `InVPC(vpc)` on wrappers for
scoped resources.

**Combined-URI Refs for project-scoped Get/Delete** — `List` and builder `InProject`
only need the project Ref (`projectRef(id)`). `Get` and `Delete` on project-scoped
resources require a single Ref encoding *both* the project and resource IDs. Declare
a file-local helper for each resource:

```go
func cloudServerRef(projectID, serverID string) aruba.Ref {
    return aruba.URI("/projects/" + projectID +
        "/providers/Aruba.Compute/cloudServers/" + serverID)
}
```

**Storage family sub-clients and Refs** — The four storage sub-clients are
`Volumes()`, `Snapshots()`, `Backups()`, `Restores()`. There is no `BlockStorages()`
alias. Each file declares a file-local Ref helper:

| Helper | Defined in | URI template |
|---|---|---|
| `volumeRef(pid, vid)` | `storage.blockstorage.go` | `/projects/<pid>/providers/Aruba.Storage/blockStorages/<vid>` |
| `snapshotRef(pid, sid)` | `storage.snapshot.go` | `/projects/<pid>/providers/Aruba.Storage/snapshots/<sid>` |
| `backupRef(pid, bid)` | `storage.backup.go` | `/projects/<pid>/providers/Aruba.Storage/backups/<bid>` |
| `restoreRef(pid, bid, rid)` | `storage.restore.go` | `/projects/<pid>/providers/Aruba.Storage/backups/<bid>/restores/<rid>` |

Note: path segments use **all-lowercase** (`blockStorages`, `snapshots`, `backups`,
`restores`) matching `internal/clients/storage/path.go`.

**Cross-family pre-validation** — `StorageBackup` Create fetches the source volume
(`Volumes().Get(ctx, volumeRef(...))`) before building the backup wrapper. `StorageRestore`
Create fetches both the parent backup (`Backups().Get`) **and** the target volume
(`Volumes().Get`) before building the restore wrapper.

**`StorageRestore` diverges from the project-scoped pattern** in two ways:
- Builder uses `IntoBackup(bk)` (not `IntoProject`) — the backup Ref carries the project ID implicitly.
- `Restores().List(ctx, backup Ref, ...)` is **backup-scoped**, not project-scoped.

```go
// Backup Create — cross-family pre-validation then builder:
vol, err := client.FromStorage().Volumes().Get(ctx, volumeRef(projectID, volumeID))
bk := aruba.NewStorageBackup().InProject(projectRef(projectID)).Named(name).
    InRegion(aruba.Region(region)).OfType(aruba.StorageBackupType(t)).FromVolume(vol)

// Restore Create — dual cross-family pre-validation:
bk, err  := client.FromStorage().Backups().Get(ctx, backupRef(projectID, backupID))
target, err := client.FromStorage().Volumes().Get(ctx, volumeRef(projectID, volumeID))
rs := aruba.NewStorageRestore().IntoBackup(bk).Named(name).InRegion(aruba.Region(region)).ToVolume(target)

// Restore List — backup-scoped:
list, err := client.FromStorage().Restores().List(ctx, backupRef(projectID, backupID), listOpts(cmd)...)
```

All four storage wrappers' `Raw()` returns the **full** typed response (e.g. `*types.StorageBackupResponse`). No `RawHTTP()` re-parse is needed, unlike `*aruba.Project`.

**Multi-level nested Refs (network family)** — VPC-scoped resources (Subnet,
SecurityGroup, VPCPeering) require 3-segment Refs; deeper resources (SecurityRule,
VPCPeeringRoute, VPNRoute) require 4-segment Refs encoding the full ancestry. The
path-segment casing matters: `subnets`, `securityGroups`, `securityRules`,
`loadBalancers`, `keyPairs` are camelCase; `vpcPeerings`, `vpcPeeringRoutes`,
`vpnTunnels`, `vpnRoutes`, `elasticIps` are also camelCase (matches
`internal/clients/network/path.go`, which is `internal/` and not importable).
Each file declares a file-local `<resource>Ref` helper and reuses the parent Ref
helper from the sibling file where it is defined once:

| Helper | Defined in | URI template |
|---|---|---|
| `vpcRef(pid, vid)` | `network.vpc.go` | `/projects/<pid>/providers/Aruba.Network/vpcs/<vid>` |
| `securityGroupRef(pid, vid, sgid)` | `network.securitygroup.go` | `…/vpcs/<vid>/securityGroups/<sgid>` |
| `vpcPeeringRef(pid, vid, peerid)` | `network.vpcpeering.go` | `…/vpcs/<vid>/vpcPeerings/<peerid>` |
| `vpnTunnelRef(pid, tid)` | `network.vpntunnel.go` | `…/vpnTunnels/<tid>` |

**Read-only sub-client** — `LoadBalancersClient` exposes only `List` and `Get`.
There is no `NewLoadBalancer()` factory and no Create/Update/Delete. The
`network.loadbalancer.go` command file reflects this: it has only `list` and `get`
subcommands.

**Deeply-nested VPN sub-builders** — `aruba.NewVPNTunnel()` composes four
independent sub-builders: `NewVPNIPConfig()`, `NewVPNIKE()`, `NewVPNESP()`,
`NewVPNPSK()`. Each is constructed separately and attached via
`WithIPConfig`/`WithIKESettings`/`WithESPSettings`/`WithPSKSettings`.

**VPN `fromResponse` does not rehydrate sub-builders** — `VPNTunnel.fromResponse()`
only populates top-level fields (`vpnType`, `vpnClientProtocol`, `billingPeriod`,
`peerClientPublicIP`). A naïve wrapper `Update` would drop the IKE/ESP/PSK/IPConfig
sub-builders from the PUT body. `network.vpntunnel.go` works around this with a
file-local `vpnTunnelReattachSettings(cur *aruba.VPNTunnel)` helper that
reconstructs the sub-builders from `cur.Raw().Properties.*` and re-attaches them
before calling `Update`.

**VPN crypto enums split per direction** — v0.3.0 provides per-direction types
(introduced in v0.2.0, carried forward): `aruba.IKEEncryption` / `aruba.ESPEncryption`,
`aruba.IKEHash` / `aruba.ESPHash`, `aruba.IKEDHGroup`, `aruba.IKEDPDAction`,
`aruba.ESPPFSGroup`. The CLI exposes seven separate `[]string` enum slices
(`vpnIKEEncryptionAlgorithms`, `vpnESPEncryptionAlgorithms`, etc.) and keys each
`--ike-*` / `--esp-*` flag to the correct family in the validation table.

**Operational methods on hydrated wrappers** — some operations (`PowerOn`, `PowerOff`,
`SetPassword`) are methods on `*aruba.<T>`, not on the sub-client. They require a
prior hydrating `Get`; calling them on a freshly-constructed `New<T>()` wrapper will
fail. `PowerOn`/`PowerOff` re-hydrate the wrapper from the response; `SetPassword`
does not — render from the pre-`Get` result:

```go
cs, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(pid, id))
if err != nil { return fmt.Errorf("powering on cloud server: %w", apiErrFromV2(err)) }
if err := cs.PowerOn(ctx); err != nil {
    return fmt.Errorf("powering on cloud server: %w", apiErrFromV2(err))
}
// cs is now re-hydrated; render from cs.Raw()
```

**Error handling** — non-2xx responses surface as `*aruba.HTTPError` in the error
return. There is no `response.IsError()` check. Use `apiErrFromV2(err)` (in
`cmd/root.go`) to format HTTP errors while passing transport errors through. Always
wrap with a verb prefix:

```go
result, err := client.From<Svc>().<Resource>().Op(ctx, ...)
if err != nil {
    return fmt.Errorf("<verb> <resource>: %w", apiErrFromV2(err))
}
```

**Rendering** — wrapper types (`*aruba.<T>`) carry only unexported fields and are not
directly JSON-marshalable via `encoding/json`. For table columns the wrapper exposes
(`.ID()`, `.Name()`, `.State()`, `.CreatedAt()`, …) use the accessors directly.

`PrintOutput` delegates full-payload rendering (`-o json` / `-o yaml`) through three
dispatch branches in `printJSON` / `printYAML`:

1. **`rawMarshaler`** (preferred) — `RawJSON()` / `RawYAML()` called on the wrapper.
   All wrapper types except `*aruba.Project` satisfy this interface in v0.3.0.
   **Always pass the SDK wrapper (not `.Raw()`) as the first arg to `PrintOutput`.**
2. **`rawHTTPer`** (legacy, `*aruba.Project` only) — falls back to `RawHTTP()` for
   re-parsing the raw HTTP response body. Used only in `management.project.go` until
   sdk-go ships `RawJSON()/RawYAML()` on `*aruba.Project`.
3. **`json.MarshalIndent`** (anonymous structs) — delete-confirmation result rows use
   small anonymous structs (`struct{ID, Status string}{...}`) that satisfy neither
   interface. These fall through to `json.MarshalIndent`. Intentional — delete handlers
   never need SDK-shaped full-payload output.

This keeps the **single-import principle**: only `github.com/Arubacloud/sdk-go/pkg/aruba`
is imported in non-test `cmd/` files.

For list `-o json`/`-o yaml`, `*aruba.List[T]` also satisfies `rawMarshaler`; no
separate `<resource>ListPayload` helper is needed for the output call itself, though
file-local helpers may still be used for extracting typed sub-fields.

**`ctx`** — use `newCtx()` (30-second timeout, in `cmd/root.go`) for all SDK calls.
Completion functions that run interactively may keep `context.Background()`.

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
    list, err := client.FromStorage().Volumes().List(ctx, projectRef(projectID))
    if err != nil { return nil, cobra.ShellCompDirectiveNoFileComp }

    var completions []string
    if list != nil {
        for _, v := range list.Items() {
            id := v.ID()
            if toComplete == "" || strings.HasPrefix(id, toComplete) {
                completions = append(completions, fmt.Sprintf("%s\t%s", id, v.Name()))
            }
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

---

## Database Family

`client.FromDatabase()` returns a `DatabaseClient` with four sub-clients:

| Sub-client | Method | Resources |
|---|---|---|
| `DBaaS()` | `FromDatabase().DBaaS()` | DBaaS instances (project-scoped) |
| `Databases()` | `FromDatabase().Databases()` | Databases inside a DBaaS (dbaas-scoped) |
| `Backups()` | `FromDatabase().Backups()` | DBaaS Backups (project-scoped) |
| `Users()` | `FromDatabase().Users()` | Users inside a DBaaS (dbaas-scoped) |

**Note:** There is no `BlockStorages()` alias — `Volumes()` is the correct name for storage. Similarly, `FromDatabase().Backups()` (database backups) is distinct from `FromStorage().Backups()` (storage backups).

### Two wrapper families

**Family A** (DBaaS, DBaaSBackup) — standard `Metadata/Properties/Status` envelope. Uses `InProject(projectRef(projectID))` for the builder.

**Family B** (Database, User) — flat request, name/username IS the path identifier, no `Metadata.ID` in responses. Uses `IntoDBaaS(dbaasRef(projectID, dbaasID))` for the builder. `Database.ID()` returns the name; `User.ID()` returns the username.

### Ref helpers (declared per file)

```go
func dbaasRef(projectID, dbaasID string) aruba.Ref {
    return aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID)
}
func databaseRef(projectID, dbaasID, name string) aruba.Ref {
    return aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/databases/" + name)
}
func userRef(projectID, dbaasID, username string) aruba.Ref {
    return aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/users/" + username)
}
func databaseBackupRef(projectID, backupID string) aruba.Ref {
    return aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/backups/" + backupID)
}
```

### List scoping

- **DBaaS / Backup List** — project-scoped: `List(ctx, projectRef(projectID), listOpts(cmd)...)`
- **Database / User List** — dbaas-scoped: `List(ctx, dbaasRef(projectID, dbaasID), listOpts(cmd)...)`

### DBaaS Update — round-trip

`DBaaS.fromResponse()` back-populates `engine`, `flavor`, `sizeGB`, `billingPeriod`, and networking refs automatically. The update pattern is just:

```go
current, err := client.FromDatabase().DBaaS().Get(ctx, dbaasRef(projectID, dbaasID))
if name != "" { current.Named(name) }
if cmd.Flags().Changed("size-gb") { current.WithSizeGB(sizeGB) }
updated, err := client.FromDatabase().DBaaS().Update(ctx, current)
```

No manual field reconstruction is needed — all unchanged fields survive the round-trip.

### Backup Create — no pre-validation Gets

`FromDBaaS(Ref)` and `FromDatabase(Ref)` on the `DBaaSBackup` builder accept any `aruba.Ref` and call `.URI()` on them. Pass the constructed Refs directly — no pre-validation Gets needed:

```go
bk := aruba.NewDBaaSBackup().
    InProject(projectRef(projectID)).
    Named(name).
    InRegion(aruba.Region(region)).
    FromDBaaS(dbaasRef(projectID, dbaasID)).
    FromDatabase(databaseRef(projectID, dbaasID, databaseName)).
    BilledBy(aruba.BillingPeriod(billingPeriod))
created, err := client.FromDatabase().Backups().Create(ctx, bk)
```

The old v0.1.x code performed 2 cross-family Gets just to obtain URIs (and used a malformed URI for the database ref). The v0.2.0 builder eliminated both round-trips; v0.3.0 renames the setter to `BilledBy`.

### BackupsClient — no Update

`BackupsClient` exposes only `List / Get / Create / Delete`. There is no Update method. The `backupUpdateCmd` stub in `database.backup.go` returns an informational error without calling the SDK.

### Identity accessors

- DBaaS: `d.DBaaSID()` (equivalent to `d.ID()`)
- DBaaSBackup: `b.DBaaSBackupID()` (use this, not `b.ID()`)
- Database: `db.DatabaseID()` (returns the name, which is the path identifier)
- User: `u.Username()` (returns the username, which is the path identifier)

---

## Container Family

`client.FromContainer()` returns a `ContainerClient` with two sub-clients:

| Sub-client | Method | Resource |
|---|---|---|
| `KaaS()` | `FromContainer().KaaS()` | Kubernetes-as-a-Service clusters (project-scoped) |
| `ContainerRegistry()` | `FromContainer().ContainerRegistry()` | Container registries (project-scoped) |

### Ref helpers (declared per file)

```go
func kaasRef(projectID, kaasID string) aruba.Ref {
    return aruba.URI("/projects/" + projectID + "/providers/Aruba.Container/kaas/" + kaasID)
}
func containerRegistryRef(projectID, registryID string) aruba.Ref {
    return aruba.URI("/projects/" + projectID + "/providers/Aruba.Container/registries/" + registryID)
}
```

**Path note**: ContainerRegistry uses `registries` (not `containerregistries`) — matches `internal/clients/container/path.go`.

### Identity accessors

- KaaS: `k.KaaSID()` (use this, not `k.ID()`)
- ContainerRegistry: `r.ContainerRegistryID()` (equivalent to `r.ID()`)

### `DownloadKubeconfig` — method on `*KaaS`, not on `KaaSClient`

`KaaSClient` interface exposes only `List / Get / Create / Update / Delete`. The kubeconfig download requires a **two-step** approach:

```go
got, err := client.FromContainer().KaaS().Get(ctx, kaasRef(projectID, kaasID))
if err != nil { return fmt.Errorf("getting KaaS cluster: %w", apiErrFromV2(err)) }
kubeconfigBytes, err := got.DownloadKubeconfig(ctx)
```

Wire path: `GET /projects/{p}/providers/Aruba.Container/kaas/{id}/download` → returns `types.KaaSKubeconfigResponse{Name, Content}`. `DownloadKubeconfig` returns `[]byte(resp.Data.Content)`. The content is base64-encoded YAML; decode before writing to disk.

### `WithSecurityGroup` on KaaS requires `*aruba.SecurityGroup`

`KaaS.WithSecurityGroup(sg Ref)` performs a type assertion `sg.(*SecurityGroup)` internally and fails at runtime if the assertion fails. Do **not** pass `aruba.URI(...)`. Always construct a minimal wrapper:

```go
sg := aruba.NewSecurityGroup().Named(securityGroupName)
k.WithSecurityGroup(sg)
```

The KaaS API stores only the SG name, not a URI. This diverges from `ContainerRegistry.WithSecurityGroup(sg Ref)` which accepts any `Ref` (plain `aruba.URI(...)`).

### NodePool sub-builder

```go
np := aruba.NewNodePool().
    Named(poolName).
    WithCount(n).
    OfInstance(aruba.NodePoolInstance(instance)).
    InZone(aruba.Zone(zone))
if autoscaling { np.WithAutoscaling(minCount, maxCount) }
k.AddNodePool(np)
```

`AddNodePool` **appends** to pools already set by `fromResponse()` on an existing wrapper. There is no `ClearNodePools` or `ReplaceNodePools`. When updating via Get→mutate→Update, adding a pool appends rather than replaces.

### `ContainerRegistry.OfSize` — replaces `ConcurrentUsers` string

The old v0.1.x `ConcurrentUsers *string` field is replaced by a typed enum:

```go
r.OfSize(aruba.ContainerRegistrySizeFlavor(concurrentUsers)) // "Small", "Medium", "HighPerf"
```

Use `aruba.ContainerRegistrySizeFlavorSmall`, `ContainerRegistrySizeFlavorMedium`, `ContainerRegistrySizeFlavorHighPerf` constants.

### `ContainerRegistry.WithElasticIP` — replaces `PublicIp.URI`

```go
r.WithElasticIP(aruba.URI(publicIPURI))
```

The response field is still `resource.Properties.PublicIp.URI` (unchanged wire format).

### KubernetesVersion constants

`KubernetesVersion1313` was removed in v0.2.0 and is absent in v0.3.0. Available: `aruba.KubernetesVersion1323`, `KubernetesVersion1332`, `KubernetesVersion1341`. The CLI accepts any string via `aruba.KubernetesVersion(version)` — validation is left to the API.

## Schedule Family

`cmd/schedule.job.go` — one resource, managed via `client.FromSchedule().Jobs()`.

### Ref helpers

```go
func jobRef(projectID, jobID string) aruba.Ref {
    return aruba.URI("/projects/" + projectID + "/providers/Aruba.Schedule/jobs/" + jobID)
}
func jobFromRaw(j *aruba.Job) *types.JobResponse { return j.Raw() }
func jobListPayload(l *aruba.List[*aruba.Job]) any {
    if r, ok := l.Raw().(*types.Response[types.JobList]); ok && r != nil {
        return r.Data
    }
    return nil
}
```

### Identity accessors

`j.JobID()` (not `j.ID()`), `j.Name()`, `j.Region()` (→ `aruba.Region`, cast with `string()`), `j.JobType()` (→ `types.JobType`), `j.Enabled()` bool, `j.State()` string.

### Schedule modes — mutually exclusive

```go
// One-shot:
j.OneShotAt(t)

// Recurring:
j.WithCron(cronExpr).RecurringUntil(endTime)
```

Call `j.Err()` after builder setup to surface any validation errors before the Create call.

### `Enabled()`/`Disabled()` setters (v0.3.0)

In v0.3.0 the boolean `WithEnabled(bool)` setter is replaced by two explicit methods:
`Enabled()` and `Disabled()`. Use `j.Enabled()` or `j.Disabled()` to set the job
state — the `omitempty` wire issue that affected the old `WithEnabled(false)` call
is resolved upstream in v0.3.0.

## Security (KMS) Family

`cmd/security.kms.go` — migrated to v0.2.0 in #108 (alongside schedule).

### Ref helpers

```go
func kmsRef(projectID, kmsID string) aruba.Ref {
    return aruba.URI("/projects/" + projectID + "/providers/Aruba.Security/kms/" + kmsID)
}
```

### Identity accessor

`k.KMSID()` (not `k.ID()`). Other accessors: `k.Name()`, `k.Region()`, `k.State()`, `k.Raw()` → `*types.KMSResponse`.
