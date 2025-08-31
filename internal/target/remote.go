/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package target

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type remote struct {
	Host   string
	Port   uint
	User   string
	auth   goph.Auth
	client *goph.Client
	format Format
	dryRun bool
}

func (s *remote) Run(cmd string) ([]byte, error) {
	printProgress(fmt.Sprintf("Running '%s' as %s@%s", cmd, s.User, s.Host), s.format)
	if s.dryRun {
		return nil, nil
	}

	return s.client.Run(cmd)
}

func (s *remote) Copy(src, dest, mode, owner string) error {
	printProgress(fmt.Sprintf("Copying from '%s' to '%s' as %s@%s", src, dest, s.User, s.Host), s.format)
	if s.dryRun {
		return nil
	}

	raw, err := s.Run("mktemp -p /tmp -d finch-XXXXXX")
	if err != nil {
		return err
	}
	tmpdest := strings.TrimSpace(string(raw))
	defer func() {
		_, _ = s.Run("rm -rf " + tmpdest)
	}()

	if err := s.client.Upload(src, tmpdest+"/file"); err != nil {
		return err
	}

	_, err = s.Run(fmt.Sprintf("sudo mv %s %s", tmpdest+"/file", dest))
	if err != nil {
		return err
	}

	if mode != "" {
		_, err = s.Run(fmt.Sprintf("sudo chmod %s %s", mode, dest))
		if err != nil {
			return err
		}
	}

	if owner != "" {
		_, err = s.Run(fmt.Sprintf("sudo chown %s %s", owner, dest))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *remote) Request(method string, url *url.URL, data []byte) ([]byte, error) {
	return nil, fmt.Errorf("Request method not implemented")
}

func NewRemote(host *url.URL, format Format, dryRun bool) (Target, error) {
	auth, err := authorize()
	if err != nil {
		return nil, fmt.Errorf("failed to authorize: %w", err)
	}

	port, err := func() (uint, error) {
		if host.Port() == "" {
			return 22, nil
		}
		port, err := strconv.Atoi(host.Port())
		if err != nil {
			return 0, err
		}
		return uint(port), nil
	}()
	if err != nil {
		return nil, err
	}

	client, err := goph.NewConn(&goph.Config{
		User:     host.User.Username(),
		Addr:     host.Hostname(),
		Port:     port,
		Auth:     auth,
		Callback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, err
	}

	return &remote{
		Host:   host.Hostname(),
		Port:   port,
		User:   host.User.Username(),
		auth:   auth,
		client: client,
		format: format,
		dryRun: dryRun,
	}, nil
}

func authorize() (goph.Auth, error) {
	var auth goph.Auth
	var err error

	switch {

	case goph.HasAgent():
		auth, err = goph.UseAgent()
		if err != nil {
			return nil, err
		}

	default:
		password, err := ask("Enter SSH password: ")
		if err != nil {
			return nil, err
		}
		auth = goph.Password(password)

	}

	return auth, nil
}

func ask(prompt string) (string, error) {
	fmt.Print(prompt)
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println("")
	return strings.TrimSpace(string(pass)), nil
}
