/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"io/fs"
	"os"
)

func (a *agent) configFile() error {
	if err := a.target.Copy(a.config, "/etc/alloy/alloy.config", "400", "root:root"); err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (a *agent) configService() error {
	services := []string{
		"alloy.service",
		"node_exporter.service",
		"node_exporter.socket",
	}

	for _, service := range services {
		content, err := fs.ReadFile(Assets, service)
		if err != nil {
			return &DeployAgentError{Message: err.Error(), Reason: ""}
		}

		f, err := os.CreateTemp("", service)
		if err != nil {
			return &DeployAgentError{Message: err.Error(), Reason: ""}
		}
		defer func() {
			_ = os.Remove(f.Name())
		}()
		if _, err := f.Write(content); err != nil {
			return &DeployAgentError{Message: err.Error(), Reason: ""}
		}

		dest := "/etc/systemd/system/" + service
		if err := a.target.Copy(f.Name(), dest, "444", "root:root"); err != nil {
			return &DeployAgentError{Message: err.Error(), Reason: ""}
		}
	}

	return nil
}

func (a *agent) configSetup() error {
	if err := a.configFile(); err != nil {
		return err
	}

	if err := a.configService(); err != nil {
		return err
	}

	return nil
}
