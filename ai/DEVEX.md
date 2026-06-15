# DEVEX.md — Development & Testing

## Build

```bash
make build              # Build for current platform
make build-all          # Build for all platforms (Windows, Linux, macOS Intel/ARM)
```

## Lint & Format

```bash
make lint               # Run fmt + vet
make lint-check         # CI-mode formatting check (fails if code needs formatting)
```

## Test

```bash
make test               # Run all unit tests (verbose)
make test-short         # Quick test run
make test-coverage      # Generate HTML coverage report
make test-race          # Run with race detector
make test-skip-client   # Skip tests requiring live API credentials
```

## E2E Tests

```bash
make e2e-test           # All E2E tests (requires credentials)
make e2e-management     # E2E tests for management resources
make e2e-storage        # E2E tests for storage resources
make e2e-network        # E2E tests for network resources
```

## CI

```bash
make ci                 # lint + mod-verify + test-skip-client
make pre-commit         # fmt + vet + tests
```

## Release

Releases are published by GoReleaser via the `release.yml` workflow on `v*` tag push.

```bash
goreleaser check                # validate .goreleaser.yaml without building
make release-snapshot           # build all artifacts locally (no tag/publish required)
                                # artifacts land in dist/
```

Install GoReleaser: `go install github.com/goreleaser/goreleaser/v2@latest`

## Testing Conventions

- Unit tests use `t.TempDir()` for file isolation.
- Set `ACLOUD_TEST_SKIP_CLIENT=true` to skip tests that require live API credentials (used in CI).
- E2E tests are bash scripts under `e2e/` organized by resource category.
- Test files: `<file>_test.go` for standard tests; `<file>_test_enhanced.go` for extended fixtures/scenarios.

## AI Agent Pre-PR Gate

When an AI agent (Copilot/Claude) prepares a PR, it must complete this gate **before** opening a new issue branch or moving to the next issue:

1. Run local tests for touched packages and full repository tests:
    ```bash
    go test ./... -count=1
    ```
2. Generate coverage and verify touched code paths are exercised:
    ```bash
    go test ./cmd -covermode=count -coverprofile=/tmp/cmd.cover.out
    go tool cover -func=/tmp/cmd.cover.out
    ```
3. After pushing and opening the PR, verify check status and wait for completion:
    ```bash
    gh pr checks <PR_NUMBER> --repo Arubacloud/acloud-cli
    ```
4. If `codecov/patch` is failing, add targeted tests for changed lines and push again.
5. Do not start the next issue until `codecov/patch` and required checks are green (or maintainer explicitly approves an exception).
