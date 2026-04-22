# Gestione Chiavi KMS

Le chiavi KMS (Key Management System) forniscono gestione delle chiavi di crittografia per proteggere i tuoi dati e risorse in Aruba Cloud.

## Comandi Disponibili

- `acloud security kms create` - Crea una nuova chiave KMS
- `acloud security kms list` - Elenca tutte le chiavi KMS
- `acloud security kms get` - Ottieni i dettagli di una chiave KMS specifica
- `acloud security kms update` - Aggiorna nome e tag chiave KMS
- `acloud security kms delete` - Elimina una chiave KMS

## Crea Chiave KMS

Crea una nuova chiave KMS nel tuo progetto.

### Utilizzo

```bash
acloud security kms create --name <name> --region <region> [flags]
```

### Flag Richiesti

- `--name` - Nome per la chiave KMS
- `--region` - Codice regione (es. "ITBG-Bergamo")

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--billing-period` - Periodo di fatturazione: Hour, Month, Year (default: "Hour")
- `--tags` - Tag (separati da virgola)

### Esempio

```bash
acloud security kms create \
  --name "my-encryption-key" \
  --region "ITBG-Bergamo" \
  --billing-period "Hour" \
  --tags "production,encryption"
```

## Elenca Chiavi KMS

Elenca tutte le chiavi KMS nel tuo progetto.

### Utilizzo

```bash
acloud security kms list [flags]
```

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud security kms list
```

## Ottieni Dettagli Chiave KMS

Recupera informazioni dettagliate su una chiave KMS specifica.

### Utilizzo

```bash
acloud security kms get <kms-id> [flags]
```

### Argomenti

- `kms-id` (richiesto): L'ID univoco della chiave KMS

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud security kms get 69455aa70d0972656501d45d
```

## Aggiorna Chiave KMS

Aggiorna nome e tag chiave KMS.

### Utilizzo

```bash
acloud security kms update <kms-id> [flags]
```

### Argomenti

- `kms-id` (richiesto): L'ID univoco della chiave KMS

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--name` - Nuovo nome per la chiave KMS
- `--tags` - Nuovi tag (separati da virgola)

### Esempio

```bash
acloud security kms update 69455aa70d0972656501d45d \
  --name "updated-key-name" \
  --tags "production,updated"
```

## Elimina Chiave KMS

Elimina una chiave KMS.

### Utilizzo

```bash
acloud security kms delete <kms-id> [--yes] [flags]
```

### Argomenti

- `kms-id` (richiesto): L'ID univoco della chiave KMS

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma

### Esempio

```bash
acloud security kms delete 69455aa70d0972656501d45d --yes
```

## Best Practices di Sicurezza

- Usa nomi descrittivi per le chiavi KMS
- Organizza le chiavi usando tag
- Ruota le chiavi regolarmente secondo la tua politica di sicurezza
- Monitora l'utilizzo e l'accesso alle chiavi
- Usa chiavi diverse per ambienti diversi (dev, staging, production)
- Non condividere mai o esporre materiale chiave

## Risorse Correlate

- [Risorse Database](../database/dbaas.md) - Usa chiavi KMS con database
- [Risorse Storage](../storage/blockstorage.md) - Crittografa volumi storage

