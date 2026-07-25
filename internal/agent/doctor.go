/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"errors"
)

type Health struct {
	Requirement string
	Status      string
	Optional    bool
}

const (
	ALLOY_HTTP_PORT      = "12345"
	ALLOY_LOKI_PORT      = "3100"
	ALLOY_MIMIR_PORT     = "9091"
	ALLOY_PYROSCOPE_PORT = "4040"
)

func (a *Agent) __doctorRequirements() (*[]Health, error) {
	var list []Health
	var errs error

	verify := func(f func() error, r, s, d string) {
		t := s

		if err := f(); err != nil {
			t = d
			errs = errors.Join(errs, convertError(err, &DoctorAgentError{}))
		}

		list = append(list, Health{Requirement: r, Status: t, Optional: false})
	}

	verify(a.__requirementsHasSudo, "sudo", "available", "not available")
	verify(a.__requirementsHasSudoPermission, "superuser permission", "sufficient", "insufficient")

	return &list, errs
}

func (a *Agent) __doctorOptionals() (*[]Health, error) {
	var list []Health
	var errs error

	verify := func(f func() bool, r, s, d string) {
		t := s

		if !f() {
			t = d
			errs = errors.Join(errs, &DoctorAgentError{Message: r + " " + d, Reason: ""})
		}

		list = append(list, Health{Requirement: r, Status: t, Optional: true})
	}

	verify(a.__additionsHasCurl, "curl", "available", "not available")
	verify(a.__additionsHasUnzip, "unzip", "available", "not available")
	verify(a.__additionsGitHubConnection, "GitHub connection", "established", "not established")

	return &list, errs
}

func (a *Agent) __doctorPorts(machine *MachineInfo) (*[]Health, error) {
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
		list = append(list, Health{"port check", cmd + " not found", false})
		return &list, err
	}

	ports := map[string]bool{
		ALLOY_HTTP_PORT:      false,
		ALLOY_LOKI_PORT:      false,
		ALLOY_MIMIR_PORT:     true,
		ALLOY_PYROSCOPE_PORT: true,
	}

	for port, optional := range ports {
		_, err = a.target.Run(a.ctx, exec)
		if err == nil {
			errs = errors.Join(errs, &DoctorAgentError{Message: "port " + port + " already bound", Reason: ""})
			status = "bound"
		} else {
			status = "unbound"
		}

		list = append(list, Health{"port " + port, status, optional})
	}

	return &list, errs
}

func (a *Agent) __doctor(checkOptionals, checkPorts bool) (*[]Health, error) {
	var list []Health
	var err error

	machine, err := a.machineInfo()
	if err != nil {
		return &list, convertError(err, &DoctorAgentError{})
	}

	requirements, err := a.__doctorRequirements()
	list = append(list, *requirements...)
	if err != nil {
		return &list, err
	}

	if checkOptionals {
		optionals, err := a.__doctorOptionals()
		list = append(list, *optionals...)
		if err != nil {
			return &list, err
		}
	}

	if checkPorts {
		ports, e := a.__doctorPorts(machine)
		err = e
		list = append(list, *ports...)
	}

	return &list, err
}
