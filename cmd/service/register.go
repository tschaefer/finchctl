/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/service"
)

var registerCmd = &cobra.Command{
	Use:   "register [user@]host[:port]",
	Short: "Register the client with a service on a remote host",
	Args:  cobra.ExactArgs(1),
	Run:   runRegisterCmd,
}

func init() {
	registerCmd.Flags().String("run.format", "progress", "output format")
	registerCmd.Flags().Bool("run.dry-run", false, "do not register, just print the commands that would be run")

	_ = registerCmd.RegisterFlagCompletionFunc("run.format", completion.CompleteRunFormat)
}

func runRegisterCmd(cmd *cobra.Command, args []string) {
	targetUrl := args[0]

	formatName, _ := cmd.Flags().GetString("run.format")
	formatType, err := format.GetRunFormat(formatName)
	cobra.CheckErr(err)
	dryRun, _ := cmd.Flags().GetBool("run.dry-run")

	s, err := service.New(nil, targetUrl, formatType, dryRun)
	errors.CheckErr(err, formatType)

	err = s.Register()
	errors.CheckErr(err, formatType)
}
