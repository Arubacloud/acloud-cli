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

**Cross-resource reference flags** — when a resource requires a reference to another
resource, accept the **resource ID** (not a full URI) via a `--<resource>-id` flag,
then build the Ref with the appropriate SDK helper:

| Flag | SDK Ref helper |
|------|---------------|
| `--vpc-id` | `aruba.VPCRef(projectID, vpcID)` |
| `--subnet-id` | `aruba.SubnetRef(projectID, vpcID, subnetID)` (requires `--vpc-id`) |
| `--security-group-id` | `aruba.SecurityGroupRef(projectID, vpcID, sgID)` (requires `--vpc-id`) |
| `--elastic-ip-id` / `--public-ip-id` | `aruba.ElasticIPRef(projectID, eipID)` |
| `--volume-id` / `--boot-disk-id` | `volumeRef(projectID, volID)` (file-local helper) |
| `--snapshot-id` | `snapshotRef(projectID, snapID)` (file-local helper) |
| `--keypair-id` | `keypairRef(projectID, name)` (file-local helper) |

Do not expose URI-shaped flags (`--vpc-uri`, `--subnet-uri`, etc.) — callers know
resource IDs, not full URI paths.

---

## Cobra Command Struct Fields

**Always set:** `Use`, `Short`, `RunE` (points to the named `<Family><Resource><Action>Run` wiring function — never an anonymous closure)

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
    RunE:  StorageBlockStorageGetRun, // named wiring function, never anonymous
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

    "github.com/Arubacloud/sdk-go/pkg/aruba"  // external: alphabetical
    "github.com/spf13/cobra"
)
```

Two groups: stdlib, then external. Each group is alphabetically ordered.

**Single-import policy** — non-test `cmd/` files must only import
`github.com/Arubacloud/sdk-go/pkg/aruba`. Importing `pkg/types` directly from
non-test cmd files is not permitted. The one documented exception (TD-033) is
`container.kaas.go`, which references `types.KaaSAPIServerAccessProfilePropertiesRequest`
until sdk-go provides an `aruba`-level constructor for that type. All `pkg/types`
references in non-test cmd files must be annotated with `// TECH_DEBT: TD-033`.

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

Follow the Args / Operation / Run decomposition for every handler (see section below).

1. Create `cmd/<category>.<resource>.go`. Define all subcommand vars at package level.
2. For each action (Create / Get / Update / Delete / List / …):
   a. Define `<Family><Resource><Action>Args` struct with typed fields.
   b. Implement `ParseFromCobraCommand(cmd)` — reads flags, resolves `ProjectID` via
      `GetProjectID(cmd)`, casts strings to SDK enum types.
   c. Implement `Validate()` — pure function, no SDK, no I/O. Checks required fields,
      string lengths, `slices.Contains` against `valid*` slices in `cmd/enums.go`.
   d. Implement `New<...>ArgsFromCobraCommand(cmd)` constructor that calls Parse then
      Validate, wrapping errors with `ErrParsingFailed` / `ErrValidationFailed`.
   e. Implement `<Family><Resource><Action>(ctx, client aruba.Client, args <...>Args) error`
      — the pure operation (SDK call + rendering). Must not call `GetArubaClient`,
      `GetProjectID`, or `cmd.Flags()`.
   f. Implement `<Family><Resource><Action>Run(cmd, _ []string) error` — thin wiring:
      `New…Args` → `GetArubaClient` → `newCtx` → operation. Delete adds `confirmDelete`
      and `--dry-run` here.
3. Register in `init()`:
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
       resourceDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

       // Completion
       resourceGetCmd.ValidArgsFunction    = completeResourceID
       resourceUpdateCmd.ValidArgsFunction = completeResourceID
       resourceDeleteCmd.ValidArgsFunction = completeResourceID
   }
   ```
4. Wire `RunE` to the named wiring function: `RunE: <Family><Resource><Action>Run`.
5. Implement `completeResourceID` following the shell completion pattern in `ARCHITECTURE.md`.
6. Register the parent category command in `cmd/<category>.go`'s `init()` if it doesn't already exist.

---

## Args / Operation / Run Decomposition

Every Cobra `RunE` handler is decomposed into **four named symbols** per action. This
makes the operation function testable with a mocked SDK client and Validate() testable
as a pure Go function — no rootCmd.Execute() required.

SDK v1.0.0 uses a fluent wrapper layer. `client.From<Svc>().<Resource>()` returns a
typed client whose CRUD methods take/return hydrated wrapper types (`*aruba.<T>`,
`*aruba.List[*aruba.<T>]`) rather than raw request/response structs. Non-2xx
responses surface as `*aruba.HTTPError` in the error return — there is no separate
`response.IsError()` check. Use `apiErrFromV2(err)` to format HTTP errors; wrap with
a verb prefix for all error sites.

**Wrapper note:** `*aruba.<T>` wrapper types carry only unexported fields and are not
directly JSON-marshalable via `encoding/json`. For single-resource rendering and
`-o json`/`-o yaml` payloads, `PrintOutput` delegates to the SDK's `RawJSON()`/
`RawYAML()` methods (via the local `rawMarshaler` interface). All wrapper types
including `*aruba.Project` satisfy `rawMarshaler` in sdk-go v1.0.0. The old
`rawHTTPer` fallback was deleted (TD-032 / TD-036).
For project-scoped resources, build the create wrapper with `.InProject(projectRef(projectID))`
and use a file-local `<resource>Ref(projectID, id)` helper (encoding both project +
resource IDs in the URI) for `Get` and `Delete`.

**Multi-segment ancestry Refs** — For nested resources (e.g. `Subnet` inside `VPC`,
`SecurityRule` inside `SecurityGroup` inside `VPC`), each file declares a file-local
`<resource>Ref(...)` helper that hand-builds the full URI string including all
ancestor IDs. Parent Ref helpers (`vpcRef`, `securityGroupRef`, `vpcPeeringRef`,
`vpnTunnelRef`) are defined once in their respective files and imported by sibling
files — never redefined. Path-segment casing must match the API exactly (see
`ARCHITECTURE.md` for the casing table).

**Read-only resources** — Resources that the API exposes as read-only (e.g.
`LoadBalancer`) have only `list` and `get` subcommands; no `create`, `update`, or
`delete`. Do not add a `NewLoadBalancer()` builder call or Create/Update/Delete
command vars.

### Naming convention

`<Family>` is the Cobra category prefix: `Compute`, `Storage`, `Network`, `Database`,
`Container`, `Security`, `Schedule`, `Management`, `Config`, `Context`. `<Resource>` is
the resource name in PascalCase (`CloudServer`, `BlockStorage`, `VPC`, …). `<Action>` is
the verb: `Create`, `Get`, `List`, `Update`, `Delete`, plus any non-CRUD verbs already in
the tree (`PowerOn`, `PowerOff`, `SetPassword`, `Connect`, `Set`, `Use`, `Current`, `Show`).

### 1. Args struct

```go
type ComputeCloudServerCreateArgs struct {
    ProjectID        string
    Name             string
    Region           aruba.Region           // typed enum, not string
    Zone             aruba.Zone
    Flavor           aruba.CloudServerFlavor
    BootDiskID       string
    VPCID            string
    SubnetIDs        []string               // IDs; operation builds aruba.Ref
    SecurityGroupIDs []string
    ElasticIPID      string
    KeypairID        string
    Tags             []string
    BillingPeriod    aruba.BillingPeriod
    UserDataFile     string
}
```

Rules:
- Enum fields use the SDK type, not `string`. The string→type cast happens in
  `ParseFromCobraCommand`; value-set membership is checked in `Validate`.
- `ProjectID` is the resolved value from `GetProjectID(cmd)`, not the raw flag.
- Cross-resource refs stay as ID strings; the operation function builds `aruba.Ref`.
- Delete args include `DryRun bool` and `SkipConfirm bool`; the Run wiring reads them
  and short-circuits before calling the operation.

### 2. Constructor + ParseFromCobraCommand + Validate

```go
// Errors sentinel package-vars are in cmd/args.go.
func NewComputeCloudServerCreateArgsFromCobraCommand(cmd *cobra.Command) (*ComputeCloudServerCreateArgs, error) {
    args := &ComputeCloudServerCreateArgs{}
    if err := args.ParseFromCobraCommand(cmd); err != nil {
        return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
    }
    if err := args.Validate(); err != nil {
        return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
    }
    return args, nil
}

func (a *ComputeCloudServerCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
    var errs []error
    var err error
    if a.ProjectID, err = GetProjectID(cmd); err != nil { errs = append(errs, err) }
    if a.Name, err = cmd.Flags().GetString("name"); err != nil { errs = append(errs, err) }
    if s, err := cmd.Flags().GetString("region"); err == nil {
        a.Region = aruba.Region(s) // cast here; value-set check in Validate
    } else { errs = append(errs, err) }
    // ... remaining fields ...
    return errors.Join(errs...)
}

func (a *ComputeCloudServerCreateArgs) Validate() error {
    var errs []error
    if len(a.Name) < 3 { errs = append(errs, errors.New("--name must be at least 3 characters")) }
    if len(a.Name) > 64 { errs = append(errs, errors.New("--name must be at most 64 characters")) }
    if !slices.Contains(validRegions, a.Region) {
        errs = append(errs, fmt.Errorf("--region %q: must be one of %v", a.Region, validRegions))
    }
    if !slices.Contains(validBillingPeriods, a.BillingPeriod) {
        errs = append(errs, fmt.Errorf("--billing-period %q: must be one of %v", a.BillingPeriod, validBillingPeriods))
    }
    if a.VPCID == "" { errs = append(errs, errors.New("--vpc-id is required")) }
    if len(a.SubnetIDs) == 0 { errs = append(errs, errors.New("--subnet-id requires at least one value")) }
    // ...
    return errors.Join(errs...)
}
```

**Enum-validation rule:** For enums where the SDK exposes no fixed set (e.g.
`aruba.KubernetesVersion`), `Validate()` checks `!= ""` only — never
`slices.Contains`. Annotate inline. Valid-value slices for cross-resource enums are
declared in `cmd/enums.go`; single-resource enums stay in the resource file.

### 3. Operation function

```go
func ComputeCloudServerCreate(ctx context.Context, client aruba.Client, args ComputeCloudServerCreateArgs) error {
    subnetRefs := make([]aruba.Ref, len(args.SubnetIDs))
    for i, s := range args.SubnetIDs { subnetRefs[i] = aruba.SubnetRef(args.ProjectID, args.VPCID, s) }

    server := aruba.NewCloudServer().
        InProject(projectRef(args.ProjectID)).
        Named(args.Name).
        InRegion(args.Region).
        InZone(args.Zone).
        OfFlavor(args.Flavor).
        BootingFrom(volumeRef(args.ProjectID, args.BootDiskID)).
        WithVPC(aruba.VPCRef(args.ProjectID, args.VPCID)).
        OnSubnets(subnetRefs...).
        BilledBy(args.BillingPeriod)
    // ... optional fields ...

    resp, err := client.FromCompute().CloudServers().Create(ctx, server)
    if err != nil { return fmt.Errorf("creating cloud server: %w", apiErrFromV2(err)) }

    if resp != nil && resp.Raw() != nil {
        // ... build headers + row, call PrintOutput(resp, headers, rows) ...
    } else {
        fmt.Println(msgCreatedAsync("Cloud server", args.Name))
    }
    return nil
}
```

**Constraints on the operation function:**
- Must not call `GetArubaClient()`, `GetProjectID()`, `cmd.Flags()`, or read `os.Stdin`.
- Takes `ctx context.Context` and `client aruba.Client` as its first two arguments.
- Accepts `args <Family><Resource><Action>Args` by value (not pointer).
- Config-less resources (Config, Context) omit the `client` parameter.

### 4. Run wiring

```go
func ComputeCloudServerCreateRun(cmd *cobra.Command, _ []string) error {
    args, err := NewComputeCloudServerCreateArgsFromCobraCommand(cmd)
    if err != nil { return fmt.Errorf("checking args: %w", err) }

    client, err := GetArubaClient()
    if err != nil { return fmt.Errorf("initializing client: %w", err) }

    ctx, cancel := newCtx()
    defer cancel()

    if err := ComputeCloudServerCreate(ctx, client, *args); err != nil {
        return fmt.Errorf("running command: %w", err)
    }
    return nil
}

var cloudserverCreateCmd = &cobra.Command{
    // ...
    RunE: ComputeCloudServerCreateRun, // named, not anonymous
}
```

Delete wiring is the only variant — `confirmDelete` and `--dry-run` live here:

```go
func ComputeCloudServerDeleteRun(cmd *cobra.Command, args []string) error {
    a, err := NewComputeCloudServerDeleteArgsFromCobraCommand(cmd, args)
    if err != nil { return fmt.Errorf("checking args: %w", err) }

    if !a.SkipConfirm {
        ok, err := confirmDelete("cloud server", a.ID)
        if err != nil { return err }
        if !ok { return nil }
    }

    client, err := GetArubaClient()
    if err != nil { return fmt.Errorf("initializing client: %w", err) }
    ctx, cancel := newCtx(); defer cancel()

    if a.DryRun {
        if _, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(a.ProjectID, a.ID)); err != nil {
            return fmt.Errorf("dry-run: cloud server not found or inaccessible: %w", apiErrFromV2(err))
        }
        fmt.Println(msgDryRun("cloud server", a.ID))
        return nil
    }

    if err := ComputeCloudServerDelete(ctx, client, *a); err != nil {
        return fmt.Errorf("running command: %w", err)
    }
    return nil
}
```

`confirmDelete(resourceType, id string) (bool, error)` is a helper in `cmd/root.go` that
detects non-interactive stdin and skips the prompt when `--yes` is set or when stdin is
not a terminal. Use it — do not inline the prompt.

### File layout

Keep all four symbols for a resource in the same file (e.g. `cmd/compute.cloudserver.go`).
The file grows but remains self-contained and grep-able. Tests live in the matching
`*_test.go` file; no separate `*_args_test.go`.

### Test layers

Three layers per resource (see `ARCHITECTURE.md` "Args / Operation / Run"):
- **Layer 1 — Validate:** table-driven pure-Go tests, one per action. Assert
  `errors.Is(err, ErrValidationFailed)` after the constructor wraps.
- **Layer 2 — Operation:** call the operation function directly with `newArubaTestServer(t).Client()`.
  Use `captureStdout` (root_test.go:277) for output assertions.
- **Layer 3 — Invocation:** existing `Test*Cmd` table tests via `runCmd` kept as-is.
  One case per action is enough to prove the RunE wiring compiles and runs end-to-end.

---

## Output Formatting

Always pass the **SDK wrapper** (not `.Raw()`) as the first argument to `PrintOutput`:

```go
// Correct — wrapper satisfies rawMarshaler; -o json/yaml use SDK's RawJSON()/RawYAML()
created, err := client.From<Svc>().<Resource>().Create(ctx, wrapper)
PrintOutput(created, headers, [][]string{row})

// Wrong — *types.XxxResponse does not implement rawMarshaler;
// -o json falls back to json.MarshalIndent on a pointer-heavy struct
raw := created.Raw()
PrintOutput(raw, headers, [][]string{row})
```

Table rows are still built from `raw.*` fields (via `.Raw()`) for now — the SDK does
not yet expose every field through wrapper accessors. Only the `PrintOutput` first arg
changes.

One intentional exception:
- **Anonymous struct delete results** — `PrintOutput(struct{ID, Status string}{...}, ...)`
  is intentional; these small structs fall through to `json.MarshalIndent` which is fine
  for the trimmed delete-confirmation shape.

Note: the `*aruba.Project` / `rawHTTPer` exception existed in v0.3.0 and was removed
in the sdk-go v1.0.0 migration (TD-032). `*aruba.Project` now satisfies `rawMarshaler`.

---

## Test Conventions

- Tests live in `package cmd` (same package as the code).
- Use `t.TempDir()` for isolated file paths; override `HOME` (or `USERPROFILE` on Windows) to redirect config/context files.
- Use `defer cleanup()` to restore environment variables.
- Skip live-API tests with `ACLOUD_TEST_SKIP_CLIENT=true`.
- Table-driven tests are preferred for multiple input/output cases.

### Test client — httptest harness (v0.2.0+, current v1.0.0)

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

---

## Storage-Specific Patterns

The storage family has three patterns that diverge from the canonical command bodies above.

### Cross-family pre-validation in Backup Create

`StorageBackup.Create` requires a source volume. The file validates the volume exists
before building the backup wrapper. The volume GET response **must** include a URI
field so the SDK can carry it through the builder chain.

```go
// Pre-validate source volume (cross-family Get)
vol, err := client.FromStorage().Volumes().Get(ctx, volumeRef(projectID, volumeID))
if err != nil { return fmt.Errorf("getting volume: %w", apiErrFromV2(err)) }

bk := aruba.NewStorageBackup().
    InProject(projectRef(projectID)).
    Named(name).
    InRegion(aruba.Region(region)).
    BilledBy(aruba.BillingPeriod(billingPeriod)).
    FromVolume(vol)
if retentionDays > 0 { bk.WithRetentionDays(int(retentionDays)) }

created, err := client.FromStorage().Backups().Create(ctx, bk)
if err != nil { return fmt.Errorf("creating backup: %w", apiErrFromV2(err)) }
```

### Restore Create: dual cross-family Gets + IntoBackup/ToVolume

`StorageRestore` is parented on a Backup (not a project). Use `IntoBackup(bk)` (not
`InProject`) — it extracts projectID and backupID from the backup's URI. Both the
backup GET and volume GET responses **must** include a URI field. `ToVolume(target)`
extracts the URI from the volume wrapper.

```go
// Pre-validate parent backup AND target volume
bk, err := client.FromStorage().Backups().Get(ctx, backupRef(projectID, backupID))
if err != nil { return fmt.Errorf("getting backup: %w", apiErrFromV2(err)) }
target, err := client.FromStorage().Volumes().Get(ctx, volumeRef(projectID, volumeID))
if err != nil { return fmt.Errorf("getting volume: %w", apiErrFromV2(err)) }

rs := aruba.NewStorageRestore().
    IntoBackup(bk).       // parents on backup, NOT InProject(...)
    Named(name).
    InRegion(aruba.Region(region)).
    ToVolume(target)

created, err := client.FromStorage().Restores().Create(ctx, rs)
if err != nil { return fmt.Errorf("creating restore: %w", apiErrFromV2(err)) }
```

### Restore List: backup-scoped (not project-scoped)

Unlike all other storage resources, `StorageRestoreClient.List` takes a **backup Ref**
as its first argument, not a project Ref.

```go
// Use backupRef(projectID, backupID), NOT projectRef(projectID)
list, err := client.FromStorage().Restores().List(ctx, backupRef(projectID, backupID), listOpts(cmd)...)
if err != nil { return fmt.Errorf("listing restores: %w", apiErrFromV2(err)) }
```

In tests, register the list route under the backup-scoped path:
`/projects/{p}/providers/Aruba.Storage/backups/{bid}/restores`

---

## Database-specific patterns

### Family B create (Database, User)

Family B resources (Database, User) use `IntoDBaaS(dbaasRef(projectID, dbaasID))` rather than `InProject`. Name/username is the path identifier.

```go
// Database create
db := aruba.NewDatabase().
    IntoDBaaS(dbaasRef(projectID, dbaasID)).
    Named(name)
created, err := client.FromDatabase().Databases().Create(ctx, db)

// User create
u := aruba.NewUser().
    IntoDBaaS(dbaasRef(projectID, dbaasID)).
    WithUsername(username).
    WithPassword(password)
created, err := client.FromDatabase().Users().Create(ctx, u)
```

### DBaaSBackup create (no pre-validation Gets)

Pass constructed Refs directly — `FromDBaaS` and `FromDatabase` accept any `aruba.Ref`:

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

### Dbaas-scoped List

Database and User List are scoped to a DBaaS instance, not a project:

```go
list, err := client.FromDatabase().Databases().List(ctx, dbaasRef(projectID, dbaasID), listOpts(cmd)...)
list, err := client.FromDatabase().Users().List(ctx, dbaasRef(projectID, dbaasID), listOpts(cmd)...)
```

DBaaS and Backup List are project-scoped (standard pattern):

```go
list, err := client.FromDatabase().DBaaS().List(ctx, projectRef(projectID), listOpts(cmd)...)
list, err := client.FromDatabase().Backups().List(ctx, projectRef(projectID), listOpts(cmd)...)
```

---

## Container-Specific Patterns

### KaaS Create — SecurityGroup wrapper required

`KaaS.WithSecurityGroup` does a type assertion to `*aruba.SecurityGroup`. Pass a named wrapper, not a URI Ref:

```go
sg := aruba.NewSecurityGroup().Named(securityGroupName) // *aruba.SecurityGroup, not aruba.URI(...)
k := aruba.NewKaaS().
    InProject(projectRef(projectID)).
    Named(name).
    InRegion(aruba.Region(region)).
    WithKubernetesVersion(aruba.KubernetesVersion(kubernetesVersion)).
    WithNodeCIDR(nodeCIDRAddress, nodeCIDRName).
    WithSecurityGroup(sg).
    WithVPC(aruba.VPCRef(projectID, vpcID)).
    WithSubnet(aruba.SubnetRef(projectID, vpcID, subnetID)).
    AddNodePool(nodePool)
created, err := client.FromContainer().KaaS().Create(ctx, k)
```

### KaaS Connect — two-step kubeconfig download

`DownloadKubeconfig` lives on `*KaaS` (not `KaaSClient`). Must `Get` first to obtain a hydrated wrapper:

```go
got, err := client.FromContainer().KaaS().Get(ctx, kaasRef(projectID, kaasID))
if err != nil { return fmt.Errorf("getting KaaS cluster: %w", apiErrFromV2(err)) }
kubeconfigBytes, err := got.DownloadKubeconfig(ctx)
if err != nil { return fmt.Errorf("downloading kubeconfig: %w", apiErrFromV2(err)) }
// kubeconfigBytes is []byte(resp.Data.Content) — base64-encoded YAML; decode before writing
decodedContent, err := base64.StdEncoding.DecodeString(string(kubeconfigBytes))
if err != nil { decodedContent = kubeconfigBytes } // already raw if decode fails
```

### ContainerRegistry Create — SDK Ref helpers for all network resources

Unlike KaaS, `ContainerRegistry.WithSecurityGroup` accepts any `Ref` (no type assertion). Use SDK Ref helpers rather than `aruba.URI(...)`:

```go
r := aruba.NewContainerRegistry().
    InProject(projectRef(projectID)).
    Named(name).
    InRegion(aruba.Region(region)).
    WithElasticIP(aruba.ElasticIPRef(projectID, publicIPID)).
    WithVPC(aruba.VPCRef(projectID, vpcID)).
    WithSubnet(aruba.SubnetRef(projectID, vpcID, subnetID)).
    WithSecurityGroup(aruba.SecurityGroupRef(projectID, vpcID, sgID)).
    WithBlockStorage(volumeRef(projectID, blockStorageID))
if concurrentUsers != "" { r.OfSize(aruba.ContainerRegistrySizeFlavor(concurrentUsers)) }
created, err := client.FromContainer().ContainerRegistry().Create(ctx, r)
```

## Schedule-Specific Patterns

### Job Create — OneShot vs Recurring (mutually exclusive)

```go
j := aruba.NewJob().
    InProject(projectRef(projectID)).
    Named(name).
    InRegion(aruba.Region(region)).
    OfType(aruba.JobType(jobType)).
    Enabled()  // or .Disabled() — explicit state setters (since v0.3.0)

if cronExpr != "" {
    endTime, _ := time.Parse(time.RFC3339, endTimeStr)
    j.WithCron(cronExpr).RecurringUntil(endTime)
} else {
    shotTime, _ := time.Parse(time.RFC3339, shotTimeStr)
    j.OneShotAt(shotTime)
}

if err := j.Err(); err != nil {
    return fmt.Errorf("invalid job configuration: %w", err)
}

ctx, cancel := newCtx(); defer cancel()
created, err := client.FromSchedule().Jobs().Create(ctx, j)
if err != nil { return fmt.Errorf("creating schedule job: %w", apiErrFromV2(err)) }
```

`j.Err()` surfaces builder validation failures (e.g. mixing OneShot and Recurring). Always check it before the Create call.

### Job Update

```go
cur, err := client.FromSchedule().Jobs().Get(ctx, jobRef(projectID, jobID))
if err != nil { return fmt.Errorf("getting schedule job: %w", apiErrFromV2(err)) }

if name != "" { cur.Named(name) }
if enabledSet {
    if enabled { cur.Enabled() } else { cur.Disabled() }  // explicit state setters (since v0.3.0)
}
if cmd.Flags().Changed("tags") { cur.RetaggedAs(tags...) }

updated, err := client.FromSchedule().Jobs().Update(ctx, cur)
```

The `omitempty` issue that affected `WithEnabled(false)` in earlier SDK versions was resolved in v0.3.0 via the explicit `Enabled()`/`Disabled()` setters.
