# VPC Peering

Il VPC Peering ti permette di collegare due Virtual Private Cloud (VPC) in Aruba Cloud, abilitando traffico di rete privato tra di essi. Le connessioni di peering sono utili per condividere risorse o abilitare comunicazione tra ambienti o progetti diversi.

## Comandi

### Elenca VPC Peering

Elenca tutte le connessioni VPC peering per un VPC.

```bash
acloud network vpcpeering list <vpc-id>
```

**Argomenti:**
- `vpc-id` - L'ID del VPC

**Esempio:**
```bash
acloud network vpcpeering list 689307f4745108d3c6343b5a
```

**Output:**
```
NAME         ID                        PEER VPC                  REGION        STATUS
prod-peer    6949666e4d0cdc87949b7204  /.../vpcs/69485a584d0cdc87949b6ff8  ITBG-Bergamo  Active
```

### Ottieni Dettagli VPC Peering

Ottieni dettagli su una connessione VPC peering specifica.

```bash
acloud network vpcpeering get <vpc-id> <peering-id>
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `peering-id` - L'ID della connessione VPC peering

**Esempio:**
```bash
acloud network vpcpeering get 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204
```

**Output:**
```
VPC Peering Details:
====================
ID:              6949666e4d0cdc87949b7204
Name:            prod-peer
Peer VPC:        /.../vpcs/69485a584d0cdc87949b6ff8
Region:          ITBG-Bergamo
Creation Date:   06-08-2025 07:44:52
Created By:      aru-297647
Tags:            [production peering]
Status:          Active
```

### Crea VPC Peering

Crea una nuova connessione VPC peering.

```bash
acloud network vpcpeering create <vpc-id> --peer-vpc-id <peer-vpc-id> --name <name> --region <region>
```

**Flag Richiesti:**
- `--peer-vpc-id string` - L'URI del VPC peer
- `--name string` - Nome per la connessione peering
- `--region string` - Codice regione (es. ITBG-Bergamo)

**Flag Opzionali:**
- `--tags strings` - Tag per il peering (separati da virgola)

**Esempio:**
```bash
acloud network vpcpeering create 689307f4745108d3c6343b5a --peer-vpc-id /projects/.../vpcs/69485a584d0cdc87949b6ff8 --name prod-peer --region ITBG-Bergamo
```

**Output:**
```
VPC Peering created successfully!
ID:      6949666e4d0cdc87949b7204
Name:    prod-peer
Peer VPC: /.../vpcs/69485a584d0cdc87949b6ff8
Region:  ITBG-Bergamo
```

### Aggiorna VPC Peering

Aggiorna una connessione VPC peering esistente.

```bash
acloud network vpcpeering update <vpc-id> <peering-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `peering-id` - L'ID della connessione VPC peering

**Flag:**
- `--name string` - Nuovo nome per il peering
- `--tags strings` - Nuovi tag per il peering (separati da virgola)

**Esempio:**
```bash
acloud network vpcpeering update 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204 --name "new-peer-name"
```

**Output:**
```
VPC Peering updated successfully!
ID:      6949666e4d0cdc87949b7204
Name:    new-peer-name
```

### Elimina VPC Peering

Elimina una connessione VPC peering.

```bash
acloud network vpcpeering delete <vpc-id> <peering-id>
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `peering-id` - L'ID della connessione VPC peering

**Esempio:**
```bash
acloud network vpcpeering delete 689307f4745108d3c6343b5a 6949666e4d0cdc87949b7204
```

**Output:**
```
VPC Peering 6949666e4d0cdc87949b7204 deleted successfully!
```

## Auto-completamento Shell

I comandi VPC Peering supportano auto-completamento per ID VPC e ID peering.

## Best Practices

- Usa nomi descrittivi per le connessioni peering.
- Tagga i peering per ambiente o scopo.

## Risoluzione dei Problemi

- Assicurati che entrambi i VPC siano nella stessa regione e progetto se richiesto.
- Controlla lo stato del peering prima di aggiornare o eliminare.

## Comandi Correlati

- [VPC Peering Route](vpcpeeringroute.md) - Gestisci route per VPC peering
- [VPC](vpc.md) - Gestisci VPC
