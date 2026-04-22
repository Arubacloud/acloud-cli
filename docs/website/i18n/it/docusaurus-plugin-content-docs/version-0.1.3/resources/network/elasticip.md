# Elastic IP

Gli Elastic IP sono indirizzi IP pubblici statici che possono essere assegnati alle tue risorse cloud.

## Comandi

### Elenca Elastic IP

Elenca tutti gli Elastic IP nel tuo progetto.

```bash
acloud network elasticip list [flags]
```

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
# Elenca Elastic IP usando il contesto
acloud network elasticip list

# Elenca con ID progetto esplicito
acloud network elasticip list --project-id 68398923fb2cb026400d4d31
```

**Output:**
```
NAME                    ID                        ADDRESS          STATUS
prod-api-ip             68ffa0ddce76e7da20465721  209.227.232.229  InUse
staging-web-ip          69007f71ce76e7da20465a52  95.110.142.229   InUse
dev-test-ip             6908820cf974c5deb5decd6c  209.227.232.182  NotUsed
```

### Ottieni Dettagli Elastic IP

Ottieni informazioni dettagliate su un Elastic IP specifico.

```bash
acloud network elasticip get <elastic-ip-id> [flags]
```

**Argomenti:**
- `elastic-ip-id` - L'ID dell'Elastic IP (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud network elasticip get 68ffa0ddce76e7da20465721
```

**Output:**
```
Elastic IP Details:
===================
ID:              68ffa0ddce76e7da20465721
URI:             /projects/.../elasticIps/68ffa0ddce76e7da20465721
Name:            prod-api-ip
Address:         209.227.232.229
Billing Period:  Hour
Linked Resources: 2
Creation Date:   27-10-2025 16:42:05
Created By:      aru-297647
Tags:            [production api public]
Status:          InUse
```

### Crea Elastic IP

Crea un nuovo Elastic IP.

```bash
acloud network elasticip create [flags]
```

**Flag Richiesti:**
- `--name string` - Nome per l'Elastic IP
- `--region string` - Codice regione (es. ITBG-Bergamo)
- `--billing-period string` - Periodo di fatturazione: Hour, Month o Year

**Flag Opzionali:**
- `--tags strings` - Tag per l'Elastic IP (separati da virgola)
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempi:**
```bash
# Crea Elastic IP fatturato a ore
acloud network elasticip create \
  --name "my-elastic-ip" \
  --region ITBG-Bergamo \
  --billing-period Hour

# Crea Elastic IP fatturato mensilmente con tag
acloud network elasticip create \
  --name "prod-api-ip" \
  --region ITBG-Bergamo \
  --billing-period Month \
  --tags production,api,public

# Crea Elastic IP fatturato annualmente
acloud network elasticip create \
  --name "long-term-ip" \
  --region ITBG-Bergamo \
  --billing-period Year \
  --tags production,critical
```

**Output:**
```
Elastic IP created successfully!
ID:      69485a704d0cdc87949b6ffe
Name:    my-elastic-ip
```

**Note:**
- Gli Elastic IP sono inizialmente in stato **InCreation**
- L'indirizzo IP viene assegnato automaticamente
- La fatturazione inizia una volta creato l'IP

**Opzioni Periodo Fatturazione:**
- `Hour` - Paga per ora (flessibile, costo unitario più alto)
- `Month` - Impegno mensile (conveniente per utilizzo costante)
- `Year` - Impegno annuale (più conveniente per uso a lungo termine)

### Aggiorna Elastic IP

Aggiorna il nome e/o i tag di un Elastic IP esistente.

```bash
acloud network elasticip update <elastic-ip-id> [flags]
```

**Argomenti:**
- `elastic-ip-id` - L'ID dell'Elastic IP (supporta auto-completamento)

**Flag:**
- `--name string` - Nuovo nome per l'Elastic IP
- `--tags strings` - Nuovi tag per l'Elastic IP (separati da virgola)
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Nota:** Almeno uno tra `--name` o `--tags` deve essere fornito.

**Esempi:**
```bash
# Aggiorna nome Elastic IP
acloud network elasticip update 68ffa0ddce76e7da20465721 --name "prod-api-gateway"

# Aggiorna tag Elastic IP
acloud network elasticip update 68ffa0ddce76e7da20465721 --tags production,api,critical

# Aggiorna sia nome che tag
acloud network elasticip update 68ffa0ddce76e7da20465721 \
  --name "prod-frontend-ip" \
  --tags production,frontend,load-balanced
```

**Output:**
```
Elastic IP updated successfully!
ID:      68ffa0ddce76e7da20465721
Name:    prod-frontend-ip
Tags:    [production frontend load-balanced]
```

**Restrizioni:**
- Non è possibile aggiornare Elastic IP in stato **InCreation**
- Non è possibile cambiare l'indirizzo IP o il periodo di fatturazione dopo la creazione
- Attendi lo stato **NotUsed** o **InUse** prima di aggiornare

### Elimina Elastic IP

Elimina un Elastic IP.

```bash
acloud network elasticip delete <elastic-ip-id> [flags]
```

**Argomenti:**
- `elastic-ip-id` - L'ID dell'Elastic IP (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)
- `-y, --yes` - Salta il prompt di conferma

**Esempi:**
```bash
# Elimina con prompt di conferma
acloud network elasticip delete 68ffa0ddce76e7da20465721

# Elimina senza conferma
acloud network elasticip delete 68ffa0ddce76e7da20465721 --yes

# Elimina con ID progetto esplicito
acloud network elasticip delete 68ffa0ddce76e7da20465721 \
  --project-id 68398923fb2cb026400d4d31 \
  --yes
```

**Prompt di Conferma:**
```
Are you sure you want to delete Elastic IP 68ffa0ddce76e7da20465721? This action cannot be undone.
Type 'yes' to confirm: yes

Elastic IP 68ffa0ddce76e7da20465721 deleted successfully!
```

**Note:**
- Scollega l'Elastic IP dalle risorse prima dell'eliminazione
- L'eliminazione è immediata e non può essere annullata
- La fatturazione si ferma dopo l'eliminazione

## Auto-completamento Shell

I comandi Elastic IP supportano auto-completamento intelligente per gli ID Elastic IP:

```bash
# Abilita completamento (bash)
source <(acloud completion bash)

# Digita il comando e premi TAB per vedere gli ID Elastic IP disponibili
acloud network elasticip get <TAB>
acloud network elasticip update <TAB>
acloud network elasticip delete <TAB>
```

L'auto-completamento mostra gli ID Elastic IP con i loro nomi:
```
68ffa0ddce76e7da20465721    prod-api-ip
69007f71ce76e7da20465a52    staging-web-ip
6908820cf974c5deb5decd6c    dev-test-ip
```

## Stati Elastic IP

Gli Elastic IP possono essere nei seguenti stati:

| Stato | Descrizione | Posso Aggiornare? | Posso Eliminare? |
|-------|-------------|-------------------|------------------|
| InCreation | L'IP è in fase di creazione | ❌ No | ❌ No |
| NotUsed | IP creato ma non collegato | ✅ Sì | ✅ Sì |
| InUse | L'IP è collegato a una risorsa | ✅ Sì | ⚠️ Scollega prima |

## Proprietà Elastic IP

### Indirizzo IP

L'indirizzo IP viene assegnato automaticamente da Aruba Cloud:
- Non può essere specificato durante la creazione
- Non può essere cambiato dopo la creazione
- Rimane costante per tutta la durata dell'Elastic IP
- Viene rilasciato quando l'Elastic IP viene eliminato

### Periodo di Fatturazione

Il periodo di fatturazione viene impostato durante la creazione e non può essere cambiato:
- **Hour**: Più flessibile, fatturato a ore
- **Month**: Impegno mensile, migliore valore
- **Year**: Impegno annuale, miglior valore

Per cambiare il periodo di fatturazione, devi eliminare e ricreare l'Elastic IP.

### Risorse Collegate

Gli Elastic IP possono essere collegati a risorse:
- Macchine virtuali
- Load balancer
- Interfacce di rete

Il conteggio `Linked Resources` mostra quante risorse stanno usando questo IP.

## Workflow Comuni

### Creazione e Configurazione di un Elastic IP

```bash
# 1. Crea l'Elastic IP
EIP_ID=$(acloud network elasticip create \
  --name "prod-api-ip" \
  --region ITBG-Bergamo \
  --billing-period Month \
  --tags production | grep "ID:" | awk '{print $2}')

# 2. Attendi il completamento della creazione
while true; do
  STATUS=$(acloud network elasticip get $EIP_ID | grep "Status:" | awk '{print $2}')
  if [ "$STATUS" = "NotUsed" ] || [ "$STATUS" = "InUse" ]; then
    break
  fi
  echo "Waiting for Elastic IP to be ready... (current: $STATUS)"
  sleep 5
done

# 3. Ottieni l'indirizzo IP assegnato
IP_ADDR=$(acloud network elasticip get $EIP_ID | grep "Address:" | awk '{print $2}')
echo "Elastic IP ready: $IP_ADDR"

# 4. Aggiorna con tag aggiuntivi
acloud network elasticip update $EIP_ID --tags production,api,public

# 5. Verifica la configurazione
acloud network elasticip get $EIP_ID
```

### Gestione di Più Elastic IP

```bash
# Elenca tutti gli Elastic IP
acloud network elasticip list

# Tagga gli IP per scopo
acloud network elasticip update <eip-id-1> --tags production,api
acloud network elasticip update <eip-id-2> --tags staging,web
acloud network elasticip update <eip-id-3> --tags development,testing

# Ottieni dettagli di tutti gli Elastic IP
for eip_id in $(acloud network elasticip list | tail -n +2 | awk '{print $2}'); do
  echo "=== Elastic IP: $eip_id ==="
  acloud network elasticip get $eip_id
  echo ""
done
```

### Pulizia Elastic IP Non Utilizzati

```bash
# Elenca tutti gli Elastic IP
acloud network elasticip list

# Identifica IP non utilizzati (Status: NotUsed)
acloud network elasticip list | grep "NotUsed"

# Elimina Elastic IP non utilizzati
acloud network elasticip delete <unused-eip-id> --yes

# Verifica eliminazione
acloud network elasticip list
```

## Best Practices

1. **Scegli Periodo di Fatturazione Appropriato**
   ```bash
   # Test a breve termine
   acloud network elasticip create --name "test-ip" --region ITBG-Bergamo --billing-period Hour
   
   # Produzione a lungo termine
   acloud network elasticip create --name "prod-ip" --region ITBG-Bergamo --billing-period Year
   ```

2. **Usa Nomi Descrittivi**
   ```bash
   acloud network elasticip create \
     --name "prod-api-gateway-ip" \
     --region ITBG-Bergamo \
     --billing-period Month
   ```

3. **Tagga per Ambiente e Scopo**
   ```bash
   acloud network elasticip update <eip-id> --tags production,frontend,load-balanced
   ```

4. **Monitora l'Utilizzo**
   ```bash
   # Controlla quali IP sono in uso
   acloud network elasticip list | grep "InUse"
   
   # Controlla quali IP non sono utilizzati
   acloud network elasticip list | grep "NotUsed"
   ```

5. **Pulisci IP Non Utilizzati**
   ```bash
   # Gli IP non utilizzati generano comunque costi
   # Elimina gli IP che non stai usando
   acloud network elasticip delete <unused-eip-id> --yes
   ```

## Ottimizzazione Costi

### Selezione Periodo di Fatturazione

Scegli il periodo di fatturazione giusto in base all'utilizzo previsto:

| Durata | Periodo Consigliato | Motivo |
|--------|---------------------|--------|
| < 1 mese | Hour | Flessibilità, nessun impegno |
| 1-12 mesi | Month | Bilanciamento tra flessibilità e costo |
| > 12 mesi | Year | Risparmio massimo |

### Identificazione IP Non Utilizzati

```bash
# Trova tutti gli IP NotUsed
acloud network elasticip list | grep "NotUsed"

# Ottieni dettagli per determinare se sono necessari
acloud network elasticip get <eip-id>

# Elimina se non più necessari
acloud network elasticip delete <eip-id> --yes
```

## Risoluzione dei Problemi

### "Cannot update Elastic IP while in InCreation state"

**Problema:** Tentativo di aggiornare un Elastic IP che non ha finito di crearsi.

**Soluzione:**
```bash
# Controlla lo stato corrente
acloud network elasticip get <eip-id>

# Attendi che lo Status diventi "NotUsed" o "InUse"
# Poi riprova l'aggiornamento
acloud network elasticip update <eip-id> --name "new-name"
```

### "Error: --billing-period must be Hour, Month, or Year"

**Problema:** Periodo di fatturazione non valido specificato.

**Soluzione:**
```bash
# Usa uno dei valori validi (case-sensitive)
acloud network elasticip create \
  --name "my-ip" \
  --region ITBG-Bergamo \
  --billing-period Month  # Non "monthly" o "month"
```

### Indirizzo IP Non Mostrato Dopo la Creazione

**Problema:** Elastic IP creato ma il campo indirizzo è vuoto.

**Soluzione:**
```bash
# Attendi alcuni secondi per l'assegnazione IP
sleep 5

# Controlla di nuovo
acloud network elasticip get <eip-id>

# L'indirizzo dovrebbe ora essere popolato
```

### Non Posso Eliminare IP Collegato a Risorsa

**Problema:** L'Elastic IP è in uso e non può essere eliminato.

**Soluzione:**
```bash
# Controlla quali risorse stanno usando l'IP
acloud network elasticip get <eip-id>

# Scollega l'IP dalle risorse prima (tramite gestione risorse)
# Poi elimina l'Elastic IP
acloud network elasticip delete <eip-id> --yes
```

## Comandi Correlati

- [VPC](vpc.md) - Isolamento di rete per Elastic IP
- [Load Balancer](loadbalancer.md) - Usa Elastic IP con load balancer
- [Gestione Contesto](../../installation.md#context-management) - Gestisci contesti progetto
