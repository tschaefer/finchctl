/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

func (a *agent) userSetup() error {
	out, err := a.target.Run("sudo useradd --system --home-dir /var/lib/node_exporter --no-create-home --shell /sbin/nologin node_exporter")
	if err != nil {
		return &DeployAgentError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
