# Gestione Backup Database

I backup database forniscono copie point-in-time dei tuoi database per disaster recovery e ripristino.

## Comandi Disponibili

- `acloud database backup create` - Crea un nuovo backup database
- `acloud database backup list` - Elenca tutti i backup database
- `acloud database backup get` - Ottieni i dettagli di un backup specifico
- `acloud database backup delete` - Elimina un backup database

**Nota:** I backup database non supportano operazioni di aggiornamento.

## Crea Backup

Crea un nuovo backup di un database in un'istanza DBaaS.

### Utilizzo

```bash
acloud database backup create --name <name> --region <region> --dbaas-id <dbaas-id> --database-name <database-name> [flags]
```

### Flag Richiesti

- `--name` - Nome per il backup
- `--region` - Codice regione (es. "ITBG-Bergamo")
- `--dbaas-id` - ID istanza DBaaS
- `--database-name` - Nome del database di cui fare il backup

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--billing-period` - Periodo di fatturazione: Hour, Month, Year (default: "Hour")
- `--tags` - Tag (separati da virgola)

### Esempio

```bash
acloud database backup create \
  --name "daily-backup" \
  --region "ITBG-Bergamo" \
  --dbaas-id "69455aa70d0972656501d45d" \
  --database-name "my-database" \
  --billing-period "Hour" \
  --tags "daily,automated"
```

## Elenca Backup

Elenca tutti i backup database nel tuo progetto.

### Utilizzo

```bash
acloud database backup list [flags]
```

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database backup list
```

## Ottieni Dettagli Backup

Recupera informazioni dettagliate su un backup specifico.

### Utilizzo

```bash
acloud database backup get <backup-id> [flags]
```

### Argomenti

- `backup-id` (richiesto): L'ID univoco del backup

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud database backup get 69455aa70d0972656501d45d
```

## Elimina Backup

Elimina un backup database.

### Utilizzo

```bash
acloud database backup delete <backup-id> [--yes] [flags]
```

### Argomenti

- `backup-id` (richiesto): L'ID univoco del backup

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma

### Esempio

```bash
acloud database backup delete 69455aa70d0972656501d45d --yes
```

## Best Practices per Backup

- Crea backup regolari di database critici
- Memorizza backup in regioni diverse per disaster recovery
- Testa le procedure di ripristino regolarmente
- Mantieni più versioni di backup
- Automatizza la creazione di backup usando job pianificati

## Risorse Correlate

- [DBaaS](dbaas.md) - Gestisci istanze DBaaS
- [Database DBaaS](dbaas.database.md) - Gestisci database
- [Job Pianificati](../schedule/job.md) - Automatizza la creazione di backup
