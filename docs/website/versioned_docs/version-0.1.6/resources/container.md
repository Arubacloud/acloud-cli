# Container Resources

The `container` category provides commands for managing container resources in Aruba Cloud, including Kubernetes as a Service (KaaS) clusters and Container Registry.

## Available Resources

### [KaaS (Kubernetes as a Service)](container/kaas.md)

KaaS provides managed Kubernetes clusters for running containerized applications.

**Quick Commands:**
```bash
# List all KaaS clusters
acloud container kaas list

# Get KaaS cluster details
acloud container kaas get <cluster-id>

# Create a KaaS cluster
acloud container kaas create --name "my-cluster" --region "ITBG-Bergamo" --version "1.28.0"

# Update a KaaS cluster
acloud container kaas update <cluster-id> --name "new-name"

# Delete a KaaS cluster
acloud container kaas delete <cluster-id>
```

### [Container Registry](container/containerregistry.md)

Container Registry provides a private Docker container registry for storing and managing container images.

**Quick Commands:**
```bash
# List all container registries
acloud container containerregistry list

# Get container registry details
acloud container containerregistry get <registry-id>

# Create a container registry
acloud container containerregistry create \
  --name "my-registry" \
  --region "ITBG-Bergamo" \
  --public-ip-uri "/projects/{id}/providers/Aruba.Network/elasticIps/{eip-id}" \
  --vpc-uri "/projects/{id}/providers/Aruba.Network/vpcs/{vpc-id}" \
  --subnet-uri "/projects/{id}/providers/Aruba.Network/subnets/{subnet-id}" \
  --security-group-uri "/projects/{id}/providers/Aruba.Network/securityGroups/{sg-id}" \
  --block-storage-uri "/projects/{id}/providers/Aruba.Storage/volumes/{volume-id}"

# Update a container registry
acloud container containerregistry update <registry-id> --name "new-name"

# Delete a container registry
acloud container containerregistry delete <registry-id>
```

## Common Use Cases

### Creating a Kubernetes Cluster

1. **Create a KaaS cluster**:
   ```bash
   acloud container kaas create \
     --name "production-cluster" \
     --region "ITBG-Bergamo" \
     --version "1.28.0" \
     --tags "production,kubernetes"
   ```

2. **Wait for the cluster to be ready** and check status:
   ```bash
   acloud container kaas get <cluster-id>
   ```

3. **Configure kubectl** (after cluster is ready):
   ```bash
   # Get cluster kubeconfig from the Aruba Cloud console or API
   # Then configure kubectl:
   kubectl config use-context <cluster-context>
   ```

### Managing Cluster Metadata

```bash
# Update cluster name and tags
acloud container kaas update <cluster-id> \
  --name "production-cluster-updated" \
  --tags "production,kubernetes,updated"
```

### Creating a Container Registry

1. **Ensure required resources exist:**
   - Public IP (Elastic IP)
   - VPC
   - Subnet
   - Security Group
   - Block Storage

2. **Create the container registry:**
   ```bash
   acloud container containerregistry create \
     --name "my-registry" \
     --region "ITBG-Bergamo" \
     --public-ip-uri "/projects/{id}/providers/Aruba.Network/elasticIps/{eip-id}" \
     --vpc-uri "/projects/{id}/providers/Aruba.Network/vpcs/{vpc-id}" \
     --subnet-uri "/projects/{id}/providers/Aruba.Network/subnets/{subnet-id}" \
     --security-group-uri "/projects/{id}/providers/Aruba.Network/securityGroups/{sg-id}" \
     --block-storage-uri "/projects/{id}/providers/Aruba.Storage/volumes/{volume-id}" \
     --billing-period "Month"
   ```

3. **Wait for the registry to be ready** and check status:
   ```bash
   acloud container containerregistry get <registry-id>
   ```

## Best Practices

- **Naming**: Use descriptive names that indicate the cluster's purpose (e.g., `prod-k8s-cluster`, `staging-cluster`)
- **Tags**: Use tags to organize clusters by environment, project, or team
- **Versioning**: Keep Kubernetes versions up to date for security and features
- **Monitoring**: Monitor cluster status and health regularly
- **Resource Management**: Plan cluster resources based on workload requirements
- **Security**: Follow Kubernetes security best practices for your workloads

## Related Resources

- [Network Resources](./network.md) - Configure networking for container workloads
- [Storage Resources](./storage.md) - Persistent volumes for container applications
- [Security Resources](./security.md) - Security policies and key management

