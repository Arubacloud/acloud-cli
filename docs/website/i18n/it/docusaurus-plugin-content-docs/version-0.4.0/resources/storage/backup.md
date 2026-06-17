# Gestione Backup

I backup sono copie gestite di volumi di block storage usando la risorsa Aruba.Storage/backup. A differenza degli snapshot, i backup forniscono funzionalità avanzate come backup incrementali, politiche di retention e operazioni pianificate.

## Comandi Disponibili

- `acloud storage backup` - Crea un nuovo backup da un volume
- `acloud storage backup list` - Elenca tutti i backup
- `acloud storage backup get` - Ottieni i dettagli di un backup specifico
- `acloud storage backup update` - Aggiorna nome e tag del backup
- `acloud storage backup delete` - Elimina un backup

## Backup vs Snapshot

| Caratteristica | Backup | Snapshot |
|---------|--------|----------|
| Risorsa API | Aruba.Storage/backup | Aruba.Storage/snapshot |
| Tipo Backup | Full o Incremental | Solo Full |
| Politica Retention | Giorni configurabili | Gestione manuale |
| Periodo Fatturazione | Hour, Month, Year | Fatturazione predefinita |
| Caso d'Uso | Protezione dati a lungo termine | Copie point-in-time rapide |

## Crea Backup

Crea un backup da un volume di block storage con opzioni avanzate.

### Utilizzo

```bash
acloud storage backup <volume-id> --name <name> [flags]
```

### Argomenti

- `volume-id` - L'ID del volume di block storage da cui fare il backup

### Flag Richiesti

- `--name` - Nome per il backup

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--region` - Codice regione (default: "ITBG-Bergamo")
- `--type` - Tipo backup: Full o Incremental (default: "Full")
- `--retention-days` - Numero di giorni per mantenere il backup
- `--billing-period` - Periodo di fatturazione: Hour, Month, Year
- `--tags` - Tag (separati da virgola, max 20 caratteri per tag)

### Esempio

```bash
# Crea un backup full con retention di 7 giorni
acloud storage backup 69455aa70d0972656501d45d \
  --name "weekly-backup" \
  --region "ITBG-Bergamo" \
  --type "Full" \
  --retention-days 7 \
  --tags "weekly,production"

# Crea un backup incrementale
acloud storage backup 69455aa70d0972656501d45d \
  --name "daily-incremental" \
  --type "Incremental" \
  --retention-days 1
```

### Output

```
Creating storage backup with the following parameters:
  Name:           weekly-backup
  Type:           Full
  Region:         ITBG-Bergamo
  Volume ID:      69455aa70d0972656501d45d
  Volume URI:     /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
  Retention Days: 7
  Tags:           [weekly production]

Storage backup created successfully!
ID:              694594818f4a09c12b5e0c19
Name:            weekly-backup
Type:            Full
Creation Date:   19-12-2025 18:08:01
```

## Elenca Backup

Elenca tutti i backup nel tuo progetto.

### Utilizzo

```bash
acloud storage backup list [flags]
```

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud storage backup list
```

### Output

```
NAME              ID                          TYPE          STATUS
weekly-backup     694594818f4a09c12b5e0c19   Full          Active
daily-backup      694595918f4a09c12b5e0c20   Incremental   Active
monthly-archive   694596a18f4a09c12b5e0c21   Full          Active
```

## Ottieni Dettagli Backup

Ottieni informazioni dettagliate su un backup specifico.

### Utilizzo

```bash
acloud storage backup get <backup-id> [flags]
```

### Argomenti

- `backup-id` - L'ID del backup (supporta auto-completamento)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud storage backup get 694594818f4a09c12b5e0c19
```

### Output

```
Storage Backup Details:
=======================
ID:              694594818f4a09c12b5e0c19
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/backups/694594818f4a09c12b5e0c19
Name:            weekly-backup
Type:            Full
Source Volume:   /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
Retention Days:  7
Region:          ITBG-Bergamo
Status:          Active
Creation Date:   19-12-2025 18:08:01
Created By:      aru-297647
Tags:            [weekly production]
```

## Aggiorna Backup

Aggiorna il nome e/o i tag di un backup.

**Importante:** Il backup non può essere aggiornato mentre un'operazione di ripristino è in esecuzione sul volume associato. Il backup deve essere in stato "Active" (non "InCreation" o "Deleting").

### Utilizzo

```bash
acloud storage backup update <backup-id> [flags]
```

### Argomenti

- `backup-id` - L'ID del backup (supporta auto-completamento)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--name` - Nuovo nome per il backup
- `--tags` - Nuovi tag (separati da virgola)

**Nota:** Almeno uno tra `--name` o `--tags` deve essere fornito.

### Esempio

```bash
# Aggiorna solo il nome
acloud storage backup update 694594818f4a09c12b5e0c19 --name "backup-renamed"

# Aggiorna solo i tag
acloud storage backup update 694594818f4a09c12b5e0c19 --tags "prod,critical"

# Aggiorna entrambi
acloud storage backup update 694594818f4a09c12b5e0c19 \
  --name "production-backup" \
  --tags "prod,critical"
```

### Output

```
Backup updated successfully!
ID:              694594818f4a09c12b5e0c19
Name:            production-backup
Tags:            [prod critical]
```

## Elimina Backup

Elimina un backup. Questa azione non può essere annullata.

**Importante:** Non è possibile eliminare un backup se ci sono operazioni di ripristino attive.

### Utilizzo

```bash
acloud storage backup delete <backup-id> [flags]
```

### Argomenti

- `backup-id` - L'ID del backup (supporta auto-completamento)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `-y, --yes` - Salta il prompt di conferma

### Esempio

```bash
# Con prompt di conferma
acloud storage backup delete 694594818f4a09c12b5e0c19

# Salta conferma
acloud storage backup delete 694594818f4a09c12b5e0c19 --yes
```

### Output

```
Backup 694594818f4a09c12b5e0c19 deleted successfully!
```

## Tipi di Backup

### Backup Full
- Copia completa del volume
- Può essere usato indipendentemente per il ripristino
- Richiede più tempo e spazio di storage
- Consigliato per: Archivi settimanali/mensili, protezione dati critici

### Backup Incrementale
- Memorizza solo le modifiche dall'ultimo backup
- Più veloce e usa meno storage
- Richiede backup precedenti per il ripristino
- Consigliato per: Backup giornalieri, snapshot frequenti

## Politica di Retention

- I backup possono avere un periodo di retention in giorni
- Dopo la scadenza del periodo di retention, i backup possono essere eliminati automaticamente
- Imposta la retention in base ai tuoi requisiti di conformità e recupero
- Periodi di retention comuni: 7 giorni (settimanale), 30 giorni (mensile), 365 giorni (annuale)

## Best Practices

1. **Usa backup Full per baseline**: Crea backup full settimanali come punti di recupero
2. **Incrementali per giornalieri**: Usa backup incrementali per protezione giornaliera
3. **Tag appropriati**: Usa tag per identificare scopo e pianificazione del backup
4. **Monitora retention**: Traccia l'età dei backup e regola la retention secondo necessità
5. **Testa ripristini**: Verifica periodicamente che i backup possano essere ripristinati con successo
6. **Elimina backup vecchi**: Pulisci i backup dopo il periodo di retention

## Note

- I backup vengono creati in modo asincrono e iniziano in stato "InCreation"
- Una volta "Active", i backup possono essere usati per operazioni di ripristino
- I tag devono essere massimo 20 caratteri ciascuno
- L'ID volume può essere trovato usando `acloud storage blockstorage list`
- I backup usano la risorsa API Aruba.Storage/backup (diversa dagli snapshot)

