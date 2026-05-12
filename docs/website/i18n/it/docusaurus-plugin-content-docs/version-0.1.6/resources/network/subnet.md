# Subnet

Le subnet ti permettono di segmentare la rete VPC in sezioni più piccole e isolate. Ogni subnet può avere il proprio blocco CIDR e può essere usata per organizzare risorse, controllare il routing e applicare politiche di sicurezza all'interno di un VPC.

## Tipi di Subnet

Le subnet possono essere create in due tipi:

- **Subnet Basic**: Blocco CIDR assegnato automaticamente dal sistema. Nessuna configurazione DHCP richiesta.
- **Subnet Advanced**: Blocco CIDR personalizzato specificato dall'utente. Richiede configurazione DHCP con il flag `--dhcp-enabled`. Supporta opzionalmente route e server DNS personalizzati.

## Comandi

### Elenca Subnet

Elenca tutte le subnet in un VPC.

```bash
acloud network subnet list <vpc-id>
```

**Argomenti:**
- `vpc-id` - L'ID del VPC

**Esempio:**
```bash
acloud network subnet list 689307f4745108d3c6343b5a
```

**Output:**
```
NAME         ID                        CIDR           STATUS
subnet-1     1234567890abcdef          10.0.1.0/24    Active
subnet-2     0987654321fedcba          10.0.2.0/24    Active
```

### Ottieni Dettagli Subnet

Ottieni dettagli su una subnet specifica.

```bash
acloud network subnet get <vpc-id> <subnet-id>
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `subnet-id` - L'ID della subnet

**Esempio:**
```bash
acloud network subnet get 689307f4745108d3c6343b5a 1234567890abcdef
```

**Output:**
```
Subnet Details:
===============
ID:            1234567890abcdef
Name:          subnet-1
Type:          Advanced
CIDR:          10.0.1.0/24
DHCP Enabled:  true
DHCP Routes:
  - 0.0.0.0/0 -> 10.0.0.1
DHCP DNS:      [8.8.8.8 8.8.4.4]
Status:        Active
```

### Crea Subnet

Crea una nuova subnet in un VPC.

```bash
acloud network subnet create <vpc-id> --name <name> --region <region> [flags]
```

**Flag Richiesti:**
- `--name string` - Nome per la subnet
- `--region string` - Regione per la subnet

**Flag Opzionali:**
- `--cidr string` - Blocco CIDR per la subnet (se fornito, crea una subnet di tipo Advanced)
- `--tags stringSlice` - Tag della subnet (separati da virgola o più flag)
- `--dhcp-enabled` - Abilita DHCP per subnet di tipo Advanced (richiesto quando `--cidr` è fornito)
- `--dhcp-routes stringSlice` - Route DHCP per subnet di tipo Advanced (formato: `destination:gateway`, es. `0.0.0.0/0:10.0.0.1`)
- `--dhcp-dns stringSlice` - Server DNS DHCP per subnet di tipo Advanced (es. `8.8.8.8`, `8.8.4.4`)

**Crea Subnet Basic (CIDR assegnato automaticamente):**
```bash
acloud network subnet create 689307f4745108d3c6343b5a --name subnet-1 --region "ITBG-Bergamo"
```

**Crea Subnet Advanced (CIDR personalizzato con DHCP):**
```bash
acloud network subnet create 689307f4745108d3c6343b5a \
  --name subnet-1 \
  --region "ITBG-Bergamo" \
  --cidr 10.0.1.0/24 \
  --dhcp-enabled \
  --dhcp-routes "0.0.0.0/0:10.0.0.1" \
  --dhcp-dns "8.8.8.8" "8.8.4.4"
```

**Output:**
```
NAME         ID                        REGION          CIDR           STATUS
subnet-1     1234567890abcdef          ITBG-Bergamo    10.0.1.0/24    Active
```

### Aggiorna Subnet

Aggiorna il nome, CIDR, tag o configurazione DHCP di una subnet esistente.

```bash
acloud network subnet update <vpc-id> <subnet-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `subnet-id` - L'ID della subnet

**Flag:**
- `--name string` - Nuovo nome per la subnet
- `--cidr string` - Nuovo blocco CIDR
- `--tags stringSlice` - Tag della subnet (separati da virgola o più flag)
- `--dhcp-enabled` - Abilita/disabilita DHCP per subnet di tipo Advanced
- `--dhcp-routes stringSlice` - Route DHCP per subnet di tipo Advanced (formato: `destination:gateway`)
- `--dhcp-dns stringSlice` - Server DNS DHCP per subnet di tipo Advanced

**Esempi:**

Aggiorna il nome della subnet:
```bash
acloud network subnet update 689307f4745108d3c6343b5a 1234567890abcdef --name new-subnet-name
```

Aggiorna le route DHCP per subnet Advanced:
```bash
acloud network subnet update 689307f4745108d3c6343b5a 1234567890abcdef \
  --dhcp-routes "192.168.1.0/24:10.0.0.1" "0.0.0.0/0:10.0.0.1"
```

Aggiorna i server DNS DHCP:
```bash
acloud network subnet update 689307f4745108d3c6343b5a 1234567890abcdef \
  --dhcp-dns "1.1.1.1" "1.0.0.1"
```

**Output:**
```
NAME              ID                        CIDR           STATUS
new-subnet-name   1234567890abcdef          10.0.1.0/24    Active
```

### Elimina Subnet

Elimina una subnet da un VPC.

```bash
acloud network subnet delete <vpc-id> <subnet-id>
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `subnet-id` - L'ID della subnet

**Esempio:**
```bash
acloud network subnet delete 689307f4745108d3c6343b5a 1234567890abcdef
```

**Output:**
```
Subnet 1234567890abcdef deleted successfully!
```

## Auto-completamento Shell

I comandi subnet supportano auto-completamento per ID VPC e ID subnet.

## Best Practices

- Usa nomi descrittivi per le subnet in base al loro scopo.
- Evita blocchi CIDR sovrapposti tra subnet.
- Per subnet Advanced, abilita sempre DHCP con `--dhcp-enabled` quando fornisci un CIDR personalizzato.
- Configura route e server DNS DHCP appropriati per subnet Advanced per garantire la connettività di rete corretta.
- Usa subnet Basic quando non hai bisogno di configurazione CIDR personalizzata.

## Risoluzione dei Problemi

- Assicurati che il VPC sia **Active** prima di creare subnet.
- Controlla conflitti CIDR quando aggiungi o aggiorni subnet.

## Comandi Correlati

- [VPC](vpc.md) - Gestisci VPC
- [Security Group](securitygroup.md) - Gestisci security group
