/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"encoding/json"
	"path"

	"github.com/tschaefer/finchctl/internal/config"
)

func (s *Service) __updateSetTargetConfiguration() error {
	cfgPath := path.Join(s.libDir(), "finch.json")
	out, err := s.target.Run(s.ctx, "sudo cat "+cfgPath)
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: string(out)}
	}

	letsencrypt := false
	yaml := path.Join(s.libDir(), "traefik/etc/conf.d/letsencrypt.yaml")
	if _, err = s.target.Run(s.ctx, "test -e "+yaml); err == nil {
		letsencrypt = true
	}

	if s.dryRun {
		s.config.Hostname = ""
		return nil
	}

	var cfg FinchConfig
	if err = json.Unmarshal([]byte(out), &cfg); err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: ""}
	}

	s.config.Hostname = cfg.Hostname
	s.config.LetsEncrypt.Enabled = letsencrypt

	return nil
}

func (s *Service) __updateRecomposeDockerServices() error {
	err := s.__deployCopyComposeFile()
	if err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	out, err := s.target.Run(s.ctx, "sudo docker compose --file "+path.Join(s.libDir(), "docker-compose.yaml")+" pull --policy missing")
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

	out, err = s.target.Run(s.ctx, "sudo docker image prune --force")
	if err != nil {
		return &UpdateServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (s *Service) updateService() error {
	if err := s.__updateSetTargetConfiguration(); err != nil {
		return err
	}

	if !s.dryRun {
		if _, err := config.LookupStack(s.config.Hostname); err != nil {
			return &UpdateServiceError{Message: err.Error(), Reason: ""}
		}
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

	if err := s.__deployCopyGrafanaAlerts(); err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	if err := s.__updateRecomposeDockerServices(); err != nil {
		return err
	}

	return nil
}
