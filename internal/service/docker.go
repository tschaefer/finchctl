/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
)

func (s *Service) __dockerIsAvailable() bool {
	_, err := s.target.Run("sudo docker -v")
	return err == nil
}

func (s *Service) __dockerIsRunning() bool {
	_, err := s.target.Run("sudo docker version")
	return err == nil
}

func (s *Service) __dockerComposeIsAvailable() bool {
	_, err := s.target.Run("sudo docker compose version")
	return err == nil
}

func (s *Service) __dockerInstallService() error {
	raw, err := s.target.Run("mktemp -p /tmp -d finch-XXXXXX")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	dir := strings.TrimSpace(string(raw))
	defer func() {
		_, _ = s.target.Run(fmt.Sprintf("rm -rf %s", dir))
	}()

	out, err := s.target.Run("curl -fsSL https://get.docker.com -o " + dir + "/get-docker.sh")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: string(out)}
	}

	out, err = s.target.Run("sudo sh " + dir + "/get-docker.sh")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (s *Service) __dockerCopyConfig() error {
	dest := "/etc/docker/daemon.json"

	content, err := fs.ReadFile(Assets, "daemon.json")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", "daemon.json")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.Write(content); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	if err := s.target.Copy(f.Name(), dest, "400", "0:0"); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	if out, err := s.target.Run("sudo systemctl restart docker"); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (s *Service) dockerService() error {
	if !s.__dockerIsAvailable() {
		if err := s.__dockerInstallService(); err != nil {
			return err
		}
	}

	if !s.__dockerIsRunning() {
		return &DeployServiceError{Message: "Docker is not running", Reason: ""}
	}
	if !s.__dockerComposeIsAvailable() {
		return &DeployServiceError{Message: "Docker Compose is not available", Reason: ""}
	}

	if err := s.__dockerCopyConfig(); err != nil {
		return err
	}

	return nil
}
