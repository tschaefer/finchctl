/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"encoding/json"
	"fmt"
)

func (s *service) updateSetConfiguration() error {
	out, err := s.target.Run(fmt.Sprintf("sudo cat %s/finch.json", s.libDir()))
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: string(out)}
	}

	if s.dryRun {
		return nil
	}

	var config Config
	err = json.Unmarshal([]byte(out), &config)
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: ""}
	}

	s.config.Hostname = config.Hostname
	s.config.Username = config.Credentials.Username
	s.config.Password = config.Credentials.Password

	yaml := fmt.Sprintf("%s/traefik/etc/conf.d/letsencrypt.yaml", s.libDir())
	_, err = s.target.Run(fmt.Sprintf("test -e %s", yaml))
	if err == nil {
		s.config.LetsEncrypt.Enabled = true
	}

	return nil
}

func (s *service) updateCompose() error {
	out, err := s.target.Run(fmt.Sprintf("sudo docker compose --file %s/docker-compose.yaml down --volumes", s.libDir()))
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

	out, err = s.target.Run(fmt.Sprintf("sudo docker compose --file %s/docker-compose.yaml pull --policy always", s.libDir()))
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: string(out)}
	}

	err = s.composeRun()
	if err != nil {
		return err
	}
	err = s.composeReady()
	if err != nil {
		return err
	}

	out, err = s.target.Run("sudo docker image prune --force")
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (s *service) updateService() error {
	if err := s.updateSetConfiguration(); err != nil {
		return err
	}

	if err := s.configLoki(); err != nil {
		return err
	}

	if err := s.configTraefikHttp(); err != nil {
		return err
	}

	if err := s.configAlloy(); err != nil {
		return err
	}

	if err := s.configMimir(); err != nil {
		return err
	}

	if err := s.configGrafanaDashboards(); err != nil {
		return err
	}

	if err := s.updateCompose(); err != nil {
		return err
	}

	return nil
}
