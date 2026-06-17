# Gestione Grant DBaaS

I grant definiscono quali utenti hanno accesso a quali database all'interno di un'istanza DBaaS, e con quale livello di privilegi (es. `liteadmin`).

## Comandi Disponibili

- `acloud database dbaas grant create` - Concede a un utente l'accesso a un database
- `acloud database dbaas grant list` - Elenca tutti i grant su un database
- `acloud database dbaas grant get` - Ottieni i dettagli di un grant specifico
- `acloud database dbaas grant delete` - Revoca un grant

## Crea Grant

Concede a un utente l'accesso a un database specifico all'interno di un'istanza DBaaS.

### Utilizzo

```bash
acloud database dbaas grant create <dbaas-id> <database-name> --username <username> --role <role> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS
- `database-name` (richiesto): Il nome del database a cui concedere l'accesso

### Flag Richiesti

- `--username` - Il nome utente a cui concedere l'accesso
- `--role` - Il ruolo di privilegio (es. `liteadmin`)

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas grant create 69455aa70d0972656501d45d "my-database" \
  --username "restapi" \
  --role "liteadmin"
```

> **Nota:** L'istanza DBaaS deve essere in stato `Active` prima di creare i grant. L'operazione di aggiornamento grant (PUT) non è supportata dall'API — revoca e ricrea il grant per modificare il ruolo.

## Elenca Grant

Elenca tutti i grant su un database specifico all'interno di un'istanza DBaaS.

### Utilizzo

```bash
acloud database dbaas grant list <dbaas-id> <database-name> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS
- `database-name` (richiesto): Il nome del database

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas grant list 69455aa70d0972656501d45d "my-database"
```

## Ottieni Dettagli Grant

Recupera i dettagli di un grant specifico.

### Utilizzo

```bash
acloud database dbaas grant get <dbaas-id> <database-name> <grant-id> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS
- `database-name` (richiesto): Il nome del database
- `grant-id` (richiesto): L'ID univoco del grant

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas grant get 69455aa70d0972656501d45d "my-database" 69455aa70d0972656501d4ab
```

## Elimina Grant

Revoca l'accesso di un utente a un database.

### Utilizzo

```bash
acloud database dbaas grant delete <dbaas-id> <database-name> <grant-id> [--yes] [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS
- `database-name` (richiesto): Il nome del database
- `grant-id` (richiesto): L'ID univoco del grant da revocare

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma

### Esempio

```bash
acloud database dbaas grant delete 69455aa70d0972656501d45d "my-database" 69455aa70d0972656501d4ab --yes
```

> **Nota:** I grant vengono rimossi anche quando il database padre o l'istanza DBaaS viene eliminata.

## Flusso di Lavoro Tipico

```bash
# 1. Crea un'istanza DBaaS e attendi lo stato Active
acloud database dbaas create --name "my-dbaas" --region "ITBG-Bergamo" --zone "ITBG-1" \
  --engine-id "mysql-8.0" --flavor "DBO4A8" --storage-size 50 \
  --vpc-id "<vpc-id>" --subnet-id "<subnet-id>" --security-group-id "<sg-id>" \
  --elastic-ip-id "<eip-id>"

# 2. Crea un database
acloud database dbaas database create <dbaas-id> --name "appdb"

# 3. Crea un utente
acloud database dbaas user create <dbaas-id> --username "restapi" --password "SecurePass123!"

# 4. Concedi l'accesso all'utente
acloud database dbaas grant create <dbaas-id> "appdb" --username "restapi" --role "liteadmin"

# 5. Verifica il grant
acloud database dbaas grant list <dbaas-id> "appdb"
```

## Risorse Correlate

- [DBaaS](dbaas.md) - Gestisci istanze DBaaS
- [Database DBaaS](dbaas.database.md) - Gestisci database all'interno di istanze DBaaS
- [Utenti DBaaS](dbaas.user.md) - Gestisci utenti per istanze DBaaS
- [Backup Database](backup.md) - Crea e gestisci backup database
