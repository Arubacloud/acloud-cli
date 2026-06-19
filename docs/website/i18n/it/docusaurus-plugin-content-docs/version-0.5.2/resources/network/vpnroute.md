# VPN Tunnel Route

Le VPN Tunnel Route definiscono regole di routing per il traffico che scorre attraverso i VPN tunnel in Aruba Cloud. Queste route controllano come il traffico di rete viene diretto tra la tua rete on-premises e il tuo VPC tramite un VPN tunnel specificando subnet cloud e subnet on-premises CIDR.

## Comandi

### Elenca VPN Tunnel Route

Elenca tutte le route per un VPN tunnel specifico.

```bash
acloud network vpnroute list <vpn-tunnel-id> [flags]
```

**Argomenti:**
- `vpn-tunnel-id` - L'ID del VPN tunnel

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud network vpnroute list 1234567890abcdef
```

**Output:**
```
NAME            ID                        CLOUD SUBNET      ONPREM SUBNET     STATUS
route-1         1234567890abcdef123456   10.0.1.0/24       192.168.1.0/24    Active
route-2         1234567890abcdef123457   10.0.2.0/24       192.168.2.0/24    Active
```

### Ottieni Dettagli VPN Tunnel Route

Ottieni informazioni dettagliate su una VPN tunnel route specifica.

```bash
acloud network vpnroute get <vpn-tunnel-id> <route-id> [flags]
```

**Argomenti:**
- `vpn-tunnel-id` - L'ID del VPN tunnel
- `route-id` - L'ID della route (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud network vpnroute get 1234567890abcdef 1234567890abcdef123456
```

**Output:**
```
VPN Route Details:
==================
ID:              1234567890abcdef123456
URI:             /projects/.../vpnroutes/1234567890abcdef123456
Name:            route-1
Region:          ITBG-Bergamo
Cloud Subnet:    10.0.1.0/24
OnPrem Subnet:   192.168.1.0/24
Creation Date:   06-08-2025 07:44:52
Created By:      aru-297647
Tags:            [vpn,route,production]
Status:          Active
```

### Crea VPN Tunnel Route

Crea una nuova route per un VPN tunnel.

```bash
acloud network vpnroute create <vpn-tunnel-id> [flags]
```

**Flag Richiesti:**
- `--name string` - Nome VPN Route
- `--region string` - Codice regione (es. ITBG-Bergamo)
- `--cloud-subnet string` - CIDR della subnet cloud
- `--onprem-subnet string` - CIDR della subnet on-prem

**Flag Opzionali:**
- `--tags strings` - Tag per la VPN route (separati da virgola)
- `--project-id string` - ID progetto (usa il contesto se non specificato)
- `-v, --verbose` - Mostra informazioni di debug dettagliate

**Esempi:**
```bash
# Crea una VPN route base
acloud network vpnroute create 1234567890abcdef \
  --name "route-1" \
  --region ITBG-Bergamo \
  --cloud-subnet "10.0.1.0/24" \
  --onprem-subnet "192.168.1.0/24"

# Crea VPN route con tag
acloud network vpnroute create 1234567890abcdef \
  --name "production-route" \
  --region ITBG-Bergamo \
  --cloud-subnet "10.0.2.0/24" \
  --onprem-subnet "192.168.2.0/24" \
  --tags "vpn,production,network"
```

**Output:**
```
NAME            ID                        CLOUD SUBNET      ONPREM SUBNET     STATUS
route-1         1234567890abcdef123456   10.0.1.0/24       192.168.1.0/24    Active
```

**Note:**
- La VPN route sarà inizialmente in stato **InCreation**
- Usa `acloud network vpnroute get` per controllare quando diventa **Active**

### Aggiorna VPN Tunnel Route

Aggiorna le proprietà di una VPN tunnel route esistente.

```bash
acloud network vpnroute update <vpn-tunnel-id> <route-id> [flags]
```

**Argomenti:**
- `vpn-tunnel-id` - L'ID del VPN tunnel
- `route-id` - L'ID della route (supporta auto-completamento)

**Flag:**
- `--name string` - Nuovo nome per la VPN route
- `--tags strings` - Nuovi tag per la VPN route (separati da virgola)
- `--cloud-subnet string` - CIDR della subnet cloud
- `--onprem-subnet string` - CIDR della subnet on-prem
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Nota:** Almeno un campo deve essere fornito per l'aggiornamento.

**Esempi:**
```bash
# Aggiorna nome VPN route
acloud network vpnroute update 1234567890abcdef 1234567890abcdef123456 \
  --name "updated-route-1"

# Aggiorna subnet cloud
acloud network vpnroute update 1234567890abcdef 1234567890abcdef123456 \
  --cloud-subnet "10.0.3.0/24"

# Aggiorna più campi
acloud network vpnroute update 1234567890abcdef 1234567890abcdef123456 \
  --name "production-route" \
  --cloud-subnet "10.0.2.0/24" \
  --onprem-subnet "192.168.2.0/24" \
  --tags "vpn,production,updated"
```

**Output:**
```
NAME            ID                        CLOUD SUBNET      ONPREM SUBNET     STATUS
production-route 1234567890abcdef123456   10.0.2.0/24       192.168.2.0/24    Active
```

**Restrizioni:**
- Non è possibile aggiornare VPN route in stato **InCreation**
- Attendi che la VPN route raggiunga lo stato **Active** prima di aggiornare

### Elimina VPN Tunnel Route

Elimina una VPN tunnel route.

```bash
acloud network vpnroute delete <vpn-tunnel-id> <route-id> [flags]
```

**Argomenti:**
- `vpn-tunnel-id` - L'ID del VPN tunnel
- `route-id` - L'ID della route (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)
- `-y, --yes` - Salta il prompt di conferma

**Esempi:**
```bash
# Elimina con prompt di conferma
acloud network vpnroute delete 1234567890abcdef 1234567890abcdef123456

# Elimina senza conferma
acloud network vpnroute delete 1234567890abcdef 1234567890abcdef123456 --yes
```

**Prompt di Conferma:**
```
Are you sure you want to delete VPN route 1234567890abcdef123456? This action cannot be undone.
Type 'yes' to confirm: yes
```

**Output:**
```
ID                              STATUS
1234567890abcdef123456         deleted
```

**Note:**
- L'eliminazione non può essere annullata
- Assicurati che il VPN tunnel non dipenda dalla route prima dell'eliminazione

## Auto-completamento Shell

I comandi VPN Tunnel Route supportano auto-completamento intelligente per gli ID route:

```bash
# Abilita completamento (bash)
source <(acloud completion bash)

# Digita il comando e premi TAB per vedere gli ID route disponibili
acloud network vpnroute get <vpn-tunnel-id> <TAB>
acloud network vpnroute update <vpn-tunnel-id> <TAB>
acloud network vpnroute delete <vpn-tunnel-id> <TAB>
```

L'auto-completamento mostra gli ID route con i loro nomi:
```
1234567890abcdef123456    route-1
1234567890abcdef123457    route-2
```

## Proprietà VPN Route

### Subnet Cloud

La subnet cloud CIDR rappresenta l'intervallo di rete nel tuo VPC che dovrebbe essere accessibile tramite il VPN tunnel.

**Esempi:**
- `10.0.1.0/24` - Subnet specifica nel VPC
- `10.0.0.0/16` - Intero intervallo di rete VPC

### Subnet On-Premises

La subnet on-premises CIDR rappresenta l'intervallo di rete nella tua infrastruttura on-premises che dovrebbe essere accessibile tramite il VPN tunnel.

**Esempi:**
- `192.168.1.0/24` - Subnet on-premises specifica
- `192.168.0.0/16` - Intero intervallo di rete on-premises

## Stati VPN Route

Le VPN route possono essere nei seguenti stati:

| Stato | Descrizione | Posso Aggiornare? | Posso Eliminare? |
|-------|-------------|-------------------|------------------|
| InCreation | La VPN route è in fase di creazione | ❌ No | ❌ No |
| Active | La VPN route è pronta per l'uso | ✅ Sì | ✅ Sì |

## Workflow Comuni

### Configurazione VPN Route

```bash
# 1. Crea VPN tunnel (se non esiste)
VPN_TUNNEL_ID=$(acloud network vpntunnel create \
  --name "prod-vpn-tunnel" \
  --region ITBG-Bergamo | grep "ID:" | awk '{print $2}')

# 2. Attendi che il tunnel sia Active
while true; do
  STATUS=$(acloud network vpntunnel get $VPN_TUNNEL_ID | grep "Status:" | awk '{print $2}')
  if [ "$STATUS" = "Active" ]; then
    break
  fi
  echo "Waiting for VPN tunnel to become Active... (current: $STATUS)"
  sleep 5
done

# 3. Crea route per subnet diverse
acloud network vpnroute create $VPN_TUNNEL_ID \
  --name "vpc-subnet-1" \
  --region ITBG-Bergamo \
  --cloud-subnet "10.0.1.0/24" \
  --onprem-subnet "192.168.1.0/24"

acloud network vpnroute create $VPN_TUNNEL_ID \
  --name "vpc-subnet-2" \
  --region ITBG-Bergamo \
  --cloud-subnet "10.0.2.0/24" \
  --onprem-subnet "192.168.2.0/24"

# 4. Elenca tutte le route
acloud network vpnroute list $VPN_TUNNEL_ID
```

## Best Practices

1. **Usa Nomi Descrittivi**
   ```bash
   --name "vpc-subnet-1-to-onprem"
   --name "production-vpn-route"
   ```

2. **Tagga le Tue Route**
   ```bash
   --tags "vpn,production,network"
   --tags "vpn,development,test"
   ```

3. **Pianifica Mapping Subnet**
   - Assicurati che le subnet cloud e on-premises non si sovrappongano
   - Usa convenzioni di denominazione chiare per l'identificazione delle route

4. **Attendi lo Stato Active**
   ```bash
   # Controlla lo stato prima di aggiornare
   acloud network vpnroute get <vpn-tunnel-id> <route-id>
   # Assicurati che lo Status sia "Active"
   acloud network vpnroute update <vpn-tunnel-id> <route-id> --name "new-name"
   ```

## Risoluzione dei Problemi

### "Cannot update VPN route while in InCreation state"

**Problema:** Tentativo di aggiornare una VPN route che non ha finito di crearsi.

**Soluzione:**
```bash
# Controlla lo stato corrente
acloud network vpnroute get <vpn-tunnel-id> <route-id>

# Attendi che lo Status diventi "Active"
# Poi riprova l'aggiornamento
acloud network vpnroute update <vpn-tunnel-id> <route-id> --name "new-name"
```

### "Error: at least one field must be provided for update"

**Problema:** Comando update chiamato senza modifiche.

**Soluzione:**
```bash
# Fornisci almeno un campo da aggiornare
acloud network vpnroute update <vpn-tunnel-id> <route-id> --name "new-name"
# oppure
acloud network vpnroute update <vpn-tunnel-id> <route-id> --tags tag1,tag2
```

## Comandi Correlati

- [VPN Tunnel](vpntunnel.md) - Gestisci VPN tunnel
- [VPC](vpc.md) - Gestisci VPC
