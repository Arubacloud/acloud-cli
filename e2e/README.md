# End-to-End (E2E) Testing

This directory contains end-to-end test scripts for validating CRUD operations across all resource categories in the Aruba Cloud CLI.

## Overview

The E2E tests are organized by resource category. Each category has its own test script that validates CRUD operations for all resources in that category:

| Category | Test Script | Resources Tested |
|----------|-------------|------------------|
| **[Management](management/)** | [management/test.sh](management/test.sh) | Projects |
| **[Storage](storage/)** | [storage/test.sh](storage/test.sh) | Block Storage, Snapshots, Backups, Restores |
| **[Network](network/)** | [network/test.sh](network/test.sh) | VPCs, Subnets, Security Groups, Security Rules, Elastic IPs, VPC Peering, VPN Tunnels, VPN Routes |
| **[Database](database/)** | [database/test.sh](database/test.sh) | DBaaS, Databases, Users, Backups |
| **[Schedule](schedule/)** | [schedule/test.sh](schedule/test.sh) | OneShot Jobs, Recurring Jobs |
| **[Security](security/)** | [security/test.sh](security/test.sh) | KMS Keys |
| **[Compute](compute/)** | [compute/test.sh](compute/test.sh) | Cloud Servers, Key Pairs |
| **[Container](container/)** | [container/test.sh](container/test.sh) | KaaS Clusters, Container Registry |

For detailed information about each category, see the [Test Scripts](#test-scripts) section below.

## Prerequisites

Before running the tests, ensure you have:

1. **Configured CLI credentials:**
   ```bash
   acloud config set --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET
   ```

2. **Set required environment variables** (see each test script for specific requirements):
   ```bash
   export ACLOUD_PROJECT_ID="your-project-id"
   export ACLOUD_REGION="ITBG-Bergamo"
   # ... other variables as needed
   ```

3. **Built the CLI:**
   ```bash
   go build -o acloud .
   ```

## Running Tests

### Run All Tests

To run all E2E tests:

```bash
# From the project root
./e2e/management/test.sh
./e2e/storage/test.sh
./e2e/network/test.sh
./e2e/database/test.sh
./e2e/schedule/test.sh
./e2e/security/test.sh
./e2e/compute/test.sh
./e2e/container/test.sh
```

### Run Individual Category Tests

Each category has its own test script:

```bash
# Management resources
./e2e/management/test.sh

# Storage resources
./e2e/storage/test.sh

# Network resources
./e2e/network/test.sh

# Database resources
./e2e/database/test.sh

# Schedule resources
./e2e/schedule/test.sh

# Security resources
./e2e/security/test.sh

# Compute resources
./e2e/compute/test.sh

# Container resources
./e2e/container/test.sh
```

## Test Structure

Each test script follows a consistent structure:

1. **Configuration** - Sets up environment variables and validates prerequisites
2. **CREATE** - Creates test resources
3. **LIST** - Lists created resources
4. **GET** - Retrieves detailed information about resources
5. **UPDATE** - Updates resource properties (name, tags, etc.)
6. **DELETE** - Cleans up test resources

## Test Scripts

Each category has comprehensive test coverage. Click on the category name to jump to its detailed section:

- [Management Tests](#management-tests) - Organization-level resources
- [Storage Tests](#storage-tests) - Storage resources
- [Network Tests](#network-tests) - Network infrastructure
- [Database Tests](#database-tests) - Database services
- [Schedule Tests](#schedule-tests) - Scheduled jobs
- [Security Tests](#security-tests) - Security resources
- [Compute Tests](#compute-tests) - Compute resources
- [Container Tests](#container-tests) - Container orchestration

### Management Tests

Tests organization-level resources:
- **Projects** - Create, list, get, update, delete projects

**Test Script:** [management/test.sh](management/test.sh)

**Related Categories:** All other categories depend on projects for resource creation.

### Storage Tests

Tests storage resources:
- **Block Storage** - Volume creation, updates, deletion
- **Snapshots** - Snapshot creation from volumes
- **Backups** - Backup operations
- **Restores** - Restore operations from backups

**Test Script:** [storage/test.sh](storage/test.sh)

**Related Categories:** 
- Block Storage volumes are used by [Container Registry](#container-tests) tests
- Storage resources are foundational for many compute workloads

### Network Tests

Tests network resources:
- **VPC** - Virtual Private Cloud management
- **Subnet** - Subnet creation and management
- **Security Group** - Security group operations
- **Security Rule** - Firewall rule management
- **Elastic IP** - Public IP address management
- **VPC Peering** - VPC peering connections
- **VPC Peering Route** - Peering route management
- **VPN Tunnel** - VPN tunnel creation and management
- **VPN Route** - VPN route configuration

**Test Script:** [network/test.sh](network/test.sh)

**Related Categories:**
- VPCs and Subnets are required for [Container Tests](#container-tests) (KaaS and Container Registry)
- Elastic IPs are used by [Container Registry](#container-tests) tests
- Security Groups are used by [Container Tests](#container-tests) and [Compute Tests](#compute-tests)

### Database Tests

Tests database resources:
- **DBaaS** - Database as a Service instance management
- **DBaaS Databases** - Database creation and management within DBaaS
- **DBaaS Users** - User management for DBaaS instances
- **Database Backups** - Backup operations for databases

**Test Script:** [database/test.sh](database/test.sh)

**Related Categories:** Database resources are standalone but may integrate with [Compute Tests](#compute-tests) for application deployments.

### Schedule Tests

Tests scheduled job resources:
- **OneShot Jobs** - One-time scheduled jobs
- **Recurring Jobs** - Recurring scheduled jobs with CRON expressions

**Test Script:** [schedule/test.sh](schedule/test.sh)

**Related Categories:** Scheduled jobs can automate operations on resources from [Storage Tests](#storage-tests), [Compute Tests](#compute-tests), and other categories.

### Security Tests

Tests security resources:
- **KMS Keys** - Key Management System key operations

**Test Script:** [security/test.sh](security/test.sh)

**Related Categories:** KMS Keys can be used to encrypt resources from [Storage Tests](#storage-tests) and other categories that support encryption.

### Compute Tests

Tests compute resources:
- **Cloud Servers** - Virtual machine instances
  - Create, list, get, update, delete cloud servers
  - Power on/off operations
  - Password management
  - SSH connection information (connect command)
- **Key Pairs** - SSH key pair management

**Test Script:** [compute/test.sh](compute/test.sh)

**Related Categories:**
- Cloud Servers can use [Network Tests](#network-tests) resources (VPCs, Subnets, Security Groups, Elastic IPs)
- Cloud Servers can attach [Storage Tests](#storage-tests) volumes
- Key Pairs are used for SSH access to Cloud Servers

### Container Tests

Tests container resources:
- **KaaS Clusters** - Kubernetes as a Service cluster management
  - Create, list, get, update, delete KaaS clusters
  - Connect to clusters and configure kubectl (requires kubectl installed)
- **Container Registry** - Private Docker container registry management
  - Create, list, get, update, delete container registries

**Test Script:** [container/test.sh](container/test.sh)

**Related Categories:**
- **Requires [Network Tests](#network-tests) resources:**
  - VPCs and Subnets (required for both KaaS and Container Registry)
  - Elastic IPs (required for Container Registry)
  - Security Groups (required for Container Registry)
- **Requires [Storage Tests](#storage-tests) resources:**
  - Block Storage volumes (required for Container Registry)
- KaaS clusters can deploy applications that use [Database Tests](#database-tests) resources
- Container Registry stores images that can be used by [Compute Tests](#compute-tests) Cloud Servers

**Required Environment Variables for KaaS:**
- `ACLOUD_VPC_ID` - VPC ID for the cluster
- `ACLOUD_SUBNET_ID` - Subnet ID for the cluster
- `ACLOUD_NODE_POOL_INSTANCE` - Instance configuration name for nodes
- `ACLOUD_NODE_POOL_ZONE` - Datacenter/zone code for nodes

**Optional Environment Variables for KaaS:**
- `ACLOUD_NODE_CIDR` - Node CIDR address (default: `10.0.0.0/16`)
- `ACLOUD_NODE_CIDR_NAME` - Node CIDR name (default: `node-cidr`)
- `ACLOUD_SECURITY_GROUP_NAME` - Security group name (default: `kaas-sg`)
- `ACLOUD_NODE_POOL_NAME` - Node pool name (default: `default-pool`)
- `ACLOUD_NODE_POOL_NODES` - Number of nodes (default: `1`)
- `ACLOUD_K8S_VERSION` - Kubernetes version (default: `1.28.0`)

**Required Environment Variables for Container Registry:**
- `ACLOUD_PUBLIC_IP_ID` - Public IP (Elastic IP) ID
- `ACLOUD_VPC_ID` - VPC ID
- `ACLOUD_SUBNET_ID` - Subnet ID
- `ACLOUD_SECURITY_GROUP_ID` - Security group ID
- `ACLOUD_BLOCK_STORAGE_ID` - Block storage ID

**Note:** The KaaS connect test requires `kubectl` to be installed and available in PATH. If kubectl is not found, the connect test will be skipped.

## Environment Variables

Common environment variables used across tests:

| Variable | Description | Example |
|----------|-------------|---------|
| `ACLOUD_PROJECT_ID` | Project ID for resources | `66a10244f62b99c686572a9f` |
| `ACLOUD_REGION` | Region code | `ITBG-Bergamo` |
| `ACLOUD_VPC_ID` | VPC ID for network resources | `69495ef64d0cdc87949b71ec` |
| `ACLOUD_PEER_VPC_ID` | Peer VPC ID for peering | `69485a584d0cdc87949b6ff8` |
| `ACLOUD_VPC_ID` | VPC ID for KaaS clusters and Container Registry | `69495ef64d0cdc87949b71ec` |
| `ACLOUD_SUBNET_ID` | Subnet ID for KaaS clusters and Container Registry | `69495ef64d0cdc87949b71ed` |
| `ACLOUD_NODE_POOL_INSTANCE` | Instance type for KaaS node pool | `small` |
| `ACLOUD_NODE_POOL_ZONE` | Zone for KaaS node pool | `ITBG-Bergamo-A` |
| `ACLOUD_PUBLIC_IP_ID` | Public IP ID for Container Registry | `69495ef64d0cdc87949b71ee` |
| `ACLOUD_SECURITY_GROUP_ID` | Security group ID for Container Registry | `69495ef64d0cdc87949b71ef` |
| `ACLOUD_BLOCK_STORAGE_ID` | Block storage ID for Container Registry | `69495ef64d0cdc87949b71f0` |

See individual test scripts for category-specific variables.

## Debug Mode

To see detailed HTTP requests/responses during tests, use the `--debug` flag:

```bash
# The test scripts can be modified to add --debug to commands
# Or run individual commands with debug:
acloud --debug network vpc create --name test-vpc --region ITBG-Bergamo
```

## Test Output

Each test script provides:
- ✅ **Success indicators** - Green checkmarks for passed tests
- ❌ **Error messages** - Red text for failed operations
- 📊 **Resource information** - Tables showing created/updated resources

## Cleanup

Test scripts attempt to clean up created resources, but if a test fails:
1. Manually delete any remaining test resources
2. Check resource names/IDs in the test output
3. Use the CLI to list and delete orphaned resources

## Troubleshooting

### "Error: project ID not specified"
- Set `ACLOUD_PROJECT_ID` environment variable
- Or use `--project-id` flag in commands
- Or set up a context: `acloud context set my-prod --project-id <id>`

### "Error: Unable to determine region value"
- Ensure `ACLOUD_REGION` is set correctly
- Use the correct region format (e.g., `ITBG-Bergamo`)

### "Failed to create ... - Status: 400"
- Verify all required parameters are provided
- Check that dependent resources exist (e.g., VPC for subnet)
- Review error details in the output

### "Cannot update ... while it is in 'InCreation' state"
- Wait for resources to reach `Active` state before updating
- Some resources take time to provision

## Contributing

When adding new resources or operations:
1. Add test cases to the appropriate category script
2. Follow the existing test structure (CREATE → LIST → GET → UPDATE → DELETE)
3. Include proper cleanup in case of failures
4. Update this README with new test coverage

## Quick Reference

### Test Script Locations

All test scripts are located in their respective category directories:

- `./e2e/management/test.sh` - [Management Tests](#management-tests)
- `./e2e/storage/test.sh` - [Storage Tests](#storage-tests)
- `./e2e/network/test.sh` - [Network Tests](#network-tests)
- `./e2e/database/test.sh` - [Database Tests](#database-tests)
- `./e2e/schedule/test.sh` - [Schedule Tests](#schedule-tests)
- `./e2e/security/test.sh` - [Security Tests](#security-tests)
- `./e2e/compute/test.sh` - [Compute Tests](#compute-tests)
- `./e2e/container/test.sh` - [Container Tests](#container-tests)

### Test Dependencies

Some test categories have dependencies on resources from other categories:

1. **Container Tests** require:
   - Network resources (VPC, Subnet, Elastic IP, Security Group) - See [Network Tests](#network-tests)
   - Storage resources (Block Storage) - See [Storage Tests](#storage-tests)

2. **Compute Tests** can use:
   - Network resources (VPC, Subnet, Security Group, Elastic IP) - See [Network Tests](#network-tests)
   - Storage resources (Block Storage volumes) - See [Storage Tests](#storage-tests)

3. **All tests** require:
   - A project from [Management Tests](#management-tests)

## See Also

- [Main Documentation](../docs/)
- [Getting Started Guide](../docs/getting-started.md)
- [Resource Documentation](../docs/resources/)
- [Management Test Script](management/test.sh)
- [Storage Test Script](storage/test.sh)
- [Network Test Script](network/test.sh)
- [Database Test Script](database/test.sh)
- [Schedule Test Script](schedule/test.sh)
- [Security Test Script](security/test.sh)
- [Compute Test Script](compute/test.sh)
- [Container Test Script](container/test.sh)

