# Installazione

Questa guida copre l'installazione di Aruba Cloud CLI sulla tua piattaforma e la configurazione iniziale.

## Installazione

### Scarica il Binario Pre-compilato

Scarica l'ultima release per la tua piattaforma dalla [pagina delle release](https://github.com/Arubacloud/acloud-cli/releases).

#### Linux AMD64

**Per Ubuntu 22.04+ o distribuzioni più recenti:**
```bash
# Scarica ed estrai
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64.tar.gz
tar -xzf acloud-linux-amd64.tar.gz

# Sposta nel PATH
sudo mv acloud-linux-amd64 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

**Per Ubuntu 20.04 o distribuzioni WSL più vecchie (compatibile con GLIBC 2.31):**

Se incontri errori di versione GLIBC (es. `GLIBC_2.34 not found`), usa il binario compatibile con Ubuntu 20.04:
```bash
# Scarica ed estrai il binario compatibile con Ubuntu 20.04
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64-ubuntu20.tar.gz
tar -xzf acloud-linux-amd64-ubuntu20.tar.gz

# Sposta nel PATH
sudo mv acloud-linux-amd64-ubuntu20 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

> **Nota:** Il binario compatibile con Ubuntu 20.04 funziona su Ubuntu 20.04, 22.04, 24.04 e versioni più recenti. Usa questa versione se stai usando distribuzioni WSL più vecchie o incontri problemi di compatibilità GLIBC.

#### Linux ARM64

**Per Ubuntu 22.04+ o distribuzioni più recenti:**
```bash
# Scarica ed estrai
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-arm64.tar.gz
tar -xzf acloud-linux-arm64.tar.gz

# Sposta nel PATH
sudo mv acloud-linux-arm64 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

**Per Ubuntu 20.04 o distribuzioni più vecchie (compatibile con GLIBC 2.31):**

Se incontri errori di versione GLIBC, usa il binario compatibile con Ubuntu 20.04:
```bash
# Scarica ed estrai il binario compatibile con Ubuntu 20.04
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-arm64-ubuntu20.tar.gz
tar -xzf acloud-linux-arm64-ubuntu20.tar.gz

# Sposta nel PATH
sudo mv acloud-linux-arm64-ubuntu20 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

#### macOS (Intel)
```bash
# Scarica ed estrai
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-amd64.tar.gz
tar -xzf acloud-darwin-amd64.tar.gz

# Sposta nel PATH
sudo mv acloud-darwin-amd64 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

#### macOS (Apple Silicon)
```bash
# Scarica ed estrai
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-arm64.tar.gz
tar -xzf acloud-darwin-arm64.tar.gz

# Sposta nel PATH
sudo mv acloud-darwin-arm64 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

#### Windows

**Usando PowerShell:**
```powershell
# Scarica
Invoke-WebRequest -Uri "https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-windows-amd64.zip" -OutFile "acloud-windows-amd64.zip"

# Estrai
Expand-Archive -Path acloud-windows-amd64.zip -DestinationPath .

# Aggiungi al PATH (sessione corrente)
$env:Path += ";$PWD"

# Oppure sposta in una posizione permanente e aggiungi al PATH
# Sposta acloud-windows-amd64.exe in C:\Program Files\acloud-cli\
# Poi aggiungi C:\Program Files\acloud-cli\ al PATH di sistema
```

**Usando Command Prompt:**
1. Scarica `acloud-windows-amd64.zip` dalla [pagina delle release](https://github.com/Arubacloud/acloud-cli/releases)
2. Estrai il file ZIP
3. Sposta `acloud-windows-amd64.exe` in una cartella (es. `C:\Program Files\acloud-cli\`)
4. Aggiungi quella cartella alla variabile d'ambiente PATH di sistema
5. Rinomina `acloud-windows-amd64.exe` in `acloud.exe` per comodità

### Compila dal Sorgente

Requisiti:
- Go 1.24 o successivo

```bash
git clone https://github.com/Arubacloud/acloud-cli.git
cd acloud-cli
go build -o acloud
```

## Autenticazione

Aruba Cloud CLI richiede credenziali API per autenticarsi con i servizi Aruba Cloud.

### Configurazione delle Credenziali

1. **Ottieni le Credenziali API**: Ottieni il tuo Client ID e Client Secret dalla console Aruba Cloud.

2. **Configura la CLI**:
   ```bash
   acloud config set
   ```

3. **Inserisci le tue credenziali** quando richiesto:
   - Client ID
   - Client Secret

4. **Verifica la configurazione**:
   ```bash
   acloud config show
   ```

### File di Configurazione

Le credenziali sono memorizzate in `~/.acloud.yaml`:

```yaml
clientId: your-client-id
clientSecret: your-client-secret
```

**Nota sulla Sicurezza**: Mantieni le tue credenziali sicure. Il file di configurazione contiene informazioni sensibili.

### Variabili d'Ambiente

Puoi anche impostare le credenziali tramite variabili d'ambiente:

```bash
export ACLOUD_CLIENT_ID="your-client-id"
export ACLOUD_CLIENT_SECRET="your-client-secret"
```

## Configurazione

La configurazione della CLI ti permette di gestire le credenziali API e impostazioni opzionali come endpoint API personalizzati.

### Impostazione della Configurazione

**Impostazioni Richieste:**

Sia `--client-id` che `--client-secret` sono obbligatori e devono essere impostati insieme:

```bash
acloud config set --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET
```

**Impostazioni Opzionali:**

Puoi opzionalmente configurare endpoint API personalizzati:

```bash
# Imposta base URL (default: https://api.arubacloud.com)
acloud config set --base-url "https://api.arubacloud.com"

# Imposta token issuer URL (default: https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token)
acloud config set --token-issuer-url "https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token"
```

Puoi anche impostare tutti i valori in una volta:

```bash
acloud config set \
  --client-id YOUR_CLIENT_ID \
  --client-secret YOUR_CLIENT_SECRET \
  --base-url "https://api.arubacloud.com" \
  --token-issuer-url "https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token"
```

### Visualizzazione della Configurazione

Per visualizzare la tua configurazione corrente:

```bash
acloud config show
```

Esempio di output:
```
Current configuration:
  Client ID: your-client-id
  Client Secret: ********
  Base URL: https://api.arubacloud.com (default)
  Token Issuer URL: https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token (default)
```

### Formato del File di Configurazione

La configurazione è memorizzata in `~/.acloud.yaml`:

```yaml
clientId: your-client-id
clientSecret: your-client-secret
baseUrl: https://api.arubacloud.com  # Opzionale, usa il default se non impostato
tokenIssuerUrl: https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token  # Opzionale, usa il default se non impostato
```

**Valori Predefiniti:**

Se `baseUrl` e `tokenIssuerUrl` non sono specificati nel file di configurazione, la CLI usa questi valori predefiniti:
- **Base URL**: `https://api.arubacloud.com`
- **Token Issuer URL**: `https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token`

### Aggiornamento della Configurazione

Puoi aggiornare impostazioni individuali senza influenzare le altre:

```bash
# Aggiorna solo il client secret
acloud config set --client-secret NEW_SECRET

# Aggiorna solo il base URL
acloud config set --base-url "https://custom-api.example.com"
```

**Nota**: Sia `--client-id` che `--client-secret` devono sempre essere presenti nella configurazione. Se stai aggiornando uno, assicurati che l'altro sia già impostato o fornisci entrambi.

## Gestione del Contesto {#context-management}

La CLI fornisce la gestione del contesto per evitare di passare `--project-id` ripetutamente. I contesti ti permettono di salvare gli ID progetto e passare tra di essi facilmente.

### Impostazione di un Contesto

Crea un contesto con un ID progetto:

```bash
acloud context set my-prod --project-id "66a10244f62b99c686572a9f"
```

### Utilizzo di un Contesto

Passa a un contesto salvato:

```bash
acloud context use my-prod
```

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

I contesti sono memorizzati in `~/.acloud-context.yaml`:

```yaml
current-context: my-prod
contexts:
  my-prod:
    project-id: 66a10244f62b99c686572a9f
  my-dev:
    project-id: 66a10244f62b99c686572a9e
```

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

## Verifica dell'Installazione

Testa la tua installazione:

```bash
# Controlla la versione
acloud --version

# Visualizza i comandi disponibili
acloud --help

# Testa la connettività API
acloud management project list
```

## Prossimi Passi

- Scopri la [Gestione Progetti](resources/management/project.md)
- Esplora le [Risorse di Gestione](resources/management.md)
- Esplora le [Risorse Storage](resources/storage.md)
- Esplora la [Documentazione delle Risorse](resources/management.md)

## Modalità Debug

La CLI fornisce un flag globale `--debug` (o `-d`) che abilita il logging dettagliato per aiutare a risolvere i problemi. Quando abilitato, mostra:

- **Dettagli Richiesta/Risposta HTTP**: Tutte le richieste e risposte HTTP fatte dall'SDK
- **Payload delle richieste**: Corpi delle richieste formattati in JSON inviati all'API
- **Dettagli degli errori**: Corpi completi delle risposte di errore quando le richieste falliscono

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

### Errori di Versione GLIBC

Se vedi errori come:
```
acloud: /lib/x86_64-linux-gnu/libc.so.6: version 'GLIBC_2.34' not found
```

Questo significa che la tua distribuzione Linux ha una versione GLIBC più vecchia di quella richiesta. **Soluzione:** Usa il binario compatibile con Ubuntu 20.04:

```bash
# Scarica invece il binario compatibile con Ubuntu 20.04
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64-ubuntu20.tar.gz
tar -xzf acloud-linux-amd64-ubuntu20.tar.gz
sudo mv acloud-linux-amd64-ubuntu20 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

I binari compatibili con Ubuntu 20.04 funzionano su Ubuntu 20.04, 22.04, 24.04 e versioni più recenti.

### "Error initializing client"

Questo di solito significa che le credenziali non sono configurate. Esegui:
```bash
acloud config set
```

### "No projects found"

Assicurati che le tue credenziali abbiano i permessi corretti e che tu abbia progetti nel tuo account.

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

