package cmd

import (
	"github.com/spf13/cobra"
)

var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "Manage compute resources",
	Long:  `Manage compute resources in Aruba Cloud.`,
}

func init() {
	rootCmd.AddCommand(computeCmd)
	// CloudServer commands are registered in compute.cloudserver.go
	// KeyPair commands are registered in compute.keypair.go
}
