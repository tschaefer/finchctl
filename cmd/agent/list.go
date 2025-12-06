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
	"github.com/tschaefer/finchctl/cmd/errors"
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

	formatType, err := format.GetRunFormat("quiet")
	cobra.CheckErr(err)

	a, err := agent.New("", "localhost", formatType, false)
	errors.CheckErr(err, formatType)

	list, err := a.List(serviceName)
	errors.CheckErr(err, formatType)

	if len(*list) == 0 {
		fmt.Println("No agents registered with this service.")
		return
	}

	jsonOutput, _ := cmd.Flags().GetBool("output.json")
	if jsonOutput {
		out, err := json.MarshalIndent(list, "", "  ")
		errors.CheckErr(err, formatType)
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
