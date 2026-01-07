/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

func (s *Service) __requirementsHasSudo() error {
	if _, err := s.target.Run("command -v sudo"); err != nil {
		return &DeployServiceError{Message: "sudo is not installed", Reason: err.Error()}
	}
	return nil
}

func (s *Service) __requirementsHasCurl() error {
	if _, err := s.target.Run("command -v curl"); err != nil {
		return &DeployServiceError{Message: "curl is not installed", Reason: err.Error()}
	}
	return nil
}

func (s *Service) __requirementsHasSudoPermission() error {
	if _, err := s.target.Run("sudo -n true"); err != nil {
		return &DeployServiceError{Message: "user has no sudo permission", Reason: err.Error()}
	}
	return nil
}

func (s *Service) requirementsService() error {
	if err := s.__requirementsHasSudo(); err != nil {
		return err
	}

	if err := s.__requirementsHasCurl(); err != nil {
		return err
	}

	if err := s.__requirementsHasSudoPermission(); err != nil {
		return err
	}

	return nil
}
