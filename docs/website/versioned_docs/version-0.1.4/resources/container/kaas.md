# KaaS (Kubernetes as a Service)

KaaS provides managed Kubernetes clusters for running containerized applications in Aruba Cloud.

## Command Syntax

```bash
acloud container kaas <command> [flags] [arguments]
```

## Available Commands

### `create`

Create a new KaaS cluster.

**Syntax:**
```bash
acloud container kaas create [flags]
```

**Required Flags:**
- `--name <string>` - Name for the KaaS cluster
- `--region <string>` - Region code (e.g., `ITBG-Bergamo`)
- `--vpc-uri <string>` - VPC URI (e.g., `/projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id}`)
- `--subnet-uri <string>` - Subnet URI (e.g., `/projects/{project-id}/providers/Aruba.Network/subnets/{subnet-id}`)
- `--node-cidr-address <string>` - Node CIDR address in CIDR notation (e.g., `10.0.0.0/16`)
- `--node-cidr-name <string>` - Node CIDR name
- `--security-group-name <string>` - Security group name
- `--kubernetes-version <string>` - Kubernetes version (e.g., `1.28.0`)
- `--node-pool-name <string>` - Node pool name
- `--node-pool-nodes <int>` - Number of nodes in the node pool
- `--node-pool-instance <string>` - Instance configuration name for nodes
- `--node-pool-zone <string>` - Datacenter/zone code for nodes

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--tags <stringSlice>` - Tags (comma-separated)
- `--pod-cidr <string>` - Pod CIDR (optional)
- `--ha` - Enable high availability
- `--billing-period <string>` - Billing period: Hour, Month, Year (optional)
- `--api-server-authorized-ip-ranges <stringSlice>` - Authorized IP ranges for API server access
- `--api-server-enable-private-cluster` - Enable private cluster for API server
- `--node-pool-autoscaling` - Enable autoscaling for node pool
- `--node-pool-min-count <int>` - Minimum number of nodes for autoscaling
- `--node-pool-max-count <int>` - Maximum number of nodes for autoscaling

**Example:**
```bash
acloud container kaas create \
  --name "production-cluster" \
  --region "ITBG-Bergamo" \
  --vpc-uri "/projects/66a10244f62b99c686572a9f/providers/Aruba.Network/vpcs/69495ef64d0cdc87949b71ec" \
  --subnet-uri "/projects/66a10244f62b99c686572a9f/providers/Aruba.Network/subnets/694b05ac4d0cdc87949b75f9" \
  --node-cidr-address "10.0.0.0/16" \
  --node-cidr-name "node-cidr" \
  --security-group-name "kaas-sg" \
  --kubernetes-version "1.28.0" \
  --node-pool-name "default-pool" \
  --node-pool-nodes 3 \
  --node-pool-instance "small" \
  --node-pool-zone "ITBG-Bergamo-A" \
  --tags "production,kubernetes" \
  --ha
```

### `list`

List all KaaS clusters in the project.

**Syntax:**
```bash
acloud container kaas list [flags]
```

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**Example:**
```bash
acloud container kaas list
```

**Output:**
The command displays a table with the following columns:
- ID
- NAME
- VERSION
- REGION
- STATUS

### `get`

Get detailed information about a specific KaaS cluster.

**Syntax:**
```bash
acloud container kaas get <cluster-id> [flags]
```

**Arguments:**
- `<cluster-id>` - The ID of the KaaS cluster

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--verbose` - Show detailed JSON output

**Example:**
```bash
acloud container kaas get 69495ef64d0cdc87949b71ec
```

**Output:**
The command displays detailed information including:
- ID and URI
- Name and region
- Kubernetes version
- Status
- Creation date and creator
- Tags

### `update`

Update a KaaS cluster's metadata and properties.

**Syntax:**
```bash
acloud container kaas update <cluster-id> [flags]
```

**Arguments:**
- `<cluster-id>` - The ID of the KaaS cluster

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--name <string>` - New name for the KaaS cluster
- `--tags <stringSlice>` - New tags (comma-separated)
- `--kubernetes-version <string>` - Kubernetes version to upgrade to
- `--kubernetes-version-upgrade-date <string>` - Upgrade date for Kubernetes version (ISO 8601 format)
- `--ha` - Enable/disable high availability
- `--storage-max-cumulative-volume-size <int>` - Maximum cumulative volume size for storage
- `--billing-period <string>` - Billing period: Hour, Month, Year
- `--node-pool-name <string>` - Node pool name to update
- `--node-pool-nodes <int>` - Number of nodes in the node pool
- `--node-pool-instance <string>` - Instance configuration name for nodes
- `--node-pool-zone <string>` - Datacenter/zone code for nodes
- `--node-pool-autoscaling` - Enable autoscaling for node pool
- `--node-pool-min-count <int>` - Minimum number of nodes for autoscaling
- `--node-pool-max-count <int>` - Maximum number of nodes for autoscaling

**Example:**
```bash
# Update metadata (name and tags)
acloud container kaas update 69495ef64d0cdc87949b71ec \
  --name "production-cluster-updated" \
  --tags "production,kubernetes,updated"

# Update Kubernetes version
acloud container kaas update 69495ef64d0cdc87949b71ec \
  --kubernetes-version "1.29.0"

# Update node pool
acloud container kaas update 69495ef64d0cdc87949b71ec \
  --node-pool-name "default-pool" \
  --node-pool-nodes 5 \
  --node-pool-autoscaling \
  --node-pool-min-count 3 \
  --node-pool-max-count 10
```

**Note:** You can update metadata (name, tags) and properties (Kubernetes version, node pools, etc.) in the same command.

### `delete`

Delete a KaaS cluster.

**Syntax:**
```bash
acloud container kaas delete <cluster-id> [flags]
```

**Arguments:**
- `<cluster-id>` - The ID of the KaaS cluster

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)
- `--yes, -y` - Skip confirmation prompt

**Example:**
```bash
acloud container kaas delete 69495ef64d0cdc87949b71ec --yes
```

**Warning:** Deleting a cluster will remove all workloads and data. Ensure you have backups if needed.

### `connect`

Connect to a KaaS cluster and configure kubectl automatically.

**Syntax:**
```bash
acloud container kaas connect <cluster-id> [flags]
```

**Arguments:**
- `<cluster-id>` - The ID of the KaaS cluster

**Optional Flags:**
- `--project-id <string>` - Project ID (uses context if not specified)

**Example:**
```bash
acloud container kaas connect 69495ef64d0cdc87949b71ec
```

**What it does:**
1. Downloads the kubeconfig from the KaaS cluster
2. Creates a kubeconfig file in `$HOME/.kube/` named with the cluster name
3. Updates `$HOME/.kube/config` with the cluster configuration
4. Runs `kubectl cluster-info` to verify the connection
5. Prints success message if connection is successful, or error if it fails

**Prerequisites:**
- `kubectl` must be installed and available in your PATH
- The cluster must be in a ready state

**Output:**
```
KaaS successfully connected
Kubeconfig saved to: /home/user/.kube/production-cluster
Default config updated: /home/user/.kube/config
```

## Auto-completion

The CLI provides auto-completion for KaaS cluster IDs:

```bash
acloud container kaas get <TAB>
acloud container kaas update <TAB>
acloud container kaas delete <TAB>
acloud container kaas connect <TAB>
```

## Common Workflows

### Creating a New Cluster

1. **Create the KaaS cluster**:
   ```bash
   acloud container kaas create \
     --name "my-k8s-cluster" \
     --region "ITBG-Bergamo" \
     --version "1.28.0"
   ```

2. **Wait for the cluster to be ready**:
   ```bash
   # Check cluster status
   acloud container kaas get <cluster-id>
   # Wait until status is "Active"
   ```

3. **Connect to the cluster** (after cluster is ready):
   ```bash
   # Automatically configure kubectl
   acloud container kaas connect <cluster-id>
   
   # Verify connection:
   kubectl get nodes
   kubectl cluster-info
   ```

### Updating Cluster

```bash
# Update cluster name and tags
acloud container kaas update <cluster-id> \
  --name "production-cluster" \
  --tags "production,kubernetes,updated"

# Update Kubernetes version
acloud container kaas update <cluster-id> \
  --kubernetes-version "1.29.0"

# Scale node pool
acloud container kaas update <cluster-id> \
  --node-pool-name "default-pool" \
  --node-pool-nodes 5
```

### Connecting to a Cluster

```bash
# Connect and configure kubectl automatically
acloud container kaas connect <cluster-id>

# After connection, use kubectl normally
kubectl get nodes
kubectl get pods --all-namespaces
```

### Listing and Filtering Clusters

```bash
# List all clusters
acloud container kaas list

# Use grep to filter by name or tags
acloud container kaas list | grep "production"
```

## Best Practices

- **Naming**: Use descriptive names that indicate the cluster's purpose (e.g., `prod-k8s-cluster`, `staging-cluster`)
- **Tags**: Use tags to organize clusters by environment, project, or team
- **Versioning**: 
  - Keep Kubernetes versions up to date for security and features
  - Test version upgrades in staging before production
- **Resource Planning**: Plan cluster resources based on workload requirements
- **Monitoring**: Monitor cluster status, health, and resource usage
- **Security**: 
  - Follow Kubernetes security best practices
  - Use RBAC for access control
  - Enable network policies
  - Regularly update cluster components
- **Backup**: Implement backup strategies for persistent volumes and application data
- **Cleanup**: Delete unused clusters to avoid unnecessary costs

## Related Resources

- [Network Resources](../network.md) - Configure networking for container workloads
- [Storage Resources](../storage.md) - Persistent volumes for container applications
- [Security Resources](../security.md) - Security policies and key management
- [Compute Resources](../compute.md) - Cloud servers that can run container workloads

