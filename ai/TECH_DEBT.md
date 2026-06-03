# TECH_DEBT.md — Technical Debt & Refactoring Backlog

Issues are grouped by severity. Address Critical items before new features ship; High items before any public release.

## Resolved

| ID | Summary |
|----|---------|
| TD-001 | All `Run` handlers converted to `RunE`; `SilenceUsage: true`; `fmtAPIError` helper in `root.go` |
| TD-002 | Nil guards added for `LocationResponse`, `Metadata.ID`, `Metadata.Name` in all list/create/update responses |
| TD-003 | Errors propagated via `return fmt.Errorf(...)` instead of printing to stdout |
| TD-004 | Flag read errors checked via `RunE` return paths |
| TD-005 | `confirmDelete()` helper in `root.go` detects non-interactive stdin before prompting |
| TD-006 | `newCtx()` helper in `root.go` applies 30-second timeout to all SDK calls |
| TD-007 | `getContextFilePath` returns `(string, error)` instead of silently falling back to CWD |
| TD-008 | YAML unmarshal errors wrapped with user-friendly messages in `LoadConfig` and `LoadContext` |
| TD-009 | `MarkFlagRequired` used as the single mechanism for all required flags; redundant `if flag == ""` manual checks removed from all 19 affected commands |
| TD-011 | `readSecret()` helper added to `root.go` using `golang.org/x/term.ReadPassword`; `config set` now prompts interactively when `--client-secret` is not passed and no secret exists in config |
| TD-012 | `--debug` flag description updated to warn about credential/token exposure in HTTP headers |
| TD-013 | `Args: cobra.NoArgs` added to all `create` and `list` commands that take no positional arguments |
| TD-014 | `cmd/constants.go` created with `StateInCreation`, `DateLayout`, `FilePermConfig`, `FilePermDirAll`; all magic strings replaced |
| TD-016 | Multi-mode output implemented: 5 canonical formats (`table`, `table-json`, `table-yaml`, `json`, `yaml`) via `resolveOutputFormat` + `PrintOutput`; `PrintTable` is now a shim; `-o json`/`-o yaml` emit full SDK response (breaking change from original PR #30 flat shape) |
| TD-017 | `listParams(cmd)` helper added; `--limit`/`--offset` flags added to all 25 list commands; list RunE handlers now pass pagination params to SDK |
| TD-018 | Global client cache vars encapsulated in `clientState` struct with `resetClientState()` helper; all test reset blocks updated to use it |
| TD-010 | Table-driven `RunE` tests added for all 23 testable command files (24 including pre-existing `network.vpc_test.go`); mock infrastructure in `cmd/mock_test.go` covers all sub-clients; `security.kms_test.go` added via `arubaTestServer` harness in #109 (was previously skipped — concrete SDK type blocked concrete mocking); nil-pointer bugs in `LocationResponse.Value` and `CreationDate.IsZero()` fixed as a side effect of test authoring; redundant double nil-check blocks left by AWK generation cleaned up in 5 files |
| TD-020 | Six helper functions added to `cmd/root.go` (`msgCreated`, `msgCreatedAsync`, `msgUpdated`, `msgUpdatedAsync`, `msgDeleted`, `msgAction`); all ~91 success `fmt.Print*` calls replaced across 20 cmd files; one double-nil-check fixed in `container.containerregistry.go` as a side effect |
| TD-021 | `Long` and `Example` fields added to all 23 create commands across 22 cmd files; subnet already had a minimal `Long` which was replaced with a richer version |
| TD-019 | `--dry-run` flag added to all 24 delete commands; in dry-run mode a `Get` validates existence and access then prints `[dry-run] Would delete …` without calling `Delete`; `msgDryRun` helper added to `cmd/root.go` |
| TD-015 | Raw-JSON `response.RawBody` ID extraction removed from `cloudserver` and `keypair` list commands; typed `Metadata.ID` used directly; entries with nil/empty ID are discarded (SDK bumped to v0.1.26) |
| TD-023 | Verbose error-body dump in `storage.backup` / `storage.restore` now uses `json.MarshalIndent(response.Error, …)`; generic `map[string]interface{}` unmarshal removed |
| TD-024 | VPN crypto enums split per-direction in v0.2.0: the five unified slices (`vpnEncryptionAlgorithms`, `vpnHashAlgorithms`, `vpnDHGroups`, `vpnDPDActions`, `vpnPFSGroups`) are replaced with seven per-direction slices (`vpnIKEEncryptionAlgorithms`, `vpnESPEncryptionAlgorithms`, `vpnIKEHashAlgorithms`, `vpnESPHashAlgorithms`, `vpnIKEDHGroups`, `vpnIKEDPDActions`, `vpnESPPFSGroups`) built from `aruba.IKE*`/`aruba.ESP*` constants; each `--ike-*`/`--esp-*` flag is keyed to its correct family in the validation table; accepted string values are unchanged (CLI behaviour byte-identical) |
| TD-022 | SDK fully migrated from v0.1.x → v0.2.0 wrapper API across all families (#100–#110); bumped to v0.2.1 (#111) which resolved all four upstream papercuts (#282–#285: Job.WithEnabled omitempty, List pagination stubs, VPC.Get missing projectID backfill, projectsClientImpl.Delete missing error-body parse). v0.2.1 also added typed Ref builders (VPCRef, SubnetRef, etc.) and WithBillingPeriod/ReplaceNodePools setters — all adopted in #111. Bumped to v0.3.0 on branch `upgrade-sdk`: natural-language setter vocabulary (InProject, InVPC, RetaggedAs, BilledBy, HighlyAvailable(), Enabled()/Disabled()), URI segment casing aligned (securityGroups, securityRules, blockStorages, loadBalancers, keyPairs), both local vendor patches removed (securityGroups casing fix and List[T].Raw() JSON marshalability — now upstream in v0.3.0), single-import achieved for all non-test cmd/ files (escape hatch: container.kaas.go uses types.APIServerAccessProfileProperties until SDK provides a setter), output handlers now delegate to SDK's RawJSON()/RawYAML() via rawMarshaler interface (all 20 cmd files pass the wrapper, not `.Raw()`, to PrintOutput), and cross-resource reference flags refactored from `--<res>-uri` to `--<res>-id` using SDK Ref helpers (VPCRef, SubnetRef, SecurityGroupRef, ElasticIPRef). |
| TD-032 | `rawHTTPer` interface and both fallback branches in `printJSON`/`printYAML` deleted from `cmd/root.go`. sdk-go v1.0.0 (released 2026-05-29) added `RawJSON()/RawYAML()` to `*aruba.Project`, so it now satisfies `rawMarshaler` directly. The `"net/http"` import and the `fakeRawHTTPer` test type also removed. `management.project.go` was already passing the wrapper directly to `PrintOutput` — no change there. Tests `TestPrintJSON/rawHTTPer_*` and `TestPrintYAML/rawHTTPer_*` deleted. |
| TD-036 | sdk-go bumped from v0.3.0 → v1.0.0 (GA, released 2026-05-29) on branch `upgrade-sdk-v1`. Breaking changes in `pkg/types`: strict role-suffix naming (`*PropertiesResult`→`*PropertiesResponse`, bare `*List`→`*ListResponse`, `*Common`-suffixed nested types). Production code impact: 7 compile-error fixes in `container.containerregistry.go`, `container.kaas.go`, `management.project.go`, `network.elasticip.go`, `network.subnet.go`, `network.vpcpeeringroute.go`, `network.vpntunnel.go`. Test files: ~46 type-rename substitutions across 28 `_test.go` files. Completion functions migrated to wrapper accessors (`.ID()`, `.Name()`) in 14 cmd files, eliminating ~50 `.Raw().Metadata.ID/Name` reads. |
| TD-034 | VPN tunnel update handler manual IKE/ESP/PSK reattach block removed from `cmd/network.vpntunnel.go`. sdk-go v1.0.0 `fromResponse` now fully rehydrates IKE/ESP/PSK into the `*VPNIKE`/`*VPNESP`/`*VPNPSK` wrapper fields exposed via `IKE()`/`ESP()`/`PSK()` accessors; `toRequest()` builds the PUT body from those fields. Closed #132. Residual: `redactVPNTunnelSecrets` still mutates `t.response.Properties.VPNClientSettingsCommon.PSK.Secret` directly because `RawJSON()` serializes `t.response`; upstream `RedactPSK()` mutator still needed (tracked in #132 comment). |
| TD-035 | Subnet update DHCP preservation block migrated to wrapper accessors in `cmd/network.subnet.go`. `subnet.Type()` replaces `subnet.Raw().Properties.Type`; `subnet.DHCP()` replaces `.Raw().Properties.DHCP` with full `IsEnabled()`/`Routes()`/`DNS()` read access on `*SubnetDHCPCommon`. Zero `.Raw()` calls remain for this code path. Closed #133. |

---

## Low

### TD-023 · Remove `PrintTable` shim
`PrintTable(headers, rows)` is now a one-line shim around `PrintOutput(nil, headers, rows)`. All call sites that pass `nil` as the first arg produce `{}` for `-o json` / `-o yaml` instead of the actual resource data. Remaining direct `PrintTable` calls should be replaced with `PrintOutput(response.Data, headers, rows)` and the shim deleted.

**Fix:** Grep for `PrintTable(` and migrate each site to `PrintOutput`, passing the typed SDK response as the first argument. Remove the `PrintTable` function once all sites are updated.

---

### TD-025 · VPN tunnel subnet semantics: CLI says "existing", API says "overlap"
`cmd/network.vpntunnel.go:28-29` documents `--subnet-cidr` as *"CIDR of existing subnet"* and `--subnet-name` as an alternative lookup — i.e. a reference to a pre-existing subnet. But when a subnet with the referenced CIDR already exists in the VPC, the API responds `ipConfigurations.subnet.cidr overlaps with an existing subnet`, suggesting it interprets the field as a *provisioning* instruction rather than a lookup. The two readings are contradictory and the e2e test cannot safely pre-create the subnet.

**Fix:** Confirm with the API team whether `ipConfigurations.subnet` is a reference or a creation spec. If it is a lookup, surface a clearer error when the subnet is missing. If it is a creation field, update the CLI `Long`, flag descriptions, and `docs/website/docs/resources/network/vpntunnel.md` accordingly, and remove the pre-create step from the e2e test.

---

### TD-026 · VPC Peering Route lifecycle ends in `Failed` due to API ACL
`network vpcpeeringroute create` against the e2e tenant now returns 200 (the v0.1.26→v0.1.27 SDK URL-path fix removed the bare 403 at CREATE), but the resulting route transitions to `Failed` shortly after creation. No `errors` array is exposed via GET; the route is simply unhealthy. This points at an API-side IAM/ACL on the `Aruba.Network/vpcPeerings/{id}/vpcPeeringRoutes` provisioning step scoped to this tenant — not a CLI bug. The e2e suite handles this via `wait_for_status`'s terminal-failure short-circuit: UPDATE is skipped, DELETE runs with `|| true`, and the function returns 0 so the suite stays green.

**Fix:** Confirm required tenant/role permissions with Aruba and update the ACL accordingly. No CLI change is required — once the route reaches `Active`, the existing UPDATE/DELETE blocks in `test_vpc_peering_route` will exercise the full CRUD without further code changes.

---

### TD-027 · `acloud security key` and `acloud security kmip` commands not yet implemented
The v0.2.0 SDK exposes `KeysClient` and `KmipsClient` sub-clients under `client.FromSecurity()`, but the CLI only wraps `KMS()` today (`cmd/security.kms.go`). Per #109 scope ("match issue strictly"), Key and KMIP CLI surfaces were intentionally deferred. Users who need these features must use the SDK directly.

**Fix:** Add `cmd/security.key.go` and `cmd/security.kmip.go` following the same fluent-builder + `arubaTestServer` patterns documented in `ai/ARCHITECTURE.md` and `ai/CONVENTIONS.md`. Track as a follow-up sub-issue against #99 (or a successor parent issue post-v0.2.0 release).

---

### TD-029 · `cloudserver update` always 400 — PUT body missing subnet/SG refs (#125)
`acloud compute cloudserver update` fails on every call with HTTP 400. The SDK's `fromResponse` does not hydrate `subnetRefs`/`securityGroupRefs` from the GET response because the API returns those in different shapes (`networkInterfaces[].subnet` and `linkedResources[]`). The PUT body therefore omits them, which the API interprets as stripping all subnets/SGs → validation failure.

**Diagnosis:** Run `acloud --debug compute cloudserver update <id> --tags foo` and inspect `[DEBUG] response body:` to confirm which field(s) the API rejects.

**Fix:** Map `networkInterfaces[].subnet` → `subnetRefs` and identify security-group URIs in `linkedResources[]` in the CLI update handler (or upstream in SDK `fromResponse`). Re-inject before calling `Update`. See #125.

---

### TD-030 · VPN route Cloud Subnet always blank on `get` / `list`

`acloud network vpnroute get` prints an empty `Cloud Subnet:` row even when the route
was created with `--cloud-subnet 10.0.0.0/24`. The CLI reads
`raw.Properties.CloudSubnet.CIDR`, but the GET response places the value under a
different field path (likely `raw.Properties.CloudSubnet.Value` or a flat string).
The CLI has no way to validate the exact shape without access to a live API response.

**Root cause:** Structural mismatch between `types.VPNRoutePropertiesResponse` and the
actual API response body — an sdk-go issue.

**Fix (blocked on sdk-go):** Once the SDK aligns the type definition with the API
contract, update the field read in `cmd/network.vpnroute.go` (both `get` and `list`
output blocks).

---

### TD-031 · Schedule job create does not send `steps[]` — API rejects with 400

`acloud schedule job create` always returns HTTP 400 because the payload contains no
`steps[]`. The `--step-resource-uri` / `--step-action-uri` / `--step-http-verb` /
`--step-name` flags are declared in `init()` but the `RunE` never reads them and the
SDK builder had no `AddStep` / `WithStep` method in v0.3.0. sdk-go v1.0.0 added
`Job.Steps()` (a read accessor) but it is unclear whether a write/builder API was
added. Check `aruba.Job` in v1.0.0 for `AddStep`/`WithStep` before implementing.

**Fix:** If v1.0.0 added a step builder, read the four flags and append the step before
calling `Create`. Otherwise, the command should return a clear "not yet implemented"
error rather than silently emitting an invalid payload.

---

### TD-033 · `container.kaas.go` and `management.project.go` residual `pkg/types` and `.Raw()`

Two residual `.Raw()` / `types.*` usages remain in production code after the v1.0.0 migration:

1. **`cmd/container.kaas.go:221`** — `types.KaaSAPIServerAccessProfilePropertiesRequest` must be
   referenced directly because sdk-go v1.0.0 provides no `aruba`-level constructor for
   API server access profile settings. Remove the `pkg/types` import once sdk-go exposes
   `aruba.NewAPIServerAccessProfile()` or an equivalent fluent setter.
2. **`cmd/management.project.go`** — `p.Raw().Metadata.CreatedBy` / `.UpdatedBy` are read
   because `*aruba.Project` in v1.0.0 does not expose `CreatedBy()` / `UpdatedBy()` wrapper
   accessors. Remove once sdk-go adds these accessors to the `Project` wrapper.

**Filed as:** https://github.com/Arubacloud/acloud-cli/issues/131

---

### TD-034 · `redactVPNTunnelSecrets` still uses `.Raw()` for PSK.Secret mutation

The `redactVPNTunnelSecrets` function in `cmd/network.vpntunnel.go` mutates
`tunnel.Raw().Properties.VPNClientSettingsCommon.PSK.Secret = nil` to strip the secret
before `-o json`/`-o yaml` output. This is necessary because `RawJSON()`/`RawYAML()`
serialize `t.response` (the raw struct), not the wrapper state. The IKE/ESP/PSK
reattach block was eliminated in the v1.0.0 migration (TD-034 resolved), but this
single mutation site remains.

**Fix (blocked on sdk-go):** Once sdk-go adds a `RedactPSK()` method (or a write-through
mutator) on `*aruba.VPNTunnel` that zeros both the wrapper's PSK secret AND the
corresponding field in the cached raw struct, remove the `.Raw()` call here.

**Tracked in:** https://github.com/Arubacloud/acloud-cli/issues/132 (comment)

---

### TD-028 · User-facing migration guide for `-o json`/`-o yaml` output-shape change
TD-016 switched `-o json`/`-o yaml` to emit the full SDK response wrapper (`{ statusCode, data: { values: [...] } }`) instead of the flat shape used by acloud-cli v0.1.x. This is a **breaking change** for machine-readable consumers. There is currently no user-facing migration guide explaining the diff and recommending `jq`/`yq` patterns for downstream callers.

**Fix:** Add a `docs/website/docs/output-formats.md` page (or expand an existing one) showing:
- Before/after shape for a representative resource (e.g. `acloud network vpc list -o json`).
- `jq` snippets to recover the flat shape (`.data.values[] | {id: .metadata.id, name: .metadata.name}`).
- A note that `-o table-json` and `-o table-yaml` continue to emit table rows for table-style consumers.

Surface this prominently in the v0.2.0 GitHub Release notes.
