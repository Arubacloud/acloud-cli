---
id: containerregistry
title: Container Registry
sidebar_label: Container Registry
description: Workflow end-to-end per la creazione di un registry privato per container, l'autenticazione con Docker e le operazioni di push e pull delle immagini tramite la CLI acloud.
---

# Esempio Container Registry

Questa guida copre il workflow completo per il provisioning di un registry privato per container su Aruba Cloud: dalla configurazione della rete e dello storage, all'autenticazione Docker, alle operazioni di push/pull delle immagini e all'amministrazione del registry.

## Prerequisiti

Prima di iniziare, assicurati di avere:
- Docker installato sul tuo computer locale
- La CLI configurata con credenziali valide (`acloud config set`)
- Una VPC attiva, o segui lo Step 0 per preparare le risorse di rete

---

## Step 0: Preparare le Risorse di Rete

### Elenca le VPC disponibili

```bash
acloud network vpc list
```

Esempio output:
```
NAME       ID                        REGION         SUBNETS    STATUS
prod-vpc   69495ef64d0cdc87949b71ec  ITBG-Bergamo   3          Active
```

Annota l'`ID` della VPC. Assicurati che `STATUS` sia `Active`.

---

### Elenca o crea una Subnet

```bash
acloud network subnet list 69495ef64d0cdc87949b71ec
```

Esempio output:
```
NAME            ID                         REGION         CIDR              STATUS
registry-sub    694ba1737712ac0032dbe50a   ITBG-Bergamo   192.168.20.0/24   Active
```

Annota l'`ID` della subnet.

---

### Elenca o crea un Security Group (consenti TCP/443)

Il security group deve consentire **traffico TCP inbound sulla porta 443** (HTTPS) per il protocollo Docker registry, e consentire il traffico egress in uscita.

```bash
acloud network securitygroup list 69495ef64d0cdc87949b71ec
```

Se devi creare un nuovo security group e aggiungere la regola HTTPS:

```bash
# Crea security group (il vpc-id è il primo argomento posizionale)
acloud network securitygroup create 69495ef64d0cdc87949b71ec \
  --name "registry-sg" \
  --region "ITBG-Bergamo"

# Aggiungi regola inbound per HTTPS (porta 443)
acloud network securityrule create 69495ef64d0cdc87949b71ec <securitygroup-id> \
  --name "allow-https" \
  --region "ITBG-Bergamo" \
  --direction Ingress \
  --protocol TCP \
  --port 443 \
  --target-kind Ip \
  --target-value "0.0.0.0/0"
```

Annota l'`ID` del security group.

---

### Ottieni un Elastic IP per l'accesso esterno

Il container registry richiede un Elastic IP per l'accesso da client Docker esterni.

```bash
acloud network elasticip list
```

Annota l'`ID` dell'Elastic IP.

---

## Step 1: Crea il Block Storage per i Dati del Registry

Il container registry richiede block storage dedicato per i layer delle immagini e i metadati:

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

Esempio output:
```
Block storage created successfully!
ID:              697b389bce7dfeef91532563
Name:            registry-storage
Size (GB):       100
Type:            Performance
Region:          ITBG-Bergamo
Status:          InCreation
```

Attendi che lo stato del block storage diventi `NotUsed`:

```bash
acloud storage blockstorage list | grep "registry-storage"
```

---

## Step 2: Crea il Container Registry

Crea il container registry con tutte le risorse richieste:

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

Esempio output:
```
Container registry created successfully!
ID:              69495ef64d0cdc87949b72ab
Name:            my-registry
Region:          ITBG-Bergamo
Status:          InCreation
Creation Date:   18-06-2026 10:00:00
```

> **Nota**: Il provisioning del registry può richiedere alcuni minuti. Usa `--wait` con un timeout esteso per attendere che il registry diventi `Active`:
> ```bash
> acloud --timeout 15m container containerregistry create ... --wait
> ```

---

## Step 3: Attendi che il Registry Diventi Attivo

Monitora lo stato del provisioning:

```bash
acloud container containerregistry list
```

Attendi che `STATUS` mostri `Active`:

```
NAME          ID                        REGION         STATUS
my-registry   69495ef64d0cdc87949b72ab  ITBG-Bergamo   Active
```

Ottieni i dettagli completi incluso l'indirizzo IP pubblico:

```bash
acloud container containerregistry get 69495ef64d0cdc87949b72ab
```

Annota l'**indirizzo IP pubblico** — questo è il nome host del registry che userai per i comandi Docker.

---

## Step 4: Autenticati con il Registry

Accedi al registry usando Docker. Usa l'indirizzo IP dell'Elastic IP come nome host del registry:

```bash
docker login <indirizzo-ip-elastic> --username registryadmin
# Password: (inserisci la password admin impostata durante la creazione)
```

Output atteso:
```
Login Succeeded
```

> **Suggerimento per le pipeline CI/CD**: Passa la password in modo non interattivo:
> ```bash
> echo "$REGISTRY_PASSWORD" | docker login <indirizzo-ip-elastic> \
>   --username registryadmin --password-stdin
> ```

---

## Step 5: Carica un'Immagine nel Registry

Tagga un'immagine locale esistente con l'indirizzo del registry e caricala:

```bash
# Tagga un'immagine locale per questo registry
docker tag my-app:latest <indirizzo-ip-elastic>/my-app:latest

# Carica l'immagine taggata nel registry
docker push <indirizzo-ip-elastic>/my-app:latest
```

Esempio output:
```
The push refers to repository [<indirizzo-ip-elastic>/my-app]
abc123456789: Pushed
def234567890: Pushed
latest: digest: sha256:a1b2c3d4e5f6... size: 1234
```

---

## Step 6: Scarica un'Immagine dal Registry

Scarica un'immagine dal registry su qualsiasi host Docker che ha accesso:

```bash
docker pull <indirizzo-ip-elastic>/my-app:latest
```

Esempio output:
```
latest: Pulling from my-app
abc123456789: Pull complete
def234567890: Pull complete
Status: Downloaded newer image for <indirizzo-ip-elastic>/my-app:latest
```

---

## Step 7: Gestisci il Registry

### Elenca tutti i Registry

```bash
acloud container containerregistry list
```

### Aggiorna le Proprietà del Registry

```bash
# Cambia il periodo di fatturazione in annuale per risparmiare sui costi
acloud container containerregistry update 69495ef64d0cdc87949b72ab \
  --billing-period "Year"

# Aumenta gli utenti concorrenti per supportare team più grandi o più pipeline CI/CD
acloud container containerregistry update 69495ef64d0cdc87949b72ab \
  --concurrent-users 20

# Aggiorna i tag per il tracciamento delle risorse
acloud container containerregistry update 69495ef64d0cdc87949b72ab \
  --tags "production,registry,team-platform"
```

### Rinomina il Registry

```bash
acloud container containerregistry update 69495ef64d0cdc87949b72ab \
  --name "prod-registry"
```

---

## Best Practice

- **Security group**: Limita la porta 443 a intervalli IP noti per i registry di produzione invece di `0.0.0.0/0`
- **Dimensionamento dello storage**: Inizia con almeno 100 GB e monitora l'utilizzo man mano che le immagini si accumulano
- **Tag delle immagini**: Usa tag con versione (es. `v1.2.3`, `20260618-abc123`) invece di affidarsi solo a `latest` per abilitare i rollback
- **Utenti concorrenti**: Imposta `--concurrent-users` in base alla dimensione del team e al numero di job paralleli nelle pipeline CI/CD
- **Fatturazione**: Passa alla fatturazione `Month` o `Year` per i registry di produzione a lungo termine per ridurre i costi
- **Credenziali**: Ruota regolarmente la password admin; usa utenti separati per team o pipeline dove supportato

---

## Step 8: Pulizia

Per eliminare il container registry quando non è più necessario:

```bash
acloud container containerregistry delete 69495ef64d0cdc87949b72ab --yes
```

> **Attenzione**: L'eliminazione di un container registry rimuove permanentemente tutte le immagini memorizzate. Assicurati che le immagini siano sottoposte a backup o replicate in un altro registry prima di eliminare.

Dopo aver eliminato il registry, rilascia il block storage associato se non è più necessario:

```bash
acloud storage blockstorage delete 697b389bce7dfeef91532563 --yes
```

## Risorse Correlate

- [Container Registry](../resources/container/containerregistry.md) - Riferimento completo comandi container registry
- [KaaS](../resources/container/kaas.md) - Cluster Kubernetes per l'esecuzione di workload containerizzati
- [Risorse di Rete](../resources/network.md) - Configurazione VPC, subnet e security group
- [Risorse Storage](../resources/storage.md) - Block storage per i dati del registry
