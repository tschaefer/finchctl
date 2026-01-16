/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
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
	Cmd.AddCommand(dashboardCmd)
}
