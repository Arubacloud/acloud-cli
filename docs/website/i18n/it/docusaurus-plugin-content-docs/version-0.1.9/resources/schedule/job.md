# Gestione Job Pianificati

I job pianificati ti permettono di automatizzare attività che vengono eseguite a orari specifici o su pianificazioni ricorrenti usando espressioni CRON.

## Comandi Disponibili

- `acloud schedule job create` - Crea un nuovo job pianificato
- `acloud schedule job list` - Elenca tutti i job pianificati
- `acloud schedule job get` - Ottieni i dettagli di un job specifico
- `acloud schedule job update` - Aggiorna proprietà job
- `acloud schedule job delete` - Elimina un job pianificato

## Tipi di Job

### Job OneShot

I job OneShot vengono eseguiti una volta a una data e ora specificate. Sono utili per:
- Attività di manutenzione una tantum
- Deployment pianificati
- Automazione basata sul tempo

### Job Ricorrenti

I job ricorrenti vengono eseguiti su una pianificazione definita da un'espressione CRON. Sono utili per:
- Backup giornalieri
- Report settimanali
- Manutenzione periodica

## Crea Job OneShot

Crea un job che viene eseguito una volta a un orario specifico.

### Utilizzo

```bash
acloud schedule job create --name <name> --region <region> --job-type "OneShot" --schedule-at <datetime> [flags]
```

### Flag Richiesti

- `--name` - Nome per il job
- `--region` - Codice regione (es. "ITBG-Bergamo")
- `--job-type` - Deve essere "OneShot"
- `--schedule-at` - Data e ora in cui il job dovrebbe essere eseguito (formato ISO 8601, es. "2024-12-31T23:59:59Z")

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--enabled` - Abilita il job (default: true)
- `--tags` - Tag (separati da virgola)

### Esempio

```bash
acloud schedule job create \
  --name "deploy-release" \
  --region "ITBG-Bergamo" \
  --job-type "OneShot" \
  --schedule-at "2024-12-31T23:59:59Z" \
  --enabled true \
  --tags "deployment,production"
```

## Crea Job Ricorrente

Crea un job che viene eseguito su una pianificazione ricorrente.

### Utilizzo

```bash
acloud schedule job create --name <name> --region <region> --job-type "Recurring" --cron <cron-expression> --execute-until <datetime> [flags]
```

### Flag Richiesti

- `--name` - Nome per il job
- `--region` - Codice regione (es. "ITBG-Bergamo")
- `--job-type` - Deve essere "Recurring"
- `--cron` - Espressione CRON che definisce la pianificazione
- `--execute-until` - Data di fine fino alla quale il job può essere eseguito (formato ISO 8601)

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--enabled` - Abilita il job (default: true)
- `--tags` - Tag (separati da virgola)

### Esempio

```bash
acloud schedule job create \
  --name "daily-backup" \
  --region "ITBG-Bergamo" \
  --job-type "Recurring" \
  --cron "0 2 * * *" \
  --execute-until "2025-12-31T23:59:59Z" \
  --enabled true \
  --tags "backup,daily"
```

## Formato Espressione CRON

Le espressioni CRON seguono il formato standard:
```
┌───────────── minuto (0 - 59)
│ ┌───────────── ora (0 - 23)
│ │ ┌───────────── giorno del mese (1 - 31)
│ │ │ ┌───────────── mese (1 - 12)
│ │ │ │ ┌───────────── giorno della settimana (0 - 6) (Domenica a Sabato)
│ │ │ │ │
* * * * *
```

### Esempi CRON Comuni

- `0 0 * * *` - Ogni giorno a mezzanotte
- `0 */6 * * *` - Ogni 6 ore
- `0 0 1 * *` - Primo giorno di ogni mese a mezzanotte
- `0 0 * * 0` - Ogni domenica a mezzanotte
- `0 0 1 1 *` - 1 gennaio a mezzanotte (annuale)
- `*/15 * * * *` - Ogni 15 minuti

## Elenca Job

Elenca tutti i job pianificati nel tuo progetto.

### Utilizzo

```bash
acloud schedule job list [flags]
```

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud schedule job list
```

## Ottieni Dettagli Job

Recupera informazioni dettagliate su un job specifico.

### Utilizzo

```bash
acloud schedule job get <job-id> [flags]
```

### Argomenti

- `job-id` (richiesto): L'ID univoco del job

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud schedule job get 69455aa70d0972656501d45d
```

## Aggiorna Job

Aggiorna proprietà job come nome, stato abilitato e tag.

### Utilizzo

```bash
acloud schedule job update <job-id> [flags]
```

### Argomenti

- `job-id` (richiesto): L'ID univoco del job

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--name` - Nuovo nome per il job
- `--enabled` - Abilita o disabilita il job
- `--tags` - Nuovi tag (separati da virgola)

### Esempio

```bash
acloud schedule job update 69455aa70d0972656501d45d \
  --name "updated-job-name" \
  --enabled false \
  --tags "updated,disabled"
```

## Elimina Job

Elimina un job pianificato.

### Utilizzo

```bash
acloud schedule job delete <job-id> [--yes] [flags]
```

### Argomenti

- `job-id` (richiesto): L'ID univoco del job

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--yes, -y` - Salta il prompt di conferma

### Esempio

```bash
acloud schedule job delete 69455aa70d0972656501d45d --yes
```

## Best Practices

- Usa nomi descrittivi per i job
- Imposta date `execute-until` appropriate per job ricorrenti
- Testa espressioni CRON prima di creare job
- Monitora lo stato di esecuzione dei job
- Usa tag per organizzare e filtrare job
- Disabilita job invece di eliminarli quando temporaneamente non necessari

## Risorse Correlate

- [Backup Database](../database/backup.md) - Automatizza la creazione di backup
- [Backup Storage](../storage/backup.md) - Pianifica backup storage

