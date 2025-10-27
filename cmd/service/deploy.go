/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/cmd/completion"
	"github.com/tschaefer/finchctl/cmd/format"
	"github.com/tschaefer/finchctl/internal/service"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [user@]host[:port]",
	Short: "Deploy Finch service to a remote host",
	Args:  cobra.ExactArgs(1),
	Run:   runDeployCmd,
}

func init() {
	deployCmd.Flags().String("run.format", "progress", "output format")
	deployCmd.Flags().Bool("run.dry-run", false, "do not deploy, just print the commands that would be run")
	deployCmd.Flags().String("service.host", "", "service host (default: auto-detected from target URL)")
	deployCmd.Flags().String("service.user", "", "service user (default: 'admin')")
	deployCmd.Flags().String("service.password", "", "service password (default: random password)")
	deployCmd.Flags().Bool("service.letsencrypt", false, "use Let's Encrypt for TLS certificate (default: false)")
	deployCmd.Flags().String("service.letsencrypt.email", "", "email address for Let's Encrypt registration (required if --service.letsencrypt is true)")
	deployCmd.Flags().Bool("service.customtls", false, "use custom TLS certificate (default: false)")
	deployCmd.Flags().String("service.customtls.cert", "", "path to custom TLS certificate file (required if --service.customtls is true)")
	deployCmd.Flags().String("service.customtls.key", "", "path to custom TLS key file (required if --service.customtls is true)")

	_ = deployCmd.RegisterFlagCompletionFunc("run.format", completion.CompleteRunFormat)
}

func runDeployCmd(cmd *cobra.Command, args []string) {
	targetUrl := args[0]
	config, err := deployConfig(cmd, args)
	if err != nil {
		cobra.CheckErr(err)
	}

	dryRun, _ := cmd.Flags().GetBool("run.dry-run")

	formatName, _ := cmd.Flags().GetString("run.format")
	format, err := format.GetRunFormat(formatName)
	cobra.CheckErr(err)

	s, err := service.New(config, targetUrl, format, dryRun)
	cobra.CheckErr(err)

	err = s.Deploy()
	cobra.CheckErr(err)
}

func deployConfig(cmd *cobra.Command, args []string) (*service.ServiceConfig, error) {
	config := &service.ServiceConfig{}
	targetUrl := args[0]

	if !strings.HasPrefix(targetUrl, "ssh://") {
		targetUrl = "ssh://" + targetUrl
	}
	target, err := url.Parse(targetUrl)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("invalid target: %w", err))
	}

	hostname, _ := cmd.Flags().GetString("service.host")
	if hostname == "" {
		hostname = target.Hostname()
	}
	config.Hostname = hostname

	user, _ := cmd.Flags().GetString("service.user")
	if user == "" {
		user = "admin"
	}
	config.Username = user

	password, _ := cmd.Flags().GetString("service.password")
	if password == "" {
		password = rand.Text()
	}
	config.Password = password

	letsencrypt, _ := cmd.Flags().GetBool("service.letsencrypt")
	letsencryptEmail, _ := cmd.Flags().GetString("service.letsencrypt.email")
	customTLS, _ := cmd.Flags().GetBool("service.customtls")
	customTLSCert, _ := cmd.Flags().GetString("service.customtls.cert")
	customTLSKey, _ := cmd.Flags().GetString("service.customtls.key")

	config.LetsEncrypt.Enabled = letsencrypt
	config.LetsEncrypt.Email = letsencryptEmail
	config.CustomTLS.Enabled = customTLS
	config.CustomTLS.CertFilePath = customTLSCert
	config.CustomTLS.KeyFilePath = customTLSKey

	return config, nil
}
