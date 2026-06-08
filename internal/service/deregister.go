/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"path"

	"github.com/tschaefer/finchctl/internal/config"
)

func (s *Service) deregisterService(clientID string, keepCfg bool) error {
	if err := s.__updateSetTargetConfiguration(); err != nil {
		return convertError(err, &DeregisterServiceError{})
	}

	if _, err := config.LookupStack(s.config.Hostname); err != nil {
		return &DeregisterServiceError{Message: err.Error(), Reason: ""}
	}

	caCertPath := path.Join(s.libDir(), "traefik/etc/certs.d", clientID+".pem")
	out, err := s.target.Run(s.ctx, "rm -f "+caCertPath)
	if err != nil {
		return &DeregisterServiceError{Message: err.Error(), Reason: string(out)}
	}

	if s.dryRun {
		return nil
	}

	if !keepCfg {
		return nil
	}

	return config.RemoveStack(s.config.Hostname)
}
