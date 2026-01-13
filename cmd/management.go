package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(managementCmd)

}

var managementCmd = &cobra.Command{
	Use:   "management",
	Short: "Manage organization resources",
	Long:  `Manage organization resources in Aruba Cloud, like projects.`,
}
