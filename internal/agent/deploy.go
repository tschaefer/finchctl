/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/tschaefer/finchctl/internal/target"
)

func (a *agent) __deployMakeDirHierarchy() error {
	directories := []string{
		"/var/lib/alloy/data",
		"/etc/alloy",
	}
	for _, dir := range directories {
		out, err := a.target.Run("sudo mkdir -p " + dir)
		if err != nil {
			return &DeployAgentError{Message: err.Error(), Reason: string(out)}
		}
	}

	return nil
}

func (a *agent) __deployCopyConfigFile() error {
	if err := a.target.Copy(a.config, "/etc/alloy/alloy.config", "400", "root:root"); err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (a *agent) __deployCopySystemdServiceUnit() error {
	dest := "/etc/systemd/system/alloy.service"

	content, err := fs.ReadFile(Assets, "alloy.service")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", "alloy.service")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.Write(content); err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	if err := a.target.Copy(f.Name(), dest, "444", "root:root"); err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (a *agent) __deployDownloadRelease(release string, tmpdir string) (string, error) {
	url := fmt.Sprintf("https://github.com/grafana/alloy/releases/latest/download/%s.zip", release)
	tmpfile := fmt.Sprintf("%s/%s-%s.zip", tmpdir, release, time.Now().Format("19800212015200"))

	a.__helperPrintProgress(fmt.Sprintf("Downloading '%s'", url))
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

func (a *agent) __deployUnzipRelease(release string, file string) (string, error) {
	tmpdir := filepath.Dir(file)
	tmpfile := fmt.Sprintf("%s/%s", tmpdir, release)

	a.__helperPrintProgress(fmt.Sprintf("Unzipping '%s'", file))

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

		break
	}

	info, err := binary.Stat()
	if err != nil {
		return "", &DeployAgentError{Message: err.Error(), Reason: ""}
	}
	if info.Size() == 0 {
		return "", &DeployAgentError{Message: "Downloaded binary is empty", Reason: ""}
	}

	if err := binary.Chmod(0755); err != nil {
		return "", &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return binary.Name(), nil
}

func (a *agent) __deployInstallBinary(binary string) error {
	err := a.target.Copy(binary, "/usr/bin/alloy", "755", "root:root")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (a *agent) __deployEnableSystemdService() error {
	out, err := a.target.Run("sudo systemctl enable --now alloy")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: string(out)}
	}
	return nil
}

func (a *agent) __helperPrintProgress(message string) {
	username := "unknown"
	user, err := user.Current()
	if err == nil {
		username = user.Username
	}

	target.PrintProgress(fmt.Sprintf("%s as %s@localhost", message, username), a.format)
}

func (a *agent) deployAgent(machine *MachineInfo) error {
	if err := a.__deployMakeDirHierarchy(); err != nil {
		return err
	}

	if err := a.__deployCopyConfigFile(); err != nil {
		return err
	}

	if err := a.__deployCopySystemdServiceUnit(); err != nil {
		return err
	}

	tmpdir, err := os.MkdirTemp(os.TempDir(), "*-finch")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.RemoveAll(tmpdir)
	}()

	release := fmt.Sprintf("alloy-%s-%s", machine.Kernel, machine.Arch)
	zip, err := a.__deployDownloadRelease(release, tmpdir)
	if err != nil {
		return err
	}

	binary, err := a.__deployUnzipRelease(release, zip)
	if err != nil {
		return err
	}

	if err := a.__deployInstallBinary(binary); err != nil {
		return err
	}

	if err := a.__deployEnableSystemdService(); err != nil {
		return err
	}

	return nil
}
