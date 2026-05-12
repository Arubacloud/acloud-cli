# Risorse Storage

La categoria `storage` fornisce comandi per gestire risorse di storage in Aruba Cloud.

## Risorse Disponibili

### [Block Storage](storage/blockstorage.md)

I volumi di block storage sono dispositivi di storage persistenti che possono essere collegati a macchine virtuali.

**Comandi Rapidi:**
```bash
# Elenca tutti i volumi di block storage
acloud storage blockstorage list

# Ottieni i dettagli del volume
acloud storage blockstorage get <volume-id>

# Crea un volume
acloud storage blockstorage create --name "my-volume" --size 50

# Aggiorna un volume
acloud storage blockstorage update <volume-id> --name "new-name"

# Elimina un volume
acloud storage blockstorage delete <volume-id>
```

### [Snapshot](storage/snapshot.md)

Gli snapshot sono copie point-in-time di volumi di block storage per backup rapidi e clonazione.

**Comandi Rapidi:**
```bash
# Elenca snapshot per un volume
acloud storage snapshot list --volume-uri <volume-uri>

# Ottieni i dettagli dello snapshot
acloud storage snapshot get <snapshot-id>

# Crea uno snapshot
acloud storage snapshot create --name "backup" --region "ITBG-Bergamo" --volume-uri <uri>

# Aggiorna uno snapshot
acloud storage snapshot update <snapshot-id> --tags "important"

# Elimina uno snapshot
acloud storage snapshot delete <snapshot-id>
```

### [Backup](storage/backup.md)

I backup forniscono protezione avanzata dei dati con tipi di backup full/incrementale e politiche di retention.

**Comandi Rapidi:**
```bash
# Elenca tutti i backup
acloud storage backup list

# Ottieni i dettagli del backup
acloud storage backup get <backup-id>

# Crea un backup
acloud storage backup <volume-id> --name "weekly-backup" --type "Full" --retention-days 7

# Aggiorna un backup
acloud storage backup update <backup-id> --tags "production"

# Elimina un backup
acloud storage backup delete <backup-id>
```

### [Operazioni di Ripristino](storage/restore.md)

Le operazioni di ripristino ti permettono di ripristinare volumi di block storage da backup.

**Comandi Rapidi:**
```bash
# Elenca operazioni di ripristino per un backup
acloud storage restore list <backup-id>

# Ottieni i dettagli del ripristino
acloud storage restore get <backup-id> <restore-id>

# Crea un'operazione di ripristino
acloud storage restore <backup-id> <volume-id> --name "restore-op" --region "ITBG-Bergamo"

# Aggiorna un'operazione di ripristino
acloud storage restore update <backup-id> <restore-id> --name "new-name"

# Elimina un'operazione di ripristino
acloud storage restore delete <backup-id> <restore-id>
```

## Struttura dei Comandi

Tutti i comandi di storage seguono questa struttura:

```
acloud storage <resource> <action> [arguments] [flags]
```

Dove:
- `<resource>`: Il tipo di risorsa (es. `blockstorage`, `snapshot`, `backup`, `restore`)
- `<action>`: L'operazione da eseguire (es. `list`, `get`, `create`, `update`, `delete`)
- `[arguments]`: Argomenti richiesti (es. ID risorse)
- `[flags]`: Flag opzionali (es. `--name`, `--size`, `--type`)

## Pattern Comuni

### Elencare le Risorse

```bash
acloud storage <resource> list [arguments]
```

Elenca tutte le risorse del tipo specificato con informazioni chiave visualizzate in formato tabella.

### Ottenere i Dettagli delle Risorse

```bash
acloud storage <resource> get <resource-id> [arguments]
```

Visualizza informazioni dettagliate su una risorsa specifica.

### Creare Risorse

```bash
acloud storage <resource> <arguments> --flag1 value1 --flag2 value2
```

Crea una nuova risorsa con le proprietà specificate.

### Aggiornare Risorse

```bash
acloud storage <resource> update <resource-id> [arguments] --flag1 value1
```

Aggiorna una risorsa esistente. Solo i campi forniti vengono modificati.

### Eliminare Risorse

```bash
acloud storage <resource> delete <resource-id> [arguments]
```

Elimina la risorsa specificata. Potrebbe richiedere conferma.

## Architettura Storage

Le risorse di storage sono organizzate gerarchicamente:

```
Project
├── Block Storage Volumes
│   ├── Snapshots (copie point-in-time)
│   └── Backups (con politiche di retention)
│       └── Restore Operations (annidati sotto i backup)
```

## Relazioni tra Risorse

- **Block Storage** → **Snapshots**: Uno-a-molti (un volume può avere più snapshot)
- **Block Storage** → **Backups**: Uno-a-molti (un volume può avere più backup)
- **Backups** → **Restore Operations**: Uno-a-molti (un backup può avere più operazioni di ripristino)

## Prossimi Passi

- [Guida alla Gestione Block Storage](storage/blockstorage.md)
- [Guida alla Gestione Snapshot](storage/snapshot.md)
- [Guida alla Gestione Backup](storage/backup.md)
- [Guida alle Operazioni di Ripristino](storage/restore.md)

