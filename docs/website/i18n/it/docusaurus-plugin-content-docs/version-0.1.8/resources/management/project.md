# Gestione Progetti

I progetti sono unità organizzative che ti aiutano a raggruppare e gestire risorse correlate in Aruba Cloud. Ogni progetto può contenere più risorse e ha i propri metadati.

## Indice

- [Panoramica](#panoramica)
- [Comandi](#comandi)
  - [Elenca Progetti](#elenca-progetti)
  - [Ottieni Dettagli Progetto](#ottieni-dettagli-progetto)
  - [Crea Progetto](#crea-progetto)
  - [Aggiorna Progetto](#aggiorna-progetto)
  - [Elimina Progetto](#elimina-progetto)
- [Esempi](#esempi)
- [Best Practices](#best-practices)

## Panoramica

Un progetto in Aruba Cloud consiste di:
- **ID**: Identificatore univoco (generato automaticamente)
- **Nome**: Nome leggibile dall'uomo
- **Descrizione**: Descrizione opzionale
- **Tag**: Etichette opzionali per l'organizzazione
- **Default**: Flag che indica se è il progetto predefinito
- **Risorse**: Numero di risorse nel progetto
- **Data di Creazione**: Quando il progetto è stato creato
- **Creato Da**: Utente che ha creato il progetto

## Comandi

### Elenca Progetti

Visualizza tutti i progetti in formato tabella.

**Sintassi:**
```bash
acloud management project list
```

**Output:**
```
ID                             NAME                                     CREATION DATE   
655b2822af30f667f826994e       defaultproject                           20-11-2023      
66a10244f62b99c686572a9f       develop                                  24-07-2024      
68398923fb2cb026400d4d31       github-runner                            30-05-2025
```

**Opzioni:**
- Nessuna

---

### Ottieni Dettagli Progetto

Recupera informazioni dettagliate su un progetto specifico.

**Sintassi:**
```bash
acloud management project get <project-id>
```

**Argomenti:**
- `project-id` (richiesto): L'ID univoco del progetto

**Esempio:**
```bash
acloud management project get 655b2822af30f667f826994e
```

**Output:**
```
Project Details:
================
ID:              655b2822af30f667f826994e
Name:            defaultproject
Description:     defaultproject
Default:         false
Resources:       5
Creation Date:   20-11-2023 09:34:26
Created By:      aru-297647
Tags:            [production arubacloud-sdk]
```

**Auto-completamento:**
```bash
acloud management project get <TAB>
# Mostra elenco di ID progetto con nomi
```

---

### Crea Progetto

Crea un nuovo progetto con proprietà specificate.

**Sintassi:**
```bash
acloud management project create --name <name> [flags]
```

**Flag Richiesti:**
- `--name <string>`: Nome per il progetto

**Flag Opzionali:**
- `--description <string>`: Descrizione per il progetto
- `--tags <tag1,tag2>`: Tag separati da virgola
- `--default`: Imposta come progetto predefinito (default: false)

**Esempi:**

1. **Creazione base:**
   ```bash
   acloud management project create --name "my-project"
   ```

2. **Con descrizione:**
   ```bash
   acloud management project create \
     --name "production-env" \
     --description "Production environment resources"
   ```

3. **Con tag:**
   ```bash
   acloud management project create \
     --name "dev-project" \
     --description "Development environment" \
     --tags dev,testing,internal
   ```

4. **Imposta come predefinito:**
   ```bash
   acloud management project create \
     --name "main-project" \
     --default
   ```

**Output:**
```
Project created successfully!
ID:              69440ae8914afa1ec8b607c1
Name:            my-project
Description:     Production environment resources
Tags:            [dev testing internal]
Default:         false
Creation Date:   18-12-2025 14:08:40
```

---

### Aggiorna Progetto

Aggiorna la descrizione e/o i tag di un progetto esistente.

**Sintassi:**
```bash
acloud management project update <project-id> [flags]
```

**Argomenti:**
- `project-id` (richiesto): L'ID univoco del progetto

**Flag:**
- `--description <string>`: Nuova descrizione per il progetto
- `--tags <tag1,tag2>`: Nuovi tag per il progetto (sostituisce quelli esistenti)

**Nota:** Almeno un flag deve essere fornito. Il nome e lo stato predefinito non possono essere modificati dopo la creazione.

**Esempi:**

1. **Aggiorna descrizione:**
   ```bash
   acloud management project update 69137e295956b621e2048eab \
     --description "Updated description"
   ```

2. **Aggiorna tag:**
   ```bash
   acloud management project update 69137e295956b621e2048eab \
     --tags production,critical,monitored
   ```

3. **Aggiorna entrambi:**
   ```bash
   acloud management project update 69137e295956b621e2048eab \
     --description "Production environment" \
     --tags prod,active
   ```

**Output:**
```
Project updated successfully!
ID:              69137e295956b621e2048eab
Name:            seca-sdk-example
Description:     Updated description
Tags:            [production critical monitored]
Default:         false
```

**Auto-completamento:**
```bash
acloud management project update <TAB>
# Mostra elenco di ID progetto con nomi
```

---

### Elimina Progetto

Elimina un progetto esistente.

**Sintassi:**
```bash
acloud management project delete <project-id> [flags]
```

**Argomenti:**
- `project-id` (richiesto): L'ID univoco del progetto da eliminare

**Flag:**
- `-y, --yes`: Salta il prompt di conferma

**Esempi:**

1. **Con conferma:**
   ```bash
   acloud management project delete 69440ae8914afa1ec8b607c1
   ```
   
   Output:
   ```
   Are you sure you want to delete project 69440ae8914afa1ec8b607c1? (yes/no): yes
   
   Project 69440ae8914afa1ec8b607c1 deleted successfully!
   ```

2. **Salta conferma:**
   ```bash
   acloud management project delete 69440ae8914afa1ec8b607c1 --yes
   ```
   
   Output:
   ```
   Project 69440ae8914afa1ec8b607c1 deleted successfully!
   ```

**Auto-completamento:**
```bash
acloud management project delete <TAB>
# Mostra elenco di ID progetto con nomi
```

**Avviso:** Eliminare un progetto è permanente e non può essere annullato. Assicurati che tutte le risorse nel progetto siano gestite correttamente prima dell'eliminazione.

---

## Esempi

### Workflow Completo

1. **Crea un nuovo progetto:**
   ```bash
   acloud management project create \
     --name "webapp-project" \
     --description "Web application resources" \
     --tags web,production,frontend
   ```

2. **Elenca progetti per verificare:**
   ```bash
   acloud management project list
   ```

3. **Ottieni i dettagli del nuovo progetto:**
   ```bash
   acloud management project get 69440ae8914afa1ec8b607c1
   ```

4. **Aggiorna la descrizione del progetto:**
   ```bash
   acloud management project update 69440ae8914afa1ec8b607c1 \
     --description "Web application production environment"
   ```

5. **Aggiungi più tag:**
   ```bash
   acloud management project update 69440ae8914afa1ec8b607c1 \
     --tags web,production,frontend,monitored,critical
   ```

6. **Quando non più necessario, elimina:**
   ```bash
   acloud management project delete 69440ae8914afa1ec8b607c1 --yes
   ```

### Utilizzo degli ID Progetto in Altri Comandi

Molti comandi di risorse richiedono un flag `--project-id`. Puoi usare il completamento per trovare il progetto giusto:

```bash
# Esempio: Creazione di una risorsa compute in un progetto specifico
acloud compute cloudserver create \
  --project-id <TAB>  # L'auto-completamento mostra i tuoi progetti
  --name "my-server" \
  --flavor "small"
```

### Filtraggio Progetti per Tag

Mentre la CLI non ha filtri integrati, puoi usare strumenti Unix standard:

```bash
# Salva elenco progetti
acloud management project list > projects.txt

# Usa grep per cercare
grep "production" projects.txt
```

## Best Practices

### Convenzioni di Denominazione

Usa nomi chiari e descrittivi:
- ✅ `production-webapp`
- ✅ `dev-testing-env`
- ✅ `staging-api-services`
- ❌ `proj1`
- ❌ `test`
- ❌ `temp`

### Strategia di Tagging

Usa tag coerenti per l'organizzazione:
- **Ambiente**: `dev`, `staging`, `production`
- **Team**: `frontend`, `backend`, `devops`
- **Stato**: `active`, `archived`, `deprecated`
- **Centro di Costo**: `engineering`, `marketing`, `sales`

Esempio:
```bash
acloud management project create \
  --name "api-production" \
  --tags production,backend,active,engineering
```

### Linee Guida per le Descrizioni

Includi informazioni utili nelle descrizioni:
- Scopo del progetto
- Team o proprietario
- Note importanti

Esempio:
```bash
acloud management project create \
  --name "customer-api" \
  --description "Customer-facing REST API - Backend Team - Production Critical"
```

### Organizzazione delle Risorse

- Raggruppa risorse correlate nello stesso progetto
- Usa progetti separati per ambienti diversi (dev, staging, prod)
- Non mescolare risorse non correlate nello stesso progetto

### Pulizia

- Rivedi regolarmente ed elimina progetti non utilizzati
- Archivia progetti vecchi aggiornando la loro descrizione
- Usa tag per identificare progetti che possono essere eliminati

## Risoluzione dei Problemi

### "Error initializing client"

Assicurati di aver configurato le tue credenziali:
```bash
acloud config set
```

### "Project not found"

Verifica che l'ID progetto esista:
```bash
acloud management project list
```

### "Error: --name is required"

Il flag `--name` è obbligatorio quando si creano progetti:
```bash
acloud management project create --name "my-project"
```

### "Error: at least one of --description or --tags must be provided"

Quando aggiorni, fornisci almeno un campo da aggiornare:
```bash
acloud management project update <id> --description "New description"
```

## Risorse Correlate

- [Guida Installazione](../../installation.md)
- [Panoramica Risorse di Gestione](../management.md)
- [Documentazione API](https://www.arubacloud.com/docs)

