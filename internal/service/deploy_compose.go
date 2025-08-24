/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"text/template"
)

func (s *service) composeRender() (string, error) {
	content, err := fs.ReadFile(Assets, "docker-compose.yaml.tmpl")
	if err != nil {
		return "", &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	tmpl, err := template.New("compose").Parse(string(content))
	if err != nil {
		return "", &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	data := struct {
		RootUrl  string
		Username string
		Password string
	}{
		RootUrl:  fmt.Sprintf("https://%s", s.config.Hostname),
		Username: s.config.Username,
		Password: s.config.Password,
	}

	out := new(bytes.Buffer)
	if err := tmpl.Execute(out, data); err != nil {
		return "", &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	return out.String(), nil
}

func (s *service) composeCopy(compose string) error {
	dest := fmt.Sprintf("%s/docker-compose.yaml", s.libDir())

	f, err := os.CreateTemp("", "docker-compose.yaml")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.WriteString(compose); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	err = s.target.Copy(f.Name(), dest, "400", "root:root")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (s *service) composeRun() error {
	out, err := s.target.Run(fmt.Sprintf("sudo docker compose --file %s/docker-compose.yaml up --detach", s.libDir()))
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (s *service) composeReady() error {
	cmd := `timeout 180 bash -c 'until curl -fs -o /dev/null -w "%{http_code}" http://localhost | grep -qE "^[234][0-9]{2}$"; do sleep 2; done'`

	out, err := s.target.Run(cmd)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (s *service) composeSetup() error {
	content, err := s.composeRender()
	if err != nil {
		return err
	}

	if err := s.composeCopy(content); err != nil {
		return err
	}

	if err := s.composeRun(); err != nil {
		return err
	}

	if err := s.composeReady(); err != nil {
		return err
	}

	return nil
}
