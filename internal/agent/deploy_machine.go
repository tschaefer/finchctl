/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"fmt"
	"strings"
)

func (a *agent) machineLinuxArch(machine string) (string, error) {
	switch machine {
	case "x86_64":
		return "amd64", nil
	case "aarch64":
		return "arm64", nil
	case "ppc64le":
		return "ppc64le", nil
	case "s390x":
		return "s390x", nil
	default:
		return "", &DeployAgentError{Message: "unsupported target architecture", Reason: machine}
	}
}

func (a *agent) machineDarwinArch(machine string) (string, error) {
	switch machine {
	case "x86_64":
		return "amd64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", &DeployAgentError{Message: "unsupported target architecture", Reason: machine}
	}
}

func (a *agent) machineFreebsdArch(machine string) (string, error) {
	switch machine {
	case "amd64":
		return "amd64", nil
	default:
		return "", &DeployAgentError{Message: "unsupported target architecture", Reason: machine}
	}
}

func (a *agent) machineInfo() (map[string]string, error) {
	out, err := a.target.Run("uname -sm")
	if err != nil {
		return nil, &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	if a.dryRun {
		return nil, nil
	}

	os := strings.SplitN(strings.TrimSpace(string(out)), " ", 2)
	if len(os) != 2 {
		return nil, &DeployAgentError{Message: "unexpected target machine", Reason: string(out)}
	}
	kernel := os[0]
	machine := os[1]

	var arch string
	switch kernel {
	case "Linux":
		arch, err = a.machineLinuxArch(machine)
		if err != nil {
			return nil, err
		}
		_, err = a.target.Run("[[ -d /run/systemd/system ]]")
		if err != nil {
			return nil, fmt.Errorf("unsupported target init system: %w", err)
		}
	case "Darwin":
		_, err = a.machineDarwinArch(machine)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("yet unsupported target kernel: %s", kernel)
	case "FreeBSD":
		_, err = a.machineFreebsdArch(machine)
		if err != nil {
			return nil, err
		}
		return nil, &DeployAgentError{Message: "yet unsupported target kernel", Reason: kernel}
	default:
		return nil, &DeployAgentError{Message: "unsupported target kernel", Reason: kernel}
	}

	info := map[string]string{
		"kernel": strings.ToLower(kernel),
		"arch":   arch,
	}

	return info, nil
}
