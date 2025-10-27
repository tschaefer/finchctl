/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/agent"

	"github.com/olekukonko/tablewriter"
)

var listCmd = &cobra.Command{
	Use:               "list service-name",
	Short:             "List agents registered with a finch service",
	Args:              cobra.ExactArgs(1),
	Run:               runListCmd,
	ValidArgsFunction: completion.CompleteStackName,
}

func init() {
	listCmd.Flags().Bool("output.json", false, "output in JSON format (not implemented yet)")
}

func runListCmd(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	format, err := format.GetRunFormat("quiet")
	cobra.CheckErr(err)

	a, err := agent.New("", "localhost", format, false)
	cobra.CheckErr(err)

	list, err := a.List(serviceName)
	cobra.CheckErr(err)

	jsonOutput, _ := cmd.Flags().GetBool("output.json")
	if jsonOutput {
		out, err := json.MarshalIndent(list, "", "  ")
		if err != nil {
			cobra.CheckErr(err)
		}
		fmt.Println(string(out))
	} else {
		t := tablewriter.NewWriter(os.Stdout)
		t.Header([]string{"#", "Hostname", "Resource Identifier"})
		for i, item := range *list {
			idx := fmt.Sprintf("%d", i+1)
			_ = t.Append([]string{idx, item.Hostname, item.ResourceID})
		}
		_ = t.Render()
	}
}
