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

var rotateCertificateCmd = &cobra.Command{
	Use:   "rotate-certificate [user@]host[:port]",
	Short: "Rotate mTLS certificates of a service on a remote host",
	Args:  cobra.ExactArgs(1),
	Run:   runRotateCertificateCmd,
}

func init() {
	rotateCertificateCmd.Flags().String("run.format", "progress", "output format")
	rotateCertificateCmd.Flags().Bool("run.dry-run", false, "do not rotate certificates, just print the commands that would be run")

	_ = rotateCertificateCmd.RegisterFlagCompletionFunc("run.format", completion.CompleteRunFormat)
}

func runRotateCertificateCmd(cmd *cobra.Command, args []string) {
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

	err = s.RotateCertificate()
	errors.CheckErr(err, formatType)
}
