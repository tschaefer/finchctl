/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"archive/zip"
	"context"
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

func (a *Agent) __deployMakeDirHierarchy() error {
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

func (a *Agent) __deployCopyConfigFile() error {
	if err := a.target.Copy(a.config, "/etc/alloy/alloy.config", "400", "0:0"); err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (a *Agent) __deployCopySystemdServiceUnit() error {
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

	if err := a.target.Copy(f.Name(), dest, "444", "0:0"); err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (a *Agent) __deployCopyRcServiceFile() error {
	dest := "/etc/rc.d/alloy"

	content, err := fs.ReadFile(Assets, "alloy.rc")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", "alloy.rc")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.Write(content); err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	if err := a.target.Copy(f.Name(), dest, "444", "0:0"); err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	out, err := a.target.Run("sudo chmod +x " + dest)
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (a *Agent) __deployCopyLaunchdServiceFile() error {
	dest := "/Library/LaunchDaemons/com.github.tschaefer.finch.agent.plist"

	content, err := fs.ReadFile(Assets, "com.github.tschaefer.finch.agent.plist")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", "com.github.tschaefer.finch.agent.plist")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.Write(content); err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	if err := a.target.Copy(f.Name(), dest, "444", "0:0"); err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (a *Agent) __deployDownloadRelease(release string, version string, tmpdir string) (string, error) {
	var url string
	if version != "latest" {
		url = fmt.Sprintf("https://github.com/grafana/alloy/releases/download/%s/%s.zip", version, release)
	} else {
		url = fmt.Sprintf("https://github.com/grafana/alloy/releases/latest/download/%s.zip", release)
	}
	tmpfile := fmt.Sprintf("%s/%s-%s.zip", tmpdir, release, time.Now().Format("19800212015200"))

	a.__helperPrintProgress(fmt.Sprintf("Running 'GET %s'", url))
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

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
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

func (a *Agent) __deployUnzipRelease(release string, file string) (string, error) {
	tmpdir := filepath.Dir(file)
	tmpfile := fmt.Sprintf("%s/%s", tmpdir, release)

	a.__helperPrintProgress(fmt.Sprintf("Running 'unzip %s'", file))

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

func (a *Agent) __deployInstallBinary(binary string, machine *MachineInfo) error {
	path := "/usr/bin/alloy"
	if machine.Kernel == "darwin" {
		path = "/usr/local/bin/alloy"
		out, err := a.target.Run("sudo mkdir -p " + filepath.Dir(path))
		if err != nil {
			return &DeployAgentError{Message: err.Error(), Reason: string(out)}
		}
	}

	err := a.target.Copy(binary, path, "755", "0:0")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (a *Agent) __deployEnableSystemdService() error {
	out, err := a.target.Run("sudo systemctl enable --now alloy")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: string(out)}
	}
	return nil
}

func (a *Agent) __deployEnableRcService() error {
	out, err := a.target.Run("sudo sysrc alloy_enable=YES")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo service alloy start")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (a *Agent) __deployEnableLaunchdService() error {
	out, err := a.target.Run("sudo launchctl bootstrap system /Library/LaunchDaemons/com.github.tschaefer.finch.agent.plist")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo launchctl kickstart -k system/com.github.tschaefer.finch.agent")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (a *Agent) __helperPrintProgress(message string) {
	username := "unknown"
	user, err := user.Current()
	if err == nil {
		username = user.Username
	}

	target.PrintProgress(fmt.Sprintf("%s as %s@localhost", message, username), a.format)
}

func (a *Agent) deployAgent(machine *MachineInfo, alloyVersion string) error {
	if err := a.__deployMakeDirHierarchy(); err != nil {
		return err
	}

	if err := a.__deployCopyConfigFile(); err != nil {
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
	zip, err := a.__deployDownloadRelease(release, alloyVersion, tmpdir)
	if err != nil {
		return err
	}

	binary, err := a.__deployUnzipRelease(release, zip)
	if err != nil {
		return err
	}

	if err := a.__deployInstallBinary(binary, machine); err != nil {
		return err
	}

	switch machine.Kernel {
	case "linux":
		if err := a.__deployCopySystemdServiceUnit(); err != nil {
			return err
		}
		if err := a.__deployEnableSystemdService(); err != nil {
			return err
		}
	case "freebsd":
		if err := a.__deployCopyRcServiceFile(); err != nil {
			return err
		}
		if err := a.__deployEnableRcService(); err != nil {
			return err
		}
	case "darwin":
		if err := a.__deployCopyLaunchdServiceFile(); err != nil {
			return err
		}
		if err := a.__deployEnableLaunchdService(); err != nil {
			return err
		}
	default:
		// no-op
	}

	return nil
}
