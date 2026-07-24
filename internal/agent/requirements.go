/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

func (a *Agent) __requirementsHasSudo() error {
	if _, err := a.target.Run(a.ctx, "command -v sudo"); err != nil {
		return &DeployAgentError{Message: "sudo is not installed", Reason: err.Error()}
	}
	return nil
}

func (a *Agent) __requirementsHasSudoPermission() error {
	if _, err := a.target.Run(a.ctx, "sudo -n true"); err != nil {
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

func (a *Agent) __additionsHasCurl() bool {
	if _, err := a.target.Run(a.ctx, "command -v curl"); err != nil {
		return false
	}
	return true
}

func (a *Agent) __additionsHasUnzip() bool {
	if _, err := a.target.Run(a.ctx, "command -v unzip"); err != nil {
		return false
	}
	return true
}

func (a *Agent) __additionsGitHubConnection() bool {
	_, err := a.target.Run(a.ctx, "curl --connect-timeout 3 -sfL -o /dev/null https://github.com")
	return err == nil
}

func (a *Agent) additionsAgent() bool {
	return a.__additionsHasCurl() && a.__additionsHasUnzip() && a.__additionsGitHubConnection()
}
