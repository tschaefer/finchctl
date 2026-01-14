/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/agent"
	"github.com/tschaefer/finchctl/cmd/service"
	"github.com/tschaefer/finchctl/internal/grpc"
	"github.com/tschaefer/finchctl/internal/version"
)

var rootCmd = &cobra.Command{
	Use:   "finchctl",
	Short: "A minimal logging infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		v, _ := cmd.Flags().GetBool("version")
		if v {
			version.Print()
			return
		}
		_ = cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		skipTLSVerify, _ := cmd.Flags().GetBool("tls.skip-verify")
		if skipTLSVerify {
			_ = os.Setenv(grpc.SkipTLSVerifyEnv, "true")
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			version.Print()
		},
	}

	rootCmd.PersistentFlags().Bool("tls.skip-verify", false, "Skip TLS certificate verification (not recommended)")

	rootCmd.AddCommand(agent.Cmd)
	rootCmd.AddCommand(service.Cmd)
	rootCmd.AddCommand(versionCmd)
}
