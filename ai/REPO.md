# REPO.md — Repository Organization

## Project Overview

`acloud-cli` is the official CLI for the Aruba Cloud Management Platform, written in Go. It wraps the Aruba Cloud Go SDK (`github.com/Arubacloud/sdk-go`) to expose infrastructure resources as Cobra commands.

## Top-Level Layout

| Path | Purpose |
|------|---------|
| `main.go` | Entry point — calls `cmd.Execute()` |
| `cmd/` | All Cobra command implementations |
| `e2e/` | End-to-end bash test scripts, organized by category |
| `docs/` | Docusaurus documentation site source |
| `Makefile` | Build, test, lint, and release automation |
| `go.mod` / `go.sum` | Go module dependencies |
| `ai/` | Contextual guidance files for Claude Code |

## `cmd/` Package Structure

File naming follows a two-level convention:

- `<category>.go` — registers the parent subcommand (e.g., `storage.go`)
- `<category>.<resource>.go` — implements a specific resource (e.g., `storage.blockstorage.go`)

Shared infrastructure lives in:
- `root.go` — client caching, `GetArubaClient()`, `GetProjectID()`, `PrintOutput()`, `PrintTable()` (shim)
- `config.go` — `LoadConfig()`, `SaveConfig()` for `~/.acloud.yaml`
- `context.go` — context management for `~/.acloud-context.yaml`

## CLI Command Tree

```
acloud [--debug|-d] [--project-id]
├── config set/show
├── context set/use/current/list/delete
├── management project
├── storage blockstorage / snapshot / backup / restore
├── network vpc / subnet / securitygroup / securityrule / elasticip /
│          loadbalancer / vpntunnel / vpnroute / vpcpeering / vpcpeeringroute
├── compute cloudserver / keypair
├── container kaas / containerregistry
├── database dbaas / dbaas user / dbaas database / backup
├── schedule job
└── security kms
```

## User Configuration Files

| File | Purpose |
|------|---------|
| `~/.acloud.yaml` | Credentials: `clientId`, `clientSecret`, optional `baseUrl` / `tokenIssuerUrl` |
| `~/.acloud-context.yaml` | Named project-ID mappings (contexts) |
