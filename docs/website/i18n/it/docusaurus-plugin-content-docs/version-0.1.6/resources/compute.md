# Risorse Compute

La categoria `compute` fornisce comandi per gestire risorse compute in Aruba Cloud, inclusi cloud server e coppie di chiavi SSH.

## Risorse Disponibili

### [Cloud Server](compute/cloudserver.md)

I cloud server sono istanze di macchine virtuali che eseguono le tue applicazioni e carichi di lavoro.

**Comandi Rapidi:**
```bash
# Elenca tutti i cloud server
acloud compute cloudserver list

# Ottieni i dettagli del cloud server
acloud compute cloudserver get <server-id>

# Crea un cloud server
acloud compute cloudserver create --name "my-server" --region "ITBG-Bergamo" --flavor "small" --image <image-id>

# Aggiorna un cloud server
acloud compute cloudserver update <server-id> --name "new-name"

# Elimina un cloud server
acloud compute cloudserver delete <server-id>
```

### [Coppie di Chiavi](compute/keypair.md)

Coppie di chiavi SSH per autenticazione sicura ai cloud server.

**Comandi Rapidi:**
```bash
# Elenca tutte le coppie di chiavi
acloud compute keypair list

# Ottieni i dettagli della coppia di chiavi
acloud compute keypair get <keypair-name>

# Crea una coppia di chiavi
acloud compute keypair create --name "my-keypair" --public-key "ssh-rsa AAAAB3..."

# Aggiorna una coppia di chiavi (cambia chiave pubblica)
acloud compute keypair update <keypair-name> --public-key "ssh-rsa AAAAB3..."

# Elimina una coppia di chiavi
acloud compute keypair delete <keypair-name>
```

## Casi d'Uso Comuni

### Avvio di un Cloud Server

1. **Crea una coppia di chiavi** (se non ne hai una):
   ```bash
   acloud compute keypair create --name "my-keypair" --public-key "$(cat ~/.ssh/id_rsa.pub)"
   ```

2. **Elenca flavor e immagini disponibili**:
   ```bash
   # Controlla le risorse disponibili (potresti dover usare la console web o l'API)
   ```

3. **Crea il cloud server**:
   ```bash
   acloud compute cloudserver create \
     --name "web-server" \
     --region "ITBG-Bergamo" \
     --flavor "small" \
     --image "your-image-id" \
     --keypair "my-keypair" \
     --tags "production,web"
   ```

4. **Verifica il server**:
   ```bash
   acloud compute cloudserver list
   acloud compute cloudserver get <server-id>
   ```

### Gestione dell'Accesso SSH

1. **Elenca tutte le coppie di chiavi**:
   ```bash
   acloud compute keypair list
   ```

2. **Aggiorna una coppia di chiavi** (ruota le chiavi):
   ```bash
   acloud compute keypair update "my-keypair" --public-key "$(cat ~/.ssh/id_rsa_new.pub)"
   ```

3. **Elimina coppie di chiavi non utilizzate**:
   ```bash
   acloud compute keypair delete "old-keypair" --yes
   ```

## Best Practices

- **Coppie di Chiavi**:
  - Usa nomi descrittivi per le coppie di chiavi (es. `user-john-laptop`, `ci-cd-server`)
  - Ruota le chiavi regolarmente per la sicurezza
  - Mantieni le chiavi private sicure e non condividerle mai
  - Usa coppie di chiavi diverse per ambienti diversi

- **Cloud Server**:
  - Usa tag per organizzare i server per ambiente, progetto o scopo
  - Scegli flavor appropriati in base ai requisiti del carico di lavoro
  - Monitora lo stato del server prima di eseguire aggiornamenti
  - Usa coppie di chiavi invece dell'autenticazione con password per una migliore sicurezza

## Risorse Correlate

- [Risorse di Rete](./network.md) - Configura il networking per i cloud server
- [Risorse Storage](./storage.md) - Collega volumi di block storage ai server
- [Risorse di Sicurezza](./security.md) - Gestisci security group e regole

