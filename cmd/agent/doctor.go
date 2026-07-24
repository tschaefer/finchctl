/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"fmt"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/agent"

	"github.com/olekukonko/tablewriter"
)

var doctorCmd = &cobra.Command{
	Use:               "doctor [user@]host[:port]",
	Short:             "Verify target is ready",
	Args:              cobra.ExactArgs(1),
	Run:               runDoctorCmd,
	ValidArgsFunction: completion.CompleteHostName,
}

func init() {
	doctorCmd.Flags().Bool("output.json", false, "output in JSON format (not implemented yet)")
}

func runDoctorCmd(cmd *cobra.Command, args []string) {
	targetUrl := args[0]

	formatType, err := format.GetRunFormat("quiet")
	cobra.CheckErr(err)

	timeout, _ := cmd.Flags().GetUint("run.cmd-timeout")
	a, err := agent.New(cmd.Context(), agent.Options{
		TargetURL:  targetUrl,
		Format:     formatType,
		CmdTimeout: time.Duration(timeout) * time.Second,
	})
	errors.CheckErr(err, formatType)

	healthy, list := a.Doctor()
	jsonOutput, _ := cmd.Flags().GetBool("output.json")
	if jsonOutput {
		out, err := json.MarshalIndent(list, "", "  ")
		errors.CheckErr(err, formatType)
		fmt.Println(string(out))
	} else {
		t := tablewriter.NewWriter(os.Stdout)
		t.Header([]string{"Requirement", "Status", "Optional"})
		for _, item := range *list {
			_ = t.Append([]string{item.Requirement, item.Status, strconv.FormatBool(item.Optional)})
		}
		_ = t.Render()
	}

	if !healthy {
		os.Exit(1)
	}
}
