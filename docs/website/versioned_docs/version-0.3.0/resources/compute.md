# Compute Resources

The `compute` category provides commands for managing compute resources in Aruba Cloud, including cloud servers and SSH key pairs.

## Available Resources

### [Cloud Servers](compute/cloudserver.md)

Cloud servers are virtual machine instances that run your applications and workloads.

**Quick Commands:**
```bash
# List all cloud servers
acloud compute cloudserver list

# Get cloud server details
acloud compute cloudserver get <server-id>

# Create a cloud server
acloud compute cloudserver create --name "my-server" --region "ITBG-Bergamo" --flavor "small" --image <image-id>

# Update a cloud server
acloud compute cloudserver update <server-id> --name "new-name"

# Delete a cloud server
acloud compute cloudserver delete <server-id>
```

### [Key Pairs](compute/keypair.md)

SSH key pairs for secure authentication to cloud servers.

**Quick Commands:**
```bash
# List all key pairs
acloud compute keypair list

# Get key pair details
acloud compute keypair get <keypair-name>

# Create a key pair
acloud compute keypair create --name "my-keypair" --public-key "ssh-rsa AAAAB3..."

# Update a key pair (change public key)
acloud compute keypair update <keypair-name> --public-key "ssh-rsa AAAAB3..."

# Delete a key pair
acloud compute keypair delete <keypair-name>
```

## Common Use Cases

### Launching a Cloud Server

1. **Create a key pair** (if you don't have one):
   ```bash
   acloud compute keypair create --name "my-keypair" --public-key "$(cat ~/.ssh/id_rsa.pub)"
   ```

2. **List available flavors and images**:
   ```bash
   # Check available resources (you may need to use the web console or API)
   ```

3. **Create the cloud server**:
   ```bash
   acloud compute cloudserver create \
     --name "web-server" \
     --region "ITBG-Bergamo" \
     --flavor "small" \
     --image "your-image-id" \
     --keypair "my-keypair" \
     --tags "production,web"
   ```

4. **Verify the server**:
   ```bash
   acloud compute cloudserver list
   acloud compute cloudserver get <server-id>
   ```

### Managing SSH Access

1. **List all key pairs**:
   ```bash
   acloud compute keypair list
   ```

2. **Update a key pair** (rotate keys):
   ```bash
   acloud compute keypair update "my-keypair" --public-key "$(cat ~/.ssh/id_rsa_new.pub)"
   ```

3. **Delete unused key pairs**:
   ```bash
   acloud compute keypair delete "old-keypair" --yes
   ```

## Best Practices

- **Key Pairs**:
  - Use descriptive names for key pairs (e.g., `user-john-laptop`, `ci-cd-server`)
  - Rotate keys regularly for security
  - Keep private keys secure and never share them
  - Use different key pairs for different environments

- **Cloud Servers**:
  - Use tags to organize servers by environment, project, or purpose
  - Choose appropriate flavors based on workload requirements
  - Monitor server status before performing updates
  - Use key pairs instead of password authentication for better security

## Related Resources

- [Network Resources](./network.md) - Configure networking for cloud servers
- [Storage Resources](./storage.md) - Attach block storage volumes to servers
- [Security Resources](./security.md) - Manage security groups and rules

