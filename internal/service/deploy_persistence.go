/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import "fmt"

func (s *service) persistenceMkdir() error {
	directories := []string{
		"grafana/dashboards",
		"loki/{data,etc}",
		"alloy/{data,etc}",
		"traefik/etc/{certs.d,conf.d}",
		"prometheus/{data,etc}",
		"mimir/{data,etc}",
	}
	for _, dir := range directories {
		out, err := s.target.Run(fmt.Sprintf("sudo mkdir -p %s/%s", s.libDir(), dir))
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: string(out)}
		}
	}

	return nil
}

func (s *service) persistenceChown() error {
	ownership := map[string]string{
		"grafana":    "472:472",
		"loki":       "10001:10001",
		"alloy":      "0:0",
		"traefik":    "0:0",
		"prometheus": "65534:65534",
		"mimir":      "0:0",
	}

	for path, owner := range ownership {
		out, err := s.target.Run(fmt.Sprintf("sudo chown -R %s %s/%s", owner, s.libDir(), path))
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: string(out)}
		}
	}

	return nil
}

func (s *service) persistenceSetup() error {
	if err := s.persistenceMkdir(); err != nil {
		return err
	}

	if err := s.persistenceChown(); err != nil {
		return err
	}

	return nil
}
