package cmd

import "fmt"

func msgCreated(kind, name string) string {
	return fmt.Sprintf("%s '%s' created successfully.", kind, name)
}

func msgCreatedAsync(kind, name string) string {
	return fmt.Sprintf("%s '%s' creation initiated. Use 'get' to check status.", kind, name)
}

func msgUpdated(kind, name string) string {
	return fmt.Sprintf("%s '%s' updated successfully.", kind, name)
}

func msgUpdatedAsync(kind, name string) string {
	return fmt.Sprintf("%s '%s' update initiated. Use 'get' to check status.", kind, name)
}

func msgDeleted(kind, name string) string {
	return fmt.Sprintf("%s '%s' deleted successfully.", kind, name)
}

func msgAction(kind, name, verb string) string {
	return fmt.Sprintf("%s '%s' %s successfully.", kind, name, verb)
}

func msgDryRun(kind, id string) string {
	return fmt.Sprintf("[dry-run] Would delete %s '%s'. Resource exists and is accessible.", kind, id)
}
