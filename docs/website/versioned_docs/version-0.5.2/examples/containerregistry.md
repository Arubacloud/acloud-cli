---
id: containerregistry
title: Container Registry
sidebar_label: Container Registry
description: End-to-end workflow for creating a private container registry, authenticating with Docker, and pushing and pulling images using the acloud CLI.
---

# Container Registry Example

This guide covers the complete workflow for provisioning a private container registry on Aruba Cloud: from network and storage setup, to Docker authentication, image push/pull operations, and registry administration.

## Prerequisites

Before starting, ensure you have:
- Docker installed on your local machine
- The CLI configured with valid credentials (`acloud config set`)
- An active VPC, or follow Step 0 to prepare networking resources

---

## Step 0: Prepare Networking Resources

### List Available VPCs

```bash
acloud network vpc list
```

Example output:
```
NAME       ID                        REGION         SUBNETS    STATUS
prod-vpc   69495ef64d0cdc87949b71ec  ITBG-Bergamo   3          Active
```

Note the VPC `ID`. Ensure `STATUS` is `Active`.

---

### List or Create a Subnet

```bash
acloud network subnet list 69495ef64d0cdc87949b71ec
```

Example output:
```
NAME            ID                         REGION         CIDR              STATUS
registry-sub    694ba1737712ac0032dbe50a   ITBG-Bergamo   192.168.20.0/24   Active
```

Note the subnet `ID`.

---

### List or Create a Security Group (allow TCP/443)

The security group must allow **inbound TCP on port 443** (HTTPS) for the Docker registry protocol, and allow outbound egress.

```bash
acloud network securitygroup list 69495ef64d0cdc87949b71ec
```

If you need to create a new security group and add the HTTPS rule:

```bash
# Create security group
acloud network securitygroup create \
  --name "registry-sg" \
  --vpc-id "69495ef64d0cdc87949b71ec" \
  --region "ITBG-Bergamo"

# Add inbound rule for HTTPS (port 443)
acloud network securityrule create 69495ef64d0cdc87949b71ec <securitygroup-id> \
  --direction Inbound \
  --protocol TCP \
  --port-range-min 443 \
  --port-range-max 443 \
  --remote-ip-prefix "0.0.0.0/0"
```

Note the security group `ID`.

---

### Get an Elastic IP for Public Access

The container registry requires an Elastic IP for external Docker client access.

```bash
acloud network elasticip list
```

Note the Elastic IP `ID`.

---

## Step 1: Create Block Storage for Registry Data

The container registry requires dedicated block storage for image layers and metadata:

```bash
acloud storage blockstorage create \
  --name "registry-storage" \
  --region "ITBG-Bergamo" \
  --zone "ITBG-1" \
  --size 100 \
  --type Performance \
  --billing-period Hour \
  --tags "registry,production"
```

Example output:
```
Block storage created successfully!
ID:              697b389bce7dfeef91532563
Name:            registry-storage
Size (GB):       100
Type:            Performance
Region:          ITBG-Bergamo
Status:          InCreation
```

Wait for the block storage status to become `NotUsed`:

```bash
acloud storage blockstorage list | grep "registry-storage"
```

---

## Step 2: Create the Container Registry

Create the container registry with all required resources:

```bash
acloud container containerregistry create \
  --name "my-registry" \
  --region "ITBG-Bergamo" \
  --public-ip-id "694bb7897712ac0032dbe60c" \
  --vpc-id "69495ef64d0cdc87949b71ec" \
  --subnet-id "694ba1737712ac0032dbe50a" \
  --security-group-id "694b05ac4d0cdc87949b75f9" \
  --block-storage-id "697b389bce7dfeef91532563" \
  --admin-username "registryadmin" \
  --concurrent-users "Small" \
  --billing-period "Hour" \
  --tags "production,registry"
```

Example output:
```
Container registry created successfully!
ID:              69495ef64d0cdc87949b72ab
Name:            my-registry
Region:          ITBG-Bergamo
Status:          InCreation
Creation Date:   18-06-2026 10:00:00
```

> **Note**: Registry provisioning may take several minutes. Use `--wait` with an extended timeout to block until the registry is `Active`:
> ```bash
> acloud --timeout 15m container containerregistry create ... --wait
> ```

---

## Step 3: Wait for the Registry to Become Active

Monitor the provisioning status:

```bash
acloud container containerregistry list
```

Wait until `STATUS` shows `Active`:

```
NAME          ID                        REGION         STATUS
my-registry   69495ef64d0cdc87949b72ab  ITBG-Bergamo   Active
```

Get full details including the public IP address:

```bash
acloud container containerregistry get 69495ef64d0cdc87949b72ab
```

Note the **public IP address** — this is the registry hostname you will use for Docker commands.

---

## Step 4: Authenticate with the Registry

Log in to the registry using Docker. Use the Elastic IP address as the registry hostname:

```bash
docker login <elastic-ip-address> --username registryadmin
# Password: (enter the admin password set during creation)
```

Expected output:
```
Login Succeeded
```

> **Tip for CI/CD pipelines**: Pass the password non-interactively:
> ```bash
> echo "$REGISTRY_PASSWORD" | docker login <elastic-ip-address> \
>   --username registryadmin --password-stdin
> ```

---

## Step 5: Push an Image to the Registry

Tag an existing local image with the registry address and push it:

```bash
# Tag a local image for this registry
docker tag my-app:latest <elastic-ip-address>/my-app:latest

# Push the tagged image to the registry
docker push <elastic-ip-address>/my-app:latest
```

Example output:
```
The push refers to repository [<elastic-ip-address>/my-app]
abc123456789: Pushed
def234567890: Pushed
latest: digest: sha256:a1b2c3d4e5f6... size: 1234
```

---

## Step 6: Pull an Image from the Registry

Pull an image from the registry on any Docker host that has access:

```bash
docker pull <elastic-ip-address>/my-app:latest
```

Example output:
```
latest: Pulling from my-app
abc123456789: Pull complete
def234567890: Pull complete
Status: Downloaded newer image for <elastic-ip-address>/my-app:latest
```

---

## Step 7: Manage the Registry

### List All Registries

```bash
acloud container containerregistry list
```

### Update Registry Properties

```bash
# Change billing period to yearly for cost savings
acloud container containerregistry update 69495ef64d0cdc87949b72ab \
  --billing-period "Year"

# Increase concurrent users to support larger teams or more CI/CD pipelines
acloud container containerregistry update 69495ef64d0cdc87949b72ab \
  --concurrent-users 20

# Update tags for resource tracking
acloud container containerregistry update 69495ef64d0cdc87949b72ab \
  --tags "production,registry,team-platform"
```

### Rename the Registry

```bash
acloud container containerregistry update 69495ef64d0cdc87949b72ab \
  --name "prod-registry"
```

---

## Best Practices

- **Security groups**: Restrict port 443 to known IP ranges for production registries instead of `0.0.0.0/0`
- **Storage sizing**: Start with at least 100 GB and monitor usage as images accumulate; expand by replacing the block storage when needed
- **Image tagging**: Use versioned tags (e.g., `v1.2.3`, `20260618-abc123`) rather than relying solely on `latest` to enable rollbacks
- **Concurrent users**: Set `--concurrent-users` based on your team size and the number of parallel CI/CD pipeline jobs
- **Billing**: Switch to `Month` or `Year` billing for long-running production registries to reduce costs
- **Credentials**: Rotate the admin password regularly; use separate users per team or pipeline where supported

---

## Step 8: Cleanup

To delete the container registry when no longer needed:

```bash
acloud container containerregistry delete 69495ef64d0cdc87949b72ab --yes
```

> **Warning**: Deleting a container registry permanently removes all stored images. Ensure images are backed up or mirrored to another registry before deleting.

After deleting the registry, release the associated block storage if it is no longer needed:

```bash
acloud storage blockstorage delete 697b389bce7dfeef91532563 --yes
```

## Related Resources

- [Container Registry](../resources/container/containerregistry.md) - Full container registry command reference
- [KaaS](../resources/container/kaas.md) - Kubernetes clusters for running containerized workloads
- [Network Resources](../resources/network.md) - Configure VPCs, subnets, and security groups
- [Storage Resources](../resources/storage.md) - Block storage for registry data
