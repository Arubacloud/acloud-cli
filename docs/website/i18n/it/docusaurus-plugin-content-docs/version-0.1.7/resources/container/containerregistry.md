# Container Registry

Container Registry fornisce un registry Docker container privato per memorizzare e gestire immagini container in Aruba Cloud.

## Sintassi Comandi

```bash
acloud container containerregistry <command> [flags] [arguments]
```

## Comandi Disponibili

### `create`

Crea un nuovo container registry.

**Sintassi:**
```bash
acloud container containerregistry create [flags]
```

**Flag Richiesti:**
- `--name <string>` - Nome per il container registry
- `--region <string>` - Codice regione (es. `ITBG-Bergamo`)
- `--public-ip-uri <string>` - URI IP pubblico (es. `/projects/{project-id}/providers/Aruba.Network/elasticIps/{elasticip-id}`)
- `--vpc-uri <string>` - URI VPC (es. `/projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id}`)
- `--subnet-uri <string>` - URI Subnet (es. `/projects/{project-id}/providers/Aruba.Network/subnets/{subnet-id}`)
- `--security-group-uri <string>` - URI security group
- `--block-storage-uri <string>` - URI block storage

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--tags <stringSlice>` - Tag (separati da virgola)
- `--billing-period <string>` - Periodo di fatturazione: Hour, Month, Year (opzionale)
- `--admin-username <string>` - Nome utente amministratore (opzionale)
- `--concurrent-users <string>` - Numero di utenti concorrenti (opzionale)

**Esempio:**
```bash
acloud container containerregistry create \
  --name "my-registry" \
  --region "ITBG-Bergamo" \
  --public-ip-uri "/projects/{project-id}/providers/Aruba.Network/elasticIps/{elasticip-id}" \
  --vpc-uri "/projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id}" \
  --subnet-uri "/projects/{project-id}/providers/Aruba.Network/subnets/{subnet-id}" \
  --security-group-uri "/projects/{project-id}/providers/Aruba.Network/securityGroups/{sg-id}" \
  --block-storage-uri "/projects/{project-id}/providers/Aruba.Storage/volumes/{volume-id}" \
  --billing-period "Month" \
  --tags "production,registry"
```

### `list`

Elenca tutti i container registry nel progetto.

**Sintassi:**
```bash
acloud container containerregistry list [flags]
```

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud container containerregistry list
```

**Output:**
Il comando visualizza una tabella con le seguenti colonne:
- NAME
- ID
- REGION
- STATUS

### `get`

Ottieni informazioni dettagliate su un container registry specifico.

**Sintassi:**
```bash
acloud container containerregistry get <registry-id> [flags]
```

**Argomenti:**
- `<registry-id>` - L'ID del container registry

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--verbose` - Mostra output JSON dettagliato

**Esempio:**
```bash
acloud container containerregistry get 69495ef64d0cdc87949b71ec
```

**Output:**
Il comando visualizza informazioni dettagliate inclusi:
- ID e URI
- Nome e regione
- IP pubblico, VPC, Subnet, Security Group, Block Storage
- Piano di fatturazione
- Utente admin (se configurato)
- Utenti concorrenti
- Stato
- Data di creazione e creatore
- Tag

### `update`

Aggiorna le proprietà di un container registry (nome, tag, periodo di fatturazione, utenti concorrenti).

**Sintassi:**
```bash
acloud container containerregistry update <registry-id> [flags]
```

**Argomenti:**
- `<registry-id>` - L'ID del container registry

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--name <string>` - Nuovo nome per il container registry
- `--tags <stringSlice>` - Nuovi tag (separati da virgola)
- `--billing-period <string>` - Periodo di fatturazione: Hour, Month, Year
- `--concurrent-users <string>` - Numero di utenti concorrenti

**Esempio:**
```bash
acloud container containerregistry update 69495ef64d0cdc87949b71ec \
  --name "my-registry-updated" \
  --tags "production,registry,updated" \
  --billing-period "Year" \
  --concurrent-users 10
```

**Nota:** Almeno uno tra `--name`, `--tags`, `--billing-period`, o `--concurrent-users` deve essere fornito.

### `delete`

Elimina un container registry.

**Sintassi:**
```bash
acloud container containerregistry delete <registry-id> [flags]
```

**Argomenti:**
- `<registry-id>` - L'ID del container registry

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma

**Esempio:**
```bash
acloud container containerregistry delete 69495ef64d0cdc87949b71ec --yes
```

## Auto-completamento

La CLI fornisce auto-completamento per gli ID container registry:

```bash
acloud container containerregistry get <TAB>
acloud container containerregistry update <TAB>
acloud container containerregistry delete <TAB>
```

## Workflow Comuni

### Creazione di un Container Registry

1. **Assicurati che le risorse richieste esistano:**
   - IP Pubblico (Elastic IP)
   - VPC
   - Subnet
   - Security Group
   - Block Storage

2. **Crea il container registry:**
   ```bash
   acloud container containerregistry create \
     --name "my-registry" \
     --region "ITBG-Bergamo" \
     --public-ip-uri "/projects/{project-id}/providers/Aruba.Network/elasticIps/{elasticip-id}" \
     --vpc-uri "/projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id}" \
     --subnet-uri "/projects/{project-id}/providers/Aruba.Network/subnets/{subnet-id}" \
     --security-group-uri "/projects/{project-id}/providers/Aruba.Network/securityGroups/{sg-id}" \
     --block-storage-uri "/projects/{project-id}/providers/Aruba.Storage/volumes/{volume-id}" \
     --billing-period "Month"
   ```

3. **Attendi che il registry sia pronto** e controlla lo stato:
   ```bash
   acloud container containerregistry get <registry-id>
   ```

### Aggiornamento Metadati Registry

```bash
# Aggiorna nome e tag del registry
acloud container containerregistry update <registry-id> \
  --name "new-name" \
  --tags "production,updated"

# Aggiorna periodo di fatturazione
acloud container containerregistry update <registry-id> \
  --billing-period "Year"

# Aggiorna utenti concorrenti
acloud container containerregistry update <registry-id> \
  --concurrent-users 20
```

## Best Practices

- **Denominazione**: Usa nomi descrittivi che indicano lo scopo del registry (es. `prod-registry`, `dev-registry`)
- **Tag**: Usa tag per organizzare i registry per ambiente, progetto o team
- **Periodo di Fatturazione**: Scegli periodi di fatturazione appropriati in base all'utilizzo previsto
- **Utenti Concorrenti**: Imposta utenti concorrenti in base alla dimensione del team e ai pattern di utilizzo
- **Sicurezza**: Assicurati che i security group siano configurati correttamente per limitare l'accesso
- **Storage**: Monitora l'utilizzo del block storage e espandi secondo necessità
- **Pulizia**: Elimina registry non utilizzati per evitare costi non necessari

## Risorse Correlate

- [KaaS](./kaas.md) - Cluster Kubernetes per eseguire applicazioni containerizzate
- [Risorse di Rete](../network.md) - Configura networking, VPC e security group
- [Risorse Storage](../storage.md) - Block storage per dati registry
