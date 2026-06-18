# VPC Network (Virtual Private Cloud Network)

I Virtual Private Cloud forniscono ambienti di rete isolati per le tue risorse cloud.

## Comandi

### Elenca VPC

Elenca tutti i VPC nel tuo progetto.

```bash
acloud network vpc list [flags]
```

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
# Elenca VPC usando il contesto
acloud network vpc list

# Elenca VPC con ID progetto esplicito
acloud network vpc list --project-id 68398923fb2cb026400d4d31
```

**Output:**
```
NAME            ID                        SUBNETS    STATUS
production-vpc  689307f4745108d3c6343b5a  4          Active
test-vpc        69485a584d0cdc87949b6ff8  0          InCreation
```

### Ottieni Dettagli VPC

Ottieni informazioni dettagliate su un VPC specifico.

```bash
acloud network vpc get <vpc-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud network vpc get 689307f4745108d3c6343b5a
```

**Output:**
```
VPC Details:
============
ID:              689307f4745108d3c6343b5a
URI:             /projects/.../vpcs/689307f4745108d3c6343b5a
Name:            production-vpc
Default:         false
Linked Resources: 4
Creation Date:   06-08-2025 07:44:52
Created By:      aru-297647
Tags:            [production network critical]
Status:          Active
```

### Crea VPC

Crea un nuovo VPC.

```bash
acloud network vpc create [flags]
```

**Flag Richiesti:**
- `--name string` - Nome per il VPC
- `--region string` - Codice regione (es. ITBG-Bergamo)

**Flag Opzionali:**
- `--tags strings` - Tag per il VPC (separati da virgola)
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempi:**
```bash
# Crea un VPC base
acloud network vpc create --name "my-vpc" --region ITBG-Bergamo

# Crea VPC con tag
acloud network vpc create \
  --name "production-vpc" \
  --region ITBG-Bergamo \
  --tags production,network,critical

# Crea VPC con ID progetto esplicito
acloud network vpc create \
  --name "my-vpc" \
  --region ITBG-Bergamo \
  --project-id 68398923fb2cb026400d4d31
```

**Output:**
```
VPC created successfully!
ID:      69485a584d0cdc87949b6ff8
Name:    my-vpc
Default: false
```

**Note:**
- I VPC vengono creati con `Default: false` e `Preset: false` automaticamente
- Il VPC sarà inizialmente in stato **InCreation**
- Usa `acloud network vpc get <vpc-id>` per controllare quando diventa **Active**

### Aggiorna VPC

Aggiorna il nome e/o i tag di un VPC esistente.

```bash
acloud network vpc update <vpc-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC (supporta auto-completamento)

**Flag:**
- `--name string` - Nuovo nome per il VPC
- `--tags strings` - Nuovi tag per il VPC (separati da virgola)
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Nota:** Almeno uno tra `--name` o `--tags` deve essere fornito.

**Esempi:**
```bash
# Aggiorna nome VPC
acloud network vpc update 689307f4745108d3c6343b5a --name "new-vpc-name"

# Aggiorna tag VPC
acloud network vpc update 689307f4745108d3c6343b5a --tags production,updated,network

# Aggiorna sia nome che tag
acloud network vpc update 689307f4745108d3c6343b5a \
  --name "production-vpc" \
  --tags production,critical,frontend
```

**Output:**
```
VPC updated successfully!
ID:      689307f4745108d3c6343b5a
Name:    production-vpc
Tags:    [production critical frontend]
```

**Restrizioni:**
- Non è possibile aggiornare VPC in stato **InCreation**
- Attendi che il VPC raggiunga lo stato **Active** prima di aggiornare

### Elimina VPC

Elimina un VPC.

```bash
acloud network vpc delete <vpc-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)
- `-y, --yes` - Salta il prompt di conferma

**Esempi:**
```bash
# Elimina con prompt di conferma
acloud network vpc delete 689307f4745108d3c6343b5a

# Elimina senza conferma
acloud network vpc delete 689307f4745108d3c6343b5a --yes

# Elimina con ID progetto esplicito
acloud network vpc delete 689307f4745108d3c6343b5a \
  --project-id 68398923fb2cb026400d4d31 \
  --yes
```

**Prompt di Conferma:**
```
Are you sure you want to delete VPC 689307f4745108d3c6343b5a? This action cannot be undone.
Type 'yes' to confirm: yes

VPC 689307f4745108d3c6343b5a deleted successfully!
```

**Note:**
- I VPC eliminati mostreranno lo stato **Deleting** prima di essere rimossi
- Assicurati che nessuna risorsa stia usando il VPC prima dell'eliminazione
- L'eliminazione non può essere annullata

## Auto-completamento Shell

I comandi VPC supportano auto-completamento intelligente per gli ID VPC:

```bash
# Abilita completamento (bash)
source <(acloud completion bash)

# Digita il comando e premi TAB per vedere gli ID VPC disponibili
acloud network vpc get <TAB>
acloud network vpc update <TAB>
acloud network vpc delete <TAB>
```

L'auto-completamento mostra gli ID VPC con i loro nomi:
```
689307f4745108d3c6343b5a    production-vpc
69485a584d0cdc87949b6ff8    test-vpc
```

## Stati VPC

I VPC possono essere nei seguenti stati:

| Stato | Descrizione | Posso Aggiornare? | Posso Eliminare? |
|-------|-------------|-------------------|------------------|
| InCreation | Il VPC è in fase di creazione | ❌ No | ❌ No |
| Active | Il VPC è pronto per l'uso | ✅ Sì | ✅ Sì |
| Deleting | Il VPC è in fase di eliminazione | ❌ No | ❌ No |

## Proprietà VPC

### Proprietà Default

La proprietà `Default` indica se un VPC è il VPC predefinito per il progetto:
- Gestita automaticamente da Aruba Cloud
- Non può essere impostata dagli utenti tramite CLI
- Solo un VPC per progetto può essere predefinito

### Proprietà Preset

La proprietà `Preset` indica se un VPC usa configurazioni predefinite:
- Sempre impostata su `false` per i VPC creati dagli utenti
- Non può essere modificata tramite CLI

### Risorse Collegate

I VPC possono avere risorse collegate (subnet, interfacce di rete, ecc.):
- Mostrate come conteggio nella vista elenco
- Dettagli visibili nel comando get
- Devono essere rimosse prima dell'eliminazione del VPC

## Workflow Comuni

### Creazione e Configurazione di un VPC

```bash
# 1. Crea il VPC
VPC_ID=$(acloud network vpc create \
  --name "production-vpc" \
  --region ITBG-Bergamo \
  --tags production | grep "ID:" | awk '{print $2}')

# 2. Attendi il completamento della creazione
while true; do
  STATUS=$(acloud network vpc get $VPC_ID | grep "Status:" | awk '{print $2}')
  if [ "$STATUS" = "Active" ]; then
    break
  fi
  echo "Waiting for VPC to become Active... (current: $STATUS)"
  sleep 5
done

# 3. Aggiorna con tag aggiuntivi
acloud network vpc update $VPC_ID --tags production,critical,network

# 4. Verifica la configurazione
acloud network vpc get $VPC_ID
```

### Gestione di Più VPC

```bash
# Elenca tutti i VPC
acloud network vpc list

# Tagga i VPC per ambiente
acloud network vpc update <vpc-id-1> --tags production,backend
acloud network vpc update <vpc-id-2> --tags staging,frontend
acloud network vpc update <vpc-id-3> --tags development,testing

# Ottieni dettagli di tutti i VPC
for vpc_id in $(acloud network vpc list | tail -n +2 | awk '{print $2}'); do
  echo "=== VPC: $vpc_id ==="
  acloud network vpc get $vpc_id
  echo ""
done
```

### Pulizia VPC di Test

```bash
# Elenca tutti i VPC
acloud network vpc list

# Elimina VPC di test (salta conferma con --yes)
acloud network vpc delete <test-vpc-id> --yes

# Verifica eliminazione
acloud network vpc list
```

## Best Practices

1. **Usa Nomi Descrittivi**
   ```bash
   acloud network vpc create --name "prod-backend-vpc" --region ITBG-Bergamo
   ```

2. **Tagga per Ambiente e Scopo**
   ```bash
   acloud network vpc update <vpc-id> --tags production,backend,critical
   ```

3. **Attendi lo Stato Active Prima della Configurazione**
   ```bash
   # Controlla lo stato prima di aggiornare
   acloud network vpc get <vpc-id>
   # Assicurati che lo Status sia "Active"
   acloud network vpc update <vpc-id> --name "new-name"
   ```

4. **Usa Contesti Progetto**
   ```bash
   acloud context use prod-project
   acloud network vpc list  # Non serve --project-id
   ```

5. **Documenta lo Scopo del VPC nei Tag**
   ```bash
   acloud network vpc create \
     --name "api-vpc" \
     --region ITBG-Bergamo \
     --tags api,public-facing,load-balanced
   ```

## Risoluzione dei Problemi

### "Cannot update VPC while in InCreation state"

**Problema:** Tentativo di aggiornare un VPC che non ha finito di crearsi.

**Soluzione:**
```bash
# Controlla lo stato corrente
acloud network vpc get <vpc-id>

# Attendi che lo Status diventi "Active"
# Poi riprova l'aggiornamento
acloud network vpc update <vpc-id> --name "new-name"
```

### "Failed to create VPC - Status: 400"

**Problema:** Formato regione non valido o campi richiesti mancanti.

**Soluzione:**
```bash
# Usa il formato regione corretto
acloud network vpc create --name "my-vpc" --region ITBG-Bergamo

# Non: --region eu-west-1 (formato errato)
```

### "Error: at least one of --name or --tags must be provided"

**Problema:** Comando update chiamato senza modifiche.

**Soluzione:**
```bash
# Fornisci almeno un campo da aggiornare
acloud network vpc update <vpc-id> --name "new-name"
# oppure
acloud network vpc update <vpc-id> --tags tag1,tag2
```

### VPC Mostrato come Default ma Non Creato da Me

**Spiegazione:** Aruba Cloud crea automaticamente un VPC predefinito per ogni progetto. Questo è normale ed è gestito dalla piattaforma.

## Comandi Correlati

- [Elastic IP](elasticip.md) - Assegna IP pubblici all'interno dei VPC
- [Load Balancer](loadbalancer.md) - Distribuisci il traffico all'interno dei VPC
- [Gestione Contesto](../../installation.md#context-management) - Gestisci contesti progetto
