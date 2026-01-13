package cmd

import (
	"github.com/spf13/cobra"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage storage resources",
	Long:  `Manage storage resources in Aruba Cloud.`,
}

func init() {
	rootCmd.AddCommand(storageCmd)
}
