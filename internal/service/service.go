/*
Copyright (c) Tobias Schäfer. All rights reservem.
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

type Service struct {
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
	CustomTLS struct {
		Enabled      bool
		CertFilePath string
		KeyFilePath  string
	}
}

type FinchConfig struct {
	CreatedAt string `json:"created_at"`
	Id        string `json:"id"`
	Database  string `json:"database"`
	Profiler  string `json:"profiler"`
	Secret    string `json:"secret"`
	Hostname  string `json:"hostname"`
	Version   string `json:"version"`
}

func New(config *ServiceConfig, targetUrl string, format target.Format, dryRun bool) (*Service, error) {
	target, err := target.NewTarget(targetUrl, format, dryRun)
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = &ServiceConfig{}
	}

	return &Service{
		config: config,
		target: target,
		format: format,
		dryRun: dryRun,
	}, nil
}

func (s *Service) Teardown() error {
	defer func() {
		if s.format == target.FormatProgress {
			println()
		}
	}()

	if err := s.requirementsService(); err != nil {
		return convertError(err, &TeardownServiceError{})
	}

	if err := s.teardownService(); err != nil {
		return err
	}

	if s.dryRun {
		return nil
	}

	return config.RemoveStack(s.config.Hostname)
}

func (s *Service) Deploy() error {
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

	return nil
}

func (s *Service) Update() error {
	defer func() {
		if s.format == target.FormatProgress {
			println()
		}
	}()

	if err := s.requirementsService(); err != nil {
		return convertError(err, &UpdateServiceError{})
	}

	if err := s.updateService(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Info() (*InfoData, error) {
	defer func() {
		if s.format == target.FormatProgress {
			println()
		}
	}()

	return s.infoService()
}

func (s *Service) libDir() string {
	var dir string
	dir = os.Getenv(ServiceLibEnv)

	if dir == "" {
		dir = "/var/lib/finch"
	}

	return dir
}
