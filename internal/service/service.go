/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"context"
	"os"
	"time"

	"github.com/tschaefer/finchctl/internal/target"
)

const (
	ServiceLibEnv string = "FINCH_SERVICE_LIB"
)

type Service struct {
	ctx        context.Context
	config     *ServiceConfig
	target     target.Target
	format     target.Format
	dryRun     bool
	cmdTimeout time.Duration
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

type Options struct {
	Config     *ServiceConfig
	TargetURL  string
	Format     target.Format
	DryRun     bool
	CmdTimeout time.Duration
}

func New(ctx context.Context, opts Options) (*Service, error) {
	t, err := target.New(opts.TargetURL, target.Options{
		Format:     opts.Format,
		DryRun:     opts.DryRun,
		CmdTimeout: opts.CmdTimeout,
	})
	if err != nil {
		return nil, err
	}

	config := opts.Config
	if config == nil {
		config = &ServiceConfig{}
	}

	return &Service{
		ctx:        ctx,
		config:     config,
		target:     t,
		format:     opts.Format,
		dryRun:     opts.DryRun,
		cmdTimeout: opts.CmdTimeout,
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

	return nil
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

func (s *Service) Dashboard(sessionTimeout int32, role string, scope []string) (*DashboardData, error) {
	return s.dashboardService(sessionTimeout, role, scope)
}

func (s *Service) libDir() string {
	var dir string
	dir = os.Getenv(ServiceLibEnv)

	if dir == "" {
		dir = "/var/lib/finch"
	}

	return dir
}

func (s *Service) Register() error {
	defer func() {
		if s.format == target.FormatProgress {
			println()
		}
	}()

	if err := s.requirementsService(); err != nil {
		return convertError(err, &RegisterServiceError{})
	}

	if err := s.registerService(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Deregister() error {
	defer func() {
		if s.format == target.FormatProgress {
			println()
		}
	}()

	if err := s.requirementsService(); err != nil {
		return convertError(err, &DeregisterServiceError{})
	}

	if err := s.deregisterService(); err != nil {
		return err
	}

	return nil
}

func (s *Service) RotateCertificate() error {
	defer func() {
		if s.format == target.FormatProgress {
			println()
		}
	}()

	if err := s.requirementsService(); err != nil {
		return convertError(err, &RotateServiceCertificateError{})
	}

	if err := s.rotateCertificate(); err != nil {
		return err
	}

	return nil
}

func (s *Service) RotateSecret() error {
	defer func() {
		if s.format == target.FormatProgress {
			println()
		}
	}()

	if err := s.requirementsService(); err != nil {
		return convertError(err, &RotateServiceSecretError{})
	}

	if err := s.rotateServiceSecret(); err != nil {
		return err
	}

	return nil
}
