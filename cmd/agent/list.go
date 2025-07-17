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
	"github.com/tschaefer/finchctl/internal/agent"

	"github.com/jedib0t/go-pretty/v6/table"
)

var listCmd = &cobra.Command{
	Use:   "list service-name",
	Short: "List agents registered with a finch service",
	Args:  cobra.ExactArgs(1),
	Run:   runListCmd,
}

func init() {
	listCmd.Flags().String("run.format", "progress", "output format")
	listCmd.Flags().Bool("run.dry-run", false, "perform a dry run without register the agent")
	listCmd.Flags().Bool("output.json", false, "output in JSON format (not implemented yet)")

	_ = listCmd.RegisterFlagCompletionFunc("run.format", completeRunFormat)
}

func runListCmd(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	dryRun, _ := cmd.Flags().GetBool("run.dry-run")
	format, err := getRunFormat(cmd)
	cobra.CheckErr(err)

	a, err := agent.NewAgent("", "localhost", format, dryRun)
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

		return
	} else {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"#", "Hostname", "Resource Identifier"})
		for i, item := range *list {
			t.AppendRow([]any{i + 1, item.Hostname, item.ResourceID})
		}
		t.Render()
	}
}
