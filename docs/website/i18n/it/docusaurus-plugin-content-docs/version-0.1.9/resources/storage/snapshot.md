# Gestione Snapshot

Gli snapshot sono copie point-in-time di volumi di block storage. Catturano l'intero stato di un volume in un momento specifico.

## Comandi Disponibili

- `acloud storage snapshot create` - Crea un nuovo snapshot da un volume
- `acloud storage snapshot list` - Elenca snapshot per un volume
- `acloud storage snapshot get` - Ottieni i dettagli di uno snapshot specifico
- `acloud storage snapshot update` - Aggiorna nome e tag dello snapshot
- `acloud storage snapshot delete` - Elimina uno snapshot

## Crea Snapshot

Crea uno snapshot da un volume di block storage esistente.

### Utilizzo

```bash
acloud storage snapshot create --name <name> --region <region> --volume-uri <volume-uri> [flags]
```

### Flag Richiesti

- `--name` - Nome per lo snapshot
- `--region` - Codice regione
- `--volume-uri` - URI del volume sorgente

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--tags` - Tag (separati da virgola)

### Esempio

```bash
acloud storage snapshot create \
  --name "backup-before-upgrade" \
  --region "ITBG-Bergamo" \
  --volume-uri "/projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d" \
  --tags "backup,pre-upgrade"
```

### Output

```
Creating snapshot with the following parameters:
  Name:       backup-before-upgrade
  Region:     ITBG-Bergamo
  Volume URI: /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
  Tags:       [backup pre-upgrade]

Snapshot created successfully!
ID:              6944fd760d0972656501d431
Name:            backup-before-upgrade
Creation Date:   18-12-2025 17:23:50
```

## Elenca Snapshot

Elenca tutti gli snapshot per un volume di block storage specifico.

### Utilizzo

```bash
acloud storage snapshot list --volume-uri <volume-uri> [flags]
```

### Flag Richiesti

- `--volume-uri` - URI del volume di block storage

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `-v, --verbose` - Mostra informazioni di debug dettagliate

### Esempio

```bash
acloud storage snapshot list --volume-uri "/projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d"
```

### Output

```
NAME                      ID                          SIZE(GB)  STATUS
backup-before-upgrade     6944fd760d0972656501d431   50        Available
daily-backup-20251218     6944fe870d0972656501d432   50        Available
```

## Ottieni Dettagli Snapshot

Ottieni informazioni dettagliate su uno snapshot specifico.

### Utilizzo

```bash
acloud storage snapshot get <snapshot-id> [flags]
```

### Argomenti

- `snapshot-id` - L'ID dello snapshot (supporta auto-completamento)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud storage snapshot get 6944fd760d0972656501d431
```

### Output

```
Snapshot Details:
=================
ID:              6944fd760d0972656501d431
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/snapshots/6944fd760d0972656501d431
Name:            backup-before-upgrade
Size (GB):       50
Source Volume:   /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
Region:          ITBG-Bergamo
Status:          Available
Creation Date:   18-12-2025 17:23:50
Created By:      aru-297647
Tags:            [backup pre-upgrade]
```

## Aggiorna Snapshot

Aggiorna il nome e/o i tag di uno snapshot.

### Utilizzo

```bash
acloud storage snapshot update <snapshot-id> [flags]
```

### Argomenti

- `snapshot-id` - L'ID dello snapshot (supporta auto-completamento)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--name` - Nuovo nome per lo snapshot
- `--tags` - Nuovi tag (separati da virgola)

**Nota:** Almeno uno tra `--name` o `--tags` deve essere fornito.

### Esempio

```bash
# Aggiorna solo il nome
acloud storage snapshot update 6944fd760d0972656501d431 --name "pre-upgrade-snapshot"

# Aggiorna solo i tag
acloud storage snapshot update 6944fd760d0972656501d431 --tags "backup,important"

# Aggiorna entrambi
acloud storage snapshot update 6944fd760d0972656501d431 \
  --name "pre-upgrade-snapshot" \
  --tags "backup,important"
```

### Output

```
Snapshot updated successfully!
ID:              6944fd760d0972656501d431
Name:            pre-upgrade-snapshot
Tags:            [backup important]
```

## Elimina Snapshot

Elimina uno snapshot. Questa azione non può essere annullata.

### Utilizzo

```bash
acloud storage snapshot delete <snapshot-id> [flags]
```

### Argomenti

- `snapshot-id` - L'ID dello snapshot (supporta auto-completamento)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `-y, --yes` - Salta il prompt di conferma

### Esempio

```bash
# Con prompt di conferma
acloud storage snapshot delete 6944fd760d0972656501d431

# Salta conferma
acloud storage snapshot delete 6944fd760d0972656501d431 --yes
```

### Output

```
Snapshot 6944fd760d0972656501d431 deleted successfully!
```

## Note

- Gli snapshot sono copie point-in-time e non includono modifiche fatte dopo la creazione
- Gli snapshot sono memorizzati nella stessa regione del volume sorgente
- Puoi creare più snapshot dallo stesso volume
- Gli snapshot possono essere usati per ripristinare volumi a uno stato precedente
- L'URI del volume può essere trovato usando `acloud storage blockstorage get <volume-id>`

