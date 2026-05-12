# Cloud Server

I cloud server sono istanze di macchine virtuali che eseguono le tue applicazioni e carichi di lavoro in Aruba Cloud.

## Sintassi dei Comandi

```bash
acloud compute cloudserver <command> [flags] [arguments]
```

## Comandi Disponibili

### `create`

Crea una nuova istanza di cloud server.

**Sintassi:**
```bash
acloud compute cloudserver create [flags]
```

**Flag Richiesti:**
- `--name <string>` - Nome per il cloud server
- `--region <string>` - Codice regione (es. `ITBG-Bergamo`)
- `--flavor <string>` - Nome flavor (es. `small`, `medium`, `large`)
- `--image <string>` - ID o nome immagine

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--keypair <string>` - Nome coppia di chiavi per accesso SSH
- `--tags <stringSlice>` - Tag (separati da virgola)
- `--user-data-file <string>` - Percorso al file YAML cloud-init (verrà codificato in base64)

**Esempio:**
```bash
acloud compute cloudserver create \
  --name "web-server" \
  --region "ITBG-Bergamo" \
  --flavor "small" \
  --image "ubuntu-22.04" \
  --keypair "my-keypair" \
  --tags "production,web" \
  --user-data-file "/percorso/al/cloud-init.yaml"
```

**Nota:** Il flag `--user-data-file` accetta un percorso a un file YAML cloud-init. Il contenuto del file verrà automaticamente codificato in base64 e incluso nella richiesta di creazione del cloud server. Questo ti consente di configurare il server durante l'inizializzazione utilizzando script cloud-init.

### `list`

Elenca tutti i cloud server nel progetto.

**Sintassi:**
```bash
acloud compute cloudserver list [flags]
```

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud compute cloudserver list
```

**Output:**
Il comando visualizza una tabella con le seguenti colonne:
- NAME
- ID
- LOCATION
- FLAVOR
- STATUS

### `get`

Ottieni informazioni dettagliate su un cloud server specifico.

**Sintassi:**
```bash
acloud compute cloudserver get <server-id> [flags]
```

**Argomenti:**
- `<server-id>` - L'ID del cloud server

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--verbose` - Mostra output JSON dettagliato

**Esempio:**
```bash
acloud compute cloudserver get 69495ef64d0cdc87949b71ec
```

**Output:**
Il comando visualizza informazioni dettagliate inclusi:
- ID e URI
- Nome e regione
- Dettagli flavor (CPU, RAM, HD)
- Informazioni immagine
- Coppia di chiavi (se configurata)
- Stato
- Data di creazione e creatore
- Tag

### `update`

Aggiorna le proprietà di un cloud server (nome, tag).

**Sintassi:**
```bash
acloud compute cloudserver update <server-id> [flags]
```

**Argomenti:**
- `<server-id>` - L'ID del cloud server

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--name <string>` - Nuovo nome per il cloud server
- `--tags <stringSlice>` - Nuovi tag (separati da virgola)

**Esempio:**
```bash
acloud compute cloudserver update 69495ef64d0cdc87949b71ec \
  --name "web-server-updated" \
  --tags "production,web,updated"
```

**Nota:** Almeno uno tra `--name` o `--tags` deve essere fornito.

### `delete`

Elimina un'istanza di cloud server.

**Sintassi:**
```bash
acloud compute cloudserver delete <server-id> [flags]
```

**Argomenti:**
- `<server-id>` - L'ID del cloud server

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma

**Esempio:**
```bash
acloud compute cloudserver delete 69495ef64d0cdc87949b71ec --yes
```

### `power-on`

Accendi un cloud server.

**Sintassi:**
```bash
acloud compute cloudserver power-on <server-id> [flags]
```

**Argomenti:**
- `<server-id>` - L'ID del cloud server

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud compute cloudserver power-on 69495ef64d0cdc87949b71ec
```

### `power-off`

Spegni un cloud server.

**Sintassi:**
```bash
acloud compute cloudserver power-off <server-id> [flags]
```

**Argomenti:**
- `<server-id>` - L'ID del cloud server

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud compute cloudserver power-off 69495ef64d0cdc87949b71ec
```

### `set-password`

Imposta o cambia la password per un cloud server.

**Sintassi:**
```bash
acloud compute cloudserver set-password <server-id> [flags]
```

**Argomenti:**
- `<server-id>` - L'ID del cloud server

**Flag Richiesti:**
- `--password <string>` - Nuova password per il cloud server

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud compute cloudserver set-password 69495ef64d0cdc87949b71ec --password "MySecurePassword123!"
```

**Nota sulla Sicurezza:** Le password fornite tramite flag da riga di comando possono essere visibili negli elenchi dei processi e nella cronologia della shell. Considera l'uso di variabili d'ambiente o strumenti di gestione password sicuri.

### `connect`

Ottieni informazioni di connessione SSH per un cloud server con Elastic IP.

**Sintassi:**
```bash
acloud compute cloudserver connect <server-id> [flags]
```

**Argomenti:**
- `<server-id>` - L'ID del cloud server

**Flag Richiesti:**
- `--user <string>` - Nome utente SSH (richiesto - vedi sotto per utenti specifici dell'immagine)

**Flag Opzionali:**
- `--project-id <string>` - ID progetto (usa il contesto se non specificato)

**Nome Utente SSH per Tipo di Immagine:**
Il nome utente SSH dipende dall'immagine/template usato quando si crea il cloud server:
- **Immagini Ubuntu/Debian**: Usa `ubuntu`
- **Immagini CentOS/RHEL**: Usa `centos` o `root`
- **Altre distribuzioni Linux**: Tipicamente `root`, ma controlla la documentazione dell'immagine
- **Immagini Windows**: Non applicabile (usa RDP invece)

Per informazioni dettagliate sull'accesso ai Cloud Server e utenti predefiniti, consulta la [Knowledge Base Aruba Cloud](https://kb.arubacloud.com/cmp/en/computing/cloud-server.aspx).

**Esempio:**
```bash
# Per immagini Ubuntu/Debian
acloud compute cloudserver connect 69495ef64d0cdc87949b71ec --user ubuntu

# Per immagini CentOS/RHEL
acloud compute cloudserver connect 69495ef64d0cdc87949b71ec --user centos
```

**Output:**
Il comando:
1. Ottiene i dettagli del cloud server
2. Verifica la presenza di un Elastic IP nelle risorse collegate
3. Recupera l'indirizzo Elastic IP
4. Stampa il comando di connessione SSH

```
Connect by running: ssh ubuntu@203.0.113.42
```

**Nota:** 
- Il cloud server deve avere un Elastic IP collegato per usare questo comando. Se non viene trovato un Elastic IP, il comando visualizzerà un messaggio di errore.
- Il flag `--user` è richiesto. Se non fornito o impostato su `<user>`, il comando visualizzerà un errore con indicazioni sugli utenti SSH comuni.

## Auto-completamento

La CLI fornisce auto-completamento per gli ID dei cloud server:

```bash
acloud compute cloudserver get <TAB>
acloud compute cloudserver update <TAB>
acloud compute cloudserver delete <TAB>
acloud compute cloudserver power-on <TAB>
acloud compute cloudserver power-off <TAB>
acloud compute cloudserver set-password <TAB>
acloud compute cloudserver connect <TAB>
```

## Workflow Comuni

### Avvio di un Nuovo Server

1. **Crea una coppia di chiavi** (se necessario):
   ```bash
   acloud compute keypair create --name "my-keypair" --public-key "$(cat ~/.ssh/id_rsa.pub)"
   ```

2. **Crea il cloud server**:
   ```bash
   acloud compute cloudserver create \
     --name "app-server" \
     --region "ITBG-Bergamo" \
     --flavor "medium" \
     --image "your-image-id" \
     --keypair "my-keypair" \
     --user-data-file "/percorso/al/cloud-init.yaml"
   ```

3. **Attendi che il server sia pronto** e controlla lo stato:
   ```bash
   acloud compute cloudserver get <server-id>
   ```

### Aggiornamento Metadati Server

```bash
# Aggiorna nome e tag del server
acloud compute cloudserver update <server-id> \
  --name "new-name" \
  --tags "production,updated"
```

### Gestione Stato Alimentazione Server

```bash
# Spegni un server
acloud compute cloudserver power-off <server-id>

# Accendi un server
acloud compute cloudserver power-on <server-id>

# Controlla lo stato del server
acloud compute cloudserver get <server-id>
```

### Connessione a un Server via SSH

```bash
# Ottieni il comando di connessione SSH (l'utente è richiesto)
# Per immagini Ubuntu/Debian
acloud compute cloudserver connect <server-id> --user ubuntu

# Per immagini CentOS/RHEL
acloud compute cloudserver connect <server-id> --user centos

# Il comando produrrà: "Connect by running: ssh user@ip-address"
```

**Importante:** Il nome utente SSH dipende dall'immagine/template usato. Valori predefiniti comuni:
- Ubuntu/Debian: `ubuntu`
- CentOS/RHEL: `centos` o `root`
- Altri Linux: `root` (controlla la documentazione dell'immagine)

Consulta la [Knowledge Base Aruba Cloud](https://kb.arubacloud.com/cmp/en/computing/cloud-server.aspx) per informazioni dettagliate sull'accesso ai Cloud Server.

## Best Practices

- **Denominazione**: Usa nomi descrittivi che indicano lo scopo del server (es. `web-server-prod`, `db-server-staging`)
- **Tag**: Usa tag per organizzare i server per ambiente, progetto o team
- **Flavor**: Scegli flavor appropriati in base ai requisiti del tuo carico di lavoro
- **Coppie di Chiavi**: Usa sempre coppie di chiavi per l'accesso SSH invece delle password
- **Monitoraggio**: Controlla lo stato del server prima di eseguire operazioni
- **Pulizia**: Elimina server non utilizzati per evitare costi non necessari

## Risorse Correlate

- [Coppie di Chiavi](keypair.md) - Gestisci coppie di chiavi SSH per l'accesso ai server
- [Risorse di Rete](../network.md) - Configura networking e security group
- [Risorse Storage](../storage.md) - Collega volumi di block storage

