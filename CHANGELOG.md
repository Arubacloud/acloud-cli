# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/Arubacloud/acloud-cli/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/Arubacloud/acloud-cli/compare/v0.1.9...v0.2.0
