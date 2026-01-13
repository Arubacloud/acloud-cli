package cmd

import (
	"github.com/spf13/cobra"
)

var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Manage databases",
	Long:  `Manage database resources in Aruba Cloud.`,
}

func init() {
	rootCmd.AddCommand(databaseCmd)
	// DBaaS commands are registered in database.dbaas.go
	// DBaaS database commands are registered in database.dbaas.database.go
	// DBaaS user commands are registered in database.dbaas.user.go
	// Database backup commands are registered in database.backup.go
}
