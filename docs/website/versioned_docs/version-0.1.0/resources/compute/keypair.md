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

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**Example:**
```bash
# Using a file
acloud compute keypair create \
  --name "my-keypair" \
  --public-key "$(cat ~/.ssh/id_rsa.pub)"

# Or directly
acloud compute keypair create \
  --name "my-keypair" \
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

Update a key pair's public key (useful for key rotation).

**Syntax:**
```bash
acloud compute keypair update <keypair-name> [flags]
```

**Arguments:**
- `<keypair-name>` - The name of the key pair

**Required Flags:**
- `--public-key <string>` - New public key value

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**Example:**
```bash
acloud compute keypair update "my-keypair" \
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
     --public-key "$(cat ~/.ssh/aruba_key.pub)"
   ```

3. **Use the key pair when creating cloud servers**:
   ```bash
   acloud compute cloudserver create \
     --name "my-server" \
     --region "ITBG-Bergamo" \
     --flavor "small" \
     --image "your-image-id" \
     --keypair "aruba-key"
   ```

### Rotating Keys

1. **Generate a new key pair**:
   ```bash
   ssh-keygen -t rsa -b 4096 -f ~/.ssh/aruba_key_new -N ""
   ```

2. **Update the key pair**:
   ```bash
   acloud compute keypair update "aruba-key" \
     --public-key "$(cat ~/.ssh/aruba_key_new.pub)"
   ```

3. **Update your local SSH config** to use the new private key

4. **Test the connection** to your servers

5. **Delete the old key pair** (if no longer needed):
   ```bash
   acloud compute keypair delete "old-keypair" --yes
   ```

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

