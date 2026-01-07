/*
Copyright (c) Tobias Schäfer. All rights reserved.
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

var teardownCmd = &cobra.Command{
	Use:   "teardown [user@]host[:port]",
	Short: "Tear down Finch agent from a remote host",
	Args:  cobra.ExactArgs(1),
	Run:   runTeardownCmd,
}

func init() {
	teardownCmd.Flags().String("run.format", "progress", "output format")
	teardownCmd.Flags().Bool("run.dry-run", false, "perform a dry run without tearing down the agent")

	_ = teardownCmd.RegisterFlagCompletionFunc("run.format", completion.CompleteRunFormat)
}

func runTeardownCmd(cmd *cobra.Command, args []string) {
	formatName, _ := cmd.Flags().GetString("run.format")
	formatType, err := format.GetRunFormat(formatName)
	cobra.CheckErr(err)

	dryRun, _ := cmd.Flags().GetBool("run.dry-run")

	targetUrl := args[0]

	a, err := agent.New("", targetUrl, formatType, dryRun)
	errors.CheckErr(err, formatType)

	err = a.Teardown()
	errors.CheckErr(err, formatType)
}
