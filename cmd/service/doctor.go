/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/internal/service"
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
}

func runDoctorCmd(cmd *cobra.Command, args []string) {
	targetUrl := args[0]

	jsonOutput, _ := cmd.Flags().GetBool("output.json")
	if jsonOutput {
		_ = os.Setenv("NO_COLOR", "1")
	}

	cfg, err := doctorConfig(cmd, args, target.FormatQuiet)
	errors.CheckErr(err, target.FormatQuiet)

	timeout, _ := cmd.Flags().GetUint("run.cmd-timeout")
	s, err := service.New(cmd.Context(), service.Options{
		Config:     cfg,
		TargetURL:  targetUrl,
		Format:     target.FormatQuiet,
		DryRun:     false,
		CmdTimeout: time.Duration(timeout) * time.Second,
	})
	errors.CheckErr(err, target.FormatQuiet)

	list, ok := s.Doctor()

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

func doctorConfig(cmd *cobra.Command, args []string, formatType target.Format) (*service.ServiceConfig, error) {
	config := &service.ServiceConfig{}
	targetUrl := args[0]

	if !strings.HasPrefix(targetUrl, "ssh://") {
		targetUrl = "ssh://" + targetUrl
	}
	target, err := url.Parse(targetUrl)
	if err != nil {
		errors.CheckErr(fmt.Errorf("invalid target: %w", err), formatType)
	}

	hostname, _ := cmd.Flags().GetString("service.host")
	if hostname == "" {
		hostname = target.Hostname()
	}
	config.Hostname = hostname

	return config, nil
}
