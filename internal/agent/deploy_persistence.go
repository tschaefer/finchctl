/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

func (a *agent) persistenceSetup() error {
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
