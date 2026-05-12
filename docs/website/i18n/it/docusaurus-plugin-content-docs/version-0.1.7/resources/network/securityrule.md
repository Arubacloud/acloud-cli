# Security Rule

Le Security Rule definiscono le regole firewall per i Security Group all'interno di un VPC. Controllano il traffico in entrata e in uscita specificando direzione, protocollo, porte e target (indirizzi IP o altri Security Group).

## Comandi

### Elenca Security Rule

Elenca tutte le security rule per un security group specifico.

```bash
acloud network securityrule list <vpc-id> <securitygroup-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `securitygroup-id` - L'ID del security group

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud network securityrule list 689307f4745108d3c6343b5a 1234567890abcdef
```

**Output:**
```
NAME            ID                        DIRECTION    PROTOCOL    PORT    TARGET                    STATUS
allow-http      1234567890abcdef123456   Ingress      TCP         80      Ip:0.0.0.0/0             Active
allow-https     1234567890abcdef123457   Ingress      TCP         443     Ip:0.0.0.0/0             Active
allow-ssh       1234567890abcdef123458   Ingress      TCP         22      Ip:10.0.0.0/8            Active
```

### Ottieni Dettagli Security Rule

Ottieni informazioni dettagliate su una security rule specifica.

```bash
acloud network securityrule get <vpc-id> <securitygroup-id> <securityrule-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `securitygroup-id` - L'ID del security group
- `securityrule-id` - L'ID della security rule (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud network securityrule get 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456
```

**Output:**
```
Security Rule Details:
=====================
ID:              1234567890abcdef123456
URI:             /projects/.../securityrules/1234567890abcdef123456
Name:            allow-http
Region:          ITBG-Bergamo
Direction:       Ingress
Protocol:        TCP
Port:            80
Target Kind:     Ip
Target Value:    0.0.0.0/0
Creation Date:   06-08-2025 07:44:52
Created By:      aru-297647
Tags:            [web http public]
Status:          Active
```

### Crea Security Rule

Crea una nuova security rule per un security group.

```bash
acloud network securityrule create <vpc-id> <securitygroup-id> [flags]
```

**Flag Richiesti:**
- `--name string` - Nome Security Rule
- `--region string` - Codice regione (es. ITBG-Bergamo)
- `--direction string` - Direzione: Ingress o Egress
- `--protocol string` - Protocollo: ANY, TCP, UDP, ICMP
- `--target-kind string` - Tipo Target: Ip o SecurityGroup
- `--target-value string` - Valore Target: Se kind = Ip, il valore deve essere un indirizzo di rete valido in notazione CIDR (incluso 0.0.0.0/0). Se kind = SecurityGroup, il valore deve essere un URI valido di qualsiasi security group all'interno dello stesso VPC

**Flag Opzionali:**
- `--port string` - Porta: una porta numerica singola, un intervallo di porte o * (richiesto per TCP/UDP)
- `--tags strings` - Tag per la security rule (separati da virgola)
- `--project-id string` - ID progetto (usa il contesto se non specificato)
- `-v, --verbose` - Mostra informazioni di debug dettagliate

**Esempi:**
```bash
# Crea una regola ingress base che permette traffico HTTP
acloud network securityrule create 689307f4745108d3c6343b5a 1234567890abcdef \
  --name "allow-http" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 80 \
  --target-kind Ip \
  --target-value "0.0.0.0/0"

# Crea una regola che permette SSH da una rete specifica
acloud network securityrule create 689307f4745108d3c6343b5a 1234567890abcdef \
  --name "allow-ssh" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 22 \
  --target-kind Ip \
  --target-value "10.0.0.0/8" \
  --tags "ssh,admin,secure"

# Crea una regola che permette tutto il traffico da un altro security group
acloud network securityrule create 689307f4745108d3c6343b5a 1234567890abcdef \
  --name "allow-from-app-sg" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol ANY \
  --target-kind SecurityGroup \
  --target-value "/projects/.../securitygroups/9876543210fedcba"

# Crea una regola egress che permette tutto il traffico in uscita
acloud network securityrule create 689307f4745108d3c6343b5a 1234567890abcdef \
  --name "allow-all-outbound" \
  --region ITBG-Bergamo \
  --direction Egress \
  --protocol ANY \
  --target-kind Ip \
  --target-value "0.0.0.0/0"
```

**Output:**
```
NAME            ID                        DIRECTION    PROTOCOL    PORT    STATUS
allow-http      1234567890abcdef123456   Ingress      TCP         80      Active
```

**Note:**
- La porta è richiesta per i protocolli TCP e UDP
- La porta può essere omessa per i protocolli ANY e ICMP
- La porta può essere un singolo numero (es. "80"), un intervallo (es. "8000-9000"), o "*" per tutte le porte
- La security rule sarà inizialmente in stato **InCreation**
- Usa `acloud network securityrule get` per controllare quando diventa **Active**

### Aggiorna Security Rule

Aggiorna il nome o i tag di una security rule esistente.

**Importante:** Le security rule possono essere aggiornate solo cambiando il nome o i tag. Proprietà come direzione, protocollo, porta e target non possono essere modificate. Per cambiare queste proprietà, devi eliminare e ricreare la security rule.

```bash
acloud network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `securitygroup-id` - L'ID del security group
- `securityrule-id` - L'ID della security rule (supporta auto-completamento)

**Flag:**
- `--name string` - Nuovo nome per la security rule
- `--tags strings` - Nuovi tag per la security rule (separati da virgola)
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Nota:** Almeno un campo (--name o --tags) deve essere fornito per l'aggiornamento.

**Modalità Debug:** Usa il flag globale `--debug` per vedere informazioni dettagliate su richiesta/risposta HTTP durante la risoluzione dei problemi di aggiornamento:

```bash
acloud --debug network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --tags test
```

Questo mostrerà il payload della richiesta e i dettagli completi dell'errore se l'aggiornamento fallisce.

**Esempi:**
```bash
# Aggiorna nome security rule
acloud network securityrule update 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456 \
  --name "allow-http-updated"

# Aggiorna tag
acloud network securityrule update 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456 \
  --tags "web,https,secure"

# Aggiorna sia nome che tag
acloud network securityrule update 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456 \
  --name "allow-https" \
  --tags "web,https,secure"
```

**Output:**
```
NAME            ID                        DIRECTION    PROTOCOL    PORT    STATUS
allow-https     1234567890abcdef123456   Ingress      TCP         443     Active
```

**Restrizioni:**
- Non è possibile aggiornare security rule in stato **InCreation**
- Attendi che la security rule raggiunga lo stato **Active** prima di aggiornare
- Solo nome e tag possono essere aggiornati; le proprietà (direzione, protocollo, porta, target) non possono essere cambiate

### Elimina Security Rule

Elimina una security rule.

```bash
acloud network securityrule delete <vpc-id> <securitygroup-id> <securityrule-id> [flags]
```

**Argomenti:**
- `vpc-id` - L'ID del VPC
- `securitygroup-id` - L'ID del security group
- `securityrule-id` - L'ID della security rule (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)
- `-y, --yes` - Salta il prompt di conferma

**Esempi:**
```bash
# Elimina con prompt di conferma
acloud network securityrule delete 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456

# Elimina senza conferma
acloud network securityrule delete 689307f4745108d3c6343b5a 1234567890abcdef 1234567890abcdef123456 --yes
```

**Prompt di Conferma:**
```
Are you sure you want to delete security rule 1234567890abcdef123456? This action cannot be undone.
Type 'yes' to confirm: yes
```

**Output:**
```
ID                              STATUS
1234567890abcdef123456         deleted
```

**Note:**
- L'eliminazione non può essere annullata
- Assicurati che nessuna risorsa dipenda dalla security rule prima dell'eliminazione

## Auto-completamento Shell

I comandi security rule supportano auto-completamento intelligente per gli ID security rule:

```bash
# Abilita completamento (bash)
source <(acloud completion bash)

# Digita il comando e premi TAB per vedere gli ID security rule disponibili
acloud network securityrule get <vpc-id> <securitygroup-id> <TAB>
acloud network securityrule update <vpc-id> <securitygroup-id> <TAB>
acloud network securityrule delete <vpc-id> <securitygroup-id> <TAB>
```

L'auto-completamento mostra gli ID security rule con i loro nomi:
```
1234567890abcdef123456    allow-http
1234567890abcdef123457    allow-https
1234567890abcdef123458    allow-ssh
```

## Proprietà Security Rule

### Direzione

- **Ingress**: Traffico in entrata (traffico che entra nel security group)
- **Egress**: Traffico in uscita (traffico che esce dal security group)

### Protocollo

- **ANY**: Tutti i protocolli
- **TCP**: Transmission Control Protocol
- **UDP**: User Datagram Protocol
- **ICMP**: Internet Control Message Protocol

### Porta

- **Porta singola**: Un valore numerico (es. "80", "443", "22")
- **Intervallo porte**: Un intervallo di porte (es. "8000-9000")
- **Tutte le porte**: Usa "*" per permettere tutte le porte
- **Nota**: La porta è richiesta per TCP e UDP, opzionale per ANY e ICMP

### Target

Il target specifica l'origine (per Ingress) o la destinazione (per Egress) del traffico:

- **Target IP**: Usa notazione CIDR (es. "0.0.0.0/0" per tutti gli IP, "10.0.0.0/8" per rete privata)
- **Target Security Group**: Usa l'URI di un altro security group all'interno dello stesso VPC

## Stati Security Rule

Le security rule possono essere nei seguenti stati:

| Stato | Descrizione | Posso Aggiornare? | Posso Eliminare? |
|-------|-------------|-------------------|------------------|
| InCreation | La security rule è in fase di creazione | ❌ No | ❌ No |
| Active | La security rule è pronta per l'uso | ✅ Sì | ✅ Sì |

## Workflow Comuni

### Configurazione Security Rule per Web Server

```bash
# 1. Crea security group (se non esiste)
VPC_ID="689307f4745108d3c6343b5a"
SG_ID=$(acloud network securitygroup create $VPC_ID \
  --name "web-server-sg" \
  --region ITBG-Bergamo | grep "ID:" | awk '{print $2}')

# 2. Permetti traffico HTTP da ovunque
acloud network securityrule create $VPC_ID $SG_ID \
  --name "allow-http" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 80 \
  --target-kind Ip \
  --target-value "0.0.0.0/0"

# 3. Permetti traffico HTTPS da ovunque
acloud network securityrule create $VPC_ID $SG_ID \
  --name "allow-https" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 443 \
  --target-kind Ip \
  --target-value "0.0.0.0/0"

# 4. Permetti SSH solo dalla rete admin
acloud network securityrule create $VPC_ID $SG_ID \
  --name "allow-ssh-admin" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 22 \
  --target-kind Ip \
  --target-value "10.0.0.0/8" \
  --tags "ssh,admin,secure"

# 5. Permetti tutto il traffico in uscita
acloud network securityrule create $VPC_ID $SG_ID \
  --name "allow-all-outbound" \
  --region ITBG-Bergamo \
  --direction Egress \
  --protocol ANY \
  --target-kind Ip \
  --target-value "0.0.0.0/0"
```

### Permettere Comunicazione tra Security Group

```bash
VPC_ID="689307f4745108d3c6343b5a"
APP_SG_ID="1234567890abcdef"
DB_SG_ID="9876543210fedcba"

# Ottieni l'URI del security group database
DB_SG_URI=$(acloud network securitygroup get $VPC_ID $DB_SG_ID | grep "URI:" | awk '{print $2}')

# Permetti al security group app di accedere al security group database
acloud network securityrule create $VPC_ID $DB_SG_ID \
  --name "allow-from-app" \
  --region ITBG-Bergamo \
  --direction Ingress \
  --protocol TCP \
  --port 5432 \
  --target-kind SecurityGroup \
  --target-value "$DB_SG_URI"
```

## Best Practices

1. **Principio del Minor Privilegio**
   ```bash
   # Buono: Permetti porta specifica da rete specifica
   --port 22 --target-value "10.0.0.0/8"
   
   # Evita: Permetti tutte le porte da ovunque
   --port "*" --target-value "0.0.0.0/0"
   ```

2. **Usa Nomi Descrittivi**
   ```bash
   --name "allow-http-public"
   --name "allow-ssh-admin-only"
   --name "allow-db-from-app"
   ```

3. **Tagga le Tue Regole**
   ```bash
   --tags "web,public,http"
   --tags "admin,ssh,secure"
   --tags "database,internal"
   ```

4. **Usa Target Security Group per Comunicazione Interna**
   ```bash
   # Meglio: Riferimento security group
   --target-kind SecurityGroup --target-value "/projects/.../securitygroups/..."
   
   # Meno sicuro: Usa intervalli IP
   --target-kind Ip --target-value "10.0.0.0/8"
   ```

5. **Rivedi Regole Regolarmente**
   ```bash
   # Elenca tutte le regole per un security group
   acloud network securityrule list <vpc-id> <securitygroup-id>
   
   # Rivedi ogni regola
   acloud network securityrule get <vpc-id> <securitygroup-id> <rule-id>
   ```

6. **Documenta Regole con Tag**
   ```bash
   --tags "purpose=web-server,environment=production,managed-by=devops"
   ```

## Risoluzione dei Problemi

### "Cannot update security rule while in InCreation state"

**Problema:** Tentativo di aggiornare una security rule che non ha finito di crearsi.

**Soluzione:**
```bash
# Controlla lo stato corrente
acloud network securityrule get <vpc-id> <securitygroup-id> <securityrule-id>

# Attendi che lo Status diventi "Active"
# Poi riprova l'aggiornamento
acloud network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --name "new-name"
```

### "Error: --port is required"

**Problema:** La porta è richiesta per i protocolli TCP e UDP ma non è stata fornita.

**Soluzione:**
```bash
# Per TCP/UDP, fornisci sempre la porta
--protocol TCP --port 80
--protocol UDP --port 53

# Per ANY e ICMP, la porta può essere omessa
--protocol ANY
--protocol ICMP
```

### "Error: at least one field must be provided for update"

**Problema:** Comando update chiamato senza modifiche.

**Soluzione:**
```bash
# Fornisci almeno un campo da aggiornare
acloud network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --name "new-name"
# oppure
acloud network securityrule update <vpc-id> <securitygroup-id> <securityrule-id> --tags tag1,tag2
```

### Valore Target Non Valido

**Problema:** Il formato del valore target è errato.

**Soluzione:**
```bash
# Per target IP, usa notazione CIDR
--target-kind Ip --target-value "0.0.0.0/0"        # Tutti gli IP
--target-kind Ip --target-value "10.0.0.0/8"       # Rete privata
--target-kind Ip --target-value "192.168.1.0/24"   # Subnet specifica

# Per target Security Group, usa URI completo
--target-kind SecurityGroup --target-value "/projects/.../securitygroups/..."
```

## Comandi Correlati

- [Security Group](securitygroup.md) - Gestisci security group
- [VPC](vpc.md) - Gestisci VPC
- [Subnet](subnet.md) - Gestisci subnet
