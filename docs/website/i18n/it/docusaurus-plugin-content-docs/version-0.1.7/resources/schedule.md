# Risorse di Pianificazione

La categoria `schedule` fornisce comandi per gestire job pianificati in Aruba Cloud. I job pianificati ti permettono di automatizzare attività che vengono eseguite a orari specifici o su pianificazioni ricorrenti.

## Risorse Disponibili

### [Job](schedule/job.md)

Job pianificati che eseguono azioni a orari specificati o su pianificazioni ricorrenti usando espressioni CRON.

**Comandi Rapidi:**
```bash
# Elenca tutti i job pianificati
acloud schedule job list

# Ottieni i dettagli del job
acloud schedule job get <job-id>

# Crea un job OneShot (esegue una volta a un orario specifico)
acloud schedule job create --name "my-oneshot-job" --region "ITBG-Bergamo" --job-type "OneShot" --schedule-at "2024-12-31T23:59:59Z"

# Crea un job Ricorrente (esegue su una pianificazione)
acloud schedule job create --name "my-recurring-job" --region "ITBG-Bergamo" --job-type "Recurring" --cron "0 0 * * *" --execute-until "2025-12-31T23:59:59Z"

# Aggiorna un job
acloud schedule job update <job-id> --name "updated-name" --enabled false

# Elimina un job
acloud schedule job delete <job-id>
```

## Struttura dei Comandi

Tutti i comandi di pianificazione seguono questa struttura:

```
acloud schedule <resource> <action> [arguments] [flags]
```

Dove:
- `<resource>`: Il tipo di risorsa (es. `job`)
- `<action>`: L'operazione da eseguire (es. `list`, `get`, `create`, `update`, `delete`)
- `[arguments]`: Argomenti richiesti (es. ID risorse)
- `[flags]`: Flag opzionali (es. `--name`, `--job-type`, `--cron`)

## Tipi di Job

### Job OneShot

I job OneShot vengono eseguiti una volta a una data e ora specificate. Sono utili per:
- Attività di manutenzione una tantum
- Deployment pianificati
- Automazione basata sul tempo

**Flag richiesti:**
- `--schedule-at`: Data e ora in cui il job dovrebbe essere eseguito (formato ISO 8601)

### Job Ricorrenti

I job ricorrenti vengono eseguiti su una pianificazione definita da un'espressione CRON. Sono utili per:
- Backup giornalieri
- Report settimanali
- Manutenzione periodica

**Flag richiesti:**
- `--cron`: Espressione CRON che definisce la pianificazione
- `--execute-until`: Data di fine fino alla quale il job può essere eseguito

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

**Esempi:**
- `0 0 * * *` - Ogni giorno a mezzanotte
- `0 */6 * * *` - Ogni 6 ore
- `0 0 1 * *` - Primo giorno di ogni mese a mezzanotte
- `0 0 * * 0` - Ogni domenica a mezzanotte

## Pattern Comuni

### Elencare i Job

```bash
acloud schedule job list
```

### Ottenere i Dettagli del Job

```bash
acloud schedule job get <job-id>
```

### Creare un Job OneShot

```bash
acloud schedule job create \
  --name "backup-job" \
  --region "ITBG-Bergamo" \
  --job-type "OneShot" \
  --schedule-at "2024-12-31T23:59:59Z" \
  --enabled true \
  --tags "backup,automation"
```

### Creare un Job Ricorrente

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

### Aggiornare un Job

```bash
acloud schedule job update <job-id> \
  --name "updated-name" \
  --enabled false \
  --tags "updated,disabled"
```

### Eliminare un Job

```bash
acloud schedule job delete <job-id> [--yes]
```

## Contesto Progetto

I job pianificati sono limitati a un progetto. Puoi:

1. **Usare il flag `--project-id`:**
   ```bash
   acloud schedule job list --project-id <project-id>
   ```

2. **Impostare un contesto:**
   ```bash
   acloud context set my-prod --project-id <project-id>
   acloud schedule job list  # Usa l'ID progetto del contesto
   ```

Vedi [Installazione - Gestione Contesto](../installation.md#context-management) per maggiori informazioni.

## Prossimi Passi

- Esplora le [Risorse di Gestione](./management.md) per risorse a livello organizzativo
- Controlla le [Risorse Storage](./storage.md) per operazioni di storage
- Rivedi le [Risorse di Rete](./network.md) per capacità di networking
- Vedi le [Risorse Database](./database.md) per la gestione database

