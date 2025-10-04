/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

func (a *agent) teardownAgent() error {
	services := []string{
		"alloy.service",
		"node_exporter.socket",
		"node_exporter.service",
	}

	for _, service := range services {
		out, err := a.target.Run("sudo systemctl stop " + service)
		if err != nil {
			return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
		}

		out, err = a.target.Run("sudo systemctl disable " + service)
		if err != nil {
			return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
		}

		out, err = a.target.Run("sudo rm -f /etc/systemd/system/" + service)
		if err != nil {
			return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
		}
	}

	out, err := a.target.Run("sudo deluser node_exporter")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -rf /etc/alloy")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -f /var/lib/node_exporter")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -rf /var/lib/alloy")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
