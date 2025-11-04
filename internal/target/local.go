/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package target

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/tschaefer/finchctl/internal/config"
	"github.com/tschaefer/finchctl/internal/version"
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

func (l *local) Request(method string, url *url.URL, data []byte) ([]byte, error) {
	PrintProgress(fmt.Sprintf("%s request to %s on %s@%s", method, url.String(), l.User, l.Host), l.format)
	if l.dryRun {
		return nil, nil
	}

	username, password, err := config.LookupStackAuth(url.Host)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{}
	skipTLSVerify, ok := os.LookupEnv(SkipTLSVerifyEnv)
	if ok && skipTLSVerify == "1" || skipTLSVerify == "true" {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(method, url.String(), io.NopCloser(bytes.NewBuffer(data)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("finchctl/%s", version.Release()))
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var data map[string]string
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(payload))
		}

		return nil, fmt.Errorf("%s", data["detail"])
	}

	return payload, nil
}

func NewLocal(host *url.URL, format Format, dryRun bool) (Target, error) {
	return &local{
		Host:   host.Hostname(),
		User:   host.User.Username(),
		format: format,
		dryRun: dryRun,
	}, nil
}
