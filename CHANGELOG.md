# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **sdk-go bumped to v0.3.0** — adopts the natural-language setter vocabulary
  (`InProject`, `InVPC`, `RetaggedAs`, `BilledBy`, `HighlyAvailable()`,
  `Enabled()`/`Disabled()`) replacing the v0.2.x imperative names
  (`IntoProject`, `IntoVPC`, `ReplaceTags`, `WithBillingPeriod`, `WithHA`,
  `WithEnabled`).

- **Vendor patches removed** — both local patches carried against sdk-go are now
  resolved upstream in v0.3.0 and have been deleted from the vendor tree:
  - `securityGroups` URI segment casing fix (the old `securitygroups` lowercase
    path is now correctly emitted by the SDK).
  - `List[T].Raw()` JSON marshalability fix (the list wrapper now satisfies the
    `rawMarshaler` interface natively).

- **Single-import achieved** — all non-test `cmd/` files now import only
  `github.com/Arubacloud/sdk-go/pkg/aruba`. The sole documented escape hatch is
  `container.kaas.go`, which retains a direct `pkg/types` reference for
  `types.APIServerAccessProfileProperties` until the SDK provides a convenience
  setter for that field.

- **URI segment casing aligned** — resource path segments now use the casing
  dictated by v0.3.0: `keyPairs`, `blockStorages`, `securityGroups`,
  `securityRules`, `loadBalancers`. Previously some segments were all-lowercase,
  which required the local vendor patch to correct.

- **Output handlers delegate to SDK marshalers** — `PrintOutput` for `--output
  json` and `--output yaml` modes now calls `RawJSON()`/`RawYAML()` on the SDK
  wrapper directly (via the local `rawMarshaler` interface) rather than
  round-tripping through `encoding/json`. `*aruba.Project` uses the `rawHTTPer`
  fallback (`RawHTTP()`) because it does not expose `RawJSON`/`RawYAML` in
  v0.3.0.

### Notes

- **YAML output key casing may differ from pre-v0.3.0** — the SDK is now
  authoritative for YAML key names. Consumers of `--output yaml` that relied on
  the previous camelCase JSON→YAML round-trip casing should verify their
  downstream tooling against the new output. `--output table-yaml` (snake_case
  table rows) is unaffected.
