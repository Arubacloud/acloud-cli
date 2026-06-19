---
id: cloudserver
title: CloudServer
sidebar_label: CloudServer
description: Scopri i comandi essenziali per creare, gestire e utilizzare un Cloud Server Aruba tramite la CLI acloud.
---
# Esempio Cloud Server

Questo esempio mostra come eseguire il provisioning di un nuovo cloud server utilizzando la CLI ArubaCloud, con tutti i flag di rete richiesti e le istruzioni aggiornate.

## Step 0: Elenca le VPC disponibili

Per prima cosa, individua la VPC che vuoi utilizzare per il tuo cloud server. Elenca tutte le VPC disponibili con:

```bash
acloud network vpc list
```

Scegli una VPC con `STATUS` pari a `Active` e annota il suo `ID` per il prossimo step.

---

## Step 1: Recupera URI e stato della VPC

Assicurati che la VPC sia già creata e che il suo stato sia **Active**.

```bash
acloud network vpc get {vpc-id} | grep -E "URI|Status"
```

Procedi solo se lo stato è `Active`.

---

## Step 2: Elenca o crea una Subnet nella VPC

Per elencare le subnet nella VPC scelta:

```bash
acloud network subnet list {vpc-id}
```

Scegli una subnet con `STATUS` pari a `Active` e annota il suo `ID` e `CIDR`. Se non esiste una subnet adatta, creane una nuova.

---

## Step 3: Verifica l'ID della Subnet

Annota l'`ID` della subnet scelta dall'output del comando list. Lo userai direttamente nel comando di creazione con `--subnet-id`.

```bash
acloud network subnet get <vpc-id> <subnet-id>
```

> **Nota:** Annota l'`ID` della subnet. Procedi solo se lo `STATUS` è `Active`.

---

## Step 4: Elenca o crea un Security Group nella VPC

Per elencare i security group nella VPC:

```bash
acloud network securitygroup list <vpc-id>
```

Scegli un security group con `STATUS` pari a `Active` e annota il suo `ID`. Se non esiste, creane uno nuovo.

---

## Step 5: Annota l'ID dell'Elastic IP (se serve accesso pubblico)

Se vuoi accesso pubblico per il server, annota l'`ID` dell'Elastic IP dall'output di `acloud network elasticip list`. Puoi verificarlo con:

```bash
acloud network elasticip get <elasticip-id>
```

> **Nota:** Annota l'`ID` dell'Elastic IP. Salta questo step se non ti serve accesso pubblico.

---

## Step 6: Crea un Block Storage avviabile (opzionale)

Se vuoi creare un volume di block storage avviabile:

```bash
acloud storage blockstorage create \
  --name boot-ubuntu \
  --region ITBG-Bergamo \
  --zone ITBG-1 \
  --set-bootable \
  --billing-period Hour \
  --size 20 \
  --tags boot \
  --type Performance \
  --image LU22-001
```

Sostituisci i parametri secondo le tue esigenze.

---

## Step 7: Elenca o crea una Keypair attiva

Per elencare le keypair:

```bash
acloud compute keypair list
```

Scegli una keypair con `STATUS` pari a `Active` e annota il suo `ID` o `NAME`. Se non esiste, creane una nuova:

```bash
acloud compute keypair create --name my-keypair --public-key "$(cat ~/.ssh/id_rsa.pub)"
```

---

## Step 8: Verifica l'ID e lo stato del disco di avvio

Annota l'`ID` del block storage avviabile. Il block storage deve essere in stato `NotUsed` prima di poterlo usare come disco di avvio.

```bash
acloud storage blockstorage list
```

Esempio output:
```
NAME         ID                         SIZE(GB)  REGION         ZONE  TYPE         STATUS
boot-ubuntu  697b3a0dce7dfeef9153256a   20        ITBG-Bergamo         Performance  NotUsed
```

> **Nota:** Annota l'`ID` del disco. Procedi solo quando lo stato è `NotUsed` o `Active`.

---

## Step 9: (Opzionale) Crea un file user-data per cloud-init

Se vuoi personalizzare il server all'avvio (ad esempio, creare utenti o installare pacchetti), crea un file `cloud-init.yaml` prima del provisioning. Esempio:

```yaml
cloud-config
users:
  - name: demo
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: users, admin
    home: /home/demo
    shell: /bin/bash
    lock_passwd: false
    passwd: <hashed-password>
```

Sostituisci `<hashed-password>` con una password hash valida (consulta la documentazione di cloud-init per i dettagli su come generarla). Puoi aggiungere altre opzioni cloud-init secondo le tue necessità.

---

## Step 10: Crea il Cloud Server

Usa gli ID raccolti nei passaggi precedenti per creare il cloud server:

```bash
acloud compute cloudserver create \
  --name "my-server" \
  --region "ITBG-Bergamo" \
  --zone "ITBG-1" \
  --flavor "CSO4A8" \
  --boot-disk-id "697b3a0dce7dfeef9153256a" \
  --vpc-id "69495ef64d0cdc87949b71ec" \
  --subnet-id "694ba1737712ac0032dbe50a" \
  --security-group-id "694b05ac4d0cdc87949b75f9" \
  --keypair-id "69007ebf4e7d691466d8621e" \
  --user-data-file "cloud-init.yaml" \
  --billing-period Hour \
  --tags "production"
```

Esempio output:
```
ID                             NAME         FLAVOR    CPU   RAM(GB)   HD(GB)   REGION
697c62605b79733376b3386a       my-server    CSO4A8    4     8         0        ITBG-Bergamo
```

- Sostituisci gli ID con i tuoi valori reali di VPC, subnet, security group, keypair e disco di avvio.
- Il flag `--user-data-file` è opzionale e può essere omesso se non necessario.
- Puoi specificare più subnet o security group usando valori separati da virgola.
- Tutti i flag di rete (`--vpc-id`, `--subnet-id`, `--security-group-id`) e il flag `--zone` sono richiesti.

---

## Step 11: Attendi che il Cloud Server sia Attivo

Una volta eseguito il comando di creazione, puoi controllare lo stato del server con:

```bash
acloud compute cloudserver list | grep "my-server"
```

Quando lo stato è `Active`, il server è pronto all'uso:

```
my-server                 697c62605b79733376b3386a       ITBG-Bergamo    CSO4A8          Active
```

---

## Step 12: Connettersi al Cloud Server

Per connetterti al cloud server, usa il comando `acloud compute cloudserver connect` specificando l'utente in base all'immagine di avvio. Ad esempio, per Ubuntu usa l'utente `ubuntu`:

```bash
acloud compute cloudserver connect 697c62605b79733376b3386a --user ubuntu
```

Esempio output:
```
Connect by running: ssh ubuntu@85.235.152.94
```

> **Nota:**
> Il flag `--user` è obbligatorio e deve corrispondere all'utente predefinito dell'immagine (es. `ubuntu` per Ubuntu, `centos` per CentOS, ecc.).
> Il comando connect mostrerà il comando SSH da utilizzare.

---

## Step 13: Spegnere o Accendere il Cloud Server

Puoi spegnere o accendere il cloud server in qualsiasi momento con i seguenti comandi:

Per spegnere:

```bash
acloud compute cloudserver power-off 697c62605b79733376b3386a
```

Esempio output:
```
Cloud server powered off successfully!
Server: my-server
Status: Updating
```

Per accendere:

```bash
acloud compute cloudserver power-on 697c62605b79733376b3386a
```

Esempio output:
```
Cloud server powered on successfully!
Server: my-server
Status: Updating
```

---

## Step 14: Eliminare il Cloud Server

Per eliminare il cloud server quando non serve più, usa il comando:

```bash
acloud compute cloudserver delete 697c62605b79733376b3386a
```

Ti verrà chiesta conferma:

```
Are you sure you want to delete cloud server '697c62605b79733376b3386a'? (yes/no): yes
Cloud server '697c62605b79733376b3386a' deleted successfully.
```
