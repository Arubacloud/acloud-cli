# VPN Tunnel

I VPN Tunnel in Aruba Cloud forniscono connessioni sicure e crittografate tra il tuo VPC e reti remote (come data center on-premises o altri cloud). Puoi gestire tunnel VPN site-to-site e client, configurare protocolli e controllare il ciclo di vita del tunnel.

## Comandi

### Elenca VPN Tunnel

Elenca tutti i VPN tunnel nel tuo progetto.

```bash
acloud network vpntunnel list [flags]
```

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud network vpntunnel list
acloud network vpntunnel list --project-id 68398923fb2cb026400d4d31
```

**Output:**
```
NAME         ID                        REGION        TYPE         STATUS
vpn-prod     1234567890abcdef          ITBG-Bergamo  Site-To-Site Active
```

### Ottieni Dettagli VPN Tunnel

Ottieni informazioni dettagliate su un VPN tunnel specifico.

```bash
acloud network vpntunnel get <vpn-tunnel-id> [flags]
```

**Argomenti:**
- `vpn-tunnel-id` - L'ID del VPN tunnel (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud network vpntunnel get 1234567890abcdef
```

**Output:**
```
VPN Tunnel Details:
===================
ID:              1234567890abcdef
Name:            vpn-prod
Region:          ITBG-Bergamo
Type:            Site-To-Site
Status:          Active
Peer IP:         203.0.113.1
VPC:             /.../vpcs/689307f4745108d3c6343b5a
Subnet CIDR:     10.0.1.0/24
Creation Date:   06-08-2025 07:44:52
Tags:            [production vpn]
```

### Crea VPN Tunnel

Crea un nuovo VPN tunnel.

```bash
acloud network vpntunnel create [flags]
```

**Flag Richiesti:**
- `--name string` - Nome per il VPN tunnel
- `--region string` - Codice regione (es. ITBG-Bergamo)
- `--peer-ip string` - Indirizzo IP pubblico del client peer
- `--vpc-uri string` - URI VPC
- `--subnet-cidr string` - Subnet CIDR (es. 10.0.1.0/24)

**Flag Opzionali:**
- `--tags strings` - Tag per il VPN tunnel (separati da virgola)
- `--vpn-type string` - Tipo VPN (default: Site-To-Site)
- `--protocol string` - Protocollo VPN (default: ikev2)
- `--billing-period string` - Periodo di fatturazione: Hour, Month, Year

**Esempio:**
```bash
acloud network vpntunnel create --name vpn-prod --region ITBG-Bergamo --peer-ip 203.0.113.1 --vpc-uri /projects/.../vpcs/689307f4745108d3c6343b5a --subnet-cidr 10.0.1.0/24
```

**Output:**
```
VPN Tunnel created successfully!
ID:      1234567890abcdef
Name:    vpn-prod
Region:  ITBG-Bergamo
```

### Aggiorna VPN Tunnel

Aggiorna il nome o i tag di un VPN tunnel esistente.

```bash
acloud network vpntunnel update <vpn-tunnel-id> [flags]
```

**Argomenti:**
- `vpn-tunnel-id` - L'ID del VPN tunnel

**Flag:**
- `--name string` - Nuovo nome per il VPN tunnel
- `--tags strings` - Nuovi tag (separati da virgola)

**Esempio:**
```bash
acloud network vpntunnel update 1234567890abcdef --name "new-vpn-name"
```

**Output:**
```
VPN Tunnel updated successfully!
ID:      1234567890abcdef
Name:    new-vpn-name
```

### Elimina VPN Tunnel

Elimina un VPN tunnel.

```bash
acloud network vpntunnel delete <vpn-tunnel-id> [flags]
```

**Argomenti:**
- `vpn-tunnel-id` - L'ID del VPN tunnel

**Flag:**
- `-y, --yes` - Salta il prompt di conferma

**Esempio:**
```bash
acloud network vpntunnel delete 1234567890abcdef --yes
```

**Output:**
```
VPN Tunnel 1234567890abcdef deleted successfully!
```

## Auto-completamento Shell

I comandi VPN Tunnel supportano auto-completamento per ID VPN tunnel.

## Best Practices

- Usa nomi e tag descrittivi per i VPN tunnel.
- Rivedi regolarmente lo stato e la configurazione del tunnel.

## Risoluzione dei Problemi

- Assicurati che il VPC e la subnet siano **Active** prima di creare un tunnel.
- Controlla l'IP peer e il CIDR per correttezza.

## Comandi Correlati

- [VPN Tunnel Route](vpnroute.md) - Gestisci route per VPN tunnel
- [VPC](vpc.md) - Gestisci VPC
