/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package target

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/tschaefer/finchctl/internal/config"
	"github.com/tschaefer/finchctl/internal/version"
)

type local struct {
	Host   string
	User   string
	format Format
	dryRun bool
}

func (l *local) Run(cmd string) ([]byte, error) {
	printProgress(fmt.Sprintf("Running '%s' as %s@%s", cmd, l.User, l.Host), l.format)
	if l.dryRun {
		return nil, nil
	}

	return exec.Command("sh", "-c", cmd).CombinedOutput()
}

func (l *local) Copy(src, dest, mode, owner string) error {
	printProgress(fmt.Sprintf("Copying from '%s' to '%s' as %s@%s", src, dest, l.User, l.Host), l.format)
	if l.dryRun {
		return nil
	}

	c := exec.Command("sudo", "cp", src, dest)
	c.Stdout = nil
	c.Stderr = nil
	if err := c.Run(); err != nil {
		return err
	}

	if mode != "" {
		c = exec.Command("sudo", "chmod", mode, dest)
		c.Stdout = nil
		c.Stderr = nil
		if err := c.Run(); err != nil {
			return err
		}
	}

	if owner != "" {
		c = exec.Command("sudo", "chown", owner, dest)
		c.Stdout = nil
		c.Stderr = nil
		if err := c.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (l *local) Request(method string, url *url.URL, data []byte) ([]byte, error) {
	printProgress(fmt.Sprintf("%s request to %s on %s@%s", method, url.String(), l.User, l.Host), l.format)

	username, password, err := config.LookupStackAuth(url.Host)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
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
			return nil, err
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
