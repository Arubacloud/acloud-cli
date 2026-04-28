# Risorse di Rete

La categoria `network` fornisce comandi per gestire risorse di rete in Aruba Cloud.

## Risorse Disponibili

### [VPC](network/vpc.md)

I Virtual Private Cloud (VPC) forniscono ambienti di rete isolati per le tue risorse.

**Comandi Rapidi:**
```bash
# Elenca tutti i VPC
acloud network vpc list

# Ottieni i dettagli del VPC
acloud network vpc get <vpc-id>

# Crea un VPC
acloud network vpc create --name "my-vpc" --region ITBG-Bergamo

# Aggiorna un VPC
acloud network vpc update <vpc-id> --name "new-name" --tags tag1,tag2

# Elimina un VPC
acloud network vpc delete <vpc-id>
```

### [Elastic IP](network/elasticip.md)

Gli Elastic IP sono indirizzi IP pubblici statici che possono essere assegnati alle tue risorse.

**Comandi Rapidi:**
```bash
# Elenca tutti gli Elastic IP
acloud network elasticip list

# Ottieni i dettagli dell'Elastic IP
acloud network elasticip get <eip-id>

# Crea un Elastic IP
acloud network elasticip create --name "my-eip" --region ITBG-Bergamo --billing-period Hour

# Aggiorna un Elastic IP
acloud network elasticip update <eip-id> --name "new-name" --tags tag1,tag2

# Elimina un Elastic IP
acloud network elasticip delete <eip-id>
```

### [Security Rule](network/securityrule.md)

Le Security Rule definiscono le regole firewall per i Security Group. Specificano direzione (Ingress/Egress), protocollo, porte e target (indirizzi IP o Security Group).

**Comandi Rapidi:**
```bash
# Elenca tutte le security rule per un security group
acloud network securityrule list <vpc-id> <securitygroup-id>

# Ottieni i dettagli della security rule
acloud network securityrule get <vpc-id> <securitygroup-id> <securityrule-id>

# Crea una security rule
acloud network securityrule create <vpc-id> <securitygroup-id> \
  --name "allow-http" --region ITBG-Bergamo \
  --direction Ingress --protocol TCP --port 80 \
  --target-kind Ip --target-value "0.0.0.0/0"

# Aggiorna una security rule
acloud network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> \
  --port 443

# Elimina una security rule
acloud network securityrule delete <vpc-id> <securitygroup-id> <securityrule-id>
```

### [Load Balancer](network/loadbalancer.md)

I Load Balancer distribuiscono il traffico tra più risorse. Nota: I Load Balancer sono in sola lettura tramite la CLI.

**Comandi Rapidi:**
```bash
# Elenca tutti i Load Balancer
acloud network loadbalancer list

# Ottieni i dettagli del Load Balancer
acloud network loadbalancer get <lb-id>
```

## Pattern Comuni

### Utilizzo del Contesto Progetto

Tutti i comandi di rete supportano il contesto progetto. Imposta un contesto per evitare di specificare `--project-id` ogni volta:

```bash
# Imposta il contesto corrente
acloud context use my-project

# Ora puoi eseguire comandi senza --project-id
acloud network vpc list
acloud network elasticip list
```

### Auto-completamento Shell

La CLI fornisce auto-completamento intelligente per gli ID risorse:

```bash
# Digita il comando e premi TAB per vedere gli ID VPC disponibili
acloud network vpc get <TAB>

# Digita il comando e premi TAB per vedere gli ID Elastic IP disponibili
acloud network elasticip update <TAB>
```

### Tagging delle Risorse

Usa i tag per organizzare e categorizzare le tue risorse di rete:

```bash
# Crea con tag
acloud network vpc create --name "prod-vpc" --region ITBG-Bergamo --tags production,critical

# Aggiorna tag
acloud network vpc update <vpc-id> --tags production,updated,network
acloud network elasticip update <eip-id> --tags production,public
```

### Risorse Regionali

Le risorse di rete sono regionali. Regioni supportate:
- `ITBG-Bergamo` - Italia, Bergamo

Specifica la regione quando crei risorse:

```bash
acloud network vpc create --name "my-vpc" --region ITBG-Bergamo
acloud network elasticip create --name "my-eip" --region ITBG-Bergamo --billing-period Hour
```

## Ciclo di Vita delle Risorse

### Ciclo di Vita VPC
1. **InCreation** - Il VPC è in fase di creazione
2. **Active** - Il VPC è pronto per l'uso
3. **Deleting** - Il VPC è in fase di eliminazione

### Ciclo di Vita Elastic IP
1. **InCreation** - L'Elastic IP è in fase di creazione
2. **Active** - L'Elastic IP è pronto per l'uso
3. **Deleting** - L'Elastic IP è in fase di eliminazione

## Prossimi Passi

- [Guida VPC](network/vpc.md)
- [Guida Elastic IP](network/elasticip.md)
- [Guida Security Rule](network/securityrule.md)
- [Guida Load Balancer](network/loadbalancer.md)

