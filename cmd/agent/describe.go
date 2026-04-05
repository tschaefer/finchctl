/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"time"

	"github.com/goccy/go-yaml"
	"github.com/olekukonko/tablewriter"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/agent"
)

var describeCmd = &cobra.Command{
	Use:               "describe service-name",
	Short:             "Get an agent description from a finch service",
	Args:              cobra.ExactArgs(1),
	Run:               runDescribeCmd,
	ValidArgsFunction: completion.CompleteStackName,
}

func init() {
	describeCmd.Flags().String("agent.rid", "", "resource identifier of the agent to config")
	describeCmd.Flags().Bool("output.json", false, "output in JSON format (not implemented yet)")
	describeCmd.Flags().Bool("output.yaml", false, "output in YAML format (not implemented yet)")
}

func runDescribeCmd(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	formatType, err := format.GetRunFormat("quiet")
	cobra.CheckErr(err)

	timeout, _ := cmd.Flags().GetUint("run.cmd-timeout")

	rid, _ := cmd.Flags().GetString("agent.rid")
	if rid == "" {
		errors.CheckErr("agent resource identifier is required", formatType)
	}

	agent, err := agent.New(cmd.Context(), agent.Options{
		TargetURL:  "local",
		Format:     formatType,
		CmdTimeout: time.Duration(timeout) * time.Second,
	})
	errors.CheckErr(err, formatType)

	desc, err := agent.Describe(serviceName, rid)
	errors.CheckErr(err, formatType)

	jsonOutput, _ := cmd.Flags().GetBool("output.json")
	yamlOutput, _ := cmd.Flags().GetBool("output.yaml")

	if yamlOutput && jsonOutput {
		errors.CheckErr("only one output format can be specified", formatType)
	}

	if jsonOutput {
		out, err := json.MarshalIndent(desc, "", "  ")
		errors.CheckErr(err, formatType)
		fmt.Println(string(out))
	} else if yamlOutput {
		out, err := yaml.Marshal(desc)
		errors.CheckErr(err, formatType)
		fmt.Print(string(out))
	} else {
		t := tablewriter.NewWriter(os.Stdout)
		t.Header([]string{"Property", "Value"})
		_ = t.Append([]string{"Hostname", desc.Hostname})
		_ = t.Append([]string{"Node", desc.Node})
		if (len(desc.Labels)) > 0 {
			_ = t.Append([]string{"Labels", strings.Join(desc.Labels, ", ")})
		}
		_ = t.Append([]string{"Docker", fmt.Sprintf("%v", desc.Logs.Docker.Enable)})
		_ = t.Append([]string{"Journal", fmt.Sprintf("%v", desc.Logs.Journal.Enable)})
		if (len(desc.Logs.Files)) > 0 {
			_ = t.Append([]string{"Files", strings.Join(desc.Logs.Files, "\n")})
		}
		if (len(desc.Logs.Events)) > 0 {
			_ = t.Append([]string{"Events", strings.Join(desc.Logs.Events, "\n")})
		}
		_ = t.Append([]string{"Metrics", fmt.Sprintf("%v", desc.Metrics.Enable)})
		if (len(desc.Metrics.Targets)) > 0 {
			_ = t.Append([]string{"Metrics targets", strings.Join(desc.Metrics.Targets, "\n")})
		}
		_ = t.Append([]string{"Profiles", fmt.Sprintf("%v", desc.Profiles.Enable)})
		_ = t.Render()
	}
}
