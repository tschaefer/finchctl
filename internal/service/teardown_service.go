/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import "fmt"

func (s *service) teardownService() error {
	out, err := s.target.Run(fmt.Sprintf("sudo docker compose --file %s/docker-compose.yaml down --volumes", s.libDir()))
	if err != nil {
		return &TeardownServiceError{Message: err.Error(), Reason: string(out)}
	}

	out, err = s.target.Run("sudo rm -rf " + s.libDir())
	if err != nil {
		return &TeardownServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
