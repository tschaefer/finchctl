/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/service"
)

var dashboardCmd = &cobra.Command{
	Use:               "dashboard service-name",
	Short:             "Get service dashboard token",
	Args:              cobra.ExactArgs(1),
	Run:               runDashboardCmd,
	ValidArgsFunction: completion.CompleteStackName,
}

func init() {
	dashboardCmd.Flags().Bool("web", false, "Open dashboard in web browser")
	dashboardCmd.Flags().Int32("permission.session-timeout", 1800, "Session timeout in seconds")
	dashboardCmd.Flags().String("permission.role", "viewer", "Role for the dashboard token (viewer, operator, admin)")
	dashboardCmd.Flags().StringSlice("permission.scope", []string{}, "List of agents to limit the dashboard token access")

	_ = dashboardCmd.RegisterFlagCompletionFunc("permission.role", completion.CompleteDashboardRole)
}

func runDashboardCmd(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	formatType, err := format.GetRunFormat("quiet")
	cobra.CheckErr(err)

	config := &service.ServiceConfig{
		Hostname: serviceName,
	}
	s, err := service.New(config, "localhost", formatType, false)
	errors.CheckErr(err, formatType)

	sessionTimeout, _ := cmd.Flags().GetInt32("permission.session-timeout")
	role, _ := cmd.Flags().GetString("permission.role")
	scope, _ := cmd.Flags().GetStringSlice("permission.scope")
	data, err := s.Dashboard(sessionTimeout, role, scope)
	errors.CheckErr(err, formatType)

	openInWeb, _ := cmd.Flags().GetBool("web")
	if openInWeb {
		url := fmt.Sprintf("%s?token=%s", data.Url, data.Token)

		var err error
		switch runtime.GOOS {
		case "darwin":
			err = exec.Command("open", url).Start()
		case "windows":
			err = exec.Command("cmd", "/c", "start", "", url).Start()
		default:
			err = exec.Command("xdg-open", url).Start()
		}
		errors.CheckErr(err, formatType)
		return
	}

	fmt.Println(data.Token)
}
