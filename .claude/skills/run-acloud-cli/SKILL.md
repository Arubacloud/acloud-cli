---
name: run-acloud-cli
description: Run, build, test, and smoke-test the acloud CLI. Use this skill to start the binary, verify output, confirm a fix works, or run the unit test suite.
---

`acloud` is a statically linked Go CLI that wraps the Aruba Cloud Management Platform API via Cobra. There is no server to start and no GUI — the interaction surface is the binary's stdout/stderr.

The driver is `.claude/skills/run-acloud-cli/smoke.sh`. It builds the binary, runs representative read-only commands across every resource category, validates JSON/YAML output, and runs the unit test suite. Run it to verify a change end-to-end.

## Prerequisites

- WSL (Ubuntu-24.04) with Go at `~/go-dist/go/` — the binary in the repo root is a pre-built Linux ELF.
- Valid credentials in `~/.acloud.yaml` (clientId + clientSecret) — list/get commands hit the real API.
- The Bash tool runs in MSYS2/Windows; execute all commands via `wsl.exe`:

```bash
wsl.exe -d Ubuntu-24.04 -- bash -c "export PATH=\"\$HOME/go-dist/go/bin:\$PATH\"; cd ~/Software/acloud-cli; <cmd>"
```

## Build

```bash
wsl.exe -d Ubuntu-24.04 -- bash -c "export PATH=\"\$HOME/go-dist/go/bin:\$PATH\"; cd ~/Software/acloud-cli && make build"
# → produces ./acloud (Linux ELF, ~12 MB, statically linked)
```

## Run (agent path) — smoke driver

Run the smoke script to verify a change is working:

```bash
wsl.exe -d Ubuntu-24.04 -- bash -c "export PATH=\"\$HOME/go-dist/go/bin:\$PATH\"; cd ~/Software/acloud-cli && bash .claude/skills/run-acloud-cli/smoke.sh"
```

The script:
1. Builds the binary (`make build`)
2. Runs `--version`, `--help`, `config show`, `context list`
3. Calls `management project list` and validates table / JSON / YAML output
4. Calls `list` on every resource category (storage, compute, container, database, security, network) — accepts empty lists
5. Runs the unit test suite with `ACLOUD_TEST_SKIP_CLIENT=true`
6. Prints `Results: N passed, N failed` and exits non-zero on any failure

Expected final line: `Results: 14 passed, 0 failed`

## Run (human path)

```bash
# In WSL terminal:
cd ~/Software/acloud-cli
./acloud --help
./acloud management project list
./acloud storage blockstorage list --output json
```

## Direct invocation (for testing a single command)

```bash
wsl.exe -d Ubuntu-24.04 -- bash -c "cd ~/Software/acloud-cli && ./acloud <subcommand> [flags]"
# e.g.:
wsl.exe -d Ubuntu-24.04 -- bash -c "cd ~/Software/acloud-cli && ./acloud management project list --output yaml"
```

## Unit tests (no live credentials)

```bash
wsl.exe -d Ubuntu-24.04 -- bash -c "export PATH=\"\$HOME/go-dist/go/bin:\$PATH\"; cd ~/Software/acloud-cli && make test-skip-client"
```

## Gotchas

- **MSYS2 can't exec the Linux ELF.** The `acloud` binary in the repo root is a Linux ELF; `./acloud` in MSYS2/Git Bash gives `Exec format error`. Always wrap in `wsl.exe -d Ubuntu-24.04 -- bash -c "..."`.
- **`network vpc list` returns HTTP 404 when the active project has no VPCs.** The API returns 404 (not an empty list) for the VPC resource when there are no results. The smoke script treats 404 as acceptable for that command.
- **Active context may point at a cleaned-up e2e project.** If `context list` shows `e2e-test-context` as current and resource commands return 404/not found, the project was deleted by a previous e2e run. Switch context: `./acloud context use <other-context>` or run with `--project-id <id>`.
- **Go must be added to PATH explicitly.** The WSL shell's `~/.bashrc` doesn't add `~/go-dist/go/bin` automatically for non-interactive `wsl.exe` invocations. Always prefix with `export PATH="$HOME/go-dist/go/bin:$PATH"`.
- **Build produces a Linux binary even on Windows.** `make build` in WSL produces `./acloud` (Linux ELF). Use `make build-windows` to produce `acloud.exe` if needed for native Windows testing.

## Troubleshooting

| Symptom | Fix |
|---|---|
| `Exec format error` when running `./acloud` | You're in MSYS2/Windows — use `wsl.exe -d Ubuntu-24.04 -- bash -c "cd ... && ./acloud"` |
| `go: command not found` in WSL | Add `export PATH="$HOME/go-dist/go/bin:$PATH"` before `make build` |
| `API error (status 401)` on any command | Credentials in `~/.acloud.yaml` are missing or expired — `./acloud config set --client-id ... --client-secret ...` |
| `API error (status 404): Not Found` on `network vpc list` | Expected: no VPCs in this project. The smoke script accepts this. |
| Context commands fail / wrong project | Active context points at a deleted project — `./acloud management project list` to find a valid ID, then `./acloud context set ... --project-id <id>` |
