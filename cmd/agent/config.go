/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/internal/agent"
)

var configCmd = &cobra.Command{
	Use:               "config service-name",
	Short:             "Download an agent config from a finch service",
	Args:              cobra.ExactArgs(1),
	Run:               runConfigCmd,
	ValidArgsFunction: completeStackName,
}

func init() {
	configCmd.Flags().String("run.format", "progress", "output format")
	configCmd.Flags().String("agent.rid", "", "resource identifier of the agent to config")
	configCmd.Flags().String("agent.config", "finch-agent.cfg", "Path to the configuration file")

	_ = configCmd.RegisterFlagCompletionFunc("run.format", completeRunFormat)
}

func runConfigCmd(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	format, err := getRunFormat(cmd)
	cobra.CheckErr(err)

	rid, _ := cmd.Flags().GetString("agent.rid")
	if rid == "" {
		cobra.CheckErr("agent resource identifier is required")
	}

	agent, err := agent.New("", "local", format, false)
	cobra.CheckErr(err)

	config, err := agent.Config(serviceName, rid)
	cobra.CheckErr(err)

	configFile, _ := cmd.Flags().GetString("agent.config")

	if err := os.WriteFile(configFile, config, 0644); err != nil {
		cobra.CheckErr("failed to write configuration file: " + err.Error())
	}

}
