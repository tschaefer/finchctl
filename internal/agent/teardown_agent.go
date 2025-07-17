/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

func (a *agent) teardownAgent() error {
	out, err := a.target.Run("sudo systemctl stop alloy.service")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo systemctl disable alloy.service")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -f /etc/systemd/system/alloy.service")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -rf /etc/alloy")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -rf /var/lib/alloy")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
