/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package agent

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage Finch agent",
}

func init() {
	Cmd.AddCommand(deployCmd)
	Cmd.AddCommand(teardownCmd)
	Cmd.AddCommand(registerCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(deregisterCmd)
	Cmd.AddCommand(configCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(describeCmd)
	Cmd.AddCommand(editCmd)
}
