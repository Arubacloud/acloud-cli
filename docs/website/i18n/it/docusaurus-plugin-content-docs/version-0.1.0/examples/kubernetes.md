---
id: kubernetes
title: Kubernetes
sidebar_label: Kubernetes
description: Scopri come creare e gestire cluster Kubernetes usando acloud CLI.
---
# Esempio di Kubernetes

Questa guida mostra come effettuare il provisioning e la gestione di un cluster Kubernetes tramite Aruba Cloud CLI.

## Step 0: Elenca le VPC disponibili

Per prima cosa, individua la VPC da usare per il cluster Kubernetes. Elenca tutte le VPC disponibili con:

```bash
acloud network vpc list
```

Esempio output:
```
NAME       ID                        REGION         SUBNETS    STATUS
prova      689307f4745108d3c6343b5a  ITBG-Bergamo   5          Active
test-cli   69495ef64d0cdc87949b71ec  ITBG-Bergamo   0          Active
```

Scegli una VPC con `STATUS` pari a `Active` e annota il suo `ID` per il prossimo step.

---

## Step 1: Recupera URI e stato della VPC

Prima di creare un cluster Kubernetes, assicurati che la VPC sia già creata e in stato **Active**.

Esegui il seguente comando per ottenere l'URI della VPC e verificarne lo stato:

```bash
acloud network vpc get {vpc-id} | grep -E "URI|Status"
```

Esempio output:
```
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Network/vpcs/69495ef64d0cdc87949b71ec
Status:          Active
```

> **Nota:** Procedi solo se lo stato è `Active`. Se non lo è, attendi che la VPC diventi attiva prima di continuare.

---

## Step 2: Elenca o crea una Subnet nella VPC

Dopo aver selezionato la VPC, serve una subnet al suo interno. Puoi elencare le subnet esistenti o crearne una nuova.

Per elencare le subnet nella VPC scelta, usa l'ID della VPC:

```bash
acloud network subnet list {vpc-id}
```

Esempio output:
```
NAME                       ID                         REGION         CIDR             STATUS
test-cli                   694ba1737712ac0032dbe50a   ITBG-Bergamo   192.168.0.0/24   Active
test-cli-new               694ba7437712ac0032dbe566   ITBG-Bergamo   192.168.1.0/24   Active
test-cli-new2              694ba7977712ac0032dbe571   ITBG-Bergamo   192.168.2.0/24   Active
e2e-test-1766569838-subnet 694bb7767712ac0032dbe5fc   ITBG-Bergamo   192.168.3.0/24   Active
e2e-test-1766570350-subnet 694bb9767712ac0032dbe640   ITBG-Bergamo   192.168.4.0/24   Active
```

Scegli una subnet con `STATUS` pari a `Active` e annota il suo `ID` e `CIDR`. Se non esiste una subnet adatta, creane una nuova tramite CLI (vedi documentazione subnet).

---

## Step 3: Estrai l'URI della Subnet

Una volta scelta la subnet, estrai il suo URI per usarlo nel comando di provisioning. Esegui:

```bash
acloud network subnet get <vpc-id> <subnet-id> | grep URI
```

Esempio:
```bash
acloud network subnet get 69495ef64d0cdc87949b71ec 694ba1737712ac0032dbe50a | grep URI
```

Esempio output:
```
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Network/vpcs/69495ef64d0cdc87949b71ec/subnets/694ba1737712ac0032dbe50a
```

> **Nota:** Salva questo URI per il provisioning Kubernetes.

---


## Step 4: Determina la versione Kubernetes e i flavor disponibili per i nodepool

Prima di creare il cluster, verifica quali versioni Kubernetes sono disponibili e quali flavor di nodepool puoi usare:

- **Verifica le versioni Kubernetes disponibili:**
  - Visita: [Metadata Versioni Kubernetes](https://api.arubacloud.com/docs/metadata/#kubernetes-version)
  - Consulta l'elenco e scegli una versione supportata per il cluster (es. `1.27`).

- **Verifica i flavor disponibili per i nodepool:**
  - Visita: [Metadata Flavor Nodepool KaaS](https://api.arubacloud.com/docs/metadata/#kaas-flavors)
  - Consulta i flavor disponibili e scegli quello più adatto alle tue esigenze (es. `K8S-Standard`).

Puoi anche elencare i flavor disponibili tramite CLI:

```bash
acloud container kaas flavor list
```


## Step 5: Crea un cluster Kubernetes

Esempio reale di creazione di un cluster Kubernetes con un singolo node pool:

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

Esempio output:

```
ID                             NAME                                     VERSION              REGION               
697cde26cc725dcf1c299a30       test-cluster                             1.33.2               ITBG-Bergamo  
```


## Step 6: Ottieni i dettagli del cluster

Esempio reale di recupero dettagli di un cluster Kubernetes:

```bash
acloud container kaas get 694ff33bc2682f8c02f4956e
```

Esempio output:

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

## Step 7: Connetti al cluster KaaS

Per connetterti e configurare il kubeconfig per il cluster, usa:

```bash
acloud container kaas connect 694ff33bc2682f8c02f4956e
```

Esempio output:

```
KaaS successfully connected
Kubeconfig saved to: /home/amedeopalopoli/.kube/kubeconfig_694ff33bc2682f8c02f4956e
Default config updated: /home/amedeopalopoli/.kube/config
```

> **Nota:** Devi avere `kubectl` installato localmente. Il comando `acloud` configurerà automaticamente il client `kubectl` per usare il cluster creato quando ti connetti.

## Step 8: Elimina il cluster Kubernetes

```bash
acloud container kaas delete <cluster-id>
```

---
