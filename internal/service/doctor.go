/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"github.com/fatih/color"
)

type Health struct {
	Requirement string
	Status      string
	Optional    bool
	Ok          bool
}

const (
	TRAEFIK_HTTP_PORT  = "80"
	TRAEFIK_HTTPS_PORT = "443"
)

func (s *Service) __doctorRequirements() (*[]Health, bool) {
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
		list = append(list, Health{r, t, false, o})
	}

	verify(s.__requirementsHasSudo, "sudo", "available", "not available")
	verify(s.__requirementsHasSudoPermission, "superuser permission", "sufficient", "insufficient")
	verify(s.__requirementsHasCurl, "curl", "available", "not available")
	verify(s.__requirementsGitHubConnection, "GitHub connection", "established", "not established")

	return &list, ok
}

func (s *Service) __doctorPorts() (*[]Health, bool) {
	var list []Health

	var port string
	var status string

	cmd := "ss"
	exec := "sudo ss -H -tlpn sport = :" + port + " | grep -q ''"

	if _, err := s.target.Run(s.ctx, "command -v "+cmd); err != nil {
		list = append(list, Health{"port check", color.RedString(cmd + " not found"), false, false})
		return &list, false
	}

	ports := map[string]bool{
		TRAEFIK_HTTP_PORT:  false,
		TRAEFIK_HTTPS_PORT: false,
	}
	ok := true

	for port, optional := range ports {
		o := true
		_, err := s.target.Run(s.ctx, exec)
		if err == nil {
			o = false
			status = color.RedString("bound")
		} else {
			status = color.GreenString("unbound")
		}

		ok = ok && o
		list = append(list, Health{"port " + port, status, optional, o})
	}

	return &list, ok
}

func (s *Service) __doctor() (*[]Health, bool) {
	var list []Health

	_, err := s.target.Run(s.ctx, "uname -sm | grep -qE 'Linux (x86_64|aarch64)'")
	if err != nil {
		list = append(list, Health{"os", color.RedString("unsupported"), false, false})
		return &list, false
	}
	list = append(list, Health{"os", color.GreenString("supported"), false, true})

	requirements, ok := s.__doctorRequirements()
	list = append(list, *requirements...)
	if !ok {
		return &list, false
	}

	ports, ok := s.__doctorPorts()
	list = append(list, *ports...)
	if !ok {
		return &list, false
	}

	return &list, true
}
