/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package target

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type remote struct {
	Host       string
	Port       uint
	User       string
	auth       goph.Auth
	client     *goph.Client
	format     Format
	dryRun     bool
	cmdTimeout time.Duration
}

func (s *remote) Run(ctx context.Context, cmd string) ([]byte, error) {
	PrintProgress(fmt.Sprintf("Running '%s' as %s@%s", cmd, s.User, s.Host), s.format)
	if s.dryRun {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, s.cmdTimeout)
	defer cancel()

	return s.client.RunContext(ctx, cmd)
}

func (s *remote) Copy(ctx context.Context, src, dest, mode, owner string) error {
	PrintProgress(fmt.Sprintf("Copying from '%s' to '%s' as %s@%s", src, dest, s.User, s.Host), s.format)
	if s.dryRun {
		return nil
	}

	raw, err := s.Run(ctx, "mktemp -p /tmp -d finch-XXXXXX")
	if err != nil {
		return err
	}
	tmpdest := strings.TrimSpace(string(raw))
	defer func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = s.Run(cleanCtx, "rm -rf "+tmpdest)
	}()

	uploadCtx, cancel := context.WithTimeout(ctx, s.cmdTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- s.client.Upload(src, tmpdest+"/file")
	}()

	select {
	case err := <-done:
		if err != nil {
			return err
		}
	case <-uploadCtx.Done():
		_ = s.client.Close()
		return fmt.Errorf("upload timed out after %s", s.cmdTimeout)
	}

	_, err = s.Run(ctx, fmt.Sprintf("sudo mv -f %s %s", tmpdest+"/file", dest))
	if err != nil {
		return err
	}

	if mode != "" {
		_, err = s.Run(ctx, fmt.Sprintf("sudo chmod %s %s", mode, dest))
		if err != nil {
			return err
		}
	}

	if owner != "" {
		_, err = s.Run(ctx, fmt.Sprintf("sudo chown %s %s", owner, dest))
		if err != nil {
			return err
		}
	}

	return nil
}

func newRemote(host *url.URL, opts Options) (Target, error) {
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
		Host:       host.Hostname(),
		Port:       port,
		User:       host.User.Username(),
		auth:       auth,
		client:     client,
		format:     opts.Format,
		dryRun:     opts.DryRun,
		cmdTimeout: opts.CmdTimeout,
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
		auth = goph.KeyboardInteractive(password)

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
