---
id: kubernetes
title: Kubernetes
sidebar_label: Kubernetes
description: Learn how to create and manage Kubernetes clusters using the acloud CLI.
---
# Kubernetes Example

This guide demonstrates how to provision and manage a Kubernetes cluster using the Aruba Cloud CLI.

## Step 0: List Available VPCs

First, determine which VPC you want to use for your Kubernetes cluster. List all available VPCs with:

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

Before provisioning a Kubernetes cluster, ensure your VPC is already created and its status is **Active**.

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

## Step 3: Extract the Subnet URI

Once you have chosen a subnet, extract its URI for use in the provisioning command. Run:

```bash
acloud network subnet get <vpc-id> <subnet-id> | grep URI
```

Example:
```bash
acloud network subnet get 69495ef64d0cdc87949b71ec 694ba1737712ac0032dbe50a | grep URI
```

Example output:
```
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Network/vpcs/69495ef64d0cdc87949b71ec/subnets/694ba1737712ac0032dbe50a
```

> **Note:** Save this URI for the Kubernetes provisioning step.

---


## Step 4: Determine Kubernetes Version and List Available Nodepool Flavors

Before creating your cluster, check which Kubernetes versions are available and which nodepool flavors you can use:

- **Check available Kubernetes versions:**
  - Visit: [Kubernetes Versions Metadata](https://api.arubacloud.com/docs/metadata/#kubernetes-version)
  - Review the list and select a supported version for your cluster (e.g., `1.27`).

- **Check available nodepool flavors:**
  - Visit: [KaaS Nodepool Flavors Metadata](https://api.arubacloud.com/docs/metadata/#kaas-flavors)
  - Review the available flavors and choose the one that fits your needs (e.g., `K8S-Standard`).

You can also list available flavors using the CLI:

```bash
acloud container kaas flavor list
```


## Step 5: Create a Kubernetes Cluster

Here is a real example of creating a Kubernetes cluster with a single node pool:

```bash
acloud container kaas create \
  --name "test-cluster" \
  --region "ITBG-Bergamo" \
  --vpc-uri "/projects/68398923fb2cb026400d4d31/providers/Aruba.Network/vpcs/69495ef64d0cdc87949b71ec" \
  --subnet-uri "/projects/68398923fb2cb026400d4d31/providers/Aruba.Network/vpcs/69495ef64d0cdc87949b71ec/subnets/694ba1737712ac0032dbe50a" \
  --node-cidr-address "10.0.0.0/16" \
  --node-cidr-name "node-cidr" \
  --security-group-name "kaas-sg" \
  --kubernetes-version "1.33.2" \
  --node-pool-name "default-pool" \
  --node-pool-autoscaling \
  --node-pool-max-count 5 \
  --node-pool-min-count 1 \
  --node-pool-nodes 3 \
  --node-pool-instance "K4A8" \
  --node-pool-zone "ITBG-1" \
  --tags "production,kubernetes" \
  --ha \
  --billing-period "Hour"
```

Example output:

```
ID                             NAME                                     VERSION              REGION               
697cde26cc725dcf1c299a30       test-cluster                             1.33.2               ITBG-Bergamo  
```


## Step 6: Get Cluster Details

Here is a real example of retrieving Kubernetes cluster details:

```bash
acloud container kaas get 694ff33bc2682f8c02f4956e
```

Example output:

```
KaaS Cluster Details:
====================
ID:              694ff33bc2682f8c02f4956e
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Container/kaas/694ff33bc2682f8c02f4956e
Name:            autoscaler
Region:          ITBG-Bergamo
Kubernetes Version: 1.32.3
Status:          Active
Creation Date:   27-12-2025 14:54:51
Created By:      aru-297647
Tags:            []
```


## Step 7: Connect to the KaaS Cluster

To connect and configure your kubeconfig for the cluster, use:

```bash
acloud container kaas connect 694ff33bc2682f8c02f4956e
```

Example output:

```
KaaS successfully connected
Kubeconfig saved to: /home/amedeopalopoli/.kube/kubeconfig_694ff33bc2682f8c02f4956e
Default config updated: /home/amedeopalopoli/.kube/config
```

> **Note:** You must have `kubectl` installed locally. The `acloud` CLI will automatically configure your `kubectl` client to use the created cluster when you connect.

## Step 8: Delete the Kubernetes Cluster

```bash
acloud container kaas delete <cluster-id>
```

---
