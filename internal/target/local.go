/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package target

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"time"
)

type local struct {
	Host       string
	User       string
	format     Format
	dryRun     bool
	cmdTimeout time.Duration
}

func (l *local) Run(ctx context.Context, cmd string) ([]byte, error) {
	PrintProgress(fmt.Sprintf("Running '%s' as %s@%s", cmd, l.User, l.Host), l.format)
	if l.dryRun {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, l.cmdTimeout)
	defer cancel()

	return exec.CommandContext(ctx, "sh", "-c", cmd).CombinedOutput()
}

func (l *local) Copy(ctx context.Context, src, dest, mode, owner string) error {
	PrintProgress(fmt.Sprintf("Copying from '%s' to '%s' as %s@%s", src, dest, l.User, l.Host), l.format)
	if l.dryRun {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, l.cmdTimeout)
	defer cancel()

	c := exec.CommandContext(ctx, "sudo", "cp", "-f", src, dest)
	c.Stdout = nil
	c.Stderr = nil
	if err := c.Run(); err != nil {
		return err
	}

	if mode != "" {
		c = exec.CommandContext(ctx, "sudo", "chmod", mode, dest)
		c.Stdout = nil
		c.Stderr = nil
		if err := c.Run(); err != nil {
			return err
		}
	}

	if owner != "" {
		c = exec.CommandContext(ctx, "sudo", "chown", owner, dest)
		c.Stdout = nil
		c.Stderr = nil
		if err := c.Run(); err != nil {
			return err
		}
	}

	return nil
}

func newLocal(host *url.URL, opts Options) (Target, error) {
	return &local{
		Host:       host.Hostname(),
		User:       host.User.Username(),
		format:     opts.Format,
		dryRun:     opts.DryRun,
		cmdTimeout: opts.CmdTimeout,
	}, nil
}
