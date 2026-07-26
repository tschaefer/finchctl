/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"errors"

	"github.com/fatih/color"
)

type Health struct {
	Requirement string
	Status      string
	Optional    bool
	Ok          bool
}

const (
	ALLOY_HTTP_PORT      = "12345"
	ALLOY_LOKI_PORT      = "3100"
	ALLOY_MIMIR_PORT     = "9091"
	ALLOY_PYROSCOPE_PORT = "4040"
)

func (a *Agent) __doctorRequirements() (*[]Health, bool) {
	var list []Health
	ok := true

	verify := func(f func() error, r, s, d string) {
		t := color.GreenString(s)
		o := true

		if err := f(); err != nil {
			t = color.RedString(d)
			o = false
		}

		ok = ok && o
		list = append(list, Health{Requirement: r, Status: t, Optional: false, Ok: o})
	}

	verify(a.__requirementsHasSudo, "sudo", "available", "not available")
	verify(a.__requirementsHasSudoPermission, "superuser permission", "sufficient", "insufficient")

	return &list, ok
}

func (a *Agent) __doctorOptionals() (*[]Health, bool) {
	var list []Health
	ok := true

	verify := func(f func() bool, r, s, d string) {
		t := color.GreenString(s)
		o := true

		if !f() {
			t = color.RedString(d)
			o = false
		}

		ok = ok && o
		list = append(list, Health{Requirement: r, Status: t, Optional: true, Ok: o})
	}

	verify(a.__additionsHasCurl, "curl", "available", "not available")
	verify(a.__additionsHasUnzip, "unzip", "available", "not available")
	verify(a.__additionsGitHubConnection, "GitHub connection", "established", "not established")

	return &list, ok
}

func (a *Agent) __doctorPorts(machine *MachineInfo) (*[]Health, bool) {
	var list []Health

	var cmd string
	var err error
	var errs error
	var exec string
	var port string
	var status string

	switch machine.Kernel {
	case "linux":
		cmd = "ss"
		exec = "sudo ss -H -tlpn sport = :" + port + " | grep -q ''"
	case "freebsd":
		cmd = "sockstat"
		exec = "sudo sockstat -q -P tcp -p " + port + " -l | grep -q ''"
	case "darwin":
		cmd = "lsof"
		exec = "sudo lsof -i :" + port + " | grep -q LISTEN"
	default:
		// pass
	}

	if _, err = a.target.Run(a.ctx, "command -v "+cmd); err != nil {
		list = append(list, Health{"port check", color.RedString(cmd + " not found"), false, false})
		return &list, false
	}

	ports := map[string]bool{
		ALLOY_HTTP_PORT:      false,
		ALLOY_LOKI_PORT:      false,
		ALLOY_MIMIR_PORT:     true,
		ALLOY_PYROSCOPE_PORT: true,
	}
	ok := true

	for port, optional := range ports {
		o := true
		_, err = a.target.Run(a.ctx, exec)
		if err == nil {
			o = false
			errs = errors.Join(errs, &DoctorAgentError{Message: "port " + port + " bound", Reason: ""})
			status = color.RedString("bound")
		} else {
			status = color.GreenString("unbound")
		}

		ok = ok && o
		list = append(list, Health{"port " + port, status, optional, o})
	}

	return &list, ok
}

func (a *Agent) __doctor(checkOptionals, checkPorts bool) (*[]Health, bool) {
	var list []Health

	machine, err := a.machineInfo()
	if err != nil {
		return &list, false
	}

	requirements, ok := a.__doctorRequirements()
	list = append(list, *requirements...)
	if !ok {
		return &list, false
	}

	if checkOptionals {
		optionals, ok := a.__doctorOptionals()
		list = append(list, *optionals...)
		if !ok {
			return &list, false
		}
	}

	if checkPorts {
		ports, ok := a.__doctorPorts(machine)
		list = append(list, *ports...)
		if !ok {
			return &list, false
		}
	}

	return &list, true
}
