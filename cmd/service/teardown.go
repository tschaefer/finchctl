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
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/service"
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
	format, err := format.GetRunFormat(formatName)
	cobra.CheckErr(err)

	dryRun, _ := cmd.Flags().GetBool("run.dry-run")

	cfg, err := teardownConfig(cmd, args)
	if err != nil {
		cobra.CheckErr(err)
	}

	s, err := service.New(cfg, targetUrl, format, dryRun)
	cobra.CheckErr(err)

	err = s.Teardown()
	cobra.CheckErr(err)
}

func teardownConfig(cmd *cobra.Command, args []string) (*service.ServiceConfig, error) {
	config := &service.ServiceConfig{}
	targetUrl := args[0]

	if !strings.HasPrefix(targetUrl, "ssh://") {
		targetUrl = "ssh://" + targetUrl
	}
	target, err := url.Parse(targetUrl)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("invalid target: %w", err))
	}

	hostname, _ := cmd.Flags().GetString("service.host")
	if hostname == "" {
		hostname = target.Hostname()
	}
	config.Hostname = hostname

	return config, nil
}
