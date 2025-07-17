/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"archive/zip"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/tschaefer/finchctl/internal/target"
)

func (a *agent) serviceInstall(machine map[string]string) error {
	if a.dryRun {
		fmt.Println("Dry run: skipping service installation")
		return nil
	}

	binary := fmt.Sprintf("alloy-%s-%s", machine["kernel"], machine["arch"])
	release := fmt.Sprintf("https://github.com/grafana/alloy/releases/latest/download/%s.zip", binary)
	dest := "/usr/bin/alloy"

	user := os.Getenv("USER")
	localhost, _ := url.Parse("host://localhost")
	localhost.User = url.User(user)
	l, _ := target.NewLocal(localhost, a.format, a.dryRun)

	out, err := l.Run("mktemp -p /tmp -d finch-XXXXXX")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: string(out)}
	}
	tmpdir := strings.TrimSpace(string(out))
	defer func() {
		_, _ = l.Run("rm -rf " + tmpdir)
	}()
	tmpfile := fmt.Sprintf("%s/%s-%s.zip", tmpdir, binary, time.Now().Format("19800212015200"))

	// Using curl instead of net/http is on purpose for now.
	out, err = l.Run("curl --silent --fail-with-body --show-error --location " + release + " --output " + tmpfile)
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: string(out)}
	}

	archive, err := zip.OpenReader(tmpfile)
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = archive.Close()
	}()

	if err := func() error {
		for _, f := range archive.File {
			if f.Name != binary {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				return &DeployAgentError{Message: err.Error(), Reason: ""}
			}
			defer func() {
				_ = rc.Close()
			}()

			outFile := fmt.Sprintf("%s/%s", tmpdir, binary)
			outF, err := os.Create(outFile)
			if err != nil {
				return &DeployAgentError{Message: err.Error(), Reason: ""}
			}
			defer func() {
				_ = outF.Close()
			}()

			if _, err := io.Copy(outF, rc); err != nil {
				return &DeployAgentError{Message: err.Error(), Reason: ""}
			}

			if err := outF.Chmod(0755); err != nil {
				return &DeployAgentError{Message: err.Error(), Reason: ""}
			}
		}
		return nil
	}(); err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	err = a.target.Copy(tmpdir+"/"+binary, dest, "755", "root:root")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (a *agent) serviceEnable() error {
	out, err := a.target.Run("sudo systemctl enable --now alloy")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: string(out)}
	}
	return nil
}

func (a *agent) serviceSetup(machine map[string]string) error {
	if err := a.serviceInstall(machine); err != nil {
		return err
	}

	if err := a.serviceEnable(); err != nil {
		return err
	}

	return nil
}
