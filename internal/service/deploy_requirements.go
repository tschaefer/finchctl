/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

func (s *service) requirementsPermissions() error {
	if _, err := s.target.Run("[ \"${EUID:-$(id -u)}\" -eq 0 ] || command -v sudo"); err != nil {
		return &DeployServiceError{Message: "Insufficient permissions", Reason: err.Error()}
	}
	return nil
}

func (s *service) requirementsCurl() error {
	if _, err := s.target.Run("command -v curl"); err != nil {
		return &DeployServiceError{Message: "curl is not installed", Reason: err.Error()}
	}
	return nil
}

func (s *service) requirementsSetup() error {
	if err := s.requirementsPermissions(); err != nil {
		return err
	}

	if err := s.requirementsCurl(); err != nil {
		return err
	}

	return nil
}
