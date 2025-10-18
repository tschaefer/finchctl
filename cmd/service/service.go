package service

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "service",
	Short: "Manage Finch service",
}

func init() {
	Cmd.AddCommand(deployCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(teardownCmd)
	Cmd.AddCommand(infoCmd)
}
