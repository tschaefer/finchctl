/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/service"
)

var deregisterCmd = &cobra.Command{
	Use:   "deregister [user@]host[:port]",
	Short: "Deregister the client from a service on a remote host",
	Args:  cobra.ExactArgs(1),
	Run:   runDeregisterCmd,
}

func init() {
	deregisterCmd.Flags().String("run.format", "progress", "output format")
	deregisterCmd.Flags().Bool("run.dry-run", false, "do not deregister, just print the commands that would be run")

	_ = deregisterCmd.RegisterFlagCompletionFunc("run.format", completion.CompleteRunFormat)
}

func runDeregisterCmd(cmd *cobra.Command, args []string) {
	targetUrl := args[0]

	formatName, _ := cmd.Flags().GetString("run.format")
	formatType, err := format.GetRunFormat(formatName)
	cobra.CheckErr(err)
	dryRun, _ := cmd.Flags().GetBool("run.dry-run")

	timeout, _ := cmd.Flags().GetUint("run.cmd-timeout")
	s, err := service.New(cmd.Context(), service.Options{
		TargetURL:  targetUrl,
		Format:     formatType,
		DryRun:     dryRun,
		CmdTimeout: time.Duration(timeout) * time.Second,
	})
	errors.CheckErr(err, formatType)

	err = s.Deregister()
	errors.CheckErr(err, formatType)
}
