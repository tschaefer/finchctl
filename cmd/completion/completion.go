/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package completion

import (
	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/internal/config"
)

func CompleteRunFormat(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return []cobra.Completion{"documentation", "json", "progress", "quiet"}, cobra.ShellCompDirectiveDefault
}

func CompleteStackName(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
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
