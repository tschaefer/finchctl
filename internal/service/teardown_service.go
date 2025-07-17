/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

func (s *service) teardownService() error {
	out, err := s.target.Run("cd /var/lib/finch && sudo docker compose down --volumes")
	if err != nil {
		return &TeardownServiceError{Message: err.Error(), Reason: string(out)}
	}

	out, err = s.target.Run("sudo rm -rf /var/lib/finch")
	if err != nil {
		return &TeardownServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
