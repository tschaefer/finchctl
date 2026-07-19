/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tschaefer/finchctl/internal/target"
)

const alloyReleasesLatestAPIURL = "https://api.github.com/repos/grafana/alloy/releases/latest"

func (a *Agent) __updateServiceBinaryGetLatestTag() (string, error) {
	url := alloyReleasesLatestAPIURL
	a.__helperPrintProgress(fmt.Sprintf("Running 'GET %s'", url))

	reqCtx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		return "", &UpdateAgentError{Message: err.Error(), Reason: ""}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", &UpdateAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &UpdateAgentError{Message: "failed to fetch latest release info", Reason: fmt.Sprintf("status code: %d", resp.StatusCode)}
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &UpdateAgentError{Message: err.Error(), Reason: ""}
	}

	a.__helperPrintProgress("Running 'JSON unmarshal \"tag_name\"'")
	var data any
	if err := json.Unmarshal(out, &data); err != nil {
		return "", &UpdateAgentError{Message: err.Error(), Reason: string(out)}
	}

	return data.(map[string]any)["tag_name"].(string), nil
}

func (a *Agent) __updateServiceBinaryIsNeeded(version string, machine *MachineInfo) (bool, error) {
	if a.dryRun {
		target.PrintProgress(
			fmt.Sprintf("Skipping Alloy update check for version '%s' due to dry-run mode", version),
			a.format,
		)
		return true, nil
	}

	latestVersion := version
	if version == "latest" {
		var err error
		latestVersion, err = a.__updateServiceBinaryGetLatestTag()
		if err != nil {
			return false, err
		}
	}

	path := "/usr/bin/alloy"
	if machine.Kernel == "darwin" {
		path = "/usr/local/bin/alloy"
	}

	out, err := a.target.Run(a.ctx, path+" --version | grep -o -E 'v[0-9\\.]+'")
	if err != nil {
		return false, &UpdateAgentError{Message: err.Error(), Reason: string(out)}
	}
	currentVersion := strings.TrimSpace(string(out))

	return latestVersion != currentVersion, nil
}

func (a *Agent) __updateServiceBinary(machine *MachineInfo, version string) error {
	ok, err := a.__updateServiceBinaryIsNeeded(version, machine)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	tmpdir, err := os.MkdirTemp(os.TempDir(), "*-finch")
	if err != nil {
		return &UpdateAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.RemoveAll(tmpdir)
	}()

	release := fmt.Sprintf("alloy-%s-%s", machine.Kernel, machine.Arch)
	zip, err := a.__deployDownloadRelease(release, version, tmpdir)
	if err != nil {
		return convertError(err, &UpdateAgentError{})
	}

	binary, err := a.__deployUnzipRelease(release, zip)
	if err != nil {
		return convertError(err, &UpdateAgentError{})
	}

	if err := a.__deployInstallBinary(binary, machine); err != nil {
		return convertError(err, &UpdateAgentError{})
	}

	return nil
}

func (a *Agent) __updateServiceBinaryOnTarget(machine *MachineInfo, version string) error {
	ok, err := a.__updateServiceBinaryIsNeeded(version, machine)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	release := fmt.Sprintf("alloy-%s-%s", machine.Kernel, machine.Arch)
	zipfile, err := a.__deployDownloadReleaseOnTarget(release, version)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = a.target.Run(a.ctx, "rm -rf "+filepath.Dir(zipfile))
	}()

	if err := a.__deployInstallBinaryOnTarget(zipfile, machine, release); err != nil {
		return err
	}

	return nil
}

func (a *Agent) updateAgent(machine *MachineInfo, skipConfig bool, skipBinaries bool, version string) error {
	if !skipConfig {
		if err := a.__deployCopyConfigFile(); err != nil {
			return convertError(err, &UpdateAgentError{})
		}
	}

	if !skipBinaries {
		if a.additionsAgent() {
			if err := a.__updateServiceBinaryOnTarget(machine, version); err != nil {
				return convertError(err, &UpdateAgentError{})
			}
		} else {
			if err := a.__updateServiceBinary(machine, version); err != nil {
				return convertError(err, &UpdateAgentError{})
			}
		}
	}

	switch machine.Kernel {
	case "linux":
		out, err := a.target.Run(a.ctx, "sudo systemctl restart alloy.service")
		if err != nil {
			return &UpdateAgentError{Message: err.Error(), Reason: string(out)}
		}
	case "freebsd":
		out, err := a.target.Run(a.ctx, "sudo service alloy restart")
		if err != nil {
			return &UpdateAgentError{Message: err.Error(), Reason: string(out)}
		}
	case "darwin":
		out, err := a.target.Run(a.ctx, "sudo launchctl kickstart -k system/com.github.tschaefer.finch.agent")
		if err != nil {
			return &UpdateAgentError{Message: err.Error(), Reason: string(out)}
		}
	default:
		// no-op
	}

	return nil
}
