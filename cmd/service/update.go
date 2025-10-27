/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/service"
)

var updateCmd = &cobra.Command{
	Use:   "update [user@]host[:port]",
	Short: "Update service on a remote host",
	Args:  cobra.ExactArgs(1),
	Run:   runUpdateCmd,
}

func init() {
	updateCmd.Flags().String("run.format", "progress", "output format")
	updateCmd.Flags().Bool("run.dry-run", false, "do not deploy, just print the commands that would be run")

	_ = updateCmd.RegisterFlagCompletionFunc("run.format", completion.CompleteRunFormat)
}

func runUpdateCmd(cmd *cobra.Command, args []string) {
	targetUrl := args[0]

	formatName, _ := cmd.Flags().GetString("run.format")
	format, err := format.GetRunFormat(formatName)
	cobra.CheckErr(err)
	dryRun, _ := cmd.Flags().GetBool("run.dry-run")

	s, err := service.New(nil, targetUrl, format, dryRun)
	cobra.CheckErr(err)

	err = s.Update()
	cobra.CheckErr(err)
}
