/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"fmt"
	"strings"
)

type MachineInfo struct {
	Kernel string
	Arch   string
}

func (a *Agent) __machineGetLinuxArch(machine string) (string, error) {
	switch machine {
	case "x86_64":
		return "amd64", nil
	case "aarch64":
		return "arm64", nil
	default:
		return "", &DeployAgentError{Message: "unsupported target architecture", Reason: machine}
	}
}

func (a *Agent) __machineGetDarwinArch(machine string) (string, error) {
	switch machine {
	case "x86_64":
		return "amd64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", &DeployAgentError{Message: "unsupported target architecture", Reason: machine}
	}
}

func (a *Agent) __machineGetFreebsdArch(machine string) (string, error) {
	switch machine {
	case "amd64":
		return "amd64", nil
	default:
		return "", &DeployAgentError{Message: "unsupported target architecture", Reason: machine}
	}
}

func (a *Agent) machineInfo() (*MachineInfo, error) {
	out, err := a.target.Run("uname -sm")
	if err != nil {
		return nil, &DeployAgentError{Message: err.Error(), Reason: ""}
	}

	if a.dryRun {
		return &MachineInfo{
			Kernel: "kernel",
			Arch:   "arch",
		}, nil
	}

	os := strings.SplitN(strings.TrimSpace(string(out)), " ", 2)
	if len(os) != 2 {
		return nil, &DeployAgentError{Message: "unexpected target machine", Reason: string(out)}
	}
	kernel := strings.ToLower(os[0])
	machine := strings.ToLower(os[1])

	var arch string
	switch kernel {
	case "linux":
		arch, err = a.__machineGetLinuxArch(machine)
		if err != nil {
			return nil, err
		}
		_, err = a.target.Run("test -d /run/systemd/system")
		if err != nil {
			return nil, fmt.Errorf("unsupported target init system: %w", err)
		}
	case "darwin":
		arch, err = a.__machineGetDarwinArch(machine)
		if err != nil {
			return nil, err
		}
	case "freebsd":
		arch, err = a.__machineGetFreebsdArch(machine)
		if err != nil {
			return nil, err
		}
	default:
		return nil, &DeployAgentError{Message: "unsupported target kernel", Reason: kernel}
	}

	return &MachineInfo{
		Kernel: kernel,
		Arch:   arch,
	}, nil
}
