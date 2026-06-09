# Risorse

Questa sezione fornisce documentazione completa per tutte le risorse di Aruba Cloud CLI.

## Categorie di Risorse

### [Risorse di Gestione](resources/management.md)

Gestisci progetti e risorse organizzative.

- [Progetti](resources/management/project.md) - Crea e gestisci progetti

### [Risorse Storage](resources/storage.md)

Gestisci block storage, snapshot, backup e operazioni di ripristino.

- [Block Storage](resources/storage/blockstorage.md) - Volumi di storage persistente
- [Snapshot](resources/storage/snapshot.md) - Copie point-in-time
- [Backup](resources/storage/backup.md) - Operazioni di backup avanzate
- [Operazioni di Ripristino](resources/storage/restore.md) - Ripristina da backup

### [Risorse di Rete](resources/network.md)

Gestisci virtual private cloud, networking e sicurezza.

- [VPC](resources/network/vpc.md) - Virtual Private Clouds
- [Subnet](resources/network/subnet.md) - Subnet di rete
- [Security Group](resources/network/securitygroup.md) - Security group
- [Security Rule](resources/network/securityrule.md) - Regole firewall
- [Elastic IP](resources/network/elasticip.md) - Indirizzi IP pubblici
- [Load Balancer](resources/network/loadbalancer.md) - Bilanciamento del carico
- [VPC Peering](resources/network/vpcpeering.md) - Connessioni VPC
- [VPC Peering Route](resources/network/vpcpeeringroute.md) - Route di peering
- [VPN Tunnel](resources/network/vpntunnel.md) - Connessioni VPN
- [VPN Route](resources/network/vpnroute.md) - Routing VPN

### [Risorse Database](resources/database.md)

Gestisci servizi database, database, utenti e backup.

- [DBaaS](resources/database/dbaas.md) - Istanze Database as a Service
- [Database DBaaS](resources/database/dbaas.database.md) - Database all'interno di DBaaS
- [Utenti DBaaS](resources/database/dbaas.user.md) - Utenti database
- [Backup Database](resources/database/backup.md) - Operazioni di backup database

### [Risorse di Pianificazione](resources/schedule.md)

Gestisci job pianificati per l'automazione.

- [Job](resources/schedule/job.md) - Job pianificati (OneShot e Ricorrenti)

### [Risorse di Sicurezza](resources/security.md)

Gestisci risorse di sicurezza e crittografia.

- [Chiavi KMS](resources/security/kms.md) - Chiavi Key Management System

### [Risorse Compute](resources/compute.md)

Gestisci risorse compute e coppie di chiavi SSH.

- [Cloud Server](resources/compute/cloudserver.md) - Istanze di macchine virtuali
- [Coppie di Chiavi](resources/compute/keypair.md) - Gestione coppie di chiavi SSH

### [Risorse Container](resources/container.md)

Gestisci risorse container e Kubernetes.

- [KaaS](resources/container/kaas.md) - Cluster Kubernetes as a Service
- [Container Registry](resources/container/containerregistry.md) - Registry Docker container privato

## Riferimento Rapido

Tutte le risorse supportano operazioni CRUD standard:

- **List**: `acloud <category> <resource> list`
- **Get**: `acloud <category> <resource> get <id>`
- **Create**: `acloud <category> <resource> create [flags]`
- **Update**: `acloud <category> <resource> update <id> [flags]`
- **Delete**: `acloud <category> <resource> delete <id> [--yes]`

Per informazioni dettagliate su ogni risorsa, consulta le pagine di documentazione specifiche delle risorse.

