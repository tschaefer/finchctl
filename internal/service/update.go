/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"encoding/json"
	"fmt"
)

func (s *service) __updateSetTargetConfiguration() error {
	out, err := s.target.Run(fmt.Sprintf("sudo cat %s/finch.json", s.libDir()))
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: string(out)}
	}

	if s.dryRun {
		return nil
	}

	var config FinchConfig
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

func (s *service) __updateRecomposeDockerServices() error {
	err := s.__deployCopyComposeFile()
	if err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	out, err := s.target.Run(fmt.Sprintf("sudo docker compose --file %s/docker-compose.yaml pull --policy always", s.libDir()))
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: string(out)}
	}

	err = s.__deployComposeUp()
	if err != nil {
		return convertError(err, &UpdateServiceError{})
	}
	err = s.__deployComposeReady()
	if err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	out, err = s.target.Run("sudo docker image prune --force")
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (s *service) updateService() error {
	if err := s.__updateSetTargetConfiguration(); err != nil {
		return err
	}

	if err := s.__deployMakeDirHierarchy(); err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	if err := s.__deploySetDirHierarchyPermission(); err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	if err := s.__deployCopyLokiConfig(); err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	if err := s.__deployCopyTraefikHttpConfig(); err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	if err := s.__deployCopyAlloyConfig(); err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	if err := s.__deployCopyMimirConfig(); err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	if err := s.__deployCopyGrafanaDashboards(); err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	if err := s.__updateRecomposeDockerServices(); err != nil {
		return err
	}

	return nil
}
