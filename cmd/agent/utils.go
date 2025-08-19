/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/internal/config"
	"github.com/tschaefer/finchctl/internal/target"
)

func completeRunFormat(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return []cobra.Completion{"documentation", "progress", "quiet"}, cobra.ShellCompDirectiveDefault
}

func completeStackName(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	config, err := config.ReadConfig()
	if err != nil {
		cobra.CheckErr(err)
	}

	var stackNames []string
	for _, stack := range config.Stacks {
		stackNames = append(stackNames, stack.Name)
	}

	return stackNames, cobra.ShellCompDirectiveNoFileComp
}

func getRunFormat(cmd *cobra.Command) (target.Format, error) {
	formatter, _ := cmd.Flags().GetString("run.format")

	var format target.Format
	var err error
	switch formatter {
	case "documentation":
		format = target.FormatDocumentation
	case "quiet":
		format = target.FormatQuiet
	case "progress":
		format = target.FormatProgress
	default:
		err = fmt.Errorf("unknown format %s", formatter)
	}

	return format, err
}
