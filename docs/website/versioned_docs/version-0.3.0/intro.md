# Introduction to Aruba Cloud CLI

Aruba Cloud CLI (`acloud`) is the official command-line interface tool for managing and automating your Aruba Cloud infrastructure. Built with Go and leveraging the Aruba Cloud SDK, it provides a powerful, scriptable interface to interact with all Aruba Cloud services directly from your terminal.

## What is Aruba Cloud CLI?

The Aruba Cloud CLI is a comprehensive tool that enables you to manage your entire Aruba Cloud infrastructure through simple, intuitive commands. Whether you're provisioning resources, managing networks, configuring storage, or automating deployments, the CLI provides a consistent interface across all Aruba Cloud services.

## Key Capabilities

### Infrastructure Management
Manage the full lifecycle of your cloud resources including:
- **Compute Resources**: Cloud servers, key pairs, and container orchestration (KaaS)
- **Storage**: Block storage volumes, snapshots, backups, and restore operations
- **Networking**: Virtual Private Clouds (VPCs), subnets, security groups, elastic IPs, load balancers, and VPN configurations
- **Databases**: Database-as-a-Service (DBaaS) instances, databases, users, and backups
- **Security**: Key Management Service (KMS) keys and encryption
- **Scheduling**: Automated job scheduling and management
- **Projects**: Multi-project management and organization

### Developer Experience
- **Intuitive Commands**: Consistent, predictable command structure across all resources
- **Shell Auto-completion**: Intelligent tab completion for commands, flags, and resource IDs
- **Context Management**: Switch between projects and environments seamlessly
- **Debug Mode**: Detailed logging for troubleshooting and development
- **Scriptable**: Perfect for automation, CI/CD pipelines, and infrastructure-as-code workflows

### Enterprise Features
- **Multi-Project Support**: Manage multiple projects from a single CLI instance
- **Secure Credential Management**: Encrypted storage of API credentials
- **Comprehensive Error Handling**: Clear, actionable error messages
- **Full API Coverage**: Complete CRUD operations for all supported resources

## Why Use Aruba Cloud CLI?

- **Automation**: Integrate Aruba Cloud management into your DevOps workflows and automation scripts
- **Efficiency**: Manage resources faster than using the web console, especially for bulk operations
- **Consistency**: Reproducible infrastructure management through command-line scripts
- **Integration**: Easily integrate with other tools, CI/CD pipelines, and infrastructure management platforms
- **Power Users**: Advanced features for power users who prefer command-line interfaces

## Getting Started

Ready to start using Aruba Cloud CLI? Check out the [Installation](installation.md) guide to get set up on your platform.

## Explore Resources

Once installed, explore the available resources:
- [Management Resources](resources/management.md) - Projects and organization
- [Storage Resources](resources/storage.md) - Block storage, snapshots, backups, and restores
- [Network Resources](resources/network.md) - VPCs, subnets, security groups, elastic IPs, load balancers, and VPN
- [Database Resources](resources/database.md) - DBaaS instances, databases, and users
- [Compute Resources](resources/compute.md) - Cloud servers and key pairs
- [Container Resources](resources/container.md) - Kubernetes as a Service (KaaS)
- [Security Resources](resources/security.md) - Key Management Service
- [Schedule Resources](resources/schedule.md) - Job scheduling

