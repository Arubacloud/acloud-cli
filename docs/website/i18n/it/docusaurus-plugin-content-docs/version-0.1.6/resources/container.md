# Risorse Container

La categoria `container` fornisce comandi per gestire risorse container in Aruba Cloud, inclusi cluster Kubernetes as a Service (KaaS) e Container Registry.

## Risorse Disponibili

### [KaaS (Kubernetes as a Service)](container/kaas.md)

KaaS fornisce cluster Kubernetes gestiti per eseguire applicazioni containerizzate.

**Comandi Rapidi:**
```bash
# Elenca tutti i cluster KaaS
acloud container kaas list

# Ottieni i dettagli del cluster KaaS
acloud container kaas get <cluster-id>

# Crea un cluster KaaS
acloud container kaas create --name "my-cluster" --region "ITBG-Bergamo" --version "1.28.0"

# Aggiorna un cluster KaaS
acloud container kaas update <cluster-id> --name "new-name"

# Elimina un cluster KaaS
acloud container kaas delete <cluster-id>
```

### [Container Registry](container/containerregistry.md)

Container Registry fornisce un registry Docker container privato per memorizzare e gestire immagini container.

**Comandi Rapidi:**
```bash
# Elenca tutti i container registry
acloud container containerregistry list

# Ottieni i dettagli del container registry
acloud container containerregistry get <registry-id>

# Crea un container registry
acloud container containerregistry create \
  --name "my-registry" \
  --region "ITBG-Bergamo" \
  --public-ip-uri "/projects/{id}/providers/Aruba.Network/elasticIps/{eip-id}" \
  --vpc-uri "/projects/{id}/providers/Aruba.Network/vpcs/{vpc-id}" \
  --subnet-uri "/projects/{id}/providers/Aruba.Network/subnets/{subnet-id}" \
  --security-group-uri "/projects/{id}/providers/Aruba.Network/securityGroups/{sg-id}" \
  --block-storage-uri "/projects/{id}/providers/Aruba.Storage/volumes/{volume-id}"

# Aggiorna un container registry
acloud container containerregistry update <registry-id> --name "new-name"

# Elimina un container registry
acloud container containerregistry delete <registry-id>
```

## Casi d'Uso Comuni

### Creazione di un Cluster Kubernetes

1. **Crea un cluster KaaS**:
   ```bash
   acloud container kaas create \
     --name "production-cluster" \
     --region "ITBG-Bergamo" \
     --version "1.28.0" \
     --tags "production,kubernetes"
   ```

2. **Attendi che il cluster sia pronto** e controlla lo stato:
   ```bash
   acloud container kaas get <cluster-id>
   ```

3. **Configura kubectl** (dopo che il cluster è pronto):
   ```bash
   # Ottieni il kubeconfig del cluster dalla console Aruba Cloud o API
   # Poi configura kubectl:
   kubectl config use-context <cluster-context>
   ```

### Gestione Metadati Cluster

```bash
# Aggiorna nome e tag del cluster
acloud container kaas update <cluster-id> \
  --name "production-cluster-updated" \
  --tags "production,kubernetes,updated"
```

### Creazione di un Container Registry

1. **Assicurati che le risorse richieste esistano:**
   - IP Pubblico (Elastic IP)
   - VPC
   - Subnet
   - Security Group
   - Block Storage

2. **Crea il container registry:**
   ```bash
   acloud container containerregistry create \
     --name "my-registry" \
     --region "ITBG-Bergamo" \
     --public-ip-uri "/projects/{id}/providers/Aruba.Network/elasticIps/{eip-id}" \
     --vpc-uri "/projects/{id}/providers/Aruba.Network/vpcs/{vpc-id}" \
     --subnet-uri "/projects/{id}/providers/Aruba.Network/subnets/{subnet-id}" \
     --security-group-uri "/projects/{id}/providers/Aruba.Network/securityGroups/{sg-id}" \
     --block-storage-uri "/projects/{id}/providers/Aruba.Storage/volumes/{volume-id}" \
     --billing-period "Month"
   ```

3. **Attendi che il registry sia pronto** e controlla lo stato:
   ```bash
   acloud container containerregistry get <registry-id>
   ```

## Best Practices

- **Denominazione**: Usa nomi descrittivi che indicano lo scopo del cluster (es. `prod-k8s-cluster`, `staging-cluster`)
- **Tag**: Usa tag per organizzare i cluster per ambiente, progetto o team
- **Versioning**: Mantieni le versioni Kubernetes aggiornate per sicurezza e funzionalità
- **Monitoraggio**: Monitora regolarmente lo stato e la salute del cluster
- **Gestione Risorse**: Pianifica le risorse del cluster in base ai requisiti del carico di lavoro
- **Sicurezza**: Segui le best practice di sicurezza Kubernetes per i tuoi carichi di lavoro

## Risorse Correlate

- [Risorse di Rete](./network.md) - Configura il networking per carichi di lavoro container
- [Risorse Storage](./storage.md) - Volumi persistenti per applicazioni container
- [Risorse di Sicurezza](./security.md) - Politiche di sicurezza e gestione chiavi

