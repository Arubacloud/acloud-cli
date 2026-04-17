# Gestione Ripristino

Le operazioni di ripristino usano la risorsa Aruba.Storage/restore per ripristinare volumi di block storage da backup. A differenza dei ripristini da snapshot che creano nuovi volumi, i ripristini da backup scrivono i dati su volumi esistenti.

## Comandi Disponibili

- `acloud storage restore` - Crea un'operazione di ripristino da un backup
- `acloud storage restore list` - Elenca operazioni di ripristino per un backup
- `acloud storage restore get` - Ottieni i dettagli di un'operazione di ripristino specifica
- `acloud storage restore update` - Aggiorna nome e tag dell'operazione di ripristino
- `acloud storage restore delete` - Elimina un'operazione di ripristino

## Ripristino vs Ripristino Snapshot

| Caratteristica | Ripristino Backup | Ripristino Snapshot |
|---------|---------------|------------------|
| Risorsa API | Aruba.Storage/restore | Crea nuovo volume |
| Target | Volume esistente | Nuovo volume |
| Gerarchico | Annidato sotto backup | Standalone |
| Caso d'Uso | Ripristina sullo stesso volume | Clona su nuovo volume |

## Crea Operazione di Ripristino

Crea un'operazione di ripristino per ripristinare dati da un backup a un volume esistente.

**Importante:** L'operazione di ripristino scrive dati SU un volume esistente. Il volume target deve esistere e sarà sovrascritto.

### Utilizzo

```bash
acloud storage restore <backup-id> <volume-id> --name <name> [flags]
```

### Argomenti

- `backup-id` - L'ID del backup da cui ripristinare
- `volume-id` - L'ID del volume target (sarà sovrascritto)

### Flag Richiesti

- `--name` - Nome per l'operazione di ripristino
- `--region` - Codice regione

### Flag Opzionali

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--tags` - Tag (separati da virgola)

### Esempio

```bash
# Ripristina backup sullo stesso volume da cui è stato creato
acloud storage restore 694594818f4a09c12b5e0c19 69455aa70d0972656501d45d \
  --name "restore-after-failure" \
  --region "ITBG-Bergamo" \
  --tags "recovery,production"

# Ripristina su un volume diverso
acloud storage restore 694594818f4a09c12b5e0c19 69455bb80d0972656501d45e \
  --name "clone-to-test" \
  --region "ITBG-Bergamo"
```

### Output

```
Creating restore operation with the following parameters:
  Name:       restore-after-failure
  Region:     ITBG-Bergamo
  Backup ID:  694594818f4a09c12b5e0c19
  Backup URI: /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/backups/694594818f4a09c12b5e0c19
  Volume ID:  69455aa70d0972656501d45d
  Volume URI: /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
  Tags:       [recovery production]

Restore operation created successfully!
ID:              6945953a8f4a09c12b5e0c71
Name:            restore-after-failure
Creation Date:   19-12-2025 18:11:06
Status:          InCreation
```

## Elenca Operazioni di Ripristino

Elenca tutte le operazioni di ripristino per un backup specifico.

**Nota:** Le operazioni di ripristino sono gerarchiche e annidate sotto i backup, quindi devi specificare l'ID del backup.

### Utilizzo

```bash
acloud storage restore list <backup-id> [flags]
```

### Argomenti

- `backup-id` - L'ID del backup (supporta auto-completamento)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Esempio

```bash
acloud storage restore list 694594818f4a09c12b5e0c19
```

### Output

```
NAME                    ID                          STATUS
restore-after-failure   6945953a8f4a09c12b5e0c71   Active
test-restore            6945954b8f4a09c12b5e0c72   Active
```

## Ottieni Dettagli Ripristino

Ottieni informazioni dettagliate su un'operazione di ripristino specifica.

### Utilizzo

```bash
acloud storage restore get <backup-id> <restore-id> [flags]
```

### Argomenti

- `backup-id` - L'ID del backup (supporta auto-completamento)
- `restore-id` - L'ID dell'operazione di ripristino (auto-completa in base al backup selezionato)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)

### Auto-completamento

I comandi di ripristino supportano auto-completamento gerarchico:

1. Prima, premi TAB per vedere gli ID backup disponibili
2. Dopo aver inserito un ID backup, premi TAB di nuovo per vedere gli ID restore per quel backup

```bash
acloud storage restore get <TAB>
# Mostra ID backup:
# 67649dac8c7bb1c5d7c80631    MyBackup
# ...

acloud storage restore get 67649dac8c7bb1c5d7c80631 <TAB>
# Mostra ID restore per quel backup:
# 67664dde0aca19a92c2c48bb    RestoreOperation1
# ...
```

### Esempio

```bash
acloud storage restore get 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71
```

### Output

```
Restore Operation Details:
==========================
ID:              6945953a8f4a09c12b5e0c71
URI:             /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/backups/694594818f4a09c12b5e0c19/restores/6945953a8f4a09c12b5e0c71
Name:            restore-after-failure
Target Volume:   /projects/68398923fb2cb026400d4d31/providers/Aruba.Storage/blockStorages/69455aa70d0972656501d45d
Region:          ITBG-Bergamo
Status:          Active
Creation Date:   19-12-2025 18:11:06
Created By:      aru-297647
Tags:            [recovery production]
```

## Aggiorna Operazione di Ripristino

Aggiorna il nome e/o i tag di un'operazione di ripristino.

**Importante:** Le operazioni di ripristino possono essere aggiornate solo quando sono in stato "Active" (non "InCreation").

### Utilizzo

```bash
acloud storage restore update <backup-id> <restore-id> [flags]
```

### Argomenti

- `backup-id` - L'ID del backup (supporta auto-completamento)
- `restore-id` - L'ID dell'operazione di ripristino (auto-completa in base al backup selezionato)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `--name` - Nuovo nome per l'operazione di ripristino
- `--tags` - Nuovi tag (separati da virgola)

**Nota:** Almeno uno tra `--name` o `--tags` deve essere fornito.

### Esempio

```bash
# Aggiorna solo il nome
acloud storage restore update 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71 \
  --name "restore-renamed"

# Aggiorna solo i tag
acloud storage restore update 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71 \
  --tags "prod,final"

# Aggiorna entrambi
acloud storage restore update 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71 \
  --name "production-restore" \
  --tags "prod,final"
```

### Output

```
Restore operation updated successfully!
ID:              6945953a8f4a09c12b5e0c71
Name:            production-restore
Tags:            [prod final]
```

## Elimina Operazione di Ripristino

Elimina un record di operazione di ripristino. Questo non annulla il ripristino; elimina solo i metadati dell'operazione.

### Utilizzo

```bash
acloud storage restore delete <backup-id> <restore-id> [flags]
```

### Argomenti

- `backup-id` - L'ID del backup (supporta auto-completamento)
- `restore-id` - L'ID dell'operazione di ripristino (auto-completa in base al backup selezionato)

### Flag

- `--project-id` - ID progetto (usa il contesto se non specificato)
- `-y, --yes` - Salta il prompt di conferma

### Esempio

```bash
# Con prompt di conferma
acloud storage restore delete 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71

# Salta conferma
acloud storage restore delete 694594818f4a09c12b5e0c19 6945953a8f4a09c12b5e0c71 --yes
```

### Output

```
Restore operation 6945953a8f4a09c12b5e0c71 deleted successfully!
```

## Workflow di Ripristino

### 1. Identifica Backup
```bash
# Elenca backup disponibili
acloud storage backup list

# Ottieni dettagli backup per verificare che sia quello giusto
acloud storage backup get <backup-id>
```

### 2. Prepara Volume Target
```bash
# Assicurati che il volume target esista e sia pronto
acloud storage blockstorage get <volume-id>

# IMPORTANTE: Scollega il volume dalla VM se collegato
# Il ripristino sovrascriverà tutti i dati sul volume
```

### 3. Crea Operazione di Ripristino
```bash
acloud storage restore <backup-id> <volume-id> \
  --name "restore-$(date +%Y%m%d)" \
  --region "ITBG-Bergamo"
```

### 4. Monitora Ripristino
```bash
# Controlla lo stato del ripristino
acloud storage restore get <backup-id> <restore-id>

# Attendi fino a quando lo stato è "Active"
```

### 5. Verifica e Pulizia
```bash
# Dopo il ripristino riuscito, elimina il record dell'operazione
acloud storage restore delete <backup-id> <restore-id> --yes
```

## Stato Ripristino

- **InCreation**: L'operazione di ripristino è in corso
- **Active**: Il ripristino completato con successo
- **Failed**: Il ripristino ha incontrato un errore

## Considerazioni Importanti

### Sovrascrittura Dati
- Le operazioni di ripristino **sovrascrivono** completamente il volume target
- Tutti i dati esistenti sul volume target saranno persi
- Verifica sempre che stai ripristinando sul volume corretto
- Considera di creare uno snapshot del volume target prima del ripristino

### Requisiti Volume
- Il volume target deve esistere prima di creare il ripristino
- Il volume target dovrebbe essere scollegato dalle VM
- Il volume target deve essere nella stessa regione del backup
- La dimensione del volume target dovrebbe corrispondere o superare la dimensione del backup

### Disponibilità Backup
- Il backup deve essere in stato "Active"
- Non è possibile aggiornare il backup mentre il ripristino è in corso
- Possono essere creati più ripristini dallo stesso backup

### Struttura Gerarchica
- Le operazioni di ripristino sono annidate sotto i backup
- Tutti i comandi di ripristino richiedono il parametro backup-id
- Usa `restore list <backup-id>` per vedere tutti i ripristini per un backup

## Best Practices

1. **Testa in non-produzione prima**: Testa sempre le procedure di ripristino in un ambiente di test
2. **Verifica backup prima del ripristino**: Controlla dettagli e stato del backup prima di ripristinare
3. **Scollega volumi**: Assicurati che i volumi target siano scollegati dalle VM
4. **Crea snapshot pre-ripristino**: Fai uno snapshot del volume target prima del ripristino come misura di sicurezza
5. **Monitora stato ripristino**: Controlla lo stato del ripristino fino a quando è "Active"
6. **Valida dopo ripristino**: Verifica l'integrità dei dati dopo il completamento del ripristino
7. **Pulisci record**: Elimina i record delle operazioni di ripristino dopo la verifica
8. **Documenta procedure**: Mantieni runbook per scenari di ripristino di emergenza

## Risoluzione dei Problemi

### Ripristino Non Riesce a Crearsi
- Verifica che il backup sia in stato "Active"
- Controlla che il volume target esista e sia accessibile
- Assicurati che la regione corrisponda tra backup e volume
- Verifica che non ci siano altri ripristini in esecuzione sullo stesso volume

### Non Posso Aggiornare Backup
- Errore: "Backup can't be deleted or modified because there is a running restore operation"
- Soluzione: Attendi il completamento del ripristino o elimina l'operazione di ripristino

### Ripristino Bloccato in "InCreation"
- Attendi alcuni minuti poiché le operazioni di ripristino possono richiedere tempo
- Controlla lo stato del backup e del volume
- Contatta il supporto se lo stato non cambia dopo un periodo prolungato

## Note

- Le operazioni di ripristino sono asincrone e iniziano in stato "InCreation"
- I ripristini scrivono SU volumi esistenti (diverso dal ripristino snapshot)
- Puoi tracciare la cronologia dei ripristini mantenendo i record delle operazioni
- Le operazioni di ripristino usano la risorsa API Aruba.Storage/restore
- I tag possono aiutare a organizzare e tracciare le operazioni di ripristino per scopo o data

