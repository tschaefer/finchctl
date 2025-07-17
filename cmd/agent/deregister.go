/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/internal/agent"
)

var deregisterCmd = &cobra.Command{
	Use:   "deregister service-name",
	Short: "Deregister an agent from a finch service",
	Args:  cobra.ExactArgs(1),
	Run:   runDeregisterCmd,
}

func init() {
	deregisterCmd.Flags().String("run.format", "progress", "output format")
	deregisterCmd.Flags().String("agent.rid", "", "resource identifier of the agent to deregister")

	_ = deregisterCmd.RegisterFlagCompletionFunc("run.format", completeRunFormat)
}

func runDeregisterCmd(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	format, err := getRunFormat(cmd)
	cobra.CheckErr(err)

	rid, _ := cmd.Flags().GetString("agent.rid")
	if rid == "" {
		cobra.CheckErr("agent resource identifier is required")
	}

	agent, err := agent.NewAgent("", "local", format, false)
	cobra.CheckErr(err)

	err = agent.Deregister(serviceName, rid)
	cobra.CheckErr(err)
}
