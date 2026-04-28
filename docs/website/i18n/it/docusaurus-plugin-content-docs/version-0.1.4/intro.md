# Introduzione ad Aruba Cloud CLI

Aruba Cloud CLI (`acloud`) è lo strumento ufficiale da riga di comando per gestire e automatizzare la tua infrastruttura Aruba Cloud. Costruito con Go e basato sull'SDK Aruba Cloud, fornisce un'interfaccia potente e scriptabile per interagire con tutti i servizi Aruba Cloud direttamente dal tuo terminale.

## Cos'è Aruba Cloud CLI?

Aruba Cloud CLI è uno strumento completo che ti consente di gestire l'intera infrastruttura Aruba Cloud attraverso comandi semplici e intuitivi. Che tu stia fornendo risorse, gestendo reti, configurando storage o automatizzando deployment, la CLI fornisce un'interfaccia coerente per tutti i servizi Aruba Cloud.

## Capacità Principali

### Gestione dell'Infrastruttura
Gestisci l'intero ciclo di vita delle tue risorse cloud inclusi:
- **Risorse Compute**: Cloud server, coppie di chiavi e orchestrazione container (KaaS)
- **Storage**: Volumi di block storage, snapshot, backup e operazioni di ripristino
- **Networking**: Virtual Private Cloud (VPC), subnet, security group, elastic IP, load balancer e configurazioni VPN
- **Database**: Istanze Database-as-a-Service (DBaaS), database, utenti e backup
- **Sicurezza**: Chiavi Key Management Service (KMS) e crittografia
- **Pianificazione**: Pianificazione e gestione automatizzata dei job
- **Progetti**: Gestione e organizzazione multi-progetto

### Esperienza Sviluppatore
- **Comandi Intuitivi**: Struttura di comandi coerente e prevedibile per tutte le risorse
- **Auto-completamento Shell**: Completamento intelligente con tab per comandi, flag e ID risorse
- **Gestione Contesto**: Passa tra progetti e ambienti senza interruzioni
- **Modalità Debug**: Logging dettagliato per troubleshooting e sviluppo
- **Scriptabile**: Perfetto per automazione, pipeline CI/CD e workflow infrastructure-as-code

### Funzionalità Enterprise
- **Supporto Multi-Progetto**: Gestisci più progetti da una singola istanza CLI
- **Gestione Credenziali Sicura**: Archiviazione crittografata delle credenziali API
- **Gestione Errori Completa**: Messaggi di errore chiari e azionabili
- **Copertura API Completa**: Operazioni CRUD complete per tutte le risorse supportate

## Perché Usare Aruba Cloud CLI?

- **Automazione**: Integra la gestione Aruba Cloud nei tuoi workflow DevOps e script di automazione
- **Efficienza**: Gestisci le risorse più velocemente rispetto alla console web, specialmente per operazioni in bulk
- **Coerenza**: Gestione dell'infrastruttura riproducibile attraverso script da riga di comando
- **Integrazione**: Integra facilmente con altri strumenti, pipeline CI/CD e piattaforme di gestione dell'infrastruttura
- **Utenti Avanzati**: Funzionalità avanzate per utenti esperti che preferiscono interfacce da riga di comando

## Iniziare

Pronto a iniziare a usare Aruba Cloud CLI? Consulta la guida [Installazione](installation.md) per configurare la tua piattaforma.

## Esplora le Risorse

Una volta installato, esplora le risorse disponibili:
- [Risorse di Gestione](resources/management.md) - Progetti e organizzazione
- [Risorse Storage](resources/storage.md) - Block storage, snapshot, backup e ripristini
- [Risorse di Rete](resources/network.md) - VPC, subnet, security group, elastic IP, load balancer e VPN
- [Risorse Database](resources/database.md) - Istanze DBaaS, database e utenti
- [Risorse Compute](resources/compute.md) - Cloud server e coppie di chiavi
- [Risorse Container](resources/container.md) - Kubernetes as a Service (KaaS)
- [Risorse di Sicurezza](resources/security.md) - Key Management Service
- [Risorse di Pianificazione](resources/schedule.md) - Pianificazione job

