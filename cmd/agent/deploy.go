/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/agent"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [user@]host[:port]",
	Short: "Deploy a Finch agent to a remote host",
	Args:  cobra.ExactArgs(1),
	Run:   runDeployCmd,
}

func init() {
	deployCmd.Flags().String("agent.config", "", "path to agent configuration file")
	deployCmd.Flags().String("run.format", "progress", "output format")
	deployCmd.Flags().Bool("run.dry-run", false, "perform a dry run without deploying the agent")

	_ = deployCmd.RegisterFlagCompletionFunc("run.format", completion.CompleteRunFormat)
}

func runDeployCmd(cmd *cobra.Command, args []string) {
	config, _ := cmd.Flags().GetString("agent.config")
	if config == "" {
		cobra.CheckErr(fmt.Errorf("agent configuration file must be specified"))
	}
	dryRun, _ := cmd.Flags().GetBool("run.dry-run")

	formatName, _ := cmd.Flags().GetString("run.format")
	format, err := format.GetRunFormat(formatName)
	cobra.CheckErr(err)

	targetUrl := args[0]

	a, err := agent.New(config, targetUrl, format, dryRun)
	cobra.CheckErr(err)

	err = a.Deploy()
	cobra.CheckErr(err)
}
