# Gestione Block Storage

I volumi di block storage sono dispositivi di storage persistenti che possono essere collegati a macchine virtuali in Aruba Cloud.

## Comandi Disponibili

- `acloud storage blockstorage create` - Crea un nuovo volume di block storage
- `acloud storage blockstorage list` - Elenca tutti i volumi di block storage
- `acloud storage blockstorage get` - Ottieni i dettagli di un volume specifico
- `acloud storage blockstorage update` - Aggiorna nome e tag del volume
- `acloud storage blockstorage delete` - Elimina un volume di block storage

## Crea Block Storage

Crea un nuovo volume di block storage nel tuo progetto.

### Utilizzo

```bash
acloud storage blockstorage create --name <name> --region <region> --size <size-gb> [flags]
```

### Flag Richiesti

- `--name` - Nome per il volume di block storage
- `--size` - Dimensione in GB (deve essere maggiore di 0)

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--region` - Codice regione (default: "ITBG-Bergamo")
- `--zone` - Zona/datacenter (opzionale)
- `--type` - Tipo volume: Standard o Performance (default: "Standard")
- `--billing-period` - Periodo di fatturazione: Hour, Month, Year (default: "Hour")
- `--tags` - Tag (separati da virgola)

### Esempio

```bash
# Crea un block storage standard da 50GB
acloud storage blockstorage create \
  --name "my-data-volume" \
  --region "ITBG-Bergamo" \
  --size 50 \
  --type "Standard" \
  --billing-period "Hour" \
  --tags "env,production"
```

### Output

```
Creating block storage with:
  Name: my-data-volume
  Region: ITBG-Bergamo
  Size: 50 GB
  Type: Standard
  Billing Period: Hour
  Project ID: 68398923fb2cb026400d4d31

Block storage created successfully!
ID:              69455aa70d0972656501d45d
Name:            my-data-volume
Size (GB):       50
Type:            Standard
Zone:            DC-BG-IT-1
Region:          ITBG-Bergamo
Status:          NotUsed
Creation Date:   18-12-2025 18:49:06
```

## Elenca Block Storage

Elenca tutti i volumi di block storage nel tuo progetto.

### Utilizzo

```bash
acloud storage blockstorage list [flags]
```

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `-v, --verbose` - Mostra informazioni di debug dettagliate

### Esempio

```bash
acloud storage blockstorage list
```

### Output

```
NAME              ID                          SIZE(GB)  REGION  ZONE         TYPE       STATUS
my-data-volume    69455aa70d0972656501d45d   50        ITBG-Bergamo   DC-BG-IT-1   Standard   NotUsed
app-volume        69455bb80d0972656501d45e   100       ITBG-Bergamo   DC-BG-IT-1   Standard   Used
```

## Ottieni Dettagli Block Storage

Ottieni informazioni dettagliate su un volume di block storage specifico.

### Utilizzo

```bash
acloud storage blockstorage get <volume-id> [flags]
```

### Argomenti

- `volume-id` - L'ID del volume di block storage

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Auto-completamento

Gli ID volume supportano l'auto-completamento. Premi TAB dopo aver digitato il comando per vedere i volumi disponibili:

```bash
acloud storage blockstorage get <TAB>
# Mostra:
# 6965a6c3ffc0fd1ef8ba5612    MyVolume
# 6965a6c3ffc0fd1ef8ba5613    DataVolume
```

### Esempio

```bash
acloud storage blockstorage get 69455aa70d0972656501d45d
```

### Output

```
Block Storage Details:
======================
ID:              69455aa70d0972656501d45d
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
Name:            my-data-volume
Size (GB):       50
Type:            Standard
Zone:            DC-BG-IT-1
Region:          ITBG-Bergamo
Bootable:        false
Status:          NotUsed
Creation Date:   18-12-2025 18:49:06
Created By:      aru-297647
Tags:            [env production]
```

## Aggiorna Block Storage

Aggiorna il nome e/o i tag di un volume di block storage.

**Nota:** Gli aggiornamenti della dimensione non sono attualmente supportati dall'API. Il volume deve essere in stato "Used" o "NotUsed" per essere aggiornato.

### Utilizzo

```bash
acloud storage blockstorage update <volume-id> [flags]
```

### Argomenti

- `volume-id` - L'ID del volume di block storage (supporta auto-completamento)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--name` - Nuovo nome per il volume
- `--tags` - Nuovi tag (separati da virgola)

**Nota:** Almeno uno tra `--name` o `--tags` deve essere fornito.

### Esempio

```bash
# Aggiorna solo il nome
acloud storage blockstorage update 69455aa70d0972656501d45d --name "prod-data-volume"

# Aggiorna solo i tag
acloud storage blockstorage update 69455aa70d0972656501d45d --tags "production,critical"

# Aggiorna entrambi
acloud storage blockstorage update 69455aa70d0972656501d45d \
  --name "prod-data-volume" \
  --tags "production,critical"
```

### Output

```
Block storage updated successfully!
ID:              69455aa70d0972656501d45d
Name:            prod-data-volume
Tags:            [production critical]
Size (GB):       50
Type:            Standard
```

## Elimina Block Storage

Elimina un volume di block storage. Questa azione non può essere annullata.

### Utilizzo

```bash
acloud storage blockstorage delete <volume-id> [flags]
```

### Argomenti

- `volume-id` - L'ID del volume di block storage (supporta auto-completamento)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `-y, --yes` - Salta il prompt di conferma

### Esempio

```bash
# Con prompt di conferma
acloud storage blockstorage delete 69455aa70d0972656501d45d

# Salta conferma
acloud storage blockstorage delete 69455aa70d0972656501d45d --yes
```

### Output

```
Block storage 69455aa70d0972656501d45d deleted successfully!
```

## Note

- I volumi di block storage possono essere creati con tipi diversi (Standard o Performance)
- I volumi devono essere scollegati dalle VM prima dell'eliminazione
- La zona viene assegnata automaticamente se non specificata
- I periodi di fatturazione possono essere Hour, Month o Year
- I tag sono utili per organizzare e filtrare le risorse

