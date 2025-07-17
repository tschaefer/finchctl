/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

func (s *service) persistenceMkdir() error {
	directories := []string{
		"/var/lib/finch/grafana/dashboards",
		"/var/lib/finch/loki/{data,etc}",
		"/var/lib/finch/alloy/{data,etc}",
		"/var/lib/finch/traefik/etc/{certs.d,conf.d}",
	}
	for _, dir := range directories {
		out, err := s.target.Run("sudo mkdir -p " + dir)
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: string(out)}
		}
	}

	return nil
}

func (s *service) persistenceChown() error {
	ownership := map[string]string{
		"/var/lib/finch/grafana": "472:472",
		"/var/lib/finch/loki":    "10001:10001",
		"/var/lib/finch/alloy":   "0:0",
		"/var/lib/finch/traefik": "0:0",
	}

	for path, owner := range ownership {
		out, err := s.target.Run("sudo chown -R " + owner + " " + path)
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
