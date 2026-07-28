/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/internal/agent"
	"github.com/tschaefer/finchctl/internal/target"

	"github.com/olekukonko/tablewriter"
)

var doctorCmd = &cobra.Command{
	Use:               "doctor [user@]host[:port]",
	Short:             "Verify target is healthy",
	Args:              cobra.ExactArgs(1),
	Run:               runDoctorCmd,
	ValidArgsFunction: completion.CompleteHostName,
}

func init() {
	doctorCmd.Flags().Bool("output.json", false, "output in JSON format")
	doctorCmd.Flags().Bool("check.ports", false, "check agent listen ports")
	doctorCmd.Flags().Bool("check.optionals", false, "check agent optional tools (remote setup)")
}

func runDoctorCmd(cmd *cobra.Command, args []string) {
	targetUrl := args[0]

	jsonOutput, _ := cmd.Flags().GetBool("output.json")
	if jsonOutput {
		_ = os.Setenv("NO_COLOR", "1")
	}

	timeout, _ := cmd.Flags().GetUint("run.cmd-timeout")
	a, err := agent.New(cmd.Context(), agent.Options{
		TargetURL:  targetUrl,
		Format:     target.FormatQuiet,
		CmdTimeout: time.Duration(timeout) * time.Second,
	})
	errors.CheckErr(err, target.FormatQuiet)

	checkPorts, _ := cmd.Flags().GetBool("check.ports")
	checkOptionals, _ := cmd.Flags().GetBool("check.optionals")
	list, ok := a.Doctor(checkOptionals, checkPorts)

	if jsonOutput {
		out, e := json.MarshalIndent(list, "", "  ")
		errors.CheckErr(e, target.FormatJSON)
		fmt.Println(string(out))
	} else {
		t := tablewriter.NewWriter(os.Stdout)
		t.Header([]string{"Requirement", "Status", "Optional"})
		for _, item := range *list {
			_ = t.Append([]string{item.Requirement, item.Status, strconv.FormatBool(item.Optional)})
		}
		_ = t.Render()
	}

	if !ok {
		os.Exit(1)
	}
}
