---
id: authentication
title: Authentication
sidebar_label: Authentication
description: Configure API credentials and manage multiple profiles for the Aruba Cloud CLI.
---

# Authentication

The Aruba Cloud CLI requires API credentials to authenticate with Aruba Cloud services.

## Setting up Credentials

1. **Obtain API Credentials**: Get your Client ID and Client Secret from the Aruba Cloud console.

2. **Configure the CLI** — pass `--client-id` on the command line; the secret is read securely with echo disabled:
   ```bash
   acloud config set --client-id YOUR_CLIENT_ID
   # Enter client secret: (hidden input, does not appear in shell history)
   ```

   For CI/automation, set the secret via environment variable:
   ```bash
   ACLOUD_CLIENT_SECRET=YOUR_CLIENT_SECRET acloud config set --client-id YOUR_CLIENT_ID
   ```

   > **Security note**: `--client-secret` is intentionally not supported to avoid exposing secrets in process lists and shell history.

3. **Verify configuration**:
   ```bash
   acloud config show
   ```

## Configuration File

Credentials are stored in `~/.config/acloud/config.yaml` (XDG Base Directory, file permissions `0600`):

```yaml
profiles:
  default:
    clientId: your-client-id
    clientSecret: your-client-secret
```

> **Legacy path**: If you used an earlier version of acloud that stored credentials in `~/.acloud.yaml`, the CLI automatically migrates that file to the new location the first time it runs and prints a one-time notice. No manual action is needed.

**Security Note**: Keep your credentials secure. The configuration file contains sensitive information.

## Client Configuration

The CLI configuration allows you to manage API credentials and optional settings like custom API endpoints.

### Setting Configuration

**Required Settings:**

`--client-id` is required. `clientSecret` is sourced from `ACLOUD_CLIENT_SECRET` (automation) or prompted securely with echo disabled (interactive):

```bash
# Recommended: secret entered via hidden prompt (does not appear in shell history)
acloud config set --client-id YOUR_CLIENT_ID

# CI/automation: provide secret via environment variable
ACLOUD_CLIENT_SECRET=YOUR_CLIENT_SECRET acloud config set --client-id YOUR_CLIENT_ID
```

**Optional Settings:**

You can optionally configure custom API endpoints:

```bash
# Set base URL (default: https://api.arubacloud.com)
acloud config set --base-url "https://api.arubacloud.com"

# Set token issuer URL (default: https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token)
acloud config set --token-issuer-url "https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token"
```

You can also set all values at once:

```bash
ACLOUD_CLIENT_SECRET=YOUR_CLIENT_SECRET \
acloud config set \
  --client-id YOUR_CLIENT_ID \
  --base-url "https://api.arubacloud.com" \
  --token-issuer-url "https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token"
```

### Viewing Configuration

```bash
acloud config show
```

Output example:
```
Current configuration:
  Client ID: your-client-id
  Client Secret: ********
  Base URL: https://api.arubacloud.com (default)
  Token Issuer URL: https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token (default)
```

### Configuration File Format

The configuration is stored in `~/.config/acloud/config.yaml` using a multi-profile envelope:

```yaml
profiles:
  default:
    clientId: your-client-id
    clientSecret: your-client-secret
    baseUrl: https://api.arubacloud.com          # optional
    tokenIssuerUrl: https://login.aruba.it/...   # optional
  prod:
    clientId: prod-client-id
    clientSecret: prod-client-secret
```

**Default Values:**

If `baseUrl` and `tokenIssuerUrl` are not specified, the CLI uses these defaults:
- **Base URL**: `https://api.arubacloud.com`
- **Token Issuer URL**: `https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token`

### Updating Configuration

`acloud config set` merges changes onto the existing configuration; fields not provided are preserved.

```bash
# Rotate credentials — client-id and client-secret are a matched pair.
# Changing --client-id always asks for a new secret (hidden prompt or ACLOUD_CLIENT_SECRET).
acloud config set --client-id NEW_CLIENT_ID
# Enter client secret: (hidden input)

# Same, non-interactively via environment variable
ACLOUD_CLIENT_SECRET=NEW_SECRET acloud config set --client-id NEW_CLIENT_ID

# Update only optional fields — credentials are untouched
acloud config set --base-url "https://custom-api.example.com"
acloud config set --token-issuer-url "https://custom-idp.example.com/token"
```

**Note**: When `--client-id` is provided, the CLI always collects a new client-secret — either from `ACLOUD_CLIENT_SECRET` or via the interactive prompt. This ensures the stored credentials remain a matched pair. To update only `--base-url` or `--token-issuer-url` without touching credentials, omit `--client-id`.

## Multi-Profile Credential Management

When you work with multiple Aruba Cloud accounts — for example a personal account, a staging environment, and a production environment — profiles let you store each set of credentials under a named key and switch between them with a single flag.

### Creating a Profile

Use `acloud config profile set <name>` to create or update a named profile. The client secret is read from `ACLOUD_CLIENT_SECRET` (recommended for automation) or prompted securely with echo disabled:

```bash
# Create a "staging" profile — secret entered interactively
acloud config profile set staging --client-id YOUR_STAGING_CLIENT_ID

# Create a "prod" profile — secret from environment variable
ACLOUD_CLIENT_SECRET=YOUR_PROD_SECRET \
  acloud config profile set prod \
  --client-id YOUR_PROD_CLIENT_ID \
  --base-url "https://api.arubacloud.com"
```

You can update a single field of an existing profile without touching the other fields:

```bash
# Update only the base URL of the prod profile — credentials are preserved
acloud config profile set prod --base-url "https://custom-api.example.com"
```

> **Credential rotation**: to rotate the credentials of the default profile use `acloud config set --client-id NEW_ID` (always prompts for a new secret). For named profiles use `acloud config profile set <name> --client-id NEW_ID --client-secret NEW_SECRET` (or `ACLOUD_CLIENT_SECRET=NEW_SECRET acloud config profile set <name> --client-id NEW_ID`).

### Selecting the Active Profile

Three ways to select which profile a command uses, in order of precedence:

| Method | Example | Notes |
|--------|---------|-------|
| `--profile` flag | `acloud --profile prod network vpc list` | Highest priority; overrides the env var |
| `ACLOUD_PROFILE` env var | `ACLOUD_PROFILE=staging acloud storage blockstorage list` | Useful in CI/CD pipelines |
| Default | *(no flag or env var)* | Uses the `default` profile |

```bash
# One-off command against prod
acloud --profile prod management project list

# Set profile for the whole shell session
export ACLOUD_PROFILE=staging
acloud network vpc list
acloud storage blockstorage list

# Restore default behaviour
unset ACLOUD_PROFILE
```

### Listing Profiles

```bash
acloud config profile list
```

Example output (active profile marked with `*`):

```
PROFILE              CLIENT_ID                        BASE_URL
* default            default-client-id                https://api.arubacloud.com
  prod               prod-client-id                   https://api.arubacloud.com
  staging            staging-client-id                https://api.arubacloud.com
```

Profiles that do not have an explicit `baseUrl` in the config file display the default (`https://api.arubacloud.com`).

### Deleting a Profile

```bash
acloud config profile delete staging
# Profile "staging" deleted.
```

### Config File Format (Multi-Profile)

All profiles are stored together in `~/.config/acloud/config.yaml` under a `profiles:` key:

```yaml
profiles:
  default:
    clientId: default-client-id
    clientSecret: default-secret
  prod:
    clientId: prod-client-id
    clientSecret: prod-secret
    baseUrl: https://api.arubacloud.com
  staging:
    clientId: staging-client-id
    clientSecret: staging-secret
```

> **Backward compatibility**: Existing single-profile config files (the old flat `clientId: / clientSecret:` format) continue to work and are automatically treated as the `default` profile. They are **not** rewritten until you run `acloud config profile set` or `acloud config set`, at which point they are converted to multi-profile format.

### Using Profiles with Context Management

Profiles (credentials) and contexts (project IDs) are independent — you can combine them freely:

```bash
# Use prod credentials + a project ID from a saved context
acloud --profile prod context use my-prod-project
acloud --profile prod network vpc list

# Or pass the project ID explicitly
acloud --profile prod network vpc list --project-id YOUR_PROJECT_ID
```

## Troubleshooting

### "Error initializing client"

This usually means credentials are not configured. Run:

```bash
acloud config set
```

### "No projects found"

Ensure your credentials have the correct permissions and you have projects in your account.
