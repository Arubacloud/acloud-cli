# KaaS (Kubernetes as a Service)

KaaS fornisce cluster Kubernetes gestiti per eseguire applicazioni containerizzate in Aruba Cloud.

## Sintassi Comandi

```bash
acloud container kaas <command> [flags] [arguments]
```

## Comandi Disponibili

### `create`

Crea un nuovo cluster KaaS.

**Sintassi:**
```bash
acloud container kaas create [flags]
```

**Flag Richiesti:**
- `--name <string>` - Nome per il cluster KaaS
- `--region <string>` - Codice regione (es. `ITBG-Bergamo`)
- `--vpc-uri <string>` - URI VPC (es. `/projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id}`)
- `--subnet-uri <string>` - URI Subnet (es. `/projects/{project-id}/providers/Aruba.Network/subnets/{subnet-id}`)
- `--node-cidr-address <string>` - Indirizzo CIDR nodo in notazione CIDR (es. `10.0.0.0/16`)
- `--node-cidr-name <string>` - Nome CIDR nodo
- `--security-group-name <string>` - Nome security group
- `--kubernetes-version <string>` - Versione Kubernetes (es. `1.28.0`)
- `--node-pool-name <string>` - Nome node pool
- `--node-pool-nodes <int>` - Numero di nodi nel node pool
- `--node-pool-instance <string>` - Nome configurazione istanza per i nodi
- `--node-pool-zone <string>` - Codice datacenter/zona per i nodi

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--tags <stringSlice>` - Tag (separati da virgola)
- `--pod-cidr <string>` - Pod CIDR (opzionale)
- `--ha` - Abilita alta disponibilità
- `--billing-period <string>` - Periodo di fatturazione: Hour, Month, Year (opzionale)
- `--api-server-authorized-ip-ranges <stringSlice>` - Intervalli IP autorizzati per accesso API server
- `--api-server-enable-private-cluster` - Abilita cluster privato per API server
- `--node-pool-autoscaling` - Abilita autoscaling per node pool
- `--node-pool-min-count <int>` - Numero minimo di nodi per autoscaling
- `--node-pool-max-count <int>` - Numero massimo di nodi per autoscaling

**Esempio:**
```bash
acloud container kaas create \
  --name "production-cluster" \
  --region "ITBG-Bergamo" \
  --vpc-uri "/projects/66a10244f62b99c686572a9f/providers/Aruba.Network/vpcs/69495ef64d0cdc87949b71ec" \
  --subnet-uri "/projects/66a10244f62b99c686572a9f/providers/Aruba.Network/subnets/694b05ac4d0cdc87949b75f9" \
  --node-cidr-address "10.0.0.0/16" \
  --node-cidr-name "node-cidr" \
  --security-group-name "kaas-sg" \
  --kubernetes-version "1.28.0" \
  --node-pool-name "default-pool" \
  --node-pool-nodes 3 \
  --node-pool-instance "small" \
  --node-pool-zone "ITBG-Bergamo-A" \
  --tags "production,kubernetes" \
  --ha
```

### `list`

Elenca tutti i cluster KaaS nel progetto.

**Sintassi:**
```bash
acloud container kaas list [flags]
```

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud container kaas list
```

**Output:**
Il comando visualizza una tabella con le seguenti colonne:
- ID
- NAME
- VERSION
- REGION
- STATUS

### `get`

Ottieni informazioni dettagliate su un cluster KaaS specifico.

**Sintassi:**
```bash
acloud container kaas get <cluster-id> [flags]
```

**Argomenti:**
- `<cluster-id>` - L'ID del cluster KaaS

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--verbose` - Mostra output JSON dettagliato

**Esempio:**
```bash
acloud container kaas get 69495ef64d0cdc87949b71ec
```

**Output:**
Il comando visualizza informazioni dettagliate inclusi:
- ID e URI
- Nome e regione
- Versione Kubernetes
- Stato
- Data di creazione e creatore
- Tag

### `update`

Aggiorna metadati e proprietà di un cluster KaaS.

**Sintassi:**
```bash
acloud container kaas update <cluster-id> [flags]
```

**Argomenti:**
- `<cluster-id>` - L'ID del cluster KaaS

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--name <string>` - Nuovo nome per il cluster KaaS
- `--tags <stringSlice>` - Nuovi tag (separati da virgola)
- `--kubernetes-version <string>` - Versione Kubernetes a cui aggiornare
- `--kubernetes-version-upgrade-date <string>` - Data aggiornamento versione Kubernetes (formato ISO 8601)
- `--ha` - Abilita/disabilita alta disponibilità
- `--storage-max-cumulative-volume-size <int>` - Dimensione massima cumulativa volume per storage
- `--billing-period <string>` - Periodo di fatturazione: Hour, Month, Year
- `--node-pool-name <string>` - Nome node pool da aggiornare
- `--node-pool-nodes <int>` - Numero di nodi nel node pool
- `--node-pool-instance <string>` - Nome configurazione istanza per i nodi
- `--node-pool-zone <string>` - Codice datacenter/zona per i nodi
- `--node-pool-autoscaling` - Abilita autoscaling per node pool
- `--node-pool-min-count <int>` - Numero minimo di nodi per autoscaling
- `--node-pool-max-count <int>` - Numero massimo di nodi per autoscaling

**Esempio:**
```bash
# Aggiorna metadati (nome e tag)
acloud container kaas update 69495ef64d0cdc87949b71ec \
  --name "production-cluster-updated" \
  --tags "production,kubernetes,updated"

# Aggiorna versione Kubernetes
acloud container kaas update 69495ef64d0cdc87949b71ec \
  --kubernetes-version "1.29.0"

# Aggiorna node pool
acloud container kaas update 69495ef64d0cdc87949b71ec \
  --node-pool-name "default-pool" \
  --node-pool-nodes 5 \
  --node-pool-autoscaling \
  --node-pool-min-count 3 \
  --node-pool-max-count 10
```

**Nota:** Puoi aggiornare metadati (nome, tag) e proprietà (versione Kubernetes, node pool, ecc.) nello stesso comando.

### `delete`

Elimina un cluster KaaS.

**Sintassi:**
```bash
acloud container kaas delete <cluster-id> [flags]
```

**Argomenti:**
- `<cluster-id>` - L'ID del cluster KaaS

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma

**Esempio:**
```bash
acloud container kaas delete 69495ef64d0cdc87949b71ec --yes
```

**Avviso:** Eliminare un cluster rimuoverà tutti i carichi di lavoro e i dati. Assicurati di avere backup se necessario.

### `connect`

Connetti a un cluster KaaS e configura kubectl automaticamente.

**Sintassi:**
```bash
acloud container kaas connect <cluster-id> [flags]
```

**Argomenti:**
- `<cluster-id>` - L'ID del cluster KaaS

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud container kaas connect 69495ef64d0cdc87949b71ec
```

**Cosa fa:**
1. Scarica il kubeconfig dal cluster KaaS
2. Crea un file kubeconfig in `$HOME/.kube/` nominato con il nome del cluster
3. Aggiorna `$HOME/.kube/config` con la configurazione del cluster
4. Esegue `kubectl cluster-info` per verificare la connessione
5. Stampa messaggio di successo se la connessione è riuscita, o errore se fallisce

**Prerequisiti:**
- `kubectl` deve essere installato e disponibile nel tuo PATH
- Il cluster deve essere in uno stato pronto

**Output:**
```
KaaS successfully connected
Kubeconfig saved to: /home/user/.kube/production-cluster
Default config updated: /home/user/.kube/config
```

## Auto-completamento

La CLI fornisce auto-completamento per gli ID cluster KaaS:

```bash
acloud container kaas get <TAB>
acloud container kaas update <TAB>
acloud container kaas delete <TAB>
acloud container kaas connect <TAB>
```

## Workflow Comuni

### Creazione di un Nuovo Cluster

1. **Crea il cluster KaaS**:
   ```bash
   acloud container kaas create \
     --name "my-k8s-cluster" \
     --region "ITBG-Bergamo" \
     --version "1.28.0"
   ```

2. **Attendi che il cluster sia pronto**:
   ```bash
   # Controlla lo stato del cluster
   acloud container kaas get <cluster-id>
   # Attendi fino a quando lo stato è "Active"
   ```

3. **Connetti al cluster** (dopo che il cluster è pronto):
   ```bash
   # Configura automaticamente kubectl
   acloud container kaas connect <cluster-id>
   
   # Verifica connessione:
   kubectl get nodes
   kubectl cluster-info
   ```

### Aggiornamento Cluster

```bash
# Aggiorna nome e tag del cluster
acloud container kaas update <cluster-id> \
  --name "production-cluster" \
  --tags "production,kubernetes,updated"

# Aggiorna versione Kubernetes
acloud container kaas update <cluster-id> \
  --kubernetes-version "1.29.0"

# Scala node pool
acloud container kaas update <cluster-id> \
  --node-pool-name "default-pool" \
  --node-pool-nodes 5
```

### Connessione a un Cluster

```bash
# Connetti e configura kubectl automaticamente
acloud container kaas connect <cluster-id>

# Dopo la connessione, usa kubectl normalmente
kubectl get nodes
kubectl get pods --all-namespaces
```

## Best Practices

- **Denominazione**: Usa nomi descrittivi che indicano lo scopo del cluster (es. `prod-k8s-cluster`, `staging-cluster`)
- **Tag**: Usa tag per organizzare i cluster per ambiente, progetto o team
- **Versioning**: 
  - Mantieni le versioni Kubernetes aggiornate per sicurezza e funzionalità
  - Testa gli aggiornamenti di versione in staging prima della produzione
- **Pianificazione Risorse**: Pianifica le risorse del cluster in base ai requisiti del carico di lavoro
- **Monitoraggio**: Monitora lo stato del cluster, la salute e l'utilizzo delle risorse
- **Sicurezza**: 
  - Segui le best practice di sicurezza Kubernetes
  - Usa RBAC per il controllo degli accessi
  - Abilita network policies
  - Aggiorna regolarmente i componenti del cluster
- **Backup**: Implementa strategie di backup per volumi persistenti e dati applicazione
- **Pulizia**: Elimina cluster non utilizzati per evitare costi non necessari

## Risorse Correlate

- [Risorse di Rete](../network.md) - Configura networking per carichi di lavoro container
- [Risorse Storage](../storage.md) - Volumi persistenti per applicazioni container
- [Risorse di Sicurezza](../security.md) - Politiche di sicurezza e gestione chiavi
- [Risorse Compute](../compute.md) - Cloud server che possono eseguire carichi di lavoro container
