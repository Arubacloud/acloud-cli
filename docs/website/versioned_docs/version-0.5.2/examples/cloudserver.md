---
id: cloudserver
title: CloudServer
sidebar_label: CloudServer
description: Learn the essential commands to create, manage and use Aruba Cloud Server using the acloud CLI.
---
# Cloud Server Example

This example demonstrates how to provision a new cloud server using the CLI with all required networking flags.

## Step 0: List Available VPCs

First, determine which VPC you want to use for your cloud server. List all available VPCs with:

```bash
acloud network vpc list
```

Example output:
```
NAME       ID                        REGION         SUBNETS    STATUS
prova      689307f4745108d3c6343b5a  ITBG-Bergamo   5          Active
test-cli   69495ef64d0cdc87949b71ec  ITBG-Bergamo   0          Active
```

Choose a VPC with `STATUS` as `Active` and note its `ID` for the next step.

---

## Step 1: Retrieve the VPC URI and Status

Before provisioning a cloud server, ensure your VPC is already created and its status is **Active**.

Run the following command to get the VPC URI and check its status:

```bash
acloud network vpc get {vpc-id} | grep -E "URI|Status"
```

Example output:
```
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Network/vpcs/69495ef64d0cdc87949b71ec
Status:          Active
```

> **Note:** Only proceed if the status is `Active`. If not, wait until the VPC becomes active before continuing.

---

## Step 2: List or Create a Subnet in the VPC

After selecting your VPC, you need a subnet within it. You can list existing subnets for the VPC or create a new one.

To list subnets in your chosen VPC, use the VPC ID:

```bash
acloud network subnet list {vpc-id}
```

Example output:
```
NAME                       ID                         REGION         CIDR             STATUS
test-cli                   694ba1737712ac0032dbe50a   ITBG-Bergamo   192.168.0.0/24   Active
test-cli-new               694ba7437712ac0032dbe566   ITBG-Bergamo   192.168.1.0/24   Active
test-cli-new2              694ba7977712ac0032dbe571   ITBG-Bergamo   192.168.2.0/24   Active
e2e-test-1766569838-subnet 694bb7767712ac0032dbe5fc   ITBG-Bergamo   192.168.3.0/24   Active
e2e-test-1766570350-subnet 694bb9767712ac0032dbe640   ITBG-Bergamo   192.168.4.0/24   Active
```

Choose a subnet with `STATUS` as `Active` and note its `ID` and `CIDR`. If no suitable subnet exists, create a new one using the CLI (see documentation for subnet creation).

---

## Step 3: Confirm the Subnet ID

Note the `ID` of the chosen subnet from the list output. You will pass this directly to the cloud server create command using `--subnet-id`.

```bash
acloud network subnet get <vpc-id> <subnet-id>
```

Example:
```bash
acloud network subnet get 69495ef64d0cdc87949b71ec 694ba1737712ac0032dbe50a
```

> **Note:** Note the subnet `ID`. Only proceed if the subnet `STATUS` is `Active`.

---

## Step 4: List or Create a Security Group in the VPC

After selecting your subnet, you need a security group within the same VPC. You can list existing security groups or create a new one.

To list security groups in your chosen VPC, use the VPC ID:

```bash
acloud network securitygroup list 69495ef64d0cdc87949b71ec
```

Example output:
```
NAME                       ID                         REGION         STATUS
test-cli                   694b05ac4d0cdc87949b75f9   ITBG-Bergamo   Active
e2e-test-1766569838-sg     694bb7817712ac0032dbe604   ITBG-Bergamo   Active
e2e-test-1766570350-sg     694bb9817712ac0032dbe648   ITBG-Bergamo   Active
```

Choose a security group with `STATUS` as `Active` and note its `ID`. If no suitable security group exists, create a new one using the CLI (see documentation for security group creation).

---

## Step 5: Note the Elastic IP ID

Note the Elastic IP `ID` from `acloud network elasticip list`. You can verify it with:

```bash
acloud network elasticip get <elasticip-id>
```

Example:
```bash
acloud network elasticip get 6a35162fdf2d456b0f95fd7e
```

> **Note:** Note the Elastic IP `ID`. An Elastic IP is required to connect to your server from the public internet (used in Step 10 with `--elasticip-id`).

---

## Step 6: Create a Bootable Block Storage


If you want to create a bootable block storage volume (for example, to use a custom image), use the following command. The `--region` and `--zone` flags are required:


```bash
acloud storage blockstorage create \
  --name boot-ubuntu \
  --region ITBG-Bergamo \
  --zone ITBG-1 \
  --set-bootable \
  --billing-period Hour \
  --size 20 \
  --tags boot \
  --type Performance \
  --image LU22-001
```

Example output:
```
Block storage created successfully!
ID:              697b389bce7dfeef91532563
Name:            boot-ubuntu
Size (GB):       20
Type:            Performance
Zone:            
Region:          ITBG-Bergamo
Status:          InCreation
Creation Date:   29-01-2026 10:38:19
```

> **Note:**
> - The `--region` and `--zone` flags are required.
> - Use `--set-bootable` to ensure the volume is created as bootable (required when using the `--image` flag).
> - Replace `LU22-001` with the desired image code.
> - Adjust other parameters as needed for your use case.

---

### List of Available Images for Bootable Block Storage

Below are some of the available image codes you can use with the `--image` flag when creating a bootable block storage. For the full and up-to-date list, see the [official ArubaCloud API documentation](https://api.arubacloud.com/docs/metadata/#cloud-server-bootvolume).

| Image Code         | Description           | OS Flavor        |
|--------------------|----------------------|------------------|
| alma8              | AlmaLinux 8 64bit    | Linux            |
| alma9              | AlmaLinux 9 64bit    | Linux            |
| DE11-001           | Debian 11 64bit      | Linux            |
| DE12-001           | Debian 12 64bit      | Linux            |
| LU20-001           | Ubuntu 20.04 64bit   | Linux            |
| LU22-001           | Ubuntu 22.04 64bit   | Linux            |
| LU24-001           | Ubuntu 24.04 64bit   | Linux            |
| osuse15_2_x64_1_0  | openSUSE 15 64bit    | Linux            |
| WS19-001_W2K19_1_0 | Windows Server 2019  | Windows          |
| WS22-001_W2K22_1_0 | Windows Server 2022  | Windows          |

> **Note:** Use the value in the "Image Code" column with the `--image` flag. For example: `--image LU22-001` for Ubuntu 22.04.

---


## Step 7: List or Create an Active Keypair

Before creating your cloud server, you need an active keypair. You can list existing keypairs and select one with status `Active`, or create a new one if needed.

To list keypairs:

```bash
acloud compute keypair list
```

Example output:
```
NAME       ID                        PUBLIC_KEY                                         STATUS
amedeo     69007ebf4e7d691466d8621e  ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCzZB11JRK... Active
```

Choose a keypair with `STATUS` as `Active` and note its `ID` or `NAME`. If no suitable keypair exists, create a new one:

```bash
acloud compute keypair create --name my-keypair --public-key "$(cat ~/.ssh/id_rsa.pub)"
```

After creation, verify the new keypair is `Active` using the list command above.

To extract the keypair URI for use in the provisioning command:

```bash
acloud compute keypair get <keypair-id> | grep URI
```

Example output:
```
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Compute/keyPairs/69007ebf4e7d691466d8621e
```

> **Note:** Note the keypair `ID` (e.g., `69007ebf4e7d691466d8621e`). Pass it to `--keypair-id` in the create command.

---

## Step 8: Confirm Boot Disk ID and Status

Note the `ID` of your bootable block storage. The block storage must be in `NotUsed` status before it can be used as a boot disk.

```bash
acloud storage blockstorage list
```

Example output:
```
NAME         ID                         SIZE(GB)  REGION         ZONE  TYPE         STATUS
boot-ubuntu  697b3a0dce7dfeef9153256a   20        ITBG-Bergamo         Performance  NotUsed
```

> **Note:** Note the boot disk `ID`. Only proceed when the status is `NotUsed` or `Active`.

---


## Step 9: (Optional) Create a user-data file for cloud-init

If you want to customize your server at boot (e.g., create users, install packages), create a `cloud-init.yaml` file before provisioning. Example:

```yaml
#cloud-config
users:
  - name: demo
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: users, admin
    home: /home/demo
    shell: /bin/bash
    lock_passwd: false
    passwd: <hashed-password>
```

Replace `<hashed-password>` with a valid hashed password (see cloud-init docs for details). You can add more cloud-init options as needed.

---

## Step 10: Create the Cloud Server

Use the IDs collected in the previous steps to create the cloud server:

```bash
acloud compute cloudserver create \
  --name "my-server" \
  --region "ITBG-Bergamo" \
  --zone "ITBG-1" \
  --flavor "CSO4A8" \
  --boot-disk-id "6a3516c275300438a29f15fd" \
  --vpc-id "69fdbc398675ff21f3c1587b" \
  --subnet-id "69fdbc8f8675ff21f3c1588a" \
  --security-group-id "69fdbc8e8675ff21f3c15888" \
  --keypair-id "6a33ec6b2b2c12d76c120f5e" \
  --elasticip-id "6a35162fdf2d456b0f95fd7e" \
  --billing-period Hour \
  --tags "production"
```

Example output:
```
ID                             NAME         FLAVOR    CPU   RAM(GB)   HD(GB)   REGION
697c62605b79733376b3386a       my-server    CSO4A8    4     8         0        ITBG-Bergamo
```

- Replace the IDs above with your actual VPC, subnet, security group, keypair, boot disk, and Elastic IP IDs.
- The `--elasticip-id` flag assigns a public IP at creation time and is required to use the `connect` command afterwards.
- The `--user-data-file` flag is optional and can be omitted if not needed.
- You can specify multiple subnets or security groups using comma-separated values.
- All networking flags (`--vpc-id`, `--subnet-id`, `--security-group-id`) and the `--zone` flag are required.

---

## Step 11: Wait for the Cloud Server to Become Active

After running the create command, your server will be provisioned. You can check its status with:

```bash
acloud compute cloudserver list | grep "my-server"
```


When the status shows `Active`, the server is ready for use:

```
my-server                 697c62605b79733376b3386a       ITBG-Bergamo    CSO4A8          Active
```

---

## Step 12: Connect to the Cloud Server

To connect to your cloud server, use the `acloud compute cloudserver connect` command. You must specify the user according to the boot image. For Ubuntu images, use the `ubuntu` user:

```bash
acloud compute cloudserver connect 697c62605b79733376b3386a --user ubuntu
```

Example output:
```
Connect by running: ssh ubuntu@85.235.152.94
```
### Connecting to Your Cloud Server via SSH

To connect to your cloud server using SSH, you can use the CLI connect command. The server must have an Elastic IP assigned to be accessible from the public internet.

Example usage:

```bash
acloud compute cloudserver connect 697b4bd0377bb8332d771b39 --user ubuntu
```

If the server has an Elastic IP, you will see a message like:

```
Connect by running: ssh ubuntu@95.110.142.229
```

If the server does not have an Elastic IP, you will see:

```
No Elastic IP found for this cloud server.
The server must have an Elastic IP linked to use the connect command.
```

> **Note:**
> - The `--user` flag should match the default username for the image you used (e.g., `ubuntu` for Ubuntu images).
> - To connect from public Internet, ensure you have assigned an Elastic IP to your server during or after creation. 
You can add an Elastic IP by updating the server or specifying the `--elasticip-uri` flag at creation time.
> - The connect command will print the SSH command to use.

## Step 13: Power Off or Power On the Cloud Server

You can power off or power on your cloud server at any time using the following commands:

To power off:

```bash
acloud compute cloudserver power-off 697c62605b79733376b3386a
```

Example output:
```
Cloud server powered off successfully!
Server: my-server
Status: Updating
```

To power on:

```bash
acloud compute cloudserver power-on 697c62605b79733376b3386a
```

Example output:
```
Cloud server powered on successfully!
Server: my-server
Status: Updating
```

## Step 14: Delete the Cloud Server

To delete your cloud server when it is no longer needed, use the following command:

```bash
acloud compute cloudserver delete 697c62605b79733376b3386a
```

You will be prompted for confirmation:

```
Are you sure you want to delete cloud server '697c62605b79733376b3386a'? (yes/no): yes
Cloud server '697c62605b79733376b3386a' deleted successfully.
```
---