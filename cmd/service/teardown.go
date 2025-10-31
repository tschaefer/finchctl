/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/service"
	"github.com/tschaefer/finchctl/internal/target"
)

var teardownCmd = &cobra.Command{
	Use:   "teardown [user@]host[:port]",
	Short: "Tear down Finch service from a remote host",
	Args:  cobra.ExactArgs(1),
	Run:   runTeardownCmd,
}

func init() {
	teardownCmd.Flags().String("run.format", "progress", "output format")
	teardownCmd.Flags().Bool("run.dry-run", false, "do not deploy, just print the commands that would be run")
	teardownCmd.Flags().String("service.host", "", "service host (default: auto-detected from target URL)")

	_ = teardownCmd.RegisterFlagCompletionFunc("run.format", completion.CompleteRunFormat)
}

func runTeardownCmd(cmd *cobra.Command, args []string) {
	targetUrl := args[0]

	formatName, _ := cmd.Flags().GetString("run.format")
	formatType, err := format.GetRunFormat(formatName)
	cobra.CheckErr(err)

	dryRun, _ := cmd.Flags().GetBool("run.dry-run")

	cfg, err := teardownConfig(cmd, args, formatType)
	errors.CheckErr(err, formatType)

	s, err := service.New(cfg, targetUrl, formatType, dryRun)
	errors.CheckErr(err, formatType)

	err = s.Teardown()
	errors.CheckErr(err, formatType)
}

func teardownConfig(cmd *cobra.Command, args []string, formatType target.Format) (*service.ServiceConfig, error) {
	config := &service.ServiceConfig{}
	targetUrl := args[0]

	if !strings.HasPrefix(targetUrl, "ssh://") {
		targetUrl = "ssh://" + targetUrl
	}
	target, err := url.Parse(targetUrl)
	errors.CheckErr(fmt.Errorf("invalid target: %w", err), formatType)

	hostname, _ := cmd.Flags().GetString("service.host")
	if hostname == "" {
		hostname = target.Hostname()
	}
	config.Hostname = hostname

	return config, nil
}
