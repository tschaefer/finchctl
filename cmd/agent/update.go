/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"fmt"
	"time"

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
	updateCmd.Flags().Bool("run.dry-run", false, "perform a dry run without updating the agent")
	updateCmd.Flags().Bool("skip.config", false, "skip configuration file update")
	updateCmd.Flags().Bool("skip.binaries", false, "skip binaries update")
	updateCmd.Flags().String("alloy.version", "latest", "version of Alloy to install")

	_ = updateCmd.RegisterFlagCompletionFunc("run.format", completion.CompleteRunFormat)
}

func runUpdateCmd(cmd *cobra.Command, args []string) {
	formatName, _ := cmd.Flags().GetString("run.format")
	formatType, err := format.GetRunFormat(formatName)
	cobra.CheckErr(err)

	config, _ := cmd.Flags().GetString("agent.config")
	skipConfig, _ := cmd.Flags().GetBool("skip.config")
	skipBinaries, _ := cmd.Flags().GetBool("skip.binaries")
	alloyVersion, _ := cmd.Flags().GetString("alloy.version")

	if skipConfig && skipBinaries {
		errors.CheckErr(fmt.Errorf("at least one of --skip.config or --skip.binaries must be false"), formatType)
	}

	if config == "" && !skipConfig {
		errors.CheckErr(fmt.Errorf("agent configuration file must be specified"), formatType)
	}
	dryRun, _ := cmd.Flags().GetBool("run.dry-run")

	targetUrl := args[0]

	timeout, _ := cmd.Flags().GetUint("run.cmd-timeout")
	a, err := agent.New(cmd.Context(), agent.Options{
		Config:     config,
		TargetURL:  targetUrl,
		Format:     formatType,
		DryRun:     dryRun,
		CmdTimeout: time.Duration(timeout) * time.Second,
	})
	errors.CheckErr(err, formatType)

	err = a.Update(skipConfig, skipBinaries, alloyVersion)
	errors.CheckErr(err, formatType)
}
