/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/errors"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/agent"
	"github.com/tschaefer/finchctl/internal/target"
)

var editCmd = &cobra.Command{
	Use:               "edit service-name",
	Short:             "Edit agent config for a specific finch service",
	Args:              cobra.ExactArgs(1),
	PreRun:            runEditPreCmd,
	Run:               runEditCmd,
	ValidArgsFunction: completion.CompleteStackName,
}

func init() {
	editCmd.Flags().String("agent.rid", "", "resource identifier of the agent to edit")
	_ = viper.BindPFlag("rid", editCmd.Flags().Lookup("agent.rid"))

	editCmd.Flags().Bool("agent.logs.journal", false, "Collect journal logs")
	_ = viper.BindPFlag("logs.journal.enable", editCmd.Flags().Lookup("agent.logs.journal"))

	editCmd.Flags().Bool("agent.logs.docker", false, "Collect docker logs")
	_ = viper.BindPFlag("logs.docker.enable", editCmd.Flags().Lookup("agent.logs.docker"))

	editCmd.Flags().Bool("agent.metrics", false, "Collect node metrics")
	_ = viper.BindPFlag("metrics.enable", editCmd.Flags().Lookup("agent.metrics"))

	editCmd.Flags().StringSlice("agent.metrics.targets", nil, "Collect metrics from specific targets")
	_ = viper.BindPFlag("metrics.targets", editCmd.Flags().Lookup("agent.metrics.targets"))

	editCmd.Flags().Bool("agent.profiles", false, "Enable profiles collector")
	_ = viper.BindPFlag("profiles.enable", editCmd.Flags().Lookup("agent.profiles"))

	editCmd.Flags().StringSlice("agent.logs.files", nil, "Collect logs from file paths")
	_ = viper.BindPFlag("logs.files", editCmd.Flags().Lookup("agent.logs.files"))

	editCmd.Flags().StringSlice("agent.labels", nil, "Optional labels for identifying the agent")
	_ = viper.BindPFlag("labels", editCmd.Flags().Lookup("agent.labels"))

	editCmd.Flags().String("agent.file", "", "Path to a file containing agent data")
}

func runEditPreCmd(cmd *cobra.Command, args []string) {
	agentFile, _ := cmd.Flags().GetString("agent.file")
	err := editInitConfig(agentFile)
	cobra.CheckErr(err)
}

func runEditCmd(cmd *cobra.Command, args []string) {
	formatType, err := format.GetRunFormat("quiet")
	cobra.CheckErr(err)

	serviceName := args[0]
	data := editParseFlags(formatType)

	a, err := agent.New("", "localhost", formatType, false)
	cobra.CheckErr(err)

	err = a.Edit(serviceName, data)
	cobra.CheckErr(err)
}

func editParseFlags(formatType target.Format) *agent.EditData {
	var logSources []string

	if viper.GetBool("logs.journal.enable") {
		logSources = append(logSources, "journal://")
	}

	if viper.GetBool("logs.docker.enable") {
		logSources = append(logSources, "docker://")
	}

	logFiles := viper.GetStringSlice("logs.files")
	if len(logFiles) != 0 {
		for _, file := range logFiles {
			logSources = append(logSources, "file://"+file)
		}
	}

	if len(logSources) == 0 {
		errors.CheckErr("at least one log source must be enabled", formatType)
	}

	rid := viper.GetString("rid")
	if rid == "" {
		errors.CheckErr("agent resource identifier must be provided", formatType)
	}

	labels := viper.GetStringSlice("labels")
	metrics := viper.GetBool("metrics.enable")
	metricsTargets := viper.GetStringSlice("metrics.targets")
	profiles := viper.GetBool("profiles.enable")

	if len(metricsTargets) != 0 && !viper.IsSet("metrics.enable") {
		metrics = true
	}

	data := &agent.EditData{
		ResourceId:     rid,
		LogSources:     logSources,
		Metrics:        metrics,
		MetricsTargets: metricsTargets,
		Profiles:       profiles,
		Labels:         labels,
	}

	return data
}

func editInitConfig(cfgFile string) error {
	if cfgFile == "" {
		return nil
	}
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return fmt.Errorf("config file not found: %w", err)
		}
		return fmt.Errorf("error reading config file: %w", err)
	}

	return nil
}
