---
id: configuration
title: Configurazione
sidebar_label: Configurazione
description: Configura la gestione del contesto, l'auto-completamento della shell, i formati di output e le opzioni avanzate della CLI.
---

# Configurazione

Questa pagina copre le opzioni di configurazione della CLI oltre alle credenziali: gestione del contesto, auto-completamento della shell, formati di output e flag avanzati.

## Gestione del Contesto {#context-management}

La CLI fornisce la gestione del contesto per evitare di passare `--project-id` ripetutamente. I contesti ti permettono di salvare gli ID progetto e passare tra di essi facilmente.

### Impostazione di un Contesto

Crea un contesto con un ID progetto:

```bash
acloud context set my-prod --project-id "66a10244f62b99c686572a9f"
```

> **Scorciatoia per contesto singolo**: Quando è configurato un solo contesto, la CLI lo utilizza automaticamente — non è necessario eseguire prima `acloud context use`. Se in seguito si aggiunge un secondo contesto, sarà necessario selezionarlo esplicitamente con `acloud context use <nome>`.

### Utilizzo di un Contesto

Passa a un contesto salvato quando ne hai più di uno:

```bash
acloud context use my-prod
```

> **Cambio di profilo**: I contesti memorizzano gli ID progetto che appartengono a un tenant specifico. Quando si cambia profilo con `acloud config profile use`, la CLI chiederà se cancellare tutti i contesti (poiché gli ID progetto salvati potrebbero non essere validi per il nuovo profilo). Usa `--clear-contexts` per cancellarli senza essere interrogato:
> ```bash
> acloud config profile use staging --clear-contexts
> ```

Una volta che un contesto è attivo, puoi eseguire comandi senza specificare `--project-id`:

```bash
# Funziona senza --project-id
acloud storage blockstorage list
acloud storage snapshot list
acloud management project get <project-id>
```

### Gestione dei Contesti

**Elenca tutti i contesti:**
```bash
acloud context list
```

L'output mostra tutti i contesti con quello corrente marcato con `*`:
```
Contexts:
=========
my-prod              Project ID: 66a10244f62b99c686572a9f *
my-dev               Project ID: 66a10244f62b99c686572a9e
my-staging           Project ID: 66a10244f62b99c686572a9d

* = current context
```

**Mostra il contesto corrente:**
```bash
acloud context current
```

**Elimina un contesto:**
```bash
acloud context delete my-dev
```

### File del Contesto

I contesti sono memorizzati in `~/.config/acloud/context.yaml`:

```yaml
current-context: my-prod
contexts:
  my-prod:
    project-id: 66a10244f62b99c686572a9f
  my-dev:
    project-id: 66a10244f62b99c686572a9e
```

> **Percorso legacy**: Le versioni precedenti salvavano questo file in `~/.acloud-context.yaml`. La CLI lo migra automaticamente al primo avvio.

### Override del Contesto

Puoi sempre sovrascrivere il contesto passando esplicitamente `--project-id`:

```bash
# Usa l'ID progetto del contesto
acloud storage blockstorage list

# Sovrascrive con un ID progetto specifico
acloud storage blockstorage list --project-id "different-project-id"
```

## Auto-completamento

La CLI supporta l'auto-completamento shell per comandi, flag e ID risorse.

### Bash

#### Sessione Corrente
```bash
source <(acloud completion bash)
```

#### Installazione Permanente

**Linux:**
```bash
acloud completion bash | sudo tee /etc/bash_completion.d/acloud
```

**macOS:**
```bash
acloud completion bash > $(brew --prefix)/etc/bash_completion.d/acloud
```

Dopo l'installazione, riavvia la shell o esegui:
```bash
source ~/.bashrc  # o ~/.bash_profile su macOS
```

### Zsh

Aggiungi a `~/.zshrc`:

```bash
# Abilita il completamento
autoload -Uz compinit
compinit

# Carica il completamento acloud
source <(acloud completion zsh)
```

Oppure per installazione permanente:
```bash
acloud completion zsh > "${fpath[1]}/_acloud"
```

### Fish

```bash
acloud completion fish | source
```

Oppure per installazione permanente:
```bash
acloud completion fish > ~/.config/fish/completions/acloud.fish
```

### PowerShell

Aggiungi al tuo profilo PowerShell:

```powershell
acloud completion powershell | Out-String | Invoke-Expression
```

## Funzionalità dell'Auto-completamento

Il sistema di auto-completamento fornisce:

1. **Completamento comandi**: Completa con tab comandi e sottocomandi
   ```bash
   acloud man<TAB>  # completa in "management"
   ```

2. **Completamento flag**: Completa con tab i flag disponibili
   ```bash
   acloud config set --<TAB>  # mostra i flag disponibili
   ```

3. **Completamento ID risorse**: Completa con tab gli ID risorse con descrizioni

   **Risorse di Gestione:**
   ```bash
   acloud management project get <TAB>
   # Mostra:
   # 655b2822af30f667f826994e    defaultproject
   # 66a10244f62b99c686572a9f    develop
   # ...
   ```

   **Risorse Storage:**
   ```bash
   # Block Storage
   acloud storage blockstorage get <TAB>
   # Mostra:
   # 6965a6c3ffc0fd1ef8ba5612    MyVolume
   # 6965a6c3ffc0fd1ef8ba5613    DataVolume
   # ...

   # Snapshots
   acloud storage snapshot get <TAB>
   # Mostra:
   # 696c9edce63c1af07d60d0c7    MySnapshot
   # 696c9edce63c1af07d60d0c8    BackupSnapshot
   # ...

   # Backups
   acloud storage backup get <TAB>
   # Mostra:
   # 67649dac8c7bb1c5d7c80631    MyBackup
   # 67649dac8c7bb1c5d7c80632    DailyBackup
   # ...

   # Restores (gerarchico: backup-id poi restore-id)
   acloud storage restore get <TAB>
   # Prima mostra gli ID backup:
   # 67649dac8c7bb1c5d7c80631    MyBackup
   # ...
   acloud storage restore get 67649dac8c7bb1c5d7c80631 <TAB>
   # Poi mostra gli ID restore per quel backup:
   # 67664dde0aca19a92c2c48bb    RestoreOperation1
   # ...
   ```

   L'auto-completamento funziona con i comandi `get`, `update` e `delete` per tutte le risorse.

## Formato di Output

Tutti i comandi `list` e `get` supportano un flag globale `--output` (o `-o`) che controlla il formato di output.

| Valore | Descrizione |
|--------|-------------|
| `table` | Tabella leggibile a larghezza fissa (predefinito) |
| `table-json` | Array JSON di oggetti piatti con chiavi snake_case (uno per riga) |
| `table-yaml` | Sequenza YAML di mapping piatti con chiavi snake_case (uno per riga) |
| `json` | Risposta completa dell'SDK come JSON indentato |
| `yaml` | Risposta completa dell'SDK come YAML |

```bash
# Output tabella predefinito
acloud network vpc list

# Array JSON piatto — facile da passare a jq
acloud network vpc list -o table-json

# Risposta completa dell'SDK
acloud network vpc list -o json
```

`table-json` è la scelta migliore per gli script con strumenti come `jq`:
```bash
acloud storage blockstorage list -o table-json | jq '.[].name'
```

## Provisioning Sincrono (--wait)

Per impostazione predefinita, i comandi `create` e `update` ritornano non appena l'API accetta la richiesta — la risorsa potrebbe essere ancora in fase di provisioning in background. Usa `--wait` per attendere finché la risorsa raggiunge uno stato operativo (`Active`, `Running`, ecc.) o fallisce.

```bash
# Attendi che il VPC sia Active
acloud network vpc create --name "prod-vpc" --region "IT-BG" --wait

# Attendi che il DBaaS sia Active; estendi il timeout a 10 minuti
acloud --timeout 10m database dbaas create --name "prod-db" --region "IT-BG" \
  --engine-id "mysql-8.0" --flavor "DBO4A8" --storage-size 50 \
  --zone "ITBG-1" --wait
```

Il progresso viene stampato su stderr (per non mescolarsi con l'output `-o json`):
```
Waiting for VPC 'prod-vpc' to become Active... [12s]
VPC 'prod-vpc' is Active after 15s.
```

**Risorse supportate:** `cloudserver`, `vpc`, `subnet`, `securitygroup`, `vpntunnel`, `elasticip`, `kaas`, `containerregistry`, `dbaas`, `kms`, `blockstorage` — sia nei comandi create che update.

**Timeout:** `--wait` rispetta il flag globale `--timeout` (default `30s`). Per risorse che richiedono più tempo per il provisioning (cluster KaaS, istanze DBaaS), imposta un timeout più lungo:
```bash
acloud --timeout 15m container kaas create ... --wait
```

**Codice di uscita:** Ritorna non-zero se la risorsa raggiunge uno stato di errore (`Error`, `Failed`) o il timeout scade prima che la risorsa diventi operativa.

## Eliminazione Sicura (Dry Run)

Ogni comando di eliminazione supporta `--dry-run` per verificare che una risorsa esista senza eliminarla effettivamente:

```bash
acloud storage blockstorage delete <volume-id> --dry-run
# [dry-run] Would delete block storage '<volume-id>'. Resource exists and is accessible.
```

Usa `--yes` (o `-y`) per saltare la richiesta di conferma interattiva negli script non interattivi.

## Paginazione

Tutti i comandi `list` supportano i flag `--limit` e `--offset` per navigare tra grandi set di risultati.

| Flag | Descrizione |
|------|-------------|
| `--limit N` | Restituisce al massimo N risultati |
| `--offset N` | Salta i primi N risultati |

```bash
# Prima pagina di 10 risultati
acloud storage blockstorage list --limit 10

# Seconda pagina
acloud storage blockstorage list --limit 10 --offset 10
```

Se nessun flag viene passato, l'API restituisce il suo set di risultati predefinito.

## Modalità Debug

La CLI fornisce un flag globale `--debug` (o `-d`) che abilita il logging dettagliato per aiutare a risolvere i problemi. Quando abilitato, mostra:

- **Dettagli Richiesta/Risposta HTTP**: Tutte le richieste e risposte HTTP fatte dall'SDK
- **Payload delle richieste**: Corpi delle richieste formattati in JSON inviati all'API
- **Dettagli degli errori**: Corpi completi delle risposte di errore quando le richieste falliscono

> **Avvertenza di sicurezza**: L'output di debug può includere credenziali e token dagli header HTTP. Non usare `--debug` in sessioni terminal condivise o incollare il suo output pubblicamente.

### Utilizzo

Aggiungi il flag `--debug` a qualsiasi comando:

```bash
# Abilita il logging debug per un comando
acloud --debug network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --tags test

# Forma breve
acloud -d network vpc list
```

### Esempio di Output

Quando la modalità debug è abilitata, vedrai output aggiuntivo come:

```
[ArubaSDK] 2025-01-15 10:30:45.123456 HTTP Request: PUT https://api.arubacloud.com/...
[ArubaSDK] 2025-01-15 10:30:45.234567 Request Headers: ...
[ArubaSDK] 2025-01-15 10:30:45.345678 Request Body: {...}

=== DEBUG: Security Rule Update Request ===
VPC ID: 69495ef64d0cdc87949b71ec
Security Group ID: 694b05ac4d0cdc87949b75f9
Security Rule ID: 694b06564d0cdc87949b7608
Request Payload:
{
  "metadata": {
    "name": "my-rule",
    "tags": ["test"],
    ...
  },
  ...
}
==========================================

[ArubaSDK] 2025-01-15 10:30:46.456789 HTTP Response: 200 OK
[ArubaSDK] 2025-01-15 10:30:46.567890 Response Body: {...}
```

**Nota**: L'output di debug viene inviato a `stderr`, quindi non interferirà con l'output normale del comando e può essere reindirizzato separatamente se necessario.

## Risoluzione dei Problemi

### Debug degli Errori API

Se incontri errori API (es. 500 Internal Server Error), usa il flag `--debug` per vedere la richiesta e risposta completa:

```bash
acloud --debug network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --tags test
```

Questo mostrerà:
- Il payload esatto della richiesta inviata
- La risposta HTTP completa (inclusi i dettagli dell'errore)
- Qualsiasi logging a livello SDK

### Auto-completamento non funziona

1. Assicurati che bash-completion sia installato:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install bash-completion

   # macOS
   brew install bash-completion
   ```

2. Ricarica la configurazione della shell:
   ```bash
   source ~/.bashrc
   ```
