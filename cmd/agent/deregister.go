/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/agent"
)

var deregisterCmd = &cobra.Command{
	Use:               "deregister service-name",
	Short:             "Deregister an agent from a finch service",
	Args:              cobra.ExactArgs(1),
	Run:               runDeregisterCmd,
	ValidArgsFunction: completion.CompleteStackName,
}

func init() {
	deregisterCmd.Flags().String("agent.rid", "", "resource identifier of the agent to deregister")
}

func runDeregisterCmd(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	formatType, err := format.GetRunFormat("quiet")
	cobra.CheckErr(err)

	rid, _ := cmd.Flags().GetString("agent.rid")
	if rid == "" {
		errors.CheckErr("agent resource identifier is required", formatType)
	}

	agent, err := agent.New("", "local", formatType, false)
	errors.CheckErr(err, formatType)

	err = agent.Deregister(serviceName, rid)
	errors.CheckErr(err, formatType)
}
