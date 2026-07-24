/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

type Health struct {
	Requirement string
	Status      string
	Optional    bool
}

func (a *Agent) __doctorRequirements() (bool, *[]Health) {
	var list []Health

	ok := true
	verify := func(f func() error, r, s, d string) {
		h := true
		t := s

		if err := f(); err != nil {
			t = d
			h = false
		}

		list = append(list, Health{Requirement: r, Status: t, Optional: false})
		ok = ok && h
	}

	verify(a.__requirementsHasSudo, "sudo", "available", "not available")
	verify(a.__requirementsHasSudoPermission, "superuser permission", "sufficient", "insufficient")

	return ok, &list
}

func (a *Agent) __doctorOptionals() *[]Health {
	var list []Health

	verify := func(f func() bool, r, s, d string) {
		t := s

		if !f() {
			t = d
		}

		list = append(list, Health{Requirement: r, Status: t, Optional: true})
	}

	verify(a.__additionsHasCurl, "curl", "available", "not available")
	verify(a.__additionsHasUnzip, "unzip", "available", "not available")
	verify(a.__additionsGitHubConnection, "GitHub connection", "established", "not established")

	return &list
}

func (a *Agent) __doctor() (bool, *[]Health) {
	var list []Health

	ok, requirements := a.__doctorRequirements()
	list = append(list, *requirements...)
	optionals := a.__doctorOptionals()
	list = append(list, *optionals...)

	return ok, &list
}
