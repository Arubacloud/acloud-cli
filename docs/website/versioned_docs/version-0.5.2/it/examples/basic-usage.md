---
id: basic-usage-it
title: Basic Usage
sidebar_label: Basic Usage
description: Learn the essential commands to list, create and manage Aruba Cloud projects using the acloud CLI.
# Basic CLI Usage

---
---
Prossimi Passi

- Scopri come effettuare il deploy di istanze compute
- Esplora la configurazione avanzata dei contesti
- Consulta la reference CLI per tutti i comandi disponibili

## Project Management

### List Projects

To list all projects available in your Aruba Cloud tenant, run:

```bash
acloud management project list
```

Example output:

```text
NAME                              ID                                CREATION DATE
develop                           66a10244f62b99c686572a9f          24-07-2024
github-runner                     68398923fb2cb026400d4d31          30-05-2025
terraform-test-project            69788486c533abe1c22eda36          27-01-2026
```

:::note
You must be authenticated before you can list projects.
If not logged in, run `acloud login`.
:::

### Get Project Details

To view detailed information about a specific project, use its Project ID from the list:

```bash
acloud management project get <project-id>
```

Example:

```bash
acloud management project get 66a10244f62b99c686572a9f
```

Example output:

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

### Create a New Project

To create a new project, specify the desired name and tags:

```bash
acloud management project create --name <project-name> --tags <tag>
```

Example:

```bash
acloud management project create --name test-cli --tags test-cli
```

Example output:

```text
ID                                NAME                              DEFAULT     RESOURCES
6978c3d2c533abe1c22edaee          test-cli                          No          0
```

:::tip
You can assign multiple tags by repeating the `--tags` flag, e.g.
`acloud management project create --name my-app --tags dev --tags backend`.
:::

## Context Management

A context defines which project the CLI interacts with by default, similar to how `kubectl` contexts work in Kubernetes.
This makes it easy to manage multiple Aruba Cloud projects from one configuration.

### Set a New Context

Create and set a new CLI context, associating it with a specific project ID:

```bash
acloud context set <context-name> --project-id <project-id>
```

Example:

```bash
acloud context set my-new-context --project-id 6978c3d2c533abe1c22edaee
```

Example output:

```text
Context 'my-new-context' set with project ID: 6978c3d2c533abe1c22edaee
```

### List and Switch Contexts

You can list all defined CLI contexts and switch between them seamlessly.

#### List Contexts

```bash
acloud context list
```

Example output:

```text
Contexts:
=========
my-prod              Project ID: 66a10244f62b99c686572a9f
my-storage-project   Project ID: 68398923fb2cb026400d4d31
test                 Project ID: 68398923fb2cb026400d4d31
e2e-test-context     Project ID: 68398923fb2cb026400d4d31 *
my-new-context       Project ID: 6978c3d2c533abe1c22edaee

* = current context
```

:::note
The context marked with * is the active one.
:::

#### Switch Contexts

Switch to a different context:

```bash
acloud context use <context-name>
```

Example:

```bash
acloud context use my-prod
```

Example output:

```text
Switched to context 'my-prod'
```

:::tip
You can verify the active context at any time by running:

```bash
acloud context current
```
:::

## Show CLI Version

To check the installed version of the Aruba Cloud CLI:

```bash
acloud version
```

Example output:

```text
Aruba Cloud CLI version 1.2.0
```

---

Next Steps

- Learn how to deploy compute instances
- Explore advanced context configuration
- Visit the CLI reference for all available commands

