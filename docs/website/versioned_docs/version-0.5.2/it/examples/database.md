---
id: database-it
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

Annota l'`ID` della subnet con `STATUS` pari a `Active`.

---

### Elenca o crea un Security Group

Il security group deve consentire **traffico TCP inbound sulla porta 3306** per la connettività MySQL.

```bash
acloud network securitygroup list 69495ef64d0cdc87949b71ec
```

Se devi creare un nuovo security group con una regola MySQL inbound:

```bash
acloud network securitygroup create \
  --name "db-security-group" \
  --vpc-id "69495ef64d0cdc87949b71ec" \
  --region "ITBG-Bergamo"

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

Annota l'`ID` dell'Elastic IP. È necessario per connettersi all'istanza DBaaS dall'esterno della VPC.

---

## Step 1: Crea l'Istanza DBaaS

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

> **Nota**: Usa `--wait --timeout 15m` per attendere che l'istanza diventi `Active`.

---

## Step 2: Attendi che l'Istanza DBaaS Diventi Attiva

```bash
acloud database dbaas list
```

Attendi che `STATUS` mostri `Active`, poi ottieni i dettagli completi:

```bash
acloud database dbaas get 69455aa70d0972656501d45d
```

---

## Step 3: Crea un Database

```bash
acloud database dbaas database create 69455aa70d0972656501d45d \
  --name "appdb"
```

Verifica:

```bash
acloud database dbaas database list 69455aa70d0972656501d45d
```

---

## Step 4: Crea un Utente del Database

```bash
acloud database dbaas user create 69455aa70d0972656501d45d \
  --username "app-user" \
  --password "SecurePassword123!"
```

> **Best practice**: Usa una password robusta. Non commettere mai le password nel version control.

Verifica:

```bash
acloud database dbaas user list 69455aa70d0972656501d45d
```

---

## Step 5: Assegna i Permessi all'Utente sul Database

```bash
acloud database dbaas grant create 69455aa70d0972656501d45d "appdb" \
  --username "app-user" \
  --role "liteadmin"
```

Verifica:

```bash
acloud database dbaas grant list 69455aa70d0972656501d45d "appdb"
```

---

## Step 6: Recupera i Dettagli di Connessione

```bash
acloud database dbaas get 69455aa70d0972656501d45d
```

Annota l'indirizzo IP pubblico dall'Elastic IP assegnato alla tua istanza DBaaS.

---

## Step 7: Connettiti al Database

```bash
mysql -h <indirizzo-ip-elastic> -P 3306 -u app-user -p appdb
```

> **Nota**: Assicurati che l'IP del tuo computer sia consentito dalla regola inbound del security group sulla porta 3306.

---

## Step 8: Verifica la Configurazione

```sql
SHOW DATABASES;
USE appdb;
SELECT USER();
SHOW GRANTS FOR 'app-user'@'%';
```

---

## Best Practice

- Pianifica backup regolari usando `acloud database dbaas backup create`
- Limita la porta 3306 a intervalli IP specifici invece di `0.0.0.0/0`
- Crea utenti separati per ogni applicazione
- Ruota regolarmente le password eliminando e ricreando gli utenti
- Usa i tag per tracciare ambiente e allocazione dei costi

---

## Step 9: Pulizia

```bash
acloud database dbaas delete 69455aa70d0972656501d45d --yes
```

> **Attenzione**: L'eliminazione rimuove permanentemente tutti i database, utenti, grant e dati. Assicurati di avere backup prima di procedere.
