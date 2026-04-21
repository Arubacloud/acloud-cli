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

---

## Low

### TD-023 · Remove `PrintTable` shim
`PrintTable(headers, rows)` is now a one-line shim around `PrintOutput(nil, headers, rows)`. All call sites that pass `nil` as the first arg produce `{}` for `-o json` / `-o yaml` instead of the actual resource data. Remaining direct `PrintTable` calls should be replaced with `PrintOutput(response.Data, headers, rows)` and the shim deleted.

**Fix:** Grep for `PrintTable(` and migrate each site to `PrintOutput`, passing the typed SDK response as the first argument. Remove the `PrintTable` function once all sites are updated.

---

### TD-022 · Pre-release SDK version (v0.1.x)
`go.mod` depends on `github.com/Arubacloud/sdk-go v0.1.26`. The `0.x` major version provides no semantic versioning stability guarantee — a minor-version bump may introduce breaking changes.

**Fix:** Track the SDK release roadmap. When a `v1.0.0` is released, migrate and pin to it. Until then, pin to a specific minor version and treat any upgrade as potentially breaking.
