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

const (
	SkipTLSVerifyEnv string = "FINCH_SKIP_TLS_VERIFY"
)

type local struct {
	Host   string
	User   string
	format Format
	dryRun bool
}

func (l *local) Run(cmd string) ([]byte, error) {
	PrintProgress(fmt.Sprintf("Running '%s' as %s@%s", cmd, l.User, l.Host), l.format)
	if l.dryRun {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	return exec.CommandContext(ctx, "sh", "-c", cmd).CombinedOutput()
}

func (l *local) Copy(src, dest, mode, owner string) error {
	PrintProgress(fmt.Sprintf("Copying from '%s' to '%s' as %s@%s", src, dest, l.User, l.Host), l.format)
	if l.dryRun {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	c := exec.CommandContext(ctx, "sudo", "cp", src, dest)
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

func NewLocal(host *url.URL, format Format, dryRun bool) (Target, error) {
	return &local{
		Host:   host.Hostname(),
		User:   host.User.Username(),
		format: format,
		dryRun: dryRun,
	}, nil
}
