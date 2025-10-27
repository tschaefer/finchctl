/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"encoding/json"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/service"
)

var infoCmd = &cobra.Command{
	Use:               "info service-name",
	Short:             "Show detailed information about a finch service",
	Args:              cobra.ExactArgs(1),
	Run:               runInfoCmd,
	ValidArgsFunction: completion.CompleteStackName,
}

func init() {
	infoCmd.Flags().Bool("output.json", false, "output in JSON format (not implemented yet)")
}

func runInfoCmd(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	format, err := format.GetRunFormat("quiet")
	cobra.CheckErr(err)

	config := &service.ServiceConfig{
		Hostname: serviceName,
	}
	s, err := service.New(config, "localhost", format, false)
	cobra.CheckErr(err)

	info, err := s.Info()
	cobra.CheckErr(err)

	jsonOutput, _ := cmd.Flags().GetBool("output.json")
	if jsonOutput {
		out, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			cobra.CheckErr(err)
		}
		fmt.Println(string(out))
		return
	} else {
		t := tablewriter.NewWriter(cmd.OutOrStdout())
		_ = t.Append([]string{"Id", info.ID})
		_ = t.Append([]string{"Hostname", info.Hostname})
		_ = t.Append([]string{"Created At", info.CreatedAt})
		_ = t.Append([]string{"Release", info.Release})
		_ = t.Append([]string{"Commit", info.Commit})
		_ = t.Render()
	}
}
