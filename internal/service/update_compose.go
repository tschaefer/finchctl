/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"encoding/json"
)

func (s *service) updateReadConfiguration() (*Config, error) {
	out, err := s.target.Run("sudo cat /var/lib/finch/finch.json")
	if err != nil {
		return nil, &UpdateServiceError{Message: err.Error(), Reason: string(out)}
	}

	var config Config

	if s.dryRun {
		return &config, nil
	}

	err = json.Unmarshal([]byte(out), &config)
	if err != nil {
		return nil, &UpdateServiceError{Message: err.Error(), Reason: ""}
	}

	return &config, nil
}

func (s *service) updateCompose() error {
	cfg, err := s.updateReadConfiguration()
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: ""}
	}

	s.config.Hostname = cfg.Hostname
	s.config.Username = cfg.Credentials.Username
	s.config.Password = cfg.Credentials.Password

	out, err := s.target.Run("cd /var/lib/finch && sudo docker compose down --volumes")
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: string(out)}
	}

	out, err = s.target.Run("cd /var/lib/finch && sudo docker compose pull --policy always")
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: string(out)}
	}

	compose, err := s.composeRender()
	if err != nil {
		return err
	}
	err = s.composeCopy(compose)
	if err != nil {
		return err
	}
	err = s.composeRun()
	if err != nil {
		return err
	}
	err = s.composeReady()
	if err != nil {
		return err
	}

	return nil
}
