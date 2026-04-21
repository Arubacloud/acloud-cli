# Security Group

I Security Group fungono da firewall virtuali per le tue risorse all'interno di un VPC. Controllano il traffico in entrata e in uscita a livello di istanza, permettendoti di definire regole basate su protocolli, porte e intervalli IP di origine/destinazione.

## Comandi

### Elenca Security Group

Elenca tutti i security group in un VPC.

```bash
acloud network securitygroup list <vpc-id>
```

**Argomenti:**
- `vpc-id` - L'ID del VPC

**Esempio:**
```bash
acloud network securitygroup list 689307f4745108d3c6343b5a
```

**Output:**
```
NAME         ID                        DESCRIPTION         STATUS
web-sg       1234567890abcdef          Allow web traffic   Active
```

### Ottieni Dettagli Security Group

Ottieni dettagli su un security group specifico.

```bash
acloud network securitygroup get <vpc-id> <security-group-id>
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `security-group-id` - L'ID del security group

**Esempio:**
```bash
acloud network securitygroup get 689307f4745108d3c6343b5a 1234567890abcdef
```

**Output:**
```
Security Group Details:
======================
ID:            1234567890abcdef
Name:          web-sg
Description:   Allow web traffic
Status:        Active
Rules:         3
```

### Crea Security Group

Crea un nuovo security group in un VPC.

```bash
acloud network securitygroup create <vpc-id> --name <name> --description <desc>
```

**Flag Richiesti:**
- `--name string` - Nome per il security group
- `--description string` - Descrizione del security group

**Esempio:**
```bash
acloud network securitygroup create 689307f4745108d3c6343b5a --name web-sg --description "Allow web traffic"
```

**Output:**
```
Security Group created successfully!
ID:          1234567890abcdef
Name:        web-sg
Description: Allow web traffic
```

### Aggiorna Security Group

Aggiorna il nome o la descrizione di un security group esistente.

```bash
acloud network securitygroup update <vpc-id> <security-group-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `security-group-id` - L'ID del security group

**Flag:**
- `--name string` - Nuovo nome per il security group
- `--description string` - Nuova descrizione

**Esempio:**
```bash
acloud network securitygroup update 689307f4745108d3c6343b5a 1234567890abcdef --name "new-sg-name"
```

**Output:**
```
Security Group updated successfully!
ID:          1234567890abcdef
Name:        new-sg-name
```

### Elimina Security Group

Elimina un security group da un VPC.

```bash
acloud network securitygroup delete <vpc-id> <security-group-id>
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `security-group-id` - L'ID del security group

**Esempio:**
```bash
acloud network securitygroup delete 689307f4745108d3c6343b5a 1234567890abcdef
```

**Output:**
```
Security Group 1234567890abcdef deleted successfully!
```

## Auto-completamento Shell

I comandi security group supportano auto-completamento per ID VPC e ID security group.

## Best Practices

- Usa nomi e descrizioni descrittivi per i security group.
- Rivedi e aggiorna regolarmente le regole dei security group.

## Risoluzione dei Problemi

- Assicurati che il VPC sia **Active** prima di creare security group.
- Controlla regole conflittuali o eccessivamente permissive.

## Comandi Correlati

- [Subnet](subnet.md) - Gestisci subnet
- [VPC](vpc.md) - Gestisci VPC
