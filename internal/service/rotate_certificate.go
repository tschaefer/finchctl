/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"fmt"

	"github.com/tschaefer/finchctl/internal/config"
)

func (s *Service) rotateCertificate() error {
	if err := s.__updateSetTargetConfiguration(); err != nil {
		return convertError(err, &RotateServiceCertificateError{})
	}

	if !s.dryRun {
		if _, err := config.LookupStack(s.config.Hostname); err != nil {
			return &RotateServiceCertificateError{Message: err.Error(), Reason: ""}
		}
	}

	if err := s.__deployGenerateMTLSCertificates(); err != nil {
		return convertError(err, &RotateServiceCertificateError{})
	}

	obsoleteCACertPath := fmt.Sprintf("%s/traefik/etc/certs.d/ca.pem", s.libDir())
	out, err := s.target.Run(fmt.Sprintf("rm -f %s", obsoleteCACertPath))
	if err != nil {
		return &RotateServiceCertificateError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
