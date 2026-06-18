# Installazione

Questa guida copre l'installazione di Aruba Cloud CLI sulla tua piattaforma, la verifica dell'installazione e la gestione degli aggiornamenti e della disinstallazione.

## Prerequisiti

- Accesso a Internet per scaricare i binari o i pacchetti
- Per compilare dal sorgente: Go 1.24 o successivo

## Installazione

### macOS — Homebrew

```bash
brew tap Arubacloud/tap
brew install acloud
```

Gli aggiornamenti vengono applicati automaticamente con `brew upgrade acloud`.

### Linux — apt (Debian / Ubuntu)

Aggiungi il repository apt di Arubacloud una sola volta, poi installa e aggiorna come qualsiasi pacchetto di sistema:

```bash
# Aggiungi la chiave di firma
curl -fsSL https://arubacloud.github.io/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/arubacloud.gpg

# Aggiungi il repository
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/arubacloud.gpg] https://arubacloud.github.io/apt stable main" | \
  sudo tee /etc/apt/sources.list.d/arubacloud.list

# Installa
sudo apt update && sudo apt install acloud
```

I futuri rilasci vengono applicati con `sudo apt upgrade acloud`.

### Linux — rpm (RHEL / Fedora / Amazon Linux)

```bash
sudo rpm -i https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud_linux_amd64.rpm
```

Per sistemi ARM64:
```bash
sudo rpm -i https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud_linux_arm64.rpm
```

### Windows — Scoop

```powershell
scoop bucket add arubacloud https://github.com/Arubacloud/scoop-bucket
scoop install acloud
```

Gli aggiornamenti vengono applicati con `scoop update acloud`.

### Installazione manuale del binario

I binari statici precompilati sono disponibili sulla [pagina delle release](https://github.com/Arubacloud/acloud-cli/releases/latest). Tutti i binari sono compilati staticamente senza dipendenze di runtime esterne e funzionano su tutte le principali distribuzioni Linux.

#### Linux AMD64

**Per Ubuntu 22.04+ o distribuzioni più recenti:**
```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64
sudo install -m 755 acloud-linux-amd64 /usr/local/bin/acloud
```

**Per Ubuntu 20.04 o distribuzioni WSL più vecchie (compatibile con GLIBC 2.31):**

Se incontri errori di versione GLIBC (es. `GLIBC_2.34 not found`), usa il binario compatibile con Ubuntu 20.04:
```bash
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64-ubuntu20.tar.gz
tar -xzf acloud-linux-amd64-ubuntu20.tar.gz
sudo mv acloud-linux-amd64-ubuntu20 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

#### Linux ARM64

```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-arm64
sudo install -m 755 acloud-linux-arm64 /usr/local/bin/acloud
```

#### macOS (Intel)

```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-amd64
sudo install -m 755 acloud-darwin-amd64 /usr/local/bin/acloud
```

#### macOS (Apple Silicon)

```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-darwin-arm64
sudo install -m 755 acloud-darwin-arm64 /usr/local/bin/acloud
```

#### Windows

**Usando PowerShell:**
```powershell
Invoke-WebRequest `
  -Uri "https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-windows-amd64.exe" `
  -OutFile "acloud.exe"
```

1. Scarica `acloud-windows-amd64.zip` dalla [pagina delle release](https://github.com/Arubacloud/acloud-cli/releases/latest)
2. Estrai il file ZIP e sposta `acloud.exe` in una cartella nel tuo `PATH` (es. `C:\Program Files\acloud-cli\`)

### Compila dal Sorgente

Requisiti:
- Go 1.24 o successivo

```bash
git clone https://github.com/Arubacloud/acloud-cli.git
cd acloud-cli
go build -o acloud
```

## Verifica dell'Installazione

```bash
# Controlla la versione installata
acloud --version
# acloud version v0.1.6

# Visualizza i comandi disponibili
acloud --help

# Testa la connettività API (richiede credenziali configurate)
acloud management project list
```

## Aggiornamento

### macOS (Homebrew)

```bash
brew upgrade acloud
```

### Linux (apt)

```bash
sudo apt update && sudo apt upgrade acloud
```

### Linux (rpm)

Scarica e installa l'ultimo pacchetto con il flag di aggiornamento:

```bash
sudo rpm -U https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud_linux_amd64.rpm
```

### Windows (Scoop)

```powershell
scoop update acloud
```

### Binario Manuale

Scarica l'ultimo binario dalla [pagina delle release](https://github.com/Arubacloud/acloud-cli/releases/latest) e sostituisci quello esistente:

```bash
curl -LO https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64
sudo install -m 755 acloud-linux-amd64 /usr/local/bin/acloud
```

## Disinstallazione

### macOS (Homebrew)

```bash
brew uninstall acloud
brew untap Arubacloud/tap
```

### Linux (apt)

```bash
sudo apt remove acloud
sudo rm /etc/apt/sources.list.d/arubacloud.list
sudo rm /etc/apt/keyrings/arubacloud.gpg
sudo apt update
```

### Linux (rpm)

```bash
sudo rpm -e acloud
```

### Windows (Scoop)

```powershell
scoop uninstall acloud
scoop bucket rm arubacloud
```

### Binario Manuale

```bash
sudo rm /usr/local/bin/acloud
```

### Rimozione dei File di Configurazione

Per rimuovere completamente tutti i dati della CLI incluse credenziali e contesti:

```bash
rm -rf ~/.config/acloud
```

> **Attenzione**: Questo elimina permanentemente tutti i profili, credenziali e impostazioni del contesto salvati. Fai il backup delle credenziali prima di eseguire questo comando se pensi di reinstallare.

## Prossimi Passi

- [Configura l'autenticazione](authentication.md) — Imposta le credenziali API e gestisci i profili
- [Esplora le opzioni di configurazione](configuration.md) — Gestione del contesto, formati di output e altro
- [Risorse](resources.md) — Esplora i tipi di risorse disponibili

## Risoluzione dei Problemi

### Errori di Versione GLIBC

Se vedi errori come:
```
acloud: /lib/x86_64-linux-gnu/libc.so.6: version 'GLIBC_2.34' not found
```

Questo significa che la tua distribuzione Linux ha una versione GLIBC più vecchia di quella richiesta. **Soluzione:** Usa il binario compatibile con Ubuntu 20.04:

```bash
wget https://github.com/Arubacloud/acloud-cli/releases/latest/download/acloud-linux-amd64-ubuntu20.tar.gz
tar -xzf acloud-linux-amd64-ubuntu20.tar.gz
sudo mv acloud-linux-amd64-ubuntu20 /usr/local/bin/acloud
sudo chmod +x /usr/local/bin/acloud
```

I binari compatibili con Ubuntu 20.04 funzionano su Ubuntu 20.04, 22.04, 24.04 e versioni più recenti.
