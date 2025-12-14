/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

func (a *Agent) __requirementsHasSudo() error {
	if _, err := a.target.Run("command -v sudo"); err != nil {
		return &DeployAgentError{Message: "sudo is not installed", Reason: err.Error()}
	}
	return nil
}

func (a *Agent) __requirementsHasSudoPermission() error {
	if _, err := a.target.Run("sudo -n true"); err != nil {
		return &DeployAgentError{Message: "user has no sudo permission", Reason: err.Error()}
	}
	return nil
}

func (a *Agent) requirementsAgent() error {
	if err := a.__requirementsHasSudo(); err != nil {
		return err
	}

	if err := a.__requirementsHasSudoPermission(); err != nil {
		return err
	}

	return nil
}
