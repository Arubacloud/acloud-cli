# Risorse di Sicurezza

La categoria `security` fornisce comandi per gestire risorse di sicurezza in Aruba Cloud, inclusi chiavi Key Management System (KMS) per crittografia e sicurezza.

## Risorse Disponibili

### [Chiavi KMS](security/kms.md)

Le chiavi KMS (Key Management Service) forniscono gestione delle chiavi di crittografia per proteggere i tuoi dati e risorse.

**Comandi Rapidi:**
```bash
# Elenca tutte le chiavi KMS
acloud security kms list

# Ottieni i dettagli della chiave KMS
acloud security kms get <kms-id>

# Crea una chiave KMS
acloud security kms create --name "my-kms-key" --region "ITBG-Bergamo" --billing-period "Hour"

# Aggiorna una chiave KMS
acloud security kms update <kms-id> --name "updated-name" --tags "production"

# Elimina una chiave KMS
acloud security kms delete <kms-id>
```

## Struttura dei Comandi

Tutti i comandi di sicurezza seguono questa struttura:

```
acloud security <resource> <action> [arguments] [flags]
```

Dove:
- `<resource>`: Il tipo di risorsa (es. `kms`)
- `<action>`: L'operazione da eseguire (es. `list`, `get`, `create`, `update`, `delete`)
- `[arguments]`: Argomenti richiesti (es. ID risorse)
- `[flags]`: Flag opzionali (es. `--name`, `--region`, `--tags`)

## Pattern Comuni

### Elencare le Risorse

```bash
acloud security <resource> list
```

### Ottenere i Dettagli delle Risorse

```bash
acloud security <resource> get <resource-id>
```

### Creare Risorse

```bash
acloud security <resource> create [required-args] [flags]
```

### Aggiornare Risorse

```bash
acloud security <resource> update <resource-id> [flags]
```

### Eliminare Risorse

```bash
acloud security <resource> delete <resource-id> [--yes]
```

## Contesto Progetto

Le risorse di sicurezza sono limitate a un progetto. Puoi:

1. **Usare il flag `--project-id`:**
   ```bash
   acloud security kms list --project-id <project-id>
   ```

2. **Impostare un contesto:**
   ```bash
   acloud context set my-prod --project-id <project-id>
   acloud security kms list  # Usa l'ID progetto del contesto
   ```

Vedi [Installazione - Gestione Contesto](../installation.md#context-management) per maggiori informazioni.

## Prossimi Passi

- Esplora le [Risorse di Gestione](./management.md) per risorse a livello organizzativo
- Controlla le [Risorse Storage](./storage.md) per operazioni di storage
- Rivedi le [Risorse di Rete](./network.md) per capacità di networking
- Vedi le [Risorse Database](./database.md) per la gestione database
- Rivedi le [Risorse di Pianificazione](./schedule.md) per la pianificazione job

