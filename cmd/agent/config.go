/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/agent"
)

var configCmd = &cobra.Command{
	Use:               "config service-name",
	Short:             "Download an agent config from a finch service",
	Args:              cobra.ExactArgs(1),
	Run:               runConfigCmd,
	ValidArgsFunction: completion.CompleteStackName,
}

func init() {
	configCmd.Flags().String("agent.rid", "", "resource identifier of the agent to config")
	configCmd.Flags().String("agent.config", "finch-agent.cfg", "Path to the configuration file")
}

func runConfigCmd(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	formatType, err := format.GetRunFormat("quiet")
	cobra.CheckErr(err)

	timeout, _ := cmd.Flags().GetUint("run.cmd-timeout")

	rid, _ := cmd.Flags().GetString("agent.rid")
	if rid == "" {
		errors.CheckErr("agent resource identifier is required", formatType)
	}

	agent, err := agent.New(cmd.Context(), agent.Options{
		TargetURL:  "local",
		Format:     formatType,
		CmdTimeout: time.Duration(timeout) * time.Second,
	})
	errors.CheckErr(err, formatType)

	config, err := agent.Config(serviceName, rid)
	errors.CheckErr(err, formatType)

	configFile, _ := cmd.Flags().GetString("agent.config")

	if err := os.WriteFile(configFile, config, 0644); err != nil {
		errors.CheckErr("failed to write configuration file: "+err.Error(), formatType)
	}

}
