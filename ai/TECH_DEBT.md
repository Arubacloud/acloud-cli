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
| TD-010 | Table-driven `RunE` tests added for all 23 testable command files (24 including pre-existing `network.vpc_test.go`); mock infrastructure in `cmd/mock_test.go` covers all sub-clients; `security.kms.go` skipped (concrete SDK type, cannot mock); nil-pointer bugs in `LocationResponse.Value` and `CreationDate.IsZero()` fixed as a side effect of test authoring; redundant double nil-check blocks left by AWK generation cleaned up in 5 files |
| TD-020 | Six helper functions added to `cmd/root.go` (`msgCreated`, `msgCreatedAsync`, `msgUpdated`, `msgUpdatedAsync`, `msgDeleted`, `msgAction`); all ~91 success `fmt.Print*` calls replaced across 20 cmd files; one double-nil-check fixed in `container.containerregistry.go` as a side effect |
| TD-021 | `Long` and `Example` fields added to all 23 create commands across 22 cmd files; subnet already had a minimal `Long` which was replaced with a richer version |
| TD-019 | `--dry-run` flag added to all 24 delete commands; in dry-run mode a `Get` validates existence and access then prints `[dry-run] Would delete …` without calling `Delete`; `msgDryRun` helper added to `cmd/root.go` |
| TD-015 | Raw-JSON `response.RawBody` ID extraction removed from `cloudserver` and `keypair` list commands; typed `Metadata.ID` used directly; entries with nil/empty ID are discarded (SDK bumped to v0.1.26) |
| TD-023 | Verbose error-body dump in `storage.backup` / `storage.restore` now uses `json.MarshalIndent(response.Error, …)`; generic `map[string]interface{}` unmarshal removed |

---

## Low

### TD-023 · Remove `PrintTable` shim
`PrintTable(headers, rows)` is now a one-line shim around `PrintOutput(nil, headers, rows)`. All call sites that pass `nil` as the first arg produce `{}` for `-o json` / `-o yaml` instead of the actual resource data. Remaining direct `PrintTable` calls should be replaced with `PrintOutput(response.Data, headers, rows)` and the shim deleted.

**Fix:** Grep for `PrintTable(` and migrate each site to `PrintOutput`, passing the typed SDK response as the first argument. Remove the `PrintTable` function once all sites are updated.

---

### TD-022 · Pre-release SDK version (v0.1.x)
`go.mod` depends on `github.com/Arubacloud/sdk-go v0.1.27`. The `0.x` major version provides no semantic versioning stability guarantee — a minor-version bump may introduce breaking changes.

**Fix:** Track the SDK release roadmap. When a `v1.0.0` is released, migrate and pin to it. Until then, pin to a specific minor version and treat any upgrade as potentially breaking.

---

### TD-024 · VPN tunnel IKE/ESP/PSK enum values undocumented
`vpntunnel create` exposes `--dhgroup`, `--pfs`, `--cloud-site`, `--onprem-site`, and sibling IKE/ESP flags as free-text strings. The SDK (`sdk-go@v0.1.27/pkg/types/network.vpn-tunnel.go:34,58,64-67`) models them as opaque `*string` with no enum constants, no Swagger/OpenAPI spec, and no sample payloads anywhere in this repo, the SDK tree, or `docs/`. The current e2e run succeeds with the values pinned in test.sh (ikev2, aes256, sha1, dh group 1, pfs enable, etc.), but those constants live only in the test script — the CLI accepts any string and the SDK has no enum validation, so a typo or future API change still produces an opaque "not valid" error.

**Fix:** Obtain the API spec or a known-good payload from Aruba. Enumerate valid values for `dhGroup`, `pfs`, `cloudSite`, `onPremSite`, encode them as Go constants in `cmd/network.vpntunnel.go`, optionally validate client-side before the request, and document them in `docs/website/docs/resources/network/vpntunnel.md`. Mark resolved once the constants are documented; the e2e gate this previously referenced no longer exists.

---

### TD-025 · VPN tunnel subnet semantics: CLI says "existing", API says "overlap"
`cmd/network.vpntunnel.go:28-29` documents `--subnet-cidr` as *"CIDR of existing subnet"* and `--subnet-name` as an alternative lookup — i.e. a reference to a pre-existing subnet. But when a subnet with the referenced CIDR already exists in the VPC, the API responds `ipConfigurations.subnet.cidr overlaps with an existing subnet`, suggesting it interprets the field as a *provisioning* instruction rather than a lookup. The two readings are contradictory and the e2e test cannot safely pre-create the subnet.

**Fix:** Confirm with the API team whether `ipConfigurations.subnet` is a reference or a creation spec. If it is a lookup, surface a clearer error when the subnet is missing. If it is a creation field, update the CLI `Long`, flag descriptions, and `docs/website/docs/resources/network/vpntunnel.md` accordingly, and remove the pre-create step from the e2e test.

---

### TD-026 · VPC Peering Route lifecycle ends in `Failed` due to API ACL
`network vpcpeeringroute create` against the e2e tenant now returns 200 (the v0.1.26→v0.1.27 SDK URL-path fix removed the bare 403 at CREATE), but the resulting route transitions to `Failed` shortly after creation. No `errors` array is exposed via GET; the route is simply unhealthy. This points at an API-side IAM/ACL on the `Aruba.Network/vpcPeerings/{id}/vpcPeeringRoutes` provisioning step scoped to this tenant — not a CLI bug. The e2e suite handles this via `wait_for_status`'s terminal-failure short-circuit: UPDATE is skipped, DELETE runs with `|| true`, and the function returns 0 so the suite stays green.

**Fix:** Confirm required tenant/role permissions with Aruba and update the ACL accordingly. No CLI change is required — once the route reaches `Active`, the existing UPDATE/DELETE blocks in `test_vpc_peering_route` will exercise the full CRUD without further code changes.
