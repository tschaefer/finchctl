/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/agent"
	"github.com/tschaefer/finchctl/internal/target"
)

var registerCmd = &cobra.Command{
	Use:               "register service-name",
	Short:             "Register a new agent with a finch service",
	Args:              cobra.ExactArgs(1),
	Run:               runRegisterCmd,
	ValidArgsFunction: completion.CompleteStackName,
}

func init() {
	registerCmd.Flags().String("agent.hostname", "", "Hostname of the agent")
	registerCmd.Flags().Bool("agent.log.journal", false, "Collect journal logs")
	registerCmd.Flags().Bool("agent.log.docker", false, "Collect docker logs")
	registerCmd.Flags().Bool("agent.metrics", false, "Collect node metrics")
	registerCmd.Flags().StringSlice("agent.metrics.target", nil, "Collect metrics from specific targets")
	registerCmd.Flags().Bool("agent.profiles", false, "Enable profiles collector")
	registerCmd.Flags().StringSlice("agent.log.file", nil, "Collect logs from file paths")
	registerCmd.Flags().StringSlice("agent.labels", nil, "Optional labels for identifying the agent")
	registerCmd.Flags().String("agent.config", "finch-agent.cfg", "Path to the configuration file")
}

func runRegisterCmd(cmd *cobra.Command, args []string) {
	formatType, err := format.GetRunFormat("quiet")
	cobra.CheckErr(err)

	serviceName := args[0]
	data := parseFlags(cmd, formatType)

	a, err := agent.New("", "localhost", formatType, false)
	cobra.CheckErr(err)

	config, err := a.Register(serviceName, data)
	cobra.CheckErr(err)

	configFile, _ := cmd.Flags().GetString("agent.config")

	if err := os.WriteFile(configFile, config, 0644); err != nil {
		cobra.CheckErr("failed to write configuration file: " + err.Error())
	}
}

func parseFlags(cmd *cobra.Command, formatType target.Format) *agent.RegisterData {
	hostname, _ := cmd.Flags().GetString("agent.hostname")
	if hostname == "" {
		errors.CheckErr("agent hostname is required", formatType)
	}

	var logSources []string

	logJournal, _ := cmd.Flags().GetBool("agent.log.journal")
	if logJournal {
		logSources = append(logSources, "journal://")
	}

	logDocker, _ := cmd.Flags().GetBool("agent.log.docker")
	if logDocker {
		logSources = append(logSources, "docker://")
	}

	logFiles, _ := cmd.Flags().GetStringSlice("agent.log.file")
	if len(logFiles) != 0 {
		for _, file := range logFiles {
			logSources = append(logSources, "file://"+file)
		}
	}

	if len(logSources) == 0 {
		errors.CheckErr("at least one log source must be enabled", formatType)
	}

	labels, _ := cmd.Flags().GetStringSlice("agent.labels")
	metrics, _ := cmd.Flags().GetBool("agent.metrics")
	metricsTargets, _ := cmd.Flags().GetStringSlice("agent.metrics.target")
	profiles, _ := cmd.Flags().GetBool("agent.profiles")

	data := &agent.RegisterData{
		Hostname:       hostname,
		LogSources:     logSources,
		Metrics:        metrics,
		MetricsTargets: metricsTargets,
		Profiles:       profiles,
		Labels:         labels,
	}

	return data
}
