---
id: database
title: Database (DBaaS)
sidebar_label: Database
description: Workflow end-to-end per il provisioning di un'istanza MySQL DBaaS, la creazione di un database, la gestione di utenti e permessi e la connessione.
---

# Esempio Database (DBaaS)

Questa guida copre il workflow completo end-to-end per il provisioning di un'istanza MySQL DBaaS tramite Aruba Cloud CLI: dalla configurazione della rete alla creazione del database, alla gestione degli utenti e alla connessione.

## Prerequisiti

Prima di iniziare, assicurati di avere:
- La CLI configurata con credenziali valide (`acloud config set`)
- Un progetto attivo (verifica con `acloud management project list`)
- Una VPC attiva, o segui lo Step 0 per preparare le risorse di rete

---

## Step 0: Preparare le Risorse di Rete

### Elenca le VPC disponibili

```bash
acloud network vpc list
```

Esempio output:
```
NAME       ID                        REGION         SUBNETS    STATUS
prod-vpc   69495ef64d0cdc87949b71ec  ITBG-Bergamo   3          Active
```

Annota l'`ID` della VPC. Assicurati che `STATUS` sia `Active` prima di procedere.

---

### Elenca o crea una Subnet

```bash
acloud network subnet list 69495ef64d0cdc87949b71ec
```

Esempio output:
```
NAME        ID                         REGION         CIDR              STATUS
db-subnet   694ba1737712ac0032dbe50a   ITBG-Bergamo   192.168.10.0/24   Active
```

Annota l'`ID` della subnet.

---

### Elenca o crea un Security Group

Il security group deve consentire **traffico TCP inbound sulla porta 3306** per la connettività MySQL.

```bash
acloud network securitygroup list 69495ef64d0cdc87949b71ec
```

Se devi creare un nuovo security group con una regola MySQL inbound:

```bash
# Crea security group
acloud network securitygroup create \
  --name "db-security-group" \
  --vpc-id "69495ef64d0cdc87949b71ec" \
  --region "ITBG-Bergamo"

# Aggiungi regola inbound per MySQL (porta 3306)
acloud network securityrule create 69495ef64d0cdc87949b71ec <securitygroup-id> \
  --direction Inbound \
  --protocol TCP \
  --port-range-min 3306 \
  --port-range-max 3306 \
  --remote-ip-prefix "0.0.0.0/0"
```

Annota l'`ID` del security group.

---

### Ottieni un Elastic IP per l'accesso pubblico

```bash
acloud network elasticip list
```

Annota l'`ID` dell'Elastic IP. Un Elastic IP è necessario per connettersi all'istanza DBaaS dall'esterno della VPC.

---

## Step 1: Crea l'Istanza DBaaS

Crea una nuova istanza MySQL 8.0 DBaaS con tutti i flag di rete richiesti:

```bash
acloud database dbaas create \
  --name "prod-mysql" \
  --region "ITBG-Bergamo" \
  --zone "ITBG-1" \
  --engine-id "mysql-8.0" \
  --flavor "DBO4A8" \
  --storage-size 50 \
  --vpc-id "69495ef64d0cdc87949b71ec" \
  --subnet-id "694ba1737712ac0032dbe50a" \
  --security-group-id "694b05ac4d0cdc87949b75f9" \
  --elastic-ip-id "694bb7897712ac0032dbe60c" \
  --tags "production,mysql"
```

Esempio output:
```
DBaaS instance created successfully!
ID:              69455aa70d0972656501d45d
Name:            prod-mysql
Engine:          mysql-8.0
Flavor:          DBO4A8
Storage (GB):    50
Region:          ITBG-Bergamo
Status:          InCreation
Creation Date:   18-06-2026 10:00:00
```

> **Nota**: Il provisioning del DBaaS richiede tipicamente alcuni minuti. Usa `--wait` con un timeout esteso per attendere che l'istanza diventi `Active`:
> ```bash
> acloud --timeout 15m database dbaas create ... --wait
> ```

---

## Step 2: Attendi che l'Istanza DBaaS Diventi Attiva

Monitora lo stato del provisioning:

```bash
acloud database dbaas list
```

Attendi che `STATUS` mostri `Active`:

```
NAME         ID                        ENGINE     FLAVOR    STORAGE  REGION         STATUS
prod-mysql   69455aa70d0972656501d45d  mysql-8.0  DBO4A8    50 GB    ITBG-Bergamo   Active
```

Ottieni i dettagli completi incluso l'endpoint di connessione:

```bash
acloud database dbaas get 69455aa70d0972656501d45d
```

---

## Step 3: Crea un Database

Una volta che l'istanza DBaaS è `Active`, crea un database al suo interno:

```bash
acloud database dbaas database create 69455aa70d0972656501d45d \
  --name "appdb"
```

Esempio output:
```
Database created successfully!
DBaaS ID:   69455aa70d0972656501d45d
Name:       appdb
```

Verifica che il database sia stato creato:

```bash
acloud database dbaas database list 69455aa70d0972656501d45d
```

---

## Step 4: Crea un Utente del Database

Crea un utente che si connetterà al database:

```bash
acloud database dbaas user create 69455aa70d0972656501d45d \
  --username "app-user" \
  --password "SecurePassword123!"
```

Esempio output:
```
User created successfully!
DBaaS ID:   69455aa70d0972656501d45d
Username:   app-user
```

> **Best practice di sicurezza**: Usa una password robusta (almeno 12 caratteri, mix di maiuscole, minuscole, numeri e simboli). Non commettere mai le password nel version control.

Verifica che l'utente sia stato creato:

```bash
acloud database dbaas user list 69455aa70d0972656501d45d
```

---

## Step 5: Assegna i Permessi all'Utente sul Database

Assegna all'utente l'accesso al database con il ruolo `liteadmin`:

```bash
acloud database dbaas grant create 69455aa70d0972656501d45d "appdb" \
  --username "app-user" \
  --role "liteadmin"
```

Esempio output:
```
Grant created successfully!
DBaaS ID:      69455aa70d0972656501d45d
Database:      appdb
Username:      app-user
Role:          liteadmin
```

Verifica il grant:

```bash
acloud database dbaas grant list 69455aa70d0972656501d45d "appdb"
```

---

## Step 6: Recupera i Dettagli di Connessione

Ottieni i dettagli completi dell'istanza DBaaS per trovare l'endpoint di connessione:

```bash
acloud database dbaas get 69455aa70d0972656501d45d
```

Annota l'indirizzo IP pubblico dall'Elastic IP assegnato alla tua istanza DBaaS.

---

## Step 7: Connettiti al Database

Usa qualsiasi client MySQL standard per connetterti:

```bash
mysql -h <indirizzo-ip-elastic> -P 3306 -u app-user -p appdb
# Enter password: (input nascosto)
```

Oppure con una stringa di connessione:

```bash
mysql --host=<indirizzo-ip-elastic> --port=3306 --user=app-user --password --database=appdb
```

> **Nota**: Assicurati che l'indirizzo IP del tuo computer locale sia consentito dalla regola inbound del security group sulla porta 3306.

---

## Step 8: Verifica la Configurazione

Una volta connesso, verifica che il database e l'utente siano configurati correttamente:

```sql
-- Mostra i database disponibili
SHOW DATABASES;

-- Seleziona il database
USE appdb;

-- Mostra l'utente corrente
SELECT USER();

-- Verifica i permessi
SHOW GRANTS FOR 'app-user'@'%';
```

---

## Best Practice

- **Backup**: Pianifica backup regolari usando `acloud database dbaas backup create`
- **Security group**: Limita la porta 3306 a intervalli IP specifici invece di `0.0.0.0/0`
- **Gestione utenti**: Crea utenti separati per ogni applicazione; evita di usare direttamente l'utente admin
- **Rotazione password**: Ruota regolarmente le password eliminando e ricreando gli utenti
- **Tag**: Usa i tag per tracciare ambiente e allocazione dei costi (es. `production,billing-team-a`)
- **Flag wait**: Usa `--wait --timeout 15m` quando crei istanze DBaaS per rilevare tempestivamente gli errori di provisioning

---

## Step 9: Pulizia

Per eliminare l'istanza DBaaS e tutte le sue risorse:

```bash
# Elimina l'istanza DBaaS (utenti, database e grant vengono rimossi automaticamente)
acloud database dbaas delete 69455aa70d0972656501d45d --yes
```

> **Attenzione**: L'eliminazione di un'istanza DBaaS rimuove permanentemente tutti i database, utenti, grant e dati al suo interno. Assicurati di avere backup prima di procedere.

## Risorse Correlate

- [DBaaS](../resources/database/dbaas.md) - Riferimento completo comandi DBaaS
- [Database DBaaS](../resources/database/dbaas.database.md) - Comandi di gestione database
- [Utenti DBaaS](../resources/database/dbaas.user.md) - Comandi di gestione utenti
- [Grant DBaaS](../resources/database/dbaas.grant.md) - Comandi di gestione grant
- [Backup Database](../resources/database/backup.md) - Comandi backup e ripristino
