# Gestione Database DBaaS

I database all'interno di istanze DBaaS memorizzano i tuoi dati. Ogni istanza DBaaS può contenere più database.

## Comandi Disponibili

- `acloud database dbaas database create` - Crea un nuovo database in un'istanza DBaaS
- `acloud database dbaas database list` - Elenca tutti i database in un'istanza DBaaS
- `acloud database dbaas database get` - Ottieni i dettagli di un database specifico
- `acloud database dbaas database update` - Aggiorna nome database
- `acloud database dbaas database delete` - Elimina un database

## Crea Database

Crea un nuovo database all'interno di un'istanza DBaaS.

### Utilizzo

```bash
acloud database dbaas database create <dbaas-id> --name <name> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS

### Flag Richiesti

- `--name` - Nome per il database

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas database create 69455aa70d0972656501d45d \
  --name "my-database"
```

## Elenca Database

Elenca tutti i database in un'istanza DBaaS.

### Utilizzo

```bash
acloud database dbaas database list <dbaas-id> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas database list 69455aa70d0972656501d45d
```

## Ottieni Dettagli Database

Recupera informazioni dettagliate su un database specifico.

### Utilizzo

```bash
acloud database dbaas database get <dbaas-id> <database-name> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS
- `database-name` (richiesto): Il nome del database

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas database get 69455aa70d0972656501d45d "my-database"
```

## Aggiorna Database

Rinomina un database.

### Utilizzo

```bash
acloud database dbaas database update <dbaas-id> <database-name> --name <new-name> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS
- `database-name` (richiesto): Il nome corrente del database

### Flag Richiesti

- `--name` - Nuovo nome per il database

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas database update 69455aa70d0972656501d45d "my-database" \
  --name "renamed-database"
```

## Elimina Database

Elimina un database da un'istanza DBaaS.

### Utilizzo

```bash
acloud database dbaas database delete <dbaas-id> <database-name> [--yes] [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS
- `database-name` (richiesto): Il nome del database da eliminare

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma

### Esempio

```bash
acloud database dbaas database delete 69455aa70d0972656501d45d "my-database" --yes
```

## Risorse Correlate

- [DBaaS](dbaas.md) - Gestisci istanze DBaaS
- [Utenti DBaaS](dbaas.user.md) - Gestisci utenti per database
- [Backup Database](backup.md) - Crea backup di database
