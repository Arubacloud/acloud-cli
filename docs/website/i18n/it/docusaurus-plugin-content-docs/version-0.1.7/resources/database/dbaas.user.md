# Gestione Utenti DBaaS

Utenti che possono accedere ai database all'interno di istanze DBaaS. Ogni utente ha un nome utente e una password per l'autenticazione.

## Comandi Disponibili

- `acloud database dbaas user create` - Crea un nuovo utente in un'istanza DBaaS
- `acloud database dbaas user list` - Elenca tutti gli utenti in un'istanza DBaaS
- `acloud database dbaas user get` - Ottieni i dettagli di un utente specifico
- `acloud database dbaas user update` - Aggiorna password utente
- `acloud database dbaas user delete` - Elimina un utente

## Crea Utente

Crea un nuovo utente all'interno di un'istanza DBaaS.

### Utilizzo

```bash
acloud database dbaas user create <dbaas-id> --username <username> --password <password> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS

### Flag Richiesti

- `--username` - Nome utente per l'utente
- `--password` - Password per l'utente

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas user create 69455aa70d0972656501d45d \
  --username "app-user" \
  --password "SecurePassword123!"
```

## Elenca Utenti

Elenca tutti gli utenti in un'istanza DBaaS.

### Utilizzo

```bash
acloud database dbaas user list <dbaas-id> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas user list 69455aa70d0972656501d45d
```

## Ottieni Dettagli Utente

Recupera informazioni dettagliate su un utente specifico.

### Utilizzo

```bash
acloud database dbaas user get <dbaas-id> <username> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS
- `username` (richiesto): Il nome utente

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas user get 69455aa70d0972656501d45d "app-user"
```

## Aggiorna Utente

Cambia la password di un utente.

### Utilizzo

```bash
acloud database dbaas user update <dbaas-id> <username> --password <new-password> [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS
- `username` (richiesto): Il nome utente

### Flag Richiesti

- `--password` - Nuova password per l'utente

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database dbaas user update 69455aa70d0972656501d45d "app-user" \
  --password "NewSecurePassword456!"
```

## Elimina Utente

Elimina un utente da un'istanza DBaaS.

### Utilizzo

```bash
acloud database dbaas user delete <dbaas-id> <username> [--yes] [flags]
```

### Argomenti

- `dbaas-id` (richiesto): L'ID univoco dell'istanza DBaaS
- `username` (richiesto): Il nome utente da eliminare

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma

### Esempio

```bash
acloud database dbaas user delete 69455aa70d0972656501d45d "app-user" --yes
```

## Best Practices di Sicurezza

- Usa password forti (minimo 12 caratteri, mix di lettere, numeri e simboli)
- Ruota le password regolarmente
- Usa password diverse per utenti diversi
- Non condividere mai password o committarle nel controllo versione
- Considera l'uso di un password manager

## Risorse Correlate

- [DBaaS](dbaas.md) - Gestisci istanze DBaaS
- [Database DBaaS](dbaas.database.md) - Gestisci database
- [Backup Database](backup.md) - Crea backup di database
