package cmd

import (
	"github.com/spf13/cobra"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage networks",
	Long:  `Manage network resources in Aruba Cloud.`,
}

func init() {
	rootCmd.AddCommand(networkCmd)
}
