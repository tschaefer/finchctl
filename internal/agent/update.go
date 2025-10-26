/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

func (a *agent) updateAgent() error {
	if err := a.target.Copy(a.config, "/etc/alloy/alloy.config", "400", "root:root"); err != nil {
		return &UpdateAgentError{Message: err.Error(), Reason: ""}
	}

	out, err := a.target.Run("sudo systemctl restart alloy.service")
	if err != nil {
		return &UpdateAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
