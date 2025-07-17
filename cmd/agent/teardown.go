/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"github.com/spf13/cobra"
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

	_ = teardownCmd.RegisterFlagCompletionFunc("run.format", completeRunFormat)
}

func runTeardownCmd(cmd *cobra.Command, args []string) {
	dryRun, _ := cmd.Flags().GetBool("run.dry-run")
	format, err := getRunFormat(cmd)
	cobra.CheckErr(err)
	targetUrl := args[0]

	a, err := agent.NewAgent("", targetUrl, format, dryRun)
	cobra.CheckErr(err)

	err = a.Teardown()
	cobra.CheckErr(err)
}
