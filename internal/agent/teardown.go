/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

func (a *Agent) __teardownSystemdService() error {
	out, err := a.target.Run("sudo systemctl stop alloy.service || true")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo systemctl disable alloy.service || true")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -f /etc/systemd/system/alloy.service")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (a *Agent) __teardownRcService() error {
	out, err := a.target.Run("sudo service alloy stop || true")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo sysrc -x alloy_enable || true")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -f /etc/rc.d/alloy")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (a *Agent) teardownAgent(machine *MachineInfo) error {
	switch machine.Kernel {
	case "linux":
		if err := a.__teardownSystemdService(); err != nil {
			return err
		}
	case "freebsd":
		if err := a.__teardownRcService(); err != nil {
			return err
		}
	default:
		// no-op
	}

	out, err := a.target.Run("sudo rm -rf /etc/alloy")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -rf /var/lib/alloy")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -f /usr/bin/alloy")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
