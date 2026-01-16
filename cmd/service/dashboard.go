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
)

var dashboardCmd = &cobra.Command{
	Use:               "dashboard service-name",
	Short:             "Get dashboard URL for a service",
	Args:              cobra.ExactArgs(1),
	Run:               runDashboardCmd,
	ValidArgsFunction: completion.CompleteStackName,
}

func init() {
	dashboardCmd.Flags().Bool("web", false, "Open dashboard in web browser")
}

func runDashboardCmd(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	url := fmt.Sprintf("https://%s/grafana", serviceName)
	openInWeb, _ := cmd.Flags().GetBool("web")
	if !openInWeb {
		fmt.Println(url)
		return
	}

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "windows":
		err = exec.Command("cmd", "/c", "start", "", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	cobra.CheckErr(err)
}
