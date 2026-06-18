---
id: authentication
title: Autenticazione
sidebar_label: Autenticazione
description: Configura le credenziali API e gestisci più profili per Aruba Cloud CLI.
---

# Autenticazione

Aruba Cloud CLI richiede credenziali API per autenticarsi con i servizi Aruba Cloud.

## Configurazione delle Credenziali

1. **Ottieni le Credenziali API**: Ottieni il tuo Client ID e Client Secret dalla console Aruba Cloud.

2. **Configura la CLI** — passa `--client-id` sulla riga di comando; il secret viene letto in modo sicuro con echo disabilitato:
   ```bash
   acloud config set --client-id YOUR_CLIENT_ID
   # Enter client secret: (input nascosto, non appare nella cronologia della shell)
   ```

   Per CI/automazione, imposta il secret tramite variabile d'ambiente:
   ```bash
   ACLOUD_CLIENT_SECRET=YOUR_CLIENT_SECRET acloud config set --client-id YOUR_CLIENT_ID
   ```

   > **Nota di sicurezza**: `--client-secret` non è supportato intenzionalmente per evitare di esporre i secret nei log di processo e nella cronologia della shell.

3. **Verifica la configurazione**:
   ```bash
   acloud config show
   ```

## File di Configurazione

Le credenziali sono memorizzate in `~/.config/acloud/config.yaml` (percorso XDG Base Directory, permessi `0600`):

```yaml
profiles:
  default:
    clientId: your-client-id
    clientSecret: your-client-secret
```

> **Percorso legacy**: Se hai usato una versione precedente di acloud che salvava le credenziali in `~/.acloud.yaml`, la CLI migra automaticamente quel file alla nuova posizione al primo avvio e mostra un avviso una tantum. Non è necessaria alcuna azione manuale.

**Nota sulla Sicurezza**: Mantieni le tue credenziali sicure. Il file di configurazione contiene informazioni sensibili.

## Configurazione del Client

La configurazione della CLI permette di gestire le credenziali API e impostazioni opzionali come endpoint API personalizzati.

### Impostazione della Configurazione

**Impostazioni Richieste:**

`--client-id` è obbligatorio. `clientSecret` viene letto da `ACLOUD_CLIENT_SECRET` (automazione) o richiesto in modo sicuro con input nascosto (interattivo):

```bash
# Consigliato: secret inserito tramite prompt nascosto (non appare nella cronologia della shell)
acloud config set --client-id YOUR_CLIENT_ID

# CI/automazione: fornisci il secret tramite variabile d'ambiente
ACLOUD_CLIENT_SECRET=YOUR_CLIENT_SECRET acloud config set --client-id YOUR_CLIENT_ID
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
ACLOUD_CLIENT_SECRET=YOUR_CLIENT_SECRET \
acloud config set \
  --client-id YOUR_CLIENT_ID \
  --base-url "https://api.arubacloud.com" \
  --token-issuer-url "https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token"
```

### Visualizzazione della Configurazione

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

La configurazione è memorizzata in `~/.config/acloud/config.yaml` usando un envelope multi-profilo:

```yaml
profiles:
  default:
    clientId: your-client-id
    clientSecret: your-client-secret
    baseUrl: https://api.arubacloud.com          # opzionale
    tokenIssuerUrl: https://login.aruba.it/...   # opzionale
  prod:
    clientId: prod-client-id
    clientSecret: prod-client-secret
```

**Valori Predefiniti:**

Se `baseUrl` e `tokenIssuerUrl` non sono specificati, la CLI usa questi valori predefiniti:
- **Base URL**: `https://api.arubacloud.com`
- **Token Issuer URL**: `https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token`

### Aggiornamento della Configurazione

```bash
# Aggiorna solo il client secret
ACLOUD_CLIENT_SECRET=NEW_SECRET acloud config set --client-id YOUR_CLIENT_ID

# Aggiorna solo il base URL
acloud config set --base-url "https://custom-api.example.com"
```

**Nota**: Sia `clientId` che `clientSecret` devono sempre essere presenti nella configurazione. Se stai aggiornando uno, assicurati che l'altro sia già impostato in config/ambiente. Per aggiornamenti interattivi del secret, esegui con `--client-id` e fornisci il secret quando richiesto:

```bash
acloud config set --client-id YOUR_CLIENT_ID   # richiesto in modo sicuro
```

## Gestione dei Profili di Credenziali

Quando lavori con più account Aruba Cloud — ad esempio un account personale, un ambiente di staging e uno di produzione — i profili ti permettono di memorizzare ogni set di credenziali sotto un nome e passare dall'uno all'altro con un singolo flag.

### Creare un Profilo

Usa `acloud config profile set <nome>` per creare o aggiornare un profilo. Il client secret viene letto da `ACLOUD_CLIENT_SECRET` (consigliato per l'automazione) o richiesto in modo sicuro con input nascosto:

```bash
# Crea il profilo "staging" — secret inserito in modo interattivo
acloud config profile set staging --client-id YOUR_STAGING_CLIENT_ID

# Crea il profilo "prod" — secret dalla variabile d'ambiente
ACLOUD_CLIENT_SECRET=YOUR_PROD_SECRET \
  acloud config profile set prod \
  --client-id YOUR_PROD_CLIENT_ID \
  --base-url "https://api.arubacloud.com"
```

Puoi aggiornare un singolo campo di un profilo esistente senza toccare gli altri:

```bash
# Rinnova il client ID nel profilo prod mantenendo il secret esistente
acloud config profile set prod --client-id NEW_PROD_CLIENT_ID
```

### Selezionare il Profilo Attivo

Tre modi per selezionare il profilo che un comando utilizza, in ordine di priorità:

| Metodo | Esempio | Note |
|--------|---------|------|
| Flag `--profile` | `acloud --profile prod network vpc list` | Priorità massima; sovrascrive la variabile d'ambiente |
| Variabile `ACLOUD_PROFILE` | `ACLOUD_PROFILE=staging acloud storage blockstorage list` | Utile nelle pipeline CI/CD |
| Default | *(nessun flag o variabile)* | Usa il profilo `default` |

```bash
# Comando singolo su prod
acloud --profile prod management project list

# Imposta il profilo per l'intera sessione shell
export ACLOUD_PROFILE=staging
acloud network vpc list
acloud storage blockstorage list

# Ripristina il comportamento predefinito
unset ACLOUD_PROFILE
```

### Elencare i Profili

```bash
acloud config profile list
```

Esempio di output (il profilo attivo è contrassegnato con `*`):

```
PROFILE              CLIENT_ID                        BASE_URL
* default            default-client-id
  prod               prod-client-id                   https://api.arubacloud.com
  staging            staging-client-id
```

### Eliminare un Profilo

```bash
acloud config profile delete staging
# Profile "staging" deleted.
```

### Formato del File di Configurazione (Multi-Profilo)

Tutti i profili sono memorizzati insieme in `~/.config/acloud/config.yaml` sotto la chiave `profiles:`:

```yaml
profiles:
  default:
    clientId: default-client-id
    clientSecret: default-secret
  prod:
    clientId: prod-client-id
    clientSecret: prod-secret
    baseUrl: https://api.arubacloud.com
  staging:
    clientId: staging-client-id
    clientSecret: staging-secret
```

> **Compatibilità con le versioni precedenti**: I file di configurazione a profilo singolo (il vecchio formato piatto `clientId: / clientSecret:`) continuano a funzionare e vengono automaticamente trattati come profilo `default`. Non vengono riscritti finché non si esegue `acloud config profile set` o `acloud config set`.

### Usare i Profili con la Gestione del Contesto

I profili (credenziali) e i contesti (ID progetto) sono indipendenti — puoi combinarli liberamente:

```bash
# Usa le credenziali prod + un ID progetto da un contesto salvato
acloud --profile prod context use my-prod-project
acloud --profile prod network vpc list

# Oppure passa l'ID progetto esplicitamente
acloud --profile prod network vpc list --project-id YOUR_PROJECT_ID
```

## Risoluzione dei Problemi

### "Error initializing client"

Questo di solito significa che le credenziali non sono configurate. Esegui:

```bash
acloud config set
```

### "No projects found"

Assicurati che le tue credenziali abbiano i permessi corretti e che tu abbia progetti nel tuo account.
