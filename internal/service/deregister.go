/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"fmt"

	"github.com/tschaefer/finchctl/internal/config"
	"github.com/tschaefer/finchctl/internal/version"
)

func (s *Service) deregisterService() error {
	if err := s.__updateSetTargetConfiguration(); err != nil {
		return convertError(err, &DeregisterServiceError{})
	}

	if !s.dryRun {
		if _, err := config.LookupStack(s.config.Hostname); err != nil {
			return &DeregisterServiceError{Message: err.Error(), Reason: ""}
		}
	}

	caCertPath := fmt.Sprintf("%s/traefik/etc/certs.d/%s.pem", s.libDir(), version.ResourceID())
	out, err := s.target.Run(fmt.Sprintf("rm -f %s", caCertPath))
	if err != nil {
		return &DeregisterServiceError{Message: err.Error(), Reason: string(out)}
	}

	if s.dryRun {
		return nil
	}

	return config.RemoveStack(s.config.Hostname)
}
