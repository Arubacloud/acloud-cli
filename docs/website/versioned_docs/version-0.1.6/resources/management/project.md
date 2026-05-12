# Project Management

Projects are organizational units that help you group and manage related resources in Aruba Cloud. Each project can contain multiple resources and has its own metadata.

## Table of Contents

- [Overview](#overview)
- [Commands](#commands)
  - [List Projects](#list-projects)
  - [Get Project Details](#get-project-details)
  - [Create Project](#create-project)
  - [Update Project](#update-project)
  - [Delete Project](#delete-project)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Overview

A project in Aruba Cloud consists of:
- **ID**: Unique identifier (auto-generated)
- **Name**: Human-readable name
- **Description**: Optional description
- **Tags**: Optional labels for organization
- **Default**: Flag indicating if it's the default project
- **Resources**: Number of resources in the project
- **Creation Date**: When the project was created
- **Created By**: User who created the project

## Commands

### List Projects

Display all projects in a table format.

**Syntax:**
```bash
acloud management project list
```

**Output:**
```
ID                             NAME                                     CREATION DATE   
655b2822af30f667f826994e       defaultproject                           20-11-2023      
66a10244f62b99c686572a9f       develop                                  24-07-2024      
68398923fb2cb026400d4d31       github-runner                            30-05-2025
```

**Options:**
- None

---

### Get Project Details

Retrieve detailed information about a specific project.

**Syntax:**
```bash
acloud management project get <project-id>
```

**Arguments:**
- `project-id` (required): The unique ID of the project

**Example:**
```bash
acloud management project get 655b2822af30f667f826994e
```

**Output:**
```
Project Details:
================
ID:              655b2822af30f667f826994e
Name:            defaultproject
Description:     defaultproject
Default:         false
Resources:       5
Creation Date:   20-11-2023 09:34:26
Created By:      aru-297647
Tags:            [production arubacloud-sdk]
```

**Auto-completion:**
```bash
acloud management project get <TAB>
# Shows list of project IDs with names
```

---

### Create Project

Create a new project with specified properties.

**Syntax:**
```bash
acloud management project create --name <name> [flags]
```

**Required Flags:**
- `--name <string>`: Name for the project

**Optional Flags:**
- `--description <string>`: Description for the project
- `--tags <tag1,tag2>`: Comma-separated tags
- `--default`: Set as default project (default: false)

**Examples:**

1. **Basic creation:**
   ```bash
   acloud management project create --name "my-project"
   ```

2. **With description:**
   ```bash
   acloud management project create \
     --name "production-env" \
     --description "Production environment resources"
   ```

3. **With tags:**
   ```bash
   acloud management project create \
     --name "dev-project" \
     --description "Development environment" \
     --tags dev,testing,internal
   ```

4. **Set as default:**
   ```bash
   acloud management project create \
     --name "main-project" \
     --default
   ```

**Output:**
```
Project created successfully!
ID:              69440ae8914afa1ec8b607c1
Name:            my-project
Description:     Production environment resources
Tags:            [dev testing internal]
Default:         false
Creation Date:   18-12-2025 14:08:40
```

---

### Update Project

Update an existing project's description and/or tags.

**Syntax:**
```bash
acloud management project update <project-id> [flags]
```

**Arguments:**
- `project-id` (required): The unique ID of the project

**Flags:**
- `--description <string>`: New description for the project
- `--tags <tag1,tag2>`: New tags for the project (replaces existing)

**Note:** At least one flag must be provided. Name and default status cannot be changed after creation.

**Examples:**

1. **Update description:**
   ```bash
   acloud management project update 69137e295956b621e2048eab \
     --description "Updated description"
   ```

2. **Update tags:**
   ```bash
   acloud management project update 69137e295956b621e2048eab \
     --tags production,critical,monitored
   ```

3. **Update both:**
   ```bash
   acloud management project update 69137e295956b621e2048eab \
     --description "Production environment" \
     --tags prod,active
   ```

**Output:**
```
Project updated successfully!
ID:              69137e295956b621e2048eab
Name:            seca-sdk-example
Description:     Updated description
Tags:            [production critical monitored]
Default:         false
```

**Auto-completion:**
```bash
acloud management project update <TAB>
# Shows list of project IDs with names
```

---

### Delete Project

Delete an existing project.

**Syntax:**
```bash
acloud management project delete <project-id> [flags]
```

**Arguments:**
- `project-id` (required): The unique ID of the project to delete

**Flags:**
- `-y, --yes`: Skip confirmation prompt

**Examples:**

1. **With confirmation:**
   ```bash
   acloud management project delete 69440ae8914afa1ec8b607c1
   ```
   
   Output:
   ```
   Are you sure you want to delete project 69440ae8914afa1ec8b607c1? (yes/no): yes
   
   Project 69440ae8914afa1ec8b607c1 deleted successfully!
   ```

2. **Skip confirmation:**
   ```bash
   acloud management project delete 69440ae8914afa1ec8b607c1 --yes
   ```
   
   Output:
   ```
   Project 69440ae8914afa1ec8b607c1 deleted successfully!
   ```

**Auto-completion:**
```bash
acloud management project delete <TAB>
# Shows list of project IDs with names
```

**Warning:** Deleting a project is permanent and cannot be undone. Ensure all resources in the project are properly handled before deletion.

---

## Examples

### Complete Workflow

1. **Create a new project:**
   ```bash
   acloud management project create \
     --name "webapp-project" \
     --description "Web application resources" \
     --tags web,production,frontend
   ```

2. **List projects to verify:**
   ```bash
   acloud management project list
   ```

3. **Get the new project details:**
   ```bash
   acloud management project get 69440ae8914afa1ec8b607c1
   ```

4. **Update project description:**
   ```bash
   acloud management project update 69440ae8914afa1ec8b607c1 \
     --description "Web application production environment"
   ```

5. **Add more tags:**
   ```bash
   acloud management project update 69440ae8914afa1ec8b607c1 \
     --tags web,production,frontend,monitored,critical
   ```

6. **When no longer needed, delete:**
   ```bash
   acloud management project delete 69440ae8914afa1ec8b607c1 --yes
   ```

### Using Project IDs in Other Commands

Many resource commands require a `--project-id` flag. You can use completion to find the right project:

```bash
# Example: Creating a compute resource in a specific project
acloud compute cloudserver create \
  --project-id <TAB>  # Auto-complete shows your projects
  --name "my-server" \
  --flavor "small"
```

### Filtering Projects by Tags

While the CLI doesn't have built-in filtering, you can use standard Unix tools:

```bash
# Save project list
acloud management project list > projects.txt

# Use grep to search
grep "production" projects.txt
```

## Best Practices

### Naming Conventions

Use clear, descriptive names:
- ✅ `production-webapp`
- ✅ `dev-testing-env`
- ✅ `staging-api-services`
- ❌ `proj1`
- ❌ `test`
- ❌ `temp`

### Tagging Strategy

Use consistent tags for organization:
- **Environment**: `dev`, `staging`, `production`
- **Team**: `frontend`, `backend`, `devops`
- **Status**: `active`, `archived`, `deprecated`
- **Cost Center**: `engineering`, `marketing`, `sales`

Example:
```bash
acloud management project create \
  --name "api-production" \
  --tags production,backend,active,engineering
```

### Description Guidelines

Include useful information in descriptions:
- Purpose of the project
- Team or owner
- Important notes

Example:
```bash
acloud management project create \
  --name "customer-api" \
  --description "Customer-facing REST API - Backend Team - Production Critical"
```

### Resource Organization

- Group related resources in the same project
- Use separate projects for different environments (dev, staging, prod)
- Don't mix unrelated resources in the same project

### Cleanup

- Regularly review and delete unused projects
- Archive old projects by updating their description
- Use tags to identify projects that can be deleted

## Troubleshooting

### "Error initializing client"

Ensure you've configured your credentials:
```bash
acloud config set
```

### "Project not found"

Verify the project ID exists:
```bash
acloud management project list
```

### "Error: --name is required"

The `--name` flag is mandatory when creating projects:
```bash
acloud management project create --name "my-project"
```

### "Error: at least one of --description or --tags must be provided"

When updating, provide at least one field to update:
```bash
acloud management project update <id> --description "New description"
```

## Related Resources

- [Installation Guide](../../installation.md)
- [Management Resources Overview](../management.md)
- [API Documentation](https://www.arubacloud.com/docs)
