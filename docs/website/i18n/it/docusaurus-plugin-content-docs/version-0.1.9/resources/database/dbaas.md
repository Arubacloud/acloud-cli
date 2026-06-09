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
acloud database dbaas create --name <name> --region <region> --zone <zone> --engine-id <engine-id> --flavor <flavor> --storage-size <gb> [flags]
```

### Flag Richiesti

- `--name` - Nome per l'istanza DBaaS
- `--region` - Codice regione (es. "ITBG-Bergamo")
- `--zone` - Zona di disponibilità (es. "ITBG-1")
- `--engine-id` - Identificatore engine database (es. `mysql-8.0`)
- `--flavor` - Codice flavor/piano database (es. `DBO4A8`)
- `--storage-size` - Dimensione storage in GB (es. `50`)

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--vpc-id` - ID VPC (richiesto quando il progetto ha una VPC)
- `--subnet-id` - ID Subnet (richiesto quando il progetto ha una VPC)
- `--security-group-id` - ID security group (richiesto quando il progetto ha una VPC)
- `--elastic-ip-id` - ID Elastic IP per accesso pubblico
- `--tags` - Tag (separati da virgola)

### Esempio

```bash
acloud database dbaas create \
  --name "my-database" \
  --region "ITBG-Bergamo" \
  --zone "ITBG-1" \
  --engine-id "mysql-8.0" \
  --flavor "DBO4A8" \
  --storage-size 50 \
  --vpc-id "<vpc-id>" \
  --subnet-id "<subnet-id>" \
  --security-group-id "<sg-id>" \
  --elastic-ip-id "<eip-id>" \
  --tags "production,mysql"
```

**Nota:** Le operazioni `update` (PUT) su database e utenti non sono supportate dall'API e restituiscono HTTP 405. Per sostituire un database o un utente, eliminarlo e ricrearlo.

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

