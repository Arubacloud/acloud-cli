# Load Balancer

I Load Balancer distribuiscono il traffico di rete in entrata tra più risorse per garantire alta disponibilità e affidabilità.

**Nota:** I Load Balancer sono attualmente in sola lettura tramite la CLI. Puoi elencare e visualizzare i dettagli, ma non puoi crearli, aggiornarli o eliminarli tramite la CLI.

## Comandi

### Elenca Load Balancer

Elenca tutti i Load Balancer nel tuo progetto.

```bash
acloud network loadbalancer list [flags]
```

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
# Elenca Load Balancer usando il contesto
acloud network loadbalancer list

# Elenca con ID progetto esplicito
acloud network loadbalancer list --project-id 68398923fb2cb026400d4d31
```

**Output:**
```
NAME                                     ID                        ADDRESS          STATUS
ingress-nginx-controller                 68ffa1797912602cb16794dc  209.227.232.229  Active
api-gateway-lb                           69485b8a4d0cdc87949b7012  95.110.142.230   Active
```

### Ottieni Dettagli Load Balancer

Ottieni informazioni dettagliate su un Load Balancer specifico.

```bash
acloud network loadbalancer get <load-balancer-id> [flags]
```

**Argomenti:**
- `load-balancer-id` - L'ID del Load Balancer (supporta auto-completamento)

**Flag:**
- `--project-id string` - ID progetto (usa il contesto se non specificato)

**Esempio:**
```bash
acloud network loadbalancer get 68ffa1797912602cb16794dc
```

**Output:**
```
Load Balancer Details:
======================
ID:              68ffa1797912602cb16794dc
URI:             /projects/.../loadBalancers/68ffa1797912602cb16794dc
Name:            ingress-nginx-controller
Address:         209.227.232.229
Linked Resources: 2
VPC:             689307f4745108d3c6343b5a
Creation Date:   27-10-2025 16:44:41
Created By:      aru-297647
Tags:            [kubernetes ingress production]
Status:          Active
```

## Auto-completamento Shell

I comandi Load Balancer supportano auto-completamento intelligente per gli ID Load Balancer:

```bash
# Abilita completamento (bash)
source <(acloud completion bash)

# Digita il comando e premi TAB per vedere gli ID Load Balancer disponibili
acloud network loadbalancer get <TAB>
```

L'auto-completamento mostra gli ID Load Balancer con i loro nomi:
```
68ffa1797912602cb16794dc    ingress-nginx-controller
69485b8a4d0cdc87949b7012    api-gateway-lb
```

## Stati Load Balancer

I Load Balancer possono essere nei seguenti stati:

| Stato | Descrizione |
|-------|-------------|
| InCreation | Il Load Balancer è in fase di creazione |
| Active | Il Load Balancer è pronto e distribuisce il traffico |
| Updating | La configurazione del Load Balancer è in fase di aggiornamento |
| Deleting | Il Load Balancer è in fase di eliminazione |

## Proprietà Load Balancer

### Indirizzo

L'indirizzo IP pubblico del Load Balancer:
- Assegnato automaticamente
- Usato per instradare il traffico in entrata
- Tipicamente un Elastic IP

### Associazione VPC

I Load Balancer sono associati a un VPC:
- Mostrato come ID VPC nei dettagli
- Determina l'isolamento di rete
- Influisce su routing e security rule

### Risorse Collegate

I Load Balancer distribuiscono il traffico a risorse collegate:
- Server backend
- Target group
- Endpoint health check

Il conteggio `Linked Resources` mostra quanti backend sono configurati.

### Tag

I Load Balancer supportano tag per l'organizzazione:
- Impostati quando il Load Balancer viene creato
- Visibili tramite il comando get
- Non possono essere modificati tramite CLI (sola lettura)

## Workflow Comuni

### Visualizzazione Informazioni Load Balancer

```bash
# Elenca tutti i Load Balancer
acloud network loadbalancer list

# Ottieni dettagli di un Load Balancer specifico
acloud network loadbalancer get <lb-id>

# Visualizza Load Balancer con tag specifici
acloud network loadbalancer list | grep "production"
```

### Monitoraggio Load Balancer

```bash
# Controlla lo stato di tutti i Load Balancer
acloud network loadbalancer list

# Ottieni informazioni dettagliate inclusi risorse collegate
for lb_id in $(acloud network loadbalancer list | tail -n +2 | awk '{print $2}'); do
  echo "=== Load Balancer: $lb_id ==="
  acloud network loadbalancer get $lb_id
  echo ""
done
```

## Limitazioni

### Accesso in Sola Lettura

I Load Balancer sono in sola lettura tramite la CLI:
- ❌ Non puoi creare Load Balancer
- ❌ Non puoi aggiornare la configurazione del Load Balancer
- ❌ Non puoi eliminare Load Balancer
- ❌ Non puoi modificare tag
- ✅ Puoi elencare Load Balancer
- ✅ Puoi visualizzare i dettagli del Load Balancer

### Gestione Load Balancer

Per creare, aggiornare o eliminare Load Balancer, usa:
- Console Web Aruba Cloud
- API Aruba Cloud direttamente
- Strumenti Infrastructure as Code (Terraform, ecc.)

La CLI fornisce accesso in sola lettura per monitoraggio e riferimento.

## Comandi Correlati

- [VPC](vpc.md) - Visualizza VPC associato ai Load Balancer
- [Elastic IP](elasticip.md) - Visualizza Elastic IP usati dai Load Balancer
- [Gestione Contesto](../../installation.md#context-management) - Gestisci contesti progetto
