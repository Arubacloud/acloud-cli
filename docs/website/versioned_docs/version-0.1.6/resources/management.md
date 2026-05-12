# Management Resources

The `management` category provides commands for managing organization-level resources in Aruba Cloud.

## Available Resources

### [Projects](management/project.md)

Projects are organizational units that group related resources together. They provide a way to organize and manage your cloud resources.

**Quick Commands:**
```bash
# List all projects
acloud management project list

# Get project details
acloud management project get <project-id>

# Create a project
acloud management project create --name "my-project"

# Update a project
acloud management project update <project-id> --description "Updated description"

# Delete a project
acloud management project delete <project-id>
```

## Command Structure

All management commands follow this structure:

```
acloud management <resource> <action> [arguments] [flags]
```

Where:
- `<resource>`: The type of resource (e.g., `project`)
- `<action>`: The operation to perform (e.g., `list`, `get`, `create`, `update`, `delete`)
- `[arguments]`: Required arguments (e.g., resource IDs)
- `[flags]`: Optional flags (e.g., `--name`, `--description`)

## Common Patterns

### Listing Resources

```bash
acloud management <resource> list
```

Lists all resources of the specified type with key information displayed in a table format.

### Getting Resource Details

```bash
acloud management <resource> get <resource-id>
```

Displays detailed information about a specific resource.

### Creating Resources

```bash
acloud management <resource> create --flag1 value1 --flag2 value2
```

Creates a new resource with the specified properties.

### Updating Resources

```bash
acloud management <resource> update <resource-id> --flag1 value1
```

Updates an existing resource. Only provided fields are modified.

### Deleting Resources

```bash
acloud management <resource> delete <resource-id>
```

Deletes the specified resource. May prompt for confirmation.

## Auto-completion

Resource IDs support auto-completion. After setting up shell completion, you can:

```bash
acloud management project get <TAB>
```

This will show available project IDs with their names.

## Next Steps

- [Project Management Guide](management/project.md)
