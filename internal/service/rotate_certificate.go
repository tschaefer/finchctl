/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"fmt"
	"os"

	"github.com/tschaefer/finchctl/internal/config"
	"github.com/tschaefer/finchctl/internal/mtls"
	"github.com/tschaefer/finchctl/internal/version"
)

func (s *Service) rotateCertificate() error {
	if err := s.__updateSetTargetConfiguration(); err != nil {
		return convertError(err, &RotateServiceCertificateError{})
	}

	caCertPEM, caKeyPEM, err := mtls.GenerateCA(s.config.Hostname)
	if err != nil {
		return &RotateServiceCertificateError{Message: err.Error(), Reason: ""}
	}

	clientCertPEM, clientKeyPEM, err := mtls.GenerateClient(s.config.Hostname, caCertPEM, caKeyPEM)
	if err != nil {
		return &RotateServiceCertificateError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", "ca.pem")
	if err != nil {
		return &RotateServiceCertificateError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.Write(caCertPEM); err != nil {
		return &RotateServiceCertificateError{Message: err.Error(), Reason: ""}
	}

	caCertPath := fmt.Sprintf("%s/traefik/etc/certs.d/%s.pem", s.libDir(), version.ResourceId())
	if err := s.target.Copy(f.Name(), caCertPath, "400", "0:0"); err != nil {
		return &RotateServiceCertificateError{Message: err.Error(), Reason: ""}
	}

	obsoleteCACertPath := fmt.Sprintf("%s/traefik/etc/certs.d/ca.pem", s.libDir())
	out, err := s.target.Run(fmt.Sprintf("rm -f %s", obsoleteCACertPath))
	if err != nil {
		return &RotateServiceCertificateError{Message: err.Error(), Reason: string(out)}
	}

	if s.dryRun {
		return nil
	}

	if err := config.UpdateStack(s.config.Hostname, clientCertPEM, clientKeyPEM); err != nil {
		return &RotateServiceCertificateError{Message: err.Error(), Reason: ""}
	}

	return nil
}
