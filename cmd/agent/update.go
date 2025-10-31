/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/agent"
)

var updateCmd = &cobra.Command{
	Use:   "update [user@]host[:port]",
	Short: "Update a Finch agent on a remote host",
	Args:  cobra.ExactArgs(1),
	Run:   runUpdateCmd,
}

func init() {
	updateCmd.Flags().String("agent.config", "", "path to agent configuration file")
	updateCmd.Flags().String("run.format", "progress", "output format")
	updateCmd.Flags().Bool("run.dry-run", false, "perform a dry run without updateing the agent")

	_ = updateCmd.RegisterFlagCompletionFunc("run.format", completion.CompleteRunFormat)
}

func runUpdateCmd(cmd *cobra.Command, args []string) {
	formatName, _ := cmd.Flags().GetString("run.format")
	formatType, err := format.GetRunFormat(formatName)
	cobra.CheckErr(err)

	config, _ := cmd.Flags().GetString("agent.config")
	if config == "" {
		errors.CheckErr(fmt.Errorf("agent configuration file must be specified"), formatType)
	}
	dryRun, _ := cmd.Flags().GetBool("run.dry-run")

	targetUrl := args[0]

	a, err := agent.New(config, targetUrl, formatType, dryRun)
	errors.CheckErr(err, formatType)

	err = a.Update()
	errors.CheckErr(err, formatType)
}
