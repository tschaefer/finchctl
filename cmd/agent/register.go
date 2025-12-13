/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	PreRun:            runRegisterPreCmd,
	Run:               runRegisterCmd,
	ValidArgsFunction: completion.CompleteStackName,
}

func init() {
	registerCmd.Flags().String("agent.hostname", "", "Hostname of the agent")
	_ = viper.BindPFlag("hostname", registerCmd.Flags().Lookup("agent.hostname"))

	registerCmd.Flags().Bool("agent.log.journal", false, "Collect journal logs")
	_ = viper.BindPFlag("log.journal.enable", registerCmd.Flags().Lookup("agent.log.journal"))

	registerCmd.Flags().Bool("agent.log.docker", false, "Collect docker logs")
	_ = viper.BindPFlag("log.docker.enable", registerCmd.Flags().Lookup("agent.log.docker"))

	registerCmd.Flags().Bool("agent.metrics", false, "Collect node metrics")
	_ = viper.BindPFlag("metrics.enable", registerCmd.Flags().Lookup("agent.metrics"))

	registerCmd.Flags().StringSlice("agent.metrics.targets", nil, "Collect metrics from specific targets")
	_ = viper.BindPFlag("metrics.targets", registerCmd.Flags().Lookup("agent.metrics.target"))

	registerCmd.Flags().Bool("agent.profiles", false, "Enable profiles collector")
	_ = viper.BindPFlag("profiles.enable", registerCmd.Flags().Lookup("agent.profiles"))

	registerCmd.Flags().StringSlice("agent.log.files", nil, "Collect logs from file paths")
	_ = viper.BindPFlag("log.files", registerCmd.Flags().Lookup("agent.log.file"))

	registerCmd.Flags().StringSlice("agent.labels", nil, "Optional labels for identifying the agent")
	_ = viper.BindPFlag("labels", registerCmd.Flags().Lookup("agent.labels"))

	registerCmd.Flags().String("agent.config", "finch-agent.cfg", "Path to the configuration file")
	registerCmd.Flags().String("agent.file", "", "Path to a JSON file containing agent data")
}

func runRegisterPreCmd(cmd *cobra.Command, args []string) {
	agentFile, _ := cmd.Flags().GetString("agent.file")
	err := initConfig(agentFile)
	cobra.CheckErr(err)
}

func runRegisterCmd(cmd *cobra.Command, args []string) {
	formatType, err := format.GetRunFormat("quiet")
	cobra.CheckErr(err)

	serviceName := args[0]
	data := parseFlags(formatType)

	a, err := agent.New("", "localhost", formatType, false)
	cobra.CheckErr(err)

	config, err := a.Register(serviceName, data)
	cobra.CheckErr(err)

	configFile, _ := cmd.Flags().GetString("agent.config")

	if err := os.WriteFile(configFile, config, 0644); err != nil {
		cobra.CheckErr("failed to write configuration file: " + err.Error())
	}
}

func parseFlags(formatType target.Format) *agent.RegisterData {
	hostname := viper.GetString("hostname")
	if hostname == "" {
		errors.CheckErr("agent hostname is required", formatType)
	}

	var logSources []string

	if viper.GetBool("log.journal.enable") {
		logSources = append(logSources, "journal://")
	}

	if viper.GetBool("log.docker.enable") {
		logSources = append(logSources, "docker://")
	}

	logFiles := viper.GetStringSlice("log.files")
	if len(logFiles) != 0 {
		for _, file := range logFiles {
			logSources = append(logSources, "file://"+file)
		}
	}

	if len(logSources) == 0 {
		errors.CheckErr("at least one log source must be enabled", formatType)
	}

	labels := viper.GetStringSlice("labels")
	metrics := viper.GetBool("metrics.enable")
	metricsTargets := viper.GetStringSlice("metrics.targets")
	profiles := viper.GetBool("profiles.enable")

	if len(metricsTargets) != 0 && !viper.IsSet("metrics.enable") {
		metrics = true
	}

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

func initConfig(cfgFile string) error {
	if cfgFile == "" {
		return nil
	}
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if cfgFile != "" {
				return fmt.Errorf("config file not found: %w", err)
			}
		} else {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	return nil
}
