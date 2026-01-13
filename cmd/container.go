package cmd

import (
	"github.com/spf13/cobra"
)

var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "Manage containers",
	Long:  `Manage container resources in Aruba Cloud.`,
}

func init() {
	rootCmd.AddCommand(containerCmd)
	// KaaS commands are registered in container.kaas.go
	// ContainerRegistry commands are registered in container.containerregistry.go
}
