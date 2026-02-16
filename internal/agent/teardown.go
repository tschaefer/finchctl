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

func (a *Agent) __teardownLaunchdService() error {
	out, err := a.target.Run("sudo launchctl bootout system/com.github.tschaefer.finch.agent || true")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -f /Library/LaunchDaemons/com.github.tschaefer.finch.agent.plist")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (a *Agent) teardownAgent(machine *MachineInfo) error {
	var err error
	switch machine.Kernel {
	case "linux":
		err = a.__teardownSystemdService()
	case "freebsd":
		err = a.__teardownRcService()
	case "darwin":
		err = a.__teardownLaunchdService()
	default:
		// no-op
	}
	if err != nil {
		return err
	}

	out, err := a.target.Run("sudo rm -rf /etc/alloy")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	out, err = a.target.Run("sudo rm -rf /var/lib/alloy")
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	path := "/usr/bin/alloy"
	if machine.Kernel == "darwin" {
		path = "/usr/local/bin/alloy"
	}

	out, err = a.target.Run("sudo rm -f " + path)
	if err != nil {
		return &TeardownAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
