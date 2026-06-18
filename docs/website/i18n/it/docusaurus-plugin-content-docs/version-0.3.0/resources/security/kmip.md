# Servizi KMIP

I servizi KMIP (Key Management Interoperability Protocol) sono annidati all'interno di un'istanza KMS ed espongono un endpoint KMIP per l'autenticazione basata su certificati client. Dopo la creazione, un certificato e una coppia di chiavi private possono essere scaricati una volta che il servizio raggiunge lo stato `CertificateAvailable`.

## Comandi Disponibili

- `acloud security kmip create` - Crea un nuovo servizio KMIP in un'istanza KMS
- `acloud security kmip list` - Elenca tutti i servizi KMIP in un'istanza KMS
- `acloud security kmip get` - Ottieni i dettagli di un servizio KMIP specifico
- `acloud security kmip delete` - Elimina un servizio KMIP
- `acloud security kmip download` - Scarica il certificato e la chiave privata KMIP (PEM)

## Crea Servizio KMIP

Crea un nuovo servizio KMIP all'interno di un'istanza KMS esistente.

### Utilizzo

```bash
acloud security kmip create --kms-id <kms-id> --name <nome> [flags]
```

### Flag Richiesti

- `--kms-id` - ID dell'istanza KMS padre
- `--name` - Nome per il servizio KMIP

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--wait` - Attendi finché il certificato non è disponibile (rispetta `--timeout`)

### Esempio

```bash
acloud security kmip create \
  --kms-id "69455aa70d0972656501d45d" \
  --name "my-kmip-service" \
  --wait
```

## Elenca Servizi KMIP

Elenca tutti i servizi KMIP all'interno di un'istanza KMS.

### Utilizzo

```bash
acloud security kmip list --kms-id <kms-id> [flags]
```

### Flag Richiesti

- `--kms-id` - ID dell'istanza KMS padre

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--limit` - Numero massimo di risultati da restituire
- `--offset` - Numero di risultati da saltare

### Esempio

```bash
acloud security kmip list --kms-id "69455aa70d0972656501d45d"
```

## Ottieni Dettagli Servizio KMIP

Recupera informazioni dettagliate su un servizio KMIP specifico.

### Utilizzo

```bash
acloud security kmip get <kmip-id> --kms-id <kms-id> [flags]
```

### Argomenti

- `kmip-id` (richiesto): L'ID univoco del servizio KMIP

### Flag Richiesti

- `--kms-id` - ID dell'istanza KMS padre

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud security kmip get abc123 --kms-id "69455aa70d0972656501d45d"
```

## Elimina Servizio KMIP

Elimina un servizio KMIP.

### Utilizzo

```bash
acloud security kmip delete <kmip-id> --kms-id <kms-id> [--yes] [flags]
```

### Argomenti

- `kmip-id` (richiesto): L'ID univoco del servizio KMIP

### Flag Richiesti

- `--kms-id` - ID dell'istanza KMS padre

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma
- `--dry-run` - Valida l'esistenza della risorsa senza eliminare

### Esempio

```bash
acloud security kmip delete abc123 --kms-id "69455aa70d0972656501d45d" --yes
```

## Scarica Certificato

Scarica il certificato PEM e la chiave privata per un servizio KMIP. Il servizio deve aver raggiunto lo stato `CertificateAvailable` prima che il download sia disponibile.

### Utilizzo

```bash
acloud security kmip download <kmip-id> --kms-id <kms-id> [flags]
```

### Argomenti

- `kmip-id` (richiesto): L'ID univoco del servizio KMIP

### Flag Richiesti

- `--kms-id` - ID dell'istanza KMS padre

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud security kmip download abc123 --kms-id "69455aa70d0972656501d45d"
```

L'output contiene sia il certificato che la chiave privata in formato PEM, che puoi reindirizzare su file:

```bash
acloud security kmip download abc123 --kms-id "69455aa70d0972656501d45d" > kmip-cert.pem
```

## Risorse Correlate

- [Gestione Chiavi KMS](kms.md) - Gestisci le istanze KMS che contengono i servizi KMIP
- [Chiavi Crittografiche](key.md) - Gestisci le chiavi crittografiche annidate nelle istanze KMS
