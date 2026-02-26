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
	"strings"
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

func (l *local) Copy(ctx context.Context, src, dest, mode, owner string) ([]byte, error) {
	PrintProgress(fmt.Sprintf("Copying from '%s' to '%s' as %s@%s", src, dest, l.User, l.Host), l.format)
	if l.dryRun {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, l.cmdTimeout)
	defer cancel()

	installCmd := []string{"sudo", "install", "-m", mode, src, dest}
	if owner != "" {
		parts := strings.SplitN(owner, ":", 2)
		if len(parts) == 2 {
			installCmd = []string{"sudo", "install", "-m", mode, "-o", parts[0], "-g", parts[1], src, dest}
		}
	}
	if out, err := exec.CommandContext(ctx, installCmd[0], installCmd[1:]...).CombinedOutput(); err != nil {
		return out, err
	}

	return nil, nil
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
