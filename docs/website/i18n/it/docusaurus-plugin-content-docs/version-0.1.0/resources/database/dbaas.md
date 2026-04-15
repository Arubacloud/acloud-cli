# Gestione DBaaS

Le istanze DBaaS (Database as a Service) forniscono servizi database gestiti in Aruba Cloud.

## Comandi Disponibili

- `acloud database dbaas create` - Crea una nuova istanza DBaaS
- `acloud database dbaas list` - Elenca tutte le istanze DBaaS
- `acloud database dbaas get` - Ottieni i dettagli di un'istanza DBaaS specifica
- `acloud database dbaas update` - Aggiorna tag istanza DBaaS
- `acloud database dbaas delete` - Elimina un'istanza DBaaS

## Crea Istanza DBaaS

Crea una nuova istanza DBaaS nel tuo progetto.

### Utilizzo

```bash
acloud database dbaas create --name <name> --region <region> --engine-id <engine-id> --flavor <flavor> [flags]
```

### Flag Richiesti

- `--name` - Nome per l'istanza DBaaS
- `--region` - Codice regione (es. "ITBG-Bergamo")
- `--engine-id` - ID engine database
- `--flavor` - Nome flavor/piano database

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--tags` - Tag (separati da virgola)

### Esempio

```bash
acloud database dbaas create \
  --name "my-database" \
  --region "ITBG-Bergamo" \
  --engine-id "69455aa70d0972656501d45d" \
  --flavor "db.t3.micro" \
  --tags "production,postgresql"
```

## Elenca Istanze DBaaS

Elenca tutte le istanze DBaaS nel tuo progetto.

### Utilizzo

```bash
acloud database dbaas list [flags]
```

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas list
```

## Ottieni Dettagli Istanza DBaaS

Recupera informazioni dettagliate su un'istanza DBaaS specifica.

### Utilizzo

```bash
acloud database dbaas get <dbaas-id> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas get 69455aa70d0972656501d45d
```

## Aggiorna Istanza DBaaS

Aggiorna i tag per un'istanza DBaaS.

### Utilizzo

```bash
acloud database dbaas update <dbaas-id> --tags <tags> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--tags` - Nuovi tag (separati da virgola)

### Esempio

```bash
acloud database dbaas update 69455aa70d0972656501d45d --tags "production,updated"
```

## Elimina Istanza DBaaS

Elimina un'istanza DBaaS.

### Utilizzo

```bash
acloud database dbaas delete <dbaas-id> [--yes] [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma

### Esempio

```bash
acloud database dbaas delete 69455aa70d0972656501d45d --yes
```

## Risorse Correlate

- [Database DBaaS](dbaas.database.md) - Gestisci database all'interno di istanze DBaaS
- [Utenti DBaaS](dbaas.user.md) - Gestisci utenti per istanze DBaaS
- [Backup Database](backup.md) - Crea e gestisci backup database

