package cmd

import (
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage schedules",
	Long:  `Manage schedules in Aruba Cloud.`,
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	// Job commands are registered in schedule.job.go
}
