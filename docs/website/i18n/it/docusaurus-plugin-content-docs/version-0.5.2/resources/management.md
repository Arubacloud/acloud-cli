# Risorse di Gestione

La categoria `management` fornisce comandi per gestire risorse a livello organizzativo in Aruba Cloud.

## Risorse Disponibili

### [Progetti](management/project.md)

I progetti sono unità organizzative che raggruppano risorse correlate insieme. Forniscono un modo per organizzare e gestire le tue risorse cloud.

**Comandi Rapidi:**
```bash
# Elenca tutti i progetti
acloud management project list

# Ottieni i dettagli del progetto
acloud management project get <project-id>

# Crea un progetto
acloud management project create --name "my-project"

# Aggiorna un progetto
acloud management project update <project-id> --description "Descrizione aggiornata"

# Elimina un progetto
acloud management project delete <project-id>
```

## Struttura dei Comandi

Tutti i comandi di gestione seguono questa struttura:

```
acloud management <resource> <action> [arguments] [flags]
```

Dove:
- `<resource>`: Il tipo di risorsa (es. `project`)
- `<action>`: L'operazione da eseguire (es. `list`, `get`, `create`, `update`, `delete`)
- `[arguments]`: Argomenti richiesti (es. ID risorse)
- `[flags]`: Flag opzionali (es. `--name`, `--description`)

## Pattern Comuni

### Elencare le Risorse

```bash
acloud management <resource> list
```

Elenca tutte le risorse del tipo specificato con informazioni chiave visualizzate in formato tabella.

### Ottenere i Dettagli delle Risorse

```bash
acloud management <resource> get <resource-id>
```

Visualizza informazioni dettagliate su una risorsa specifica.

### Creare Risorse

```bash
acloud management <resource> create --flag1 value1 --flag2 value2
```

Crea una nuova risorsa con le proprietà specificate.

### Aggiornare Risorse

```bash
acloud management <resource> update <resource-id> --flag1 value1
```

Aggiorna una risorsa esistente. Solo i campi forniti vengono modificati.

### Eliminare Risorse

```bash
acloud management <resource> delete <resource-id>
```

Elimina la risorsa specificata. Potrebbe richiedere conferma.

## Auto-completamento

Gli ID risorse supportano l'auto-completamento. Dopo aver configurato il completamento shell, puoi:

```bash
acloud management project get <TAB>
```

Questo mostrerà gli ID progetto disponibili con i loro nomi.

## Prossimi Passi

- [Guida alla Gestione Progetti](management/project.md)

