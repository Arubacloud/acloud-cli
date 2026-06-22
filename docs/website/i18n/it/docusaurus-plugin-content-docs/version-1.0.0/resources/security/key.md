# Chiavi Crittografiche

Le chiavi crittografiche sono annidate all'interno di un'istanza KMS e forniscono il materiale di cifratura effettivo. Ogni chiave ha un algoritmo (AES simmetrico o RSA asimmetrico) e uno stato del ciclo di vita.

## Comandi Disponibili

- `acloud security key create` - Crea una nuova chiave crittografica in un'istanza KMS
- `acloud security key list` - Elenca tutte le chiavi in un'istanza KMS
- `acloud security key get` - Ottieni i dettagli di una chiave specifica
- `acloud security key delete` - Elimina una chiave

## Crea Chiave

Crea una nuova chiave crittografica all'interno di un'istanza KMS esistente.

### Utilizzo

```bash
acloud security key create --kms-id <kms-id> --name <nome> --algorithm <algoritmo> [flags]
```

### Flag Richiesti

- `--kms-id` - ID dell'istanza KMS padre
- `--name` - Nome per la chiave
- `--algorithm` - Algoritmo crittografico: `Aes` (simmetrico) o `Rsa` (asimmetrico)

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud security key create \
  --kms-id "69455aa70d0972656501d45d" \
  --name "my-aes-key" \
  --algorithm "Aes"
```

## Elenca Chiavi

Elenca tutte le chiavi all'interno di un'istanza KMS.

### Utilizzo

```bash
acloud security key list --kms-id <kms-id> [flags]
```

### Flag Richiesti

- `--kms-id` - ID dell'istanza KMS padre

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--limit` - Numero massimo di risultati da restituire
- `--offset` - Numero di risultati da saltare

### Esempio

```bash
acloud security key list --kms-id "69455aa70d0972656501d45d"
```

## Ottieni Dettagli Chiave

Recupera informazioni dettagliate su una chiave specifica.

### Utilizzo

```bash
acloud security key get <key-id> --kms-id <kms-id> [flags]
```

### Argomenti

- `key-id` (richiesto): L'ID univoco della chiave

### Flag Richiesti

- `--kms-id` - ID dell'istanza KMS padre

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud security key get abc123 --kms-id "69455aa70d0972656501d45d"
```

## Elimina Chiave

Elimina una chiave crittografica. Questa azione è irreversibile.

### Utilizzo

```bash
acloud security key delete <key-id> --kms-id <kms-id> [--yes] [flags]
```

### Argomenti

- `key-id` (richiesto): L'ID univoco della chiave

### Flag Richiesti

- `--kms-id` - ID dell'istanza KMS padre

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma
- `--dry-run` - Valida l'esistenza della risorsa senza eliminare

### Esempio

```bash
acloud security key delete abc123 --kms-id "69455aa70d0972656501d45d" --yes
```

## Risorse Correlate

- [Gestione Chiavi KMS](kms.md) - Gestisci le istanze KMS che contengono le chiavi
- [Servizi KMIP](kmip.md) - Gestisci i servizi KMIP annidati nelle istanze KMS
