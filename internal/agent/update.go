/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/tschaefer/finchctl/internal/target"
)

func (a *agent) __updateServiceBinaryIsNeeded() (bool, error) {
	if a.dryRun {
		target.PrintProgress("Skipping update check due to dry-run mode", a.format)
		return false, nil
	}

	url := "https://api.github.com/repos/grafana/alloy/releases/latest"
	a.__helperPrintProgress(fmt.Sprintf("Running 'GET %s'", url))
	resp, err := http.Get(url)
	if err != nil {
		return false, &UpdateAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, &UpdateAgentError{Message: "failed to fetch latest release info", Reason: fmt.Sprintf("status code: %d", resp.StatusCode)}
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, &UpdateAgentError{Message: err.Error(), Reason: ""}
	}

	a.__helperPrintProgress("Running 'JSON unmarshal \"tag_name\"'")
	var data any
	if err := json.Unmarshal(out, &data); err != nil {
		return false, &UpdateAgentError{Message: err.Error(), Reason: string(out)}
	}
	latestVersion := data.(map[string]any)["tag_name"].(string)

	out, err = a.target.Run("alloy --version | grep -o -E 'v[0-9\\.]+'")
	if err != nil {
		return false, &UpdateAgentError{Message: err.Error(), Reason: string(out)}
	}
	currentVersion := strings.TrimSpace(string(out))

	return latestVersion != currentVersion, nil
}

func (a *agent) __updateServiceBinary(machine *MachineInfo) error {
	ok, err := a.__updateServiceBinaryIsNeeded()
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
	zip, err := a.__deployDownloadRelease(release, tmpdir)
	if err != nil {
		return convertError(err, &UpdateAgentError{})
	}

	binary, err := a.__deployUnzipRelease(release, zip)
	if err != nil {
		return convertError(err, &UpdateAgentError{})
	}

	if err := a.__deployInstallBinary(binary); err != nil {
		return convertError(err, &UpdateAgentError{})
	}

	return nil
}

func (a *agent) updateAgent(machine *MachineInfo, skipConfig bool, skipBinaries bool) error {
	if !skipConfig {
		if err := a.__deployCopyConfigFile(); err != nil {
			return convertError(err, &UpdateAgentError{})
		}
	}

	if !skipBinaries {
		if err := a.__updateServiceBinary(machine); err != nil {
			return convertError(err, &UpdateAgentError{})
		}
	}

	out, err := a.target.Run("sudo systemctl restart alloy.service")
	if err != nil {
		return &UpdateAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
