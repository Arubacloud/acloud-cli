package cmd

import (
	"github.com/spf13/cobra"
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Manage security settings",
	Long:  `Manage security settings in Aruba Cloud.`,
}

func init() {
	rootCmd.AddCommand(securityCmd)
	// KMS commands are registered in security.kms.go
}
