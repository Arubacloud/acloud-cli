---
id: containerregistry-it
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

Annota l'`ID` della VPC con `STATUS` pari a `Active`.

---

### Elenca o crea una Subnet

```bash
acloud network subnet list <vpc-id>
```

Annota l'`ID` della subnet.

---

### Elenca o crea un Security Group (consenti TCP/443)

Il security group deve consentire **traffico TCP inbound sulla porta 443** (HTTPS) per il protocollo Docker registry.

```bash
acloud network securitygroup list <vpc-id>
```

Se devi creare un nuovo security group:

```bash
acloud network securitygroup create \
  --name "registry-sg" \
  --vpc-id "<vpc-id>" \
  --region "ITBG-Bergamo"

acloud network securityrule create <vpc-id> <securitygroup-id> \
  --direction Inbound \
  --protocol TCP \
  --port-range-min 443 \
  --port-range-max 443 \
  --remote-ip-prefix "0.0.0.0/0"
```

Annota l'`ID` del security group.

---

### Ottieni un Elastic IP per l'accesso esterno

```bash
acloud network elasticip list
```

Annota l'`ID` dell'Elastic IP.

---

## Step 1: Crea il Block Storage per i Dati del Registry

```bash
acloud storage blockstorage create \
  --name "registry-storage" \
  --region "ITBG-Bergamo" \
  --zone "itbg1-a" \
  --size 100 \
  --type Performance \
  --billing-period Hour \
  --tags "registry,production"
```

Attendi che lo stato diventi `NotUsed`:

```bash
acloud storage blockstorage list | grep "registry-storage"
```

---

## Step 2: Crea il Container Registry

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

> **Nota**: Usa `--wait --timeout 15m` per attendere che il registry diventi `Active`.

---

## Step 3: Attendi che il Registry Diventi Attivo

```bash
acloud container containerregistry list
```

Attendi che `STATUS` mostri `Active`, poi ottieni l'indirizzo IP pubblico:

```bash
acloud container containerregistry get 69495ef64d0cdc87949b72ab
```

Annota l'**indirizzo IP pubblico** — è il nome host del registry per i comandi Docker.

---

## Step 4: Autenticati con il Registry

```bash
docker login <indirizzo-ip-elastic> --username registryadmin
# Password: (inserisci la password admin)
```

Output atteso:
```
Login Succeeded
```

> **Suggerimento CI/CD**:
> ```bash
> echo "$REGISTRY_PASSWORD" | docker login <indirizzo-ip-elastic> \
>   --username registryadmin --password-stdin
> ```

---

## Step 5: Carica un'Immagine nel Registry

```bash
docker tag my-app:latest <indirizzo-ip-elastic>/my-app:latest
docker push <indirizzo-ip-elastic>/my-app:latest
```

---

## Step 6: Scarica un'Immagine dal Registry

```bash
docker pull <indirizzo-ip-elastic>/my-app:latest
```

---

## Step 7: Gestisci il Registry

```bash
# Elenca tutti i registry
acloud container containerregistry list

# Cambia periodo di fatturazione in annuale
acloud container containerregistry update 69495ef64d0cdc87949b72ab \
  --billing-period "Year"

# Aumenta gli utenti concorrenti
acloud container containerregistry update 69495ef64d0cdc87949b72ab \
  --concurrent-users 20

# Rinomina il registry
acloud container containerregistry update 69495ef64d0cdc87949b72ab \
  --name "prod-registry"
```

---

## Best Practice

- Limita la porta 443 a intervalli IP noti per i registry di produzione
- Inizia con almeno 100 GB di storage e monitora l'utilizzo
- Usa tag con versione (es. `v1.2.3`) invece di affidarsi solo a `latest`
- Imposta `--concurrent-users` in base alla dimensione del team
- Passa alla fatturazione `Month` o `Year` per registry a lungo termine

---

## Step 8: Pulizia

```bash
acloud container containerregistry delete 69495ef64d0cdc87949b72ab --yes

# Elimina il block storage associato se non più necessario
acloud storage blockstorage delete 697b389bce7dfeef91532563 --yes
```

> **Attenzione**: L'eliminazione rimuove permanentemente tutte le immagini memorizzate. Assicurati che le immagini siano sottoposte a backup prima di eliminare.
