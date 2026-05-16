/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package completion

import (
	"bufio"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/user"
	"slices"

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

	stacks, err := config.ListStacks()
	cobra.CheckErr(err)

	return stacks, cobra.ShellCompDirectiveNoFileComp
}

func CompleteNodeName(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return []string{"unix", "windows"}, cobra.ShellCompDirectiveNoFileComp
}

func CompleteDashboardRole(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return []string{"viewer", "operator", "admin"}, cobra.ShellCompDirectiveNoFileComp
}

func CompleteHostName(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	curUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	knownHosts := curUser.HomeDir + "/.ssh/known_hosts"
	if _, err = os.Stat(knownHosts); errors.Is(err, os.ErrNotExist) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	readFile, err := os.Open(knownHosts)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer func() {
		_ = readFile.Close()
	}()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	hosts := map[string]int{}
	for fileScanner.Scan() {
		var host string
		if _, err := fmt.Sscan(fileScanner.Text(), &host); err != nil {
			continue
		}

		hosts[host] = 1
	}
	list := slices.Sorted(maps.Keys(hosts))

	return list, cobra.ShellCompDirectiveNoFileComp
}
