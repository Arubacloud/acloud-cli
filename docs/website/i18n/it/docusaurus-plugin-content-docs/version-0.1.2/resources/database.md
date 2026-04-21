# Risorse Database

La categoria `database` fornisce comandi per gestire risorse database in Aruba Cloud, inclusi istanze DBaaS (Database as a Service), database, utenti e backup.

## Risorse Disponibili

### [DBaaS](database/dbaas.md)

Le istanze DBaaS (Database as a Service) forniscono servizi database gestiti in Aruba Cloud.

**Comandi Rapidi:**
```bash
# Elenca tutte le istanze DBaaS
acloud database dbaas list

# Ottieni i dettagli dell'istanza DBaaS
acloud database dbaas get <dbaas-id>

# Crea un'istanza DBaaS
acloud database dbaas create --name "my-dbaas" --region "ITBG-Bergamo" --engine-id <engine-id> --flavor <flavor>

# Aggiorna un'istanza DBaaS
acloud database dbaas update <dbaas-id> --tags "production,updated"

# Elimina un'istanza DBaaS
acloud database dbaas delete <dbaas-id>
```

### [Database DBaaS](database/dbaas.database.md)

Database all'interno di istanze DBaaS che memorizzano i tuoi dati.

**Comandi Rapidi:**
```bash
# Elenca tutti i database in un'istanza DBaaS
acloud database dbaas database list <dbaas-id>

# Ottieni i dettagli del database
acloud database dbaas database get <dbaas-id> <database-name>

# Crea un database
acloud database dbaas database create <dbaas-id> --name "my-database"

# Aggiorna un database (rinomina)
acloud database dbaas database update <dbaas-id> <database-name> --name "new-name"

# Elimina un database
acloud database dbaas database delete <dbaas-id> <database-name>
```

### [Utenti DBaaS](database/dbaas.user.md)

Utenti che possono accedere ai database all'interno di istanze DBaaS.

**Comandi Rapidi:**
```bash
# Elenca tutti gli utenti in un'istanza DBaaS
acloud database dbaas user list <dbaas-id>

# Ottieni i dettagli dell'utente
acloud database dbaas user get <dbaas-id> <username>

# Crea un utente
acloud database dbaas user create <dbaas-id> --username "my-user" --password "secure-password"

# Aggiorna un utente (cambia password)
acloud database dbaas user update <dbaas-id> <username> --password "new-password"

# Elimina un utente
acloud database dbaas user delete <dbaas-id> <username>
```

### [Backup Database](database/backup.md)

Backup di database per disaster recovery e ripristino point-in-time.

**Comandi Rapidi:**
```bash
# Elenca tutti i backup database
acloud database backup list

# Ottieni i dettagli del backup
acloud database backup get <backup-id>

# Crea un backup
acloud database backup create --name "my-backup" --region "ITBG-Bergamo" --dbaas-id <dbaas-id> --database-name <database-name>

# Elimina un backup
acloud database backup delete <backup-id>
```

## Struttura dei Comandi

Tutti i comandi database seguono questa struttura:

```
acloud database <resource> <action> [arguments] [flags]
```

Dove:
- `<resource>`: Il tipo di risorsa (es. `dbaas`, `dbaas database`, `dbaas user`, `backup`)
- `<action>`: L'operazione da eseguire (es. `list`, `get`, `create`, `update`, `delete`)
- `[arguments]`: Argomenti richiesti (es. ID risorse, nomi)
- `[flags]`: Flag opzionali (es. `--name`, `--region`, `--tags`)

## Pattern Comuni

### Elencare le Risorse

```bash
acloud database <resource> list
```

### Ottenere i Dettagli delle Risorse

```bash
acloud database <resource> get <resource-id>
```

### Creare Risorse

```bash
acloud database <resource> create [required-args] [flags]
```

### Aggiornare Risorse

```bash
acloud database <resource> update <resource-id> [flags]
```

### Eliminare Risorse

```bash
acloud database <resource> delete <resource-id> [--yes]
```

## Contesto Progetto

Le risorse database sono limitate a un progetto. Puoi:

1. **Usare il flag `--project-id`:**
   ```bash
   acloud database dbaas list --project-id <project-id>
   ```

2. **Impostare un contesto:**
   ```bash
   acloud context set my-prod --project-id <project-id>
   acloud database dbaas list  # Usa l'ID progetto del contesto
   ```

Vedi [Installazione - Gestione Contesto](../installation.md#context-management) per maggiori informazioni.

## Prossimi Passi

- Esplora le [Risorse di Gestione](./management.md) per risorse a livello organizzativo
- Controlla le [Risorse Storage](./storage.md) per operazioni di storage
- Rivedi le [Risorse di Rete](./network.md) per capacità di networking

