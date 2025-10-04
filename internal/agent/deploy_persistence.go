/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import "fmt"

func (a *agent) persistenceMkdir() error {
	directories := []string{
		"/var/lib/alloy/data",
		"/etc/alloy",
		"/var/lib/node_exporter/textfile_collector",
	}
	for _, dir := range directories {
		out, err := a.target.Run("sudo mkdir -p " + dir)
		if err != nil {
			return &DeployAgentError{Message: err.Error(), Reason: string(out)}
		}
	}

	return nil
}

func (a *agent) persistenceChown() error {
	ownership := map[string]string{
		"/var/lib/node_exporter": "node_exporter:node_exporter",
	}

	for path, owner := range ownership {
		out, err := a.target.Run(fmt.Sprintf("sudo chown -R %s %s", owner, path))
		if err != nil {
			return &DeployAgentError{Message: err.Error(), Reason: string(out)}
		}
	}

	return nil
}

func (a *agent) persistenceSetup() error {
	if err := a.persistenceMkdir(); err != nil {
		return err
	}

	if err := a.persistenceChown(); err != nil {
		return err
	}

	return nil
}
