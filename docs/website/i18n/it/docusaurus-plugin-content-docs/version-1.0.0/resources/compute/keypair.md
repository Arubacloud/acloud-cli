# Coppie di Chiavi

Le coppie di chiavi SSH forniscono autenticazione sicura ai cloud server senza usare password.

## Sintassi dei Comandi

```bash
acloud compute keypair <command> [flags] [arguments]
```

## Comandi Disponibili

### `create`

Crea una nuova coppia di chiavi SSH.

**Sintassi:**
```bash
acloud compute keypair create [flags]
```

**Flag Richiesti:**
- `--name <string>` - Nome per la coppia di chiavi
- `--public-key <string>` - Valore della chiave pubblica (chiave pubblica SSH)
- `--region <string>` - Codice regione (es. `ITBG-Bergamo`)

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
# Usando un file
acloud compute keypair create \
  --name "my-keypair" \
  --region "ITBG-Bergamo" \
  --public-key "$(cat ~/.ssh/id_rsa.pub)"

# O direttamente
acloud compute keypair create \
  --name "my-keypair" \
  --region "ITBG-Bergamo" \
  --public-key "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... user@host"
```

### `list`

Elenca tutte le coppie di chiavi nel progetto.

**Sintassi:**
```bash
acloud compute keypair list [flags]
```

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud compute keypair list
```

**Output:**
Il comando visualizza una tabella con le seguenti colonne:
- NAME
- ID
- PUBLIC_KEY (troncato a 50 caratteri)

### `get`

Ottieni informazioni dettagliate su una coppia di chiavi specifica.

**Sintassi:**
```bash
acloud compute keypair get <keypair-name> [flags]
```

**Argomenti:**
- `<keypair-name>` - Il nome della coppia di chiavi

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--verbose` - Mostra output JSON dettagliato

**Esempio:**
```bash
acloud compute keypair get "my-keypair"
```

**Output:**
Il comando visualizza:
- Nome
- URI
- Chiave pubblica (valore completo)
- Data di creazione e creatore

### `update`

**Non supportato.** L'API non supporta l'aggiornamento della chiave pubblica di una coppia di chiavi. L'esecuzione di questo comando stampa un errore e termina senza apportare modifiche.

Per ruotare una coppia di chiavi, eliminarla e ricrearla con lo stesso nome:

```bash
acloud compute keypair delete "my-keypair" --yes
acloud compute keypair create \
  --name "my-keypair" \
  --region "ITBG-Bergamo" \
  --public-key "$(cat ~/.ssh/id_rsa_new.pub)"
```

### `delete`

Elimina una coppia di chiavi.

**Sintassi:**
```bash
acloud compute keypair delete <keypair-name> [flags]
```

**Argomenti:**
- `<keypair-name>` - Il nome della coppia di chiavi

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma

**Esempio:**
```bash
acloud compute keypair delete "my-keypair" --yes
```

## Auto-completamento

La CLI fornisce auto-completamento per i nomi delle coppie di chiavi:

```bash
acloud compute keypair get <TAB>
acloud compute keypair update <TAB>
acloud compute keypair delete <TAB>
```

## Workflow Comuni

### Creazione di una Coppia di Chiavi da Chiave SSH Esistente

```bash
# Se hai già una coppia di chiavi SSH
acloud compute keypair create \
  --name "my-laptop-key" \
  --region "ITBG-Bergamo" \
  --public-key "$(cat ~/.ssh/id_rsa.pub)"
```

### Generazione di una Nuova Coppia di Chiavi

1. **Genera una nuova coppia di chiavi SSH** (se necessario):
   ```bash
   ssh-keygen -t rsa -b 4096 -f ~/.ssh/aruba_key -N ""
   ```

2. **Crea la coppia di chiavi in Aruba Cloud**:
   ```bash
   acloud compute keypair create \
     --name "aruba-key" \
     --region "ITBG-Bergamo" \
     --public-key "$(cat ~/.ssh/aruba_key.pub)"
   ```

3. **Usa la coppia di chiavi quando crei cloud server**:
   ```bash
   acloud compute cloudserver create \
     --name "my-server" \
     --region "ITBG-Bergamo" \
     --zone "ITBG-1" \
     --flavor "CSO4A8" \
     --boot-disk-id "<volume-id>" \
     --vpc-id "<vpc-id>" \
     --subnet-id "<subnet-id>" \
     --security-group-id "<sg-id>" \
     --keypair-id "<keypair-id>"
   ```

### Rotazione delle Chiavi

Poiché l'aggiornamento della coppia di chiavi non è supportato dall'API, la rotazione richiede eliminazione + ricreazione:

1. **Genera una nuova coppia di chiavi SSH**:
   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/aruba_key_new -N ""
   ```

2. **Elimina la coppia di chiavi esistente**:
   ```bash
   acloud compute keypair delete "aruba-key" --yes
   ```

3. **Ricreala con la nuova chiave pubblica**:
   ```bash
   acloud compute keypair create \
     --name "aruba-key" \
     --region "ITBG-Bergamo" \
     --public-key "$(cat ~/.ssh/aruba_key_new.pub)"
   ```

4. **Aggiorna la tua configurazione SSH locale** per usare la nuova chiave privata e testa la connessione.

## Best Practices

- **Denominazione**: Usa nomi descrittivi che indicano lo scopo o il proprietario della chiave (es. `user-john-laptop`, `ci-cd-server`, `admin-key`)
- **Sicurezza delle Chiavi**: 
  - Non condividere mai o esporre chiavi private
  - Usa tipi di chiave forti (RSA 4096-bit o Ed25519)
  - Proteggi le chiavi private con passphrase
- **Rotazione delle Chiavi**: Ruota le chiavi regolarmente per la sicurezza
- **Chiavi Multiple**: Usa coppie di chiavi diverse per ambienti o scopi diversi
- **Backup**: Mantieni backup sicuri delle tue chiavi private
- **Controllo Accessi**: Concedi l'accesso alla coppia di chiavi solo agli utenti autorizzati

## Risorse Correlate

- [Cloud Server](cloudserver.md) - Usa le coppie di chiavi quando crei server
- [Risorse di Sicurezza](../security.md) - Best practice di sicurezza aggiuntive

