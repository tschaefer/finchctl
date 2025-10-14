/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/tschaefer/finchctl/internal/target"
)

func (a *agent) serviceDownload(release string, tmpdir string) (string, error) {
	url := fmt.Sprintf("https://github.com/grafana/alloy/releases/latest/download/%s.zip", release)
	tmpfile := fmt.Sprintf("%s/%s-%s.zip", tmpdir, release, time.Now().Format("19800212015200"))

	a.printProgress(fmt.Sprintf("Downloading '%s'", url))
	if a.dryRun {
		return tmpfile, nil
	}

	out, err := os.Create(tmpfile)
	if err != nil {
		return "", &DeployAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = out.Close()
	}()

	resp, err := http.Get(url)
	if err != nil {
		return "", &DeployAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", &DeployAgentError{Message: fmt.Sprintf("Failed to download release: %s", resp.Status), Reason: ""}
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return tmpfile, nil
}

func (a *agent) serviceUnzip(release string, file string) (string, error) {
	tmpdir := filepath.Dir(file)
	tmpfile := fmt.Sprintf("%s/%s", tmpdir, release)

	a.printProgress(fmt.Sprintf("Unzipping '%s'", file))

	if a.dryRun {
		return tmpfile, nil
	}

	archive, err := zip.OpenReader(file)
	if err != nil {
		return "", &DeployAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = archive.Close()
	}()

	binary, err := os.Create(tmpfile)
	if err != nil {
		return "", &DeployAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = binary.Close()
	}()

	for _, part := range archive.File {
		if part.Name != release {
			continue
		}
		data, err := part.Open()
		if err != nil {
			return "", &DeployAgentError{Message: err.Error(), Reason: ""}
		}
		defer func() {
			_ = data.Close()
		}()

		if _, err := io.Copy(binary, data); err != nil {
			return "", &DeployAgentError{Message: err.Error(), Reason: ""}
		}
	}

	if err := binary.Chmod(0755); err != nil {
		return "", &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return binary.Name(), nil
}

func (a *agent) serviceInstall(machine map[string]string) error {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "*-finch")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.RemoveAll(tmpdir)
	}()

	release := fmt.Sprintf("alloy-%s-%s", machine["kernel"], machine["arch"])
	zip, err := a.serviceDownload(release, tmpdir)
	if err != nil {
		return err
	}

	binary, err := a.serviceUnzip(release, zip)
	if err != nil {
		return err
	}

	err = a.target.Copy(binary, "/usr/bin/alloy", "755", "root:root")
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

func (a *agent) printProgress(message string) {
	username := "unknown"
	user, err := user.Current()
	if err == nil {
		username = user.Username
	}

	switch a.format {
	case target.FormatProgress:
		fmt.Print(".")
	case target.FormatDocumentation:
		fmt.Printf("%s as %s@localhost\n", message, username)
	case target.FormatQuiet:
		// Do nothing
	default:
		fmt.Println(".")
	}
}
