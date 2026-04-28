# VPC Peering Route

Le VPC Peering Route definiscono regole di routing per il traffico tra VPC in peering in Aruba Cloud. Queste route controllano come il traffico di rete viene diretto tra VPC collegati tramite una connessione VPC Peering specificando indirizzi di rete locali e remoti.

## Comandi

### Elenca VPC Peering Route

Elenca tutte le route per una connessione VPC peering specifica.

```bash
acloud network vpcpeeringroute list <vpc-id> <peering-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `peering-id` - L'ID della connessione VPC peering

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud network vpcpeeringroute list 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204
```

**Output:**
```
NAME            ID                        LOCAL NETWORK     REMOTE NETWORK    STATUS
route-1         1234567890abcdef123456   10.0.1.0/24       10.1.1.0/24       Active
route-2         1234567890abcdef123457   10.0.2.0/24       10.1.2.0/24       Active
```

### Ottieni Dettagli VPC Peering Route

Ottieni informazioni dettagliate su una VPC peering route specifica.

```bash
acloud network vpcpeeringroute get <vpc-id> <peering-id> <route-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `peering-id` - L'ID della connessione VPC peering
- `route-id` - L'ID della route (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud network vpcpeeringroute get 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456
```

**Output:**
```
VPC Peering Route Details:
==========================
ID:              1234567890abcdef123456
URI:             /projects/.../vpcpeeringroutes/1234567890abcdef123456
Name:            route-1
Local Network:   10.0.1.0/24
Remote Network:  10.1.1.0/24
Billing Period:  Hour
Creation Date:   06-08-2025 07:44:52
Created By:      aru-297647
Tags:            [vpc,peering,production]
Status:          Active
```

### Crea VPC Peering Route

Crea una nuova route per una connessione VPC peering.

```bash
acloud network vpcpeeringroute create <vpc-id> <peering-id> [flags]
```

**Flag Richiesti:**
- `--name string` - Nome VPC Peering Route
- `--local-network string` - Indirizzo di rete locale in notazione CIDR
- `--remote-network string` - Indirizzo di rete remoto in notazione CIDR

**Flag Opzionali:**
- `--billing-period string` - Periodo di fatturazione: Hour, Month, Year (default: Hour)
- `--tags strings` - Tag per la VPC peering route (separati da virgola)
- `--project-id string` - ID progetto (usa il contesto se non specificato)
- `-v, --verbose` - Mostra informazioni di debug dettagliate

**Esempi:**
```bash
# Crea una VPC peering route base
acloud network vpcpeeringroute create 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 \
  --name "route-1" \
  --local-network "10.0.1.0/24" \
  --remote-network "10.1.1.0/24"

# Crea VPC peering route con periodo di fatturazione e tag
acloud network vpcpeeringroute create 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 \
  --name "production-route" \
  --local-network "10.0.2.0/24" \
  --remote-network "10.1.2.0/24" \
  --billing-period Month \
  --tags "vpc,peering,production"
```

**Output:**
```
NAME            ID                        LOCAL NETWORK     REMOTE NETWORK    STATUS
route-1         1234567890abcdef123456   10.0.1.0/24       10.1.1.0/24       Active
```

**Note:**
- La VPC peering route sarà inizialmente in stato **InCreation**
- Usa `acloud network vpcpeeringroute get` per controllare quando diventa **Active**

### Aggiorna VPC Peering Route

Aggiorna le proprietà di una VPC peering route esistente.

```bash
acloud network vpcpeeringroute update <vpc-id> <peering-id> <route-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `peering-id` - L'ID della connessione VPC peering
- `route-id` - L'ID della route (supporta auto-completamento)

**Flag:**
- `--name string` - Nuovo nome per la VPC peering route
- `--tags strings` - Nuovi tag per la VPC peering route (separati da virgola)
- `--local-network string` - Indirizzo di rete locale in notazione CIDR
- `--remote-network string` - Indirizzo di rete remoto in notazione CIDR
- `--billing-period string` - Periodo di fatturazione: Hour, Month, Year
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Nota:** Almeno un campo deve essere fornito per l'aggiornamento.

**Esempi:**
```bash
# Aggiorna nome VPC peering route
acloud network vpcpeeringroute update 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456 \
  --name "updated-route-1"

# Aggiorna rete locale
acloud network vpcpeeringroute update 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456 \
  --local-network "10.0.3.0/24"

# Aggiorna periodo di fatturazione
acloud network vpcpeeringroute update 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456 \
  --billing-period Month

# Aggiorna più campi
acloud network vpcpeeringroute update 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456 \
  --name "production-route" \
  --local-network "10.0.2.0/24" \
  --remote-network "10.1.2.0/24" \
  --billing-period Month \
  --tags "vpc,peering,production,updated"
```

**Output:**
```
NAME            ID                        LOCAL NETWORK     REMOTE NETWORK    STATUS
production-route 1234567890abcdef123456   10.0.2.0/24       10.1.2.0/24       Active
```

**Restrizioni:**
- Non è possibile aggiornare VPC peering route in stato **InCreation**
- Attendi che la VPC peering route raggiunga lo stato **Active** prima di aggiornare

### Elimina VPC Peering Route

Elimina una VPC peering route.

```bash
acloud network vpcpeeringroute delete <vpc-id> <peering-id> <route-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `peering-id` - L'ID della connessione VPC peering
- `route-id` - L'ID della route (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)
- `-y, --yes` - Salta il prompt di conferma

**Esempi:**
```bash
# Elimina con prompt di conferma
acloud network vpcpeeringroute delete 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456

# Elimina senza conferma
acloud network vpcpeeringroute delete 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 1234567890abcdef123456 --yes
```

**Prompt di Conferma:**
```
Are you sure you want to delete VPC peering route 1234567890abcdef123456? This action cannot be undone.
Type 'yes' to confirm: yes
```

**Output:**
```
ID                              STATUS
1234567890abcdef123456         deleted
```

**Note:**
- L'eliminazione non può essere annullata
- Assicurati che la connessione VPC peering non dipenda dalla route prima dell'eliminazione

## Auto-completamento Shell

I comandi VPC Peering Route supportano auto-completamento intelligente per gli ID route:

```bash
# Abilita completamento (bash)
source <(acloud completion bash)

# Digita il comando e premi TAB per vedere gli ID route disponibili
acloud network vpcpeeringroute get <vpc-id> <peering-id> <TAB>
acloud network vpcpeeringroute update <vpc-id> <peering-id> <TAB>
acloud network vpcpeeringroute delete <vpc-id> <peering-id> <TAB>
```

L'auto-completamento mostra gli ID route con i loro nomi:
```
1234567890abcdef123456    route-1
1234567890abcdef123457    route-2
```

## Proprietà VPC Peering Route

### Indirizzo di Rete Locale

L'indirizzo di rete locale (CIDR) rappresenta l'intervallo di rete nel VPC locale che dovrebbe essere accessibile tramite la connessione peering.

**Esempi:**
- `10.0.1.0/24` - Subnet specifica nel VPC locale
- `10.0.0.0/16` - Intero intervallo di rete VPC locale

### Indirizzo di Rete Remoto

L'indirizzo di rete remoto (CIDR) rappresenta l'intervallo di rete nel VPC remoto che dovrebbe essere accessibile tramite la connessione peering.

**Esempi:**
- `10.1.1.0/24` - Subnet specifica nel VPC remoto
- `10.1.0.0/16` - Intero intervallo di rete VPC remoto

### Periodo di Fatturazione

Il periodo di fatturazione determina come viene fatturata la VPC peering route:

- **Hour**: Fatturazione pay-per-hour (default)
- **Month**: Fatturazione mensile
- **Year**: Fatturazione annuale (miglior risparmio)

## Stati VPC Peering Route

Le VPC peering route possono essere nei seguenti stati:

| Stato | Descrizione | Posso Aggiornare? | Posso Eliminare? |
|-------|-------------|-------------------|------------------|
| InCreation | La VPC peering route è in fase di creazione | ❌ No | ❌ No |
| Active | La VPC peering route è pronta per l'uso | ✅ Sì | ✅ Sì |

## Workflow Comuni

### Configurazione VPC Peering Route

```bash
# 1. Crea VPC peering (se non esiste)
VPC_ID="689307f4745108d3c6343b5a"
PEERING_ID=$(acloud network vpcpeering create $VPC_ID \
  --name "prod-peering" \
  --region ITBG-Bergamo | grep "ID:" | awk '{print $2}')

# 2. Attendi che il peering sia Active
while true; do
  STATUS=$(acloud network vpcpeering get $VPC_ID $PEERING_ID | grep "Status:" | awk '{print $2}')
  if [ "$STATUS" = "Active" ]; then
    break
  fi
  echo "Waiting for VPC peering to become Active... (current: $STATUS)"
  sleep 5
done

# 3. Crea route per subnet diverse
acloud network vpcpeeringroute create $VPC_ID $PEERING_ID \
  --name "subnet-1-route" \
  --local-network "10.0.1.0/24" \
  --remote-network "10.1.1.0/24"

acloud network vpcpeeringroute create $VPC_ID $PEERING_ID \
  --name "subnet-2-route" \
  --local-network "10.0.2.0/24" \
  --remote-network "10.1.2.0/24" \
  --billing-period Month

# 4. Elenca tutte le route
acloud network vpcpeeringroute list $VPC_ID $PEERING_ID
```

## Best Practices

1. **Usa Nomi Descrittivi**
   ```bash
   --name "vpc1-subnet1-to-vpc2-subnet1"
   --name "production-peering-route"
   ```

2. **Tagga le Tue Route**
   ```bash
   --tags "vpc,peering,production"
   --tags "vpc,peering,development"
   ```

3. **Pianifica Mapping di Rete**
   - Assicurati che le reti locali e remote non si sovrappongano
   - Usa convenzioni di denominazione chiare per l'identificazione delle route
   - Documenta i mapping di rete per riferimento futuro

4. **Scegli Periodo di Fatturazione Appropriato**
   - Usa **Hour** per route temporanee o di test
   - Usa **Month** per route di produzione con utilizzo variabile
   - Usa **Year** per route stabili a lungo termine (miglior risparmio)

5. **Attendi lo Stato Active**
   ```bash
   # Controlla lo stato prima di aggiornare
   acloud network vpcpeeringroute get <vpc-id> <peering-id> <route-id>
   # Assicurati che lo Status sia "Active"
   acloud network vpcpeeringroute update <vpc-id> <peering-id> <route-id> --name "new-name"
   ```

## Risoluzione dei Problemi

### "Cannot update VPC peering route while in InCreation state"

**Problema:** Tentativo di aggiornare una VPC peering route che non ha finito di crearsi.

**Soluzione:**
```bash
# Controlla lo stato corrente
acloud network vpcpeeringroute get <vpc-id> <peering-id> <route-id>

# Attendi che lo Status diventi "Active"
# Poi riprova l'aggiornamento
acloud network vpcpeeringroute update <vpc-id> <peering-id> <route-id> --name "new-name"
```

### "Error: at least one field must be provided for update"

**Problema:** Comando update chiamato senza modifiche.

**Soluzione:**
```bash
# Fornisci almeno un campo da aggiornare
acloud network vpcpeeringroute update <vpc-id> <peering-id> <route-id> --name "new-name"
# oppure
acloud network vpcpeeringroute update <vpc-id> <peering-id> <route-id> --tags tag1,tag2
```

## Comandi Correlati

- [VPC Peering](vpcpeering.md) - Gestisci connessioni VPC peering
- [VPC](vpc.md) - Gestisci VPC
