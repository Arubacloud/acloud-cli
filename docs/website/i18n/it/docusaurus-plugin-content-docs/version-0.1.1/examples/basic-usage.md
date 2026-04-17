---
id: basic-usage
sidebar_label: Uso Base
title: Uso Base
description: Scopri i comandi essenziali per elencare, creare e gestire progetti Aruba Cloud usando acloud CLI.
---

# Uso Base della CLI

La **Aruba Cloud CLI** (`acloud`) ti permette di gestire l'ambiente Aruba Cloud direttamente da linea di comando.  
Con essa puoi **elencare**, **ispezionare** e **creare** progetti, oltre a **configurare e cambiare** tra diversi contesti CLI.

---

## Gestione Progetti

### Elenca Progetti

Per elencare tutti i progetti disponibili nel tuo tenant Aruba Cloud, esegui:

```bash
acloud management project list
```

Esempio output:

```text
NAME                              ID                                CREATION DATE
develop                           66a10244f62b99c686572a9f          24-07-2024
github-runner                     68398923fb2cb026400d4d31          30-05-2025
terraform-test-project            69788486c533abe1c22eda36          27-01-2026
```

:::note
Devi essere autenticato per elencare i progetti.
Se non hai effettuato l'accesso, esegui `acloud login`.
:::

### Dettagli Progetto

Per visualizzare i dettagli di un progetto specifico, usa il suo Project ID:

```bash
acloud management project get <project-id>
```

Esempio:

```bash
acloud management project get 66a10244f62b99c686572a9f
```

Esempio output:

```text
Project Details:
================
ID:               66a10244f62b99c686572a9f
Name:             develop
Description:      this is for my dev environment
Default:          false
Resources:        4
Creation Date:    24-07-2024 13:31:48
Created By:       aru-297647
Tags:             [develop]
```

### Crea un Nuovo Progetto

Per creare un nuovo progetto, specifica nome e tag desiderati:

```bash
acloud management project create --name <project-name> --tags <tag>
```

Esempio:

```bash
acloud management project create --name test-cli --tags test-cli
```

Esempio output:

```text
ID                                NAME                              DEFAULT     RESOURCES
6978c3d2c533abe1c22edaee          test-cli                          No          0
```

:::tip
Puoi assegnare più tag ripetendo il flag `--tags`, ad esempio:
`acloud management project create --name my-app --tags dev --tags backend`.
:::

## Gestione Contesti

Un contesto definisce con quale progetto la CLI interagisce di default, in modo simile ai contesti di `kubectl` in Kubernetes.
Questo rende semplice gestire più progetti Aruba Cloud da una sola configurazione.

### Imposta un Nuovo Contesto

Crea e imposta un nuovo contesto CLI, associandolo a uno specifico Project ID:

```bash
acloud context set <context-name> --project-id <project-id>
```

Esempio:

```bash
acloud context set my-new-context --project-id 6978c3d2c533abe1c22edaee
```

Esempio output:

```text
Context 'my-new-context' set with project ID: 6978c3d2c533abe1c22edaee
```

### Elenca e Cambia Contesto

Puoi elencare tutti i contesti CLI definiti e cambiare tra essi facilmente.

#### Elenca Contesti

```bash
acloud context list
```

Esempio output:

```text
Contexts:
=========
my-prod              Project ID: 66a10244f62b99c686572a9f
my-storage-project   Project ID: 68398923fb2cb026400d4d31
test                 Project ID: 68398923fb2cb026400d4d31
e2e-test-context     Project ID: 68398923fb2cb026400d4d31 *
my-new-context       Project ID: 6978c3d2c533abe1c22edaee

* = contesto attivo
```

:::note
Il contesto contrassegnato con * è quello attivo.
:::

#### Cambia Contesto

Cambia contesto:

```bash
acloud context use <context-name>
```

Esempio:

```bash
acloud context use my-prod
```

Esempio output:

```text
Switched to context 'my-prod'
```

:::tip
Puoi verificare il contesto attivo in qualsiasi momento con:

```bash
acloud context current
```
:::

## Mostra Versione CLI

Per controllare la versione installata della Aruba Cloud CLI:

```bash
acloud version
```

Esempio output:

```text
Aruba Cloud CLI version 1.2.0
```

---

Prossimi Passi

- Scopri come effettuare il deploy di istanze compute
- Esplora la configurazione avanzata dei contesti
- Consulta la reference CLI per tutti i comandi disponibili
