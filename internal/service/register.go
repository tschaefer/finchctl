/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import "github.com/tschaefer/finchctl/internal/config"

func (s *Service) registerService() error {
	if err := s.__updateSetTargetConfiguration(); err != nil {
		return convertError(err, &RegisterServiceError{})
	}

	if _, err := config.LookupStack(s.config.Hostname); err == nil {
		return &RegisterServiceError{Message: "service already registered", Reason: ""}
	}

	if err := s.__deployGenerateMTLSCertificates(); err != nil {
		return convertError(err, &RegisterServiceError{})
	}

	return nil
}
