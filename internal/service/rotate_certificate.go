/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"path"

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

	obsoleteCACertPath := path.Join(s.libDir(), "traefik/etc/certs.d/ca.pem")
	out, err := s.target.Run("rm -f " + obsoleteCACertPath)
	if err != nil {
		return &RotateServiceCertificateError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
