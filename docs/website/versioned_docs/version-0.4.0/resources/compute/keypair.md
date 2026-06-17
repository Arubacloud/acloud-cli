# Key Pairs

SSH key pairs provide secure authentication to cloud servers without using passwords.

## Command Syntax

```bash
acloud compute keypair <command> [flags] [arguments]
```

## Available Commands

### `create`

Create a new SSH key pair.

**Syntax:**
```bash
acloud compute keypair create [flags]
```

**Required Flags:**
- `--name <string>` - Name for the key pair
- `--public-key <string>` - Public key value (SSH public key)
- `--region <string>` - Region code (e.g., `ITBG-Bergamo`)

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**Example:**
```bash
# Using a file
acloud compute keypair create \
  --name "my-keypair" \
  --region "ITBG-Bergamo" \
  --public-key "$(cat ~/.ssh/id_rsa.pub)"

# Or directly
acloud compute keypair create \
  --name "my-keypair" \
  --region "ITBG-Bergamo" \
  --public-key "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... user@host"
```

### `list`

List all key pairs in the project.

**Syntax:**
```bash
acloud compute keypair list [flags]
```

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**Example:**
```bash
acloud compute keypair list
```

**Output:**
The command displays a table with the following columns:
- NAME
- ID
- PUBLIC_KEY (truncated to 50 characters)

### `get`

Get detailed information about a specific key pair.

**Syntax:**
```bash
acloud compute keypair get <keypair-name> [flags]
```

**Arguments:**
- `<keypair-name>` - The name of the key pair

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--verbose` - Show detailed JSON output

**Example:**
```bash
acloud compute keypair get "my-keypair"
```

**Output:**
The command displays:
- Name
- URI
- Public key (full value)
- Creation date and creator

### `update`

**Not supported.** The API does not support updating a key pair's public key. Running this command prints an error and exits without making any changes.

To rotate a key pair, delete it and recreate it with the same name:

```bash
acloud compute keypair delete "my-keypair" --yes
acloud compute keypair create \
  --name "my-keypair" \
  --region "ITBG-Bergamo" \
  --public-key "$(cat ~/.ssh/id_rsa_new.pub)"
```

### `delete`

Delete a key pair.

**Syntax:**
```bash
acloud compute keypair delete <keypair-name> [flags]
```

**Arguments:**
- `<keypair-name>` - The name of the key pair

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt

**Example:**
```bash
acloud compute keypair delete "my-keypair" --yes
```

## Auto-completion

The CLI provides auto-completion for key pair names:

```bash
acloud compute keypair get <TAB>
acloud compute keypair update <TAB>
acloud compute keypair delete <TAB>
```

## Common Workflows

### Creating a Key Pair from Existing SSH Key

```bash
# If you already have an SSH key pair
acloud compute keypair create \
  --name "my-laptop-key" \
  --region "ITBG-Bergamo" \
  --public-key "$(cat ~/.ssh/id_rsa.pub)"
```

### Generating a New Key Pair

1. **Generate a new SSH key pair** (if needed):
   ```bash
   ssh-keygen -t rsa -b 4096 -f ~/.ssh/aruba_key -N ""
   ```

2. **Create the key pair in Aruba Cloud**:
   ```bash
   acloud compute keypair create \
     --name "aruba-key" \
     --region "ITBG-Bergamo" \
     --public-key "$(cat ~/.ssh/aruba_key.pub)"
   ```

3. **Use the key pair when creating cloud servers**:
   ```bash
   acloud compute cloudserver create \
     --name "my-server" \
     --region "ITBG-Bergamo" \
     --zone "ITBG-1" \
     --flavor "CSO4A8" \
     --boot-disk-id "<volume-id>" \
     --vpc-id "<vpc-id>" \
     --subnet-id "<subnet-id>" \
     --security-group-id "<sg-id>" \
     --keypair-id "<keypair-id>"
   ```

### Rotating Keys

Since keypair update is not supported by the API, key rotation requires delete + recreate:

1. **Generate a new SSH key pair**:
   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/aruba_key_new -N ""
   ```

2. **Delete the existing key pair**:
   ```bash
   acloud compute keypair delete "aruba-key" --yes
   ```

3. **Recreate it with the new public key**:
   ```bash
   acloud compute keypair create \
     --name "aruba-key" \
     --region "ITBG-Bergamo" \
     --public-key "$(cat ~/.ssh/aruba_key_new.pub)"
   ```

4. **Update your local SSH config** to use the new private key and test the connection.

## Best Practices

- **Naming**: Use descriptive names that indicate the key's purpose or owner (e.g., `user-john-laptop`, `ci-cd-server`, `admin-key`)
- **Key Security**: 
  - Never share or expose private keys
  - Use strong key types (RSA 4096-bit or Ed25519)
  - Protect private keys with passphrases
- **Key Rotation**: Rotate keys regularly for security
- **Multiple Keys**: Use different key pairs for different environments or purposes
- **Backup**: Keep secure backups of your private keys
- **Access Control**: Only grant key pair access to authorized users

## Related Resources

- [Cloud Servers](./cloudserver.md) - Use key pairs when creating servers
- [Security Resources](../security.md) - Additional security best practices

