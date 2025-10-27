/*
Copyright (c) 2025 Tobias Schäfer. All rights reservem.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"os"

	"github.com/tschaefer/finchctl/internal/config"
	"github.com/tschaefer/finchctl/internal/target"
)

const (
	ServiceLibEnv string = "FINCH_SERVICE_LIB"
)

type Service interface {
	Deploy() error
	Update() error
	Teardown() error
	Info() (*InfoData, error)
}

type service struct {
	config *ServiceConfig
	target target.Target
	format target.Format
	dryRun bool
}

type ServiceConfig struct {
	Hostname    string
	LetsEncrypt struct {
		Enabled bool
		Email   string
	}
	Username  string
	Password  string
	CustomTLS struct {
		Enabled      bool
		CertFilePath string
		KeyFilePath  string
	}
}

type FinchConfig struct {
	CreatedAt   string `json:"created_at"`
	Id          string `json:"id"`
	Database    string `json:"database"`
	Secret      string `json:"secret"`
	Hostname    string `json:"hostname"`
	Version     string `json:"version"`
	Credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"credentials"`
}

func New(config *ServiceConfig, targetUrl string, format target.Format, dryRun bool) (Service, error) {
	target, err := target.NewTarget(targetUrl, format, dryRun)
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = &ServiceConfig{}
	}

	return &service{
		config: config,
		target: target,
		format: format,
		dryRun: dryRun,
	}, nil
}

func (s *service) Teardown() error {
	defer func() {
		if s.format == target.FormatProgress {
			println()
		}
	}()

	if !s.dryRun {
		if err := config.RemoveStackAuth(s.config.Hostname); err != nil {
			return err
		}
	}

	if err := s.teardownService(); err != nil {
		return err
	}

	return nil
}

func (s *service) Deploy() error {
	defer func() {
		if s.format == target.FormatProgress {
			println()
		}
	}()

	if err := s.requirementsService(); err != nil {
		return err
	}

	if err := s.dockerService(); err != nil {
		return err
	}

	if err := s.deployService(); err != nil {
		return err
	}

	if !s.dryRun {
		if err := config.UpdateStackAuth(s.config.Hostname, s.config.Username, s.config.Password); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) Update() error {
	defer func() {
		if s.format == target.FormatProgress {
			println()
		}
	}()

	if err := s.updateService(); err != nil {
		return err
	}

	return nil
}

func (s *service) Info() (*InfoData, error) {
	defer func() {
		if s.format == target.FormatProgress {
			println()
		}
	}()

	return s.infoService()
}

func (s *service) libDir() string {
	var dir string
	dir = os.Getenv(ServiceLibEnv)

	if dir == "" {
		dir = "/var/lib/finch"
	}

	return dir
}
