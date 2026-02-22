/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"path"

	"github.com/tschaefer/finchctl/internal/config"
)

func (s *Service) teardownService() error {
	if err := s.__updateSetTargetConfiguration(); err != nil {
		return convertError(err, &TeardownServiceError{})
	}

	if !s.dryRun {
		if _, err := config.LookupStack(s.config.Hostname); err != nil {
			return &TeardownServiceError{Message: err.Error(), Reason: ""}
		}
	}

	out, err := s.target.Run(s.ctx, "sudo docker compose --file "+path.Join(s.libDir(), "docker-compose.yaml")+" down --volumes")
	if err != nil {
		return &TeardownServiceError{Message: err.Error(), Reason: string(out)}
	}

	out, err = s.target.Run(s.ctx, "sudo rm -rf "+s.libDir())
	if err != nil {
		return &TeardownServiceError{Message: err.Error(), Reason: string(out)}
	}

	if s.dryRun {
		return nil
	}

	return config.RemoveStack(s.config.Hostname)
}
