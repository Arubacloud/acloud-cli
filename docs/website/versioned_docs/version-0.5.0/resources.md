# Resources

This section provides comprehensive documentation for all Aruba Cloud CLI resources.

## Resource Categories

### [Management Resources](resources/management.md)

Manage projects and organizational resources.

- [Projects](resources/management/project.md) - Create and manage projects

### [Storage Resources](resources/storage.md)

Manage block storage, snapshots, backups, and restore operations.

- [Block Storage](resources/storage/blockstorage.md) - Persistent storage volumes
- [Snapshots](resources/storage/snapshot.md) - Point-in-time copies
- [Backups](resources/storage/backup.md) - Advanced backup operations
- [Restore Operations](resources/storage/restore.md) - Restore from backups

### [Network Resources](resources/network.md)

Manage virtual private clouds, networking, and security.

- [VPC](resources/network/vpc.md) - Virtual Private Clouds
- [Subnet](resources/network/subnet.md) - Network subnets
- [Security Group](resources/network/securitygroup.md) - Security groups
- [Security Rule](resources/network/securityrule.md) - Firewall rules
- [Elastic IP](resources/network/elasticip.md) - Public IP addresses
- [Load Balancer](resources/network/loadbalancer.md) - Load balancing
- [VPC Peering](resources/network/vpcpeering.md) - VPC connections
- [VPC Peering Route](resources/network/vpcpeeringroute.md) - Peering routes
- [VPN Tunnel](resources/network/vpntunnel.md) - VPN connections
- [VPN Route](resources/network/vpnroute.md) - VPN routing

### [Database Resources](resources/database.md)

Manage database services, databases, users, and backups.

- [DBaaS](resources/database/dbaas.md) - Database as a Service instances
- [DBaaS Databases](resources/database/dbaas.database.md) - Databases within DBaaS
- [DBaaS Users](resources/database/dbaas.user.md) - Database users
- [Database Backups](resources/database/backup.md) - Database backup operations

### [Schedule Resources](resources/schedule.md)

Manage scheduled jobs for automation.

- [Jobs](resources/schedule/job.md) - Scheduled jobs (OneShot and Recurring)

### [Security Resources](resources/security.md)

Manage security and encryption resources.

- [KMS Keys](resources/security/kms.md) - Key Management System keys

### [Compute Resources](resources/compute.md)

Manage compute resources and SSH key pairs.

- [Cloud Servers](resources/compute/cloudserver.md) - Virtual machine instances
- [Key Pairs](resources/compute/keypair.md) - SSH key pair management

### [Container Resources](resources/container.md)

Manage container and Kubernetes resources.

- [KaaS](resources/container/kaas.md) - Kubernetes as a Service clusters
- [Container Registry](resources/container/containerregistry.md) - Private Docker container registry

## Quick Reference

All resources support standard CRUD operations:

- **List**: `acloud <category> <resource> list`
- **Get**: `acloud <category> <resource> get <id>`
- **Create**: `acloud <category> <resource> create [flags]`
- **Update**: `acloud <category> <resource> update <id> [flags]`
- **Delete**: `acloud <category> <resource> delete <id> [--yes]`

For detailed information about each resource, see the specific resource documentation pages.

