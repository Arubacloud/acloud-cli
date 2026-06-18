# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.1] - 2026-06-19

### Fixed

- **Config**: `acloud config set --client-id <new>` now always collects a new
  client-secret — via `ACLOUD_CLIENT_SECRET` or the interactive prompt — even
  when a secret is already stored in the config file. The client-id and
  client-secret are a matched credential pair; previously only the ID was
  updated, leaving the config in an inconsistent state (closes #227).
- **Config**: `acloud config profile list` now shows the effective base URL for
  every profile. Profiles that rely on the implicit default (`https://api.arubacloud.com`)
  were previously shown with a blank `BASE_URL` column (closes #228).

## [0.5.0] - 2026-06-18

### Added

- **Database use case documentation** — new end-to-end guide covering the
  complete workflow for provisioning a MySQL DBaaS instance: VPC/subnet/security
  group setup, instance creation, database creation, user management, grant
  assignment, and connecting with a MySQL client. Available in English and
  Italian.
- **Container Registry use case documentation** — new end-to-end guide covering
  the complete workflow for provisioning a private container registry: block
  storage and network setup, registry creation, Docker authentication, image
  push/pull operations, and registry administration. Available in English and
  Italian.
- **Uninstallation documentation** — new section in the Installation page
  covering uninstallation procedures for all supported platforms (Homebrew, apt,
  rpm, Scoop, manual binary) and configuration file cleanup.
- **Upgrading documentation** — explicit upgrade procedures per platform added
  to the Installation page.

### Changed

- **Documentation structure reorganised** — the monolithic Installation page has
  been split into three focused pages:
  - **Installation** — covers only installation, upgrading, uninstallation,
    verifying, and platform-specific troubleshooting (GLIBC).
  - **Authentication** — covers API credential setup, client configuration,
    multi-profile credential management, and authentication troubleshooting.
    Moved from the Installation page.
  - **Configuration** — covers context management, shell auto-completion, output
    formats, `--wait`, `--dry-run`, pagination, debug mode, and API
    troubleshooting. Moved from the Installation page.
- **New "Authentication & Configuration" sidebar section** — the previously
  flat Installation link has been replaced by a dedicated sidebar category
  grouping the Authentication and Configuration pages for easier discovery.
- **Use Cases section extended** — the Use Cases / Examples section now includes
  Database and Container Registry in addition to the existing Basic Usage,
  CloudServer, and Kubernetes scenarios.

### Improved

- **English and Italian documentation alignment** — all new and restructured
  pages are available in both English and Italian (i18n), keeping both language
  versions at feature parity.
- **Overall documentation clarity and usability** — consistent structure,
  terminology, formatting, and navigation hierarchy across all pages.

## [0.4.0] - 2026-06-17

### Added

- **`--wait` flag for create and update commands** — all resource families now
  support `--wait` to block until the provisioned resource reaches `Active`
  status, turning an asynchronous API call into a synchronous one. Combines
  with `--timeout` to control the maximum wait duration (closes #174).
- **`--timeout` global flag** — overrides the hardcoded 30-second API timeout
  on a per-invocation basis (e.g. `--timeout 5m`). Accepted by every command
  (closes #175).
- **XDG Base Directory Specification** — the config file is now stored at
  `$XDG_CONFIG_HOME/acloud/config.yaml` (defaults to
  `~/.config/acloud/config.yaml`). The legacy `~/.acloud.yaml` path is
  transparently migrated on first read (closes #176).
- **Multi-profile credential support** — a single config file can now hold
  multiple named credential profiles. Switch between dev, staging, and
  production accounts with `--profile <name>` or `ACLOUD_PROFILE=<name>`
  without editing the file (closes #180).
- **Shell completion caching** — completion results (resource IDs, names) are
  cached in memory with a short TTL, eliminating one API round-trip per
  keystroke in interactive shells (closes #179).

### Refactored

These changes improve code organisation and testability; user-facing behaviour
is unchanged.

- **`cmd/root.go` decomposed into `internal/` packages** — the 659-line
  god-file has been split into three focused packages with dedicated test files:
  `internal/errs` (error formatting), `internal/client` (SDK client caching),
  and `internal/output` (rendering). `cmd/root.go` is now 72 lines (closes
  #173).
- **`RenderList[T]` generic helper** — a new `internal/output.RenderList[T]`
  function co-locates each table column's header definition with its value
  extractor, eliminating the repeated `headers + rows + PrintOutput` block from
  all 26 list commands. A `cmd`-level type alias keeps existing call sites
  unchanged (closes #206).
- **`RenderGet` template helper** — a new `internal/output.RenderGet` function
  renders a `text/template` string against a per-resource view struct. All 22
  `get` command detail views are now driven by template constants consolidated
  in `cmd/templates.go`, replacing hand-written `fmt.Printf` blocks (closes
  #207).
- **`PersistentPreRunE` for early client init** — `GetArubaClient()` is now
  called once in `rootCmd.PersistentPreRunE`, validating credentials before any
  resource command's `RunE` fires. Commands that work without credentials
  (`config`, `context`, `completion`) are automatically skipped (closes #208).

### Fixed

- **e2e**: `cleanup()` in the `network` and `storage` suites was missing a
  terminal `echo "Cleanup completed"` guard. Without it, the bash `EXIT` trap
  propagated a non-zero exit status through the suite when `BOOTSTRAP_PROJECT_ID`
  was empty (CI environments with `ACLOUD_PROJECT_ID` pre-set), causing false
  positive suite failures.

## [0.3.0] - 2026-06-16

### Added

- **`--zone` flag for `database dbaas update`** — the API requires the original
  creation zone to be present in every PUT body; omitting it causes a 400
  "DataCenter cannot be modified" error. Pass `--zone <zone>` (e.g. `ITBG-1`)
  to supply it (closes #193).
- **golangci-lint configuration** — `.golangci.yml` added and wired to
  `make lint` target (closes #191).

### Changed

- **sdk-go bumped to v1.0.4** (GA) from v0.3.0 — breaking changes in
  `pkg/types`: strict role-suffix naming (`*PropertiesResult` →
  `*PropertiesResponse`, bare `*List` → `*ListResponse`, `*Common`-suffixed
  nested types). Production code impact: 7 compile-error fixes across container,
  database, network, and storage families. All completion functions migrated to
  wrapper accessors (`.ID()`, `.Name()`), eliminating ~50 `.Raw().Metadata`
  reads. `*aruba.Project` now satisfies `rawMarshaler` via `RawJSON()`/`RawYAML()`
  (TD-032). VPN tunnel update reattach block removed; subnet DHCP preservation
  migrated to wrapper accessors; VPN route cloud-subnet field fixed (TD-034,
  TD-035, TD-030).
- **Go module path renamed** to `github.com/Arubacloud/acloud-cli` (closes #188).
- **`--client-secret` flag removed from `config set`** — the client secret is
  now read exclusively from the `ACLOUD_CLIENT_SECRET` environment variable or
  prompted interactively via secure terminal input. This prevents secrets from
  appearing in shell history (closes #186).

### Fixed

- **Config**: `config show` now masks the client secret in all output formats
  (plain text, JSON, YAML, table-json, table-yaml) (closes #187).
- **Output**: `PrintTable` legacy shim removed; all commands consistently use
  `PrintOutput` (closes #169 / TD-016).
- **Schedule**: `job create` now correctly sends step parameters (`resource_uri`,
  `action_uri`, `http_verb`, `name`) in the API payload; previously the steps
  array was empty (closes #170 / TD-031).
- **Database**: `dbaas update` re-injects the Engine catalog ID from the GET
  response to prevent "Product not found in catalog" 400 errors on PUT.
- **Database**: `backup create` retries up to 3 times on "Specified Database
  name is not found" errors to handle the propagation delay between the DBaaS
  database API and the backup service's internal registry.
- **e2e**: Added `wait_for_removal` helper to `common.sh`; applied after every
  subnet, VPC, security group, and container registry delete to prevent
  parent-resource deletion failures ("subnet not in valid status", "project
  can't be deleted due to the presence of resources").
- **e2e**: `wait_for_status` now prints the captured error output when a polled
  command fails, making auth failures and API errors visible instead of silently
  bailing.
- **e2e/container**: `--concurrent-users` in registry UPDATE fixed from integer
  `20` to valid enum `Medium`. KaaS cleanup now skips the 600-second `wait_del`
  loop after a failed delete, prints orphaned resource IDs for manual portal
  cleanup, and accepts `Failed`/`Error` states in the pre-delete wait. VPC retry
  loop stops early on "used by another resource" instead of spamming 20 identical
  errors.
- **e2e/database**: Skips individual user/database deletes when grants exist
  (they always 409; DBaaS cascade delete handles cleanup). Sweeps untracked
  DBaaS instances before project delete to handle instances left behind by failed
  create calls. VPC retry loop exits immediately on 404 (already gone). DBaaS
  update called with `--zone` so the PUT body includes `dataCenter`. All test
  functions wired to `FAILURES` counter (previously a create failure reported
  "all checks passed").
- **e2e/compute**: `wait_for_removal` applied after subnet and VPC deletes.
- **e2e/schedule**: `bootstrap_step_resource` now self-bootstraps a full cloud
  server (VPC → subnet → SG → boot disk → server) when `ACLOUD_STEP_RESOURCE_URI`
  is not set, using the canonical URI format from the Terraform provider examples
  (`/projects/{id}/providers/Aruba.Compute/cloudServers/{id}`). Full cleanup
  chain with `wait_for_removal` at each boundary.

## [0.2.0] - 2026-06-09

### Added

- **DBaaS Grant commands** — new `database dbaas grant` sub-family: `create`,
  `list`, `get`, `delete` for managing database-level grants.
- **`--zone` flag for `database backup create`** — zone can now be specified at
  backup creation time (closes #152).
- **VPN IKE/ESP enum constants** — client-side validation for `--ike-version`,
  `--ike-encryption`, `--ike-integrity`, and `--esp-*` flags on VPN tunnel
  commands (closes #72).
- **`User-Agent` header** — HTTP requests now carry `acloud-cli/<version>` so
  server logs identify the CLI version.
- **DBaaS networking flags** — `dbaas create` now accepts networking-related
  flags.
- **Schedule `--step-*` flags** — `schedule job create` payload now includes
  step parameters.

### Changed

- **sdk-go bumped to v0.3.0** — the SDK was upgraded from v0.1.x through v0.2.x
  to v0.3.0:
  - Adopts natural-language setter vocabulary (`InProject`, `InVPC`,
    `RetaggedAs`, `BilledBy`, `HighlyAvailable()`, `Enabled()`/`Disabled()`)
    replacing the v0.2.x imperative names (`IntoProject`, `IntoVPC`,
    `ReplaceTags`, `WithBillingPeriod`, `WithHA`, `WithEnabled`).
  - All resource families migrated to the SDK wrapper API (compute, network,
    storage, database, container, schedule, security, management).
  - URI segment casing aligned to SDK v0.3.0: `keyPairs`, `blockStorages`,
    `securityGroups`, `securityRules`, `loadBalancers`.
  - Both local vendor patches removed — resolved upstream in v0.3.0 (`securityGroups`
    URI casing and `List[T].Raw()` JSON marshalability).
- **Resource IDs instead of full URIs** — all commands now accept bare resource
  IDs; full URI strings are no longer required.
- **Output marshalers** — `--output json` and `--output yaml` now delegate
  directly to SDK `RawJSON()`/`RawYAML()` rather than round-tripping through
  `encoding/json`. Note: YAML key casing is now authoritative from the SDK;
  consumers relying on the previous camelCase round-trip casing should verify
  downstream tooling. `--output table-yaml` (snake_case rows) is unaffected.
- **`RunE` decomposition** — all 137 command handlers refactored into `Args`,
  `operation`, and `Run` layers for improved testability and consistency.

### Fixed

- **Database**: re-populate engine catalog ID and zone from API response before
  DBaaS update; prevents spurious 400 errors (closes #153).
- **Container**: registry provisioning timeout extended to 900 s; skipped poll
  steps are now surfaced in output (closes #155).
- **Compute**: security groups from `linkedResources` are now re-injected on
  `cloudserver update`; previously they were silently dropped (closes #125).
- **Network**: VPN route Cloud Subnet field was always blank on `get`/`list`;
  fixed via `.CIDR` accessor (TD-030).
- **Network**: VPN tunnel `--subnet-cidr` semantics clarified; flag is a routing
  reference, not a network configuration parameter (TD-025).
- **Network**: VPN tunnel `update` now preserves existing client settings when
  not explicitly changed.
- **Compute**: `cloudserver update` tag-change detection now uses
  `Flags().Changed`; previously tags were unconditionally re-applied.
- **Compute**: `keypair create` now sets the region on the request payload.
- **Schedule**: `job create` now correctly wires `--step-*` flag values into the
  API payload; SDK constants used for flag defaults.
- **Security**: duplicate KMS `Delete` API call removed.
- **E2e**: empty-string elements in array flags removed with exact-match filter,
  replacing fragile pattern substitution (closes #154).
- **E2e**: retry loop added for VPN route cloud-subnet deletion race condition
  (closes #136).
- **E2e**: project `DELETE` failure now propagates correctly in the management
  suite (closes #128).

[Unreleased]: https://github.com/Arubacloud/acloud-cli/compare/v0.5.1...HEAD
[0.5.1]: https://github.com/Arubacloud/acloud-cli/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/Arubacloud/acloud-cli/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/Arubacloud/acloud-cli/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/Arubacloud/acloud-cli/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/Arubacloud/acloud-cli/compare/v0.1.9...v0.2.0
