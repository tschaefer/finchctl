/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *service) __deployMakeDirHierachy() error {
	directories := []string{
		"grafana/dashboards",
		"loki/{data,etc}",
		"alloy/{data,etc}",
		"traefik/etc/{certs.d,conf.d}",
		"mimir/{data,etc}",
		"pyroscope/data",
	}
	for _, dir := range directories {
		out, err := s.target.Run(fmt.Sprintf("sudo mkdir -p %s/%s", s.libDir(), dir))
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: string(out)}
		}
	}

	return nil
}

func (s *service) __deploySetDirHierachyPermission() error {
	ownership := map[string]string{
		"grafana":   "472:472",
		"loki":      "10001:10001",
		"alloy":     "0:0",
		"traefik":   "0:0",
		"mimir":     "10001:10001",
		"pyroscope": "10001:10001",
	}

	for path, owner := range ownership {
		out, err := s.target.Run(fmt.Sprintf("sudo chown -R %s %s/%s", owner, s.libDir(), path))
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: string(out)}
		}
	}

	return nil
}

func (s *service) __deployDirHierachy() error {
	if err := s.__deployMakeDirHierachy(); err != nil {
		return err
	}

	if err := s.__deploySetDirHierachyPermission(); err != nil {
		return err
	}

	return nil
}

func (s *service) __deployCopyLokiConfig() error {
	path := fmt.Sprintf("%s/loki/etc/loki.yaml", s.libDir())
	return s.__helperCopyConfig(path, "400", "10001:10001")
}

func (s *service) __deployCopyLokiUserAuthFile() error {
	path := fmt.Sprintf("%s/traefik/etc/conf.d/loki-users.yaml", s.libDir())

	hash, err := bcrypt.GenerateFromPassword([]byte(s.config.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	data := struct {
		Username string
		Password string
	}{
		Username: s.config.Username,
		Password: string(hash),
	}

	return s.__helperCopyTemplate(path, "400", "0:0", data)
}

func (s *service) __deployCopyTraefikConfig() error {
	path := fmt.Sprintf("%s/traefik/etc/traefik.yaml", s.libDir())

	letsencrypt := s.config.LetsEncrypt.Email
	if letsencrypt == "" {
		letsencrypt = "acme@example.com"
	}
	data := struct {
		Email string
	}{
		Email: letsencrypt,
	}

	return s.__helperCopyTemplate(path, "400", "0:0", data)
}

func (s *service) __deployCopyTraefikHttpConfig() error {
	path := fmt.Sprintf("%s/traefik/etc/conf.d/http.yaml", s.libDir())

	data := struct {
		HostRule string
	}{
		HostRule: "",
	}
	if s.config.LetsEncrypt.Enabled {
		data.HostRule = fmt.Sprintf("&& Host(`%s`)", s.config.Hostname)
	}

	return s.__helperCopyTemplate(path, "400", "0:0", data)
}

func (s *service) __deployCopyTraefikHttpTlsConfig() error {
	if s.config.LetsEncrypt.Enabled {
		path := fmt.Sprintf("%s/traefik/etc/conf.d/letsencrypt.yaml", s.libDir())

		data := struct {
			Host string
		}{
			Host: s.config.Hostname,
		}

		if err := s.__helperCopyTemplate(path, "400", "0:0", data); err != nil {
			return err
		}
	}

	if s.config.CustomTLS.Enabled {
		assets := map[string]string{
			"cert": s.config.CustomTLS.CertFilePath,
			"key":  s.config.CustomTLS.KeyFilePath,
		}
		for k, v := range assets {
			pem := fmt.Sprintf("%s/traefik/etc/certs.d/%s.pem", s.libDir(), k)
			if err := s.target.Copy(v, pem, "400", "0:0"); err != nil {
				return &DeployServiceError{Message: err.Error(), Reason: ""}
			}
		}
	}

	return nil
}

func (s *service) __deployCopyAlloyConfig() error {
	path := fmt.Sprintf("%s/alloy/etc/alloy.config", s.libDir())

	data := struct {
		Hostname string
	}{
		Hostname: s.config.Hostname,
	}

	return s.__helperCopyTemplate(path, "400", "0:0", data)
}

func (s *service) __deployCopyFinchConfig() error {
	path := fmt.Sprintf("%s/finch.json", s.libDir())

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	secret := base64.StdEncoding.EncodeToString(key)

	hash := sha256.Sum256([]byte(s.config.Hostname))
	data := Config{
		Id:        hex.EncodeToString(hash[:])[0:16],
		CreatedAt: time.Now().Format(time.RFC3339),
		Hostname:  s.config.Hostname,
		Database:  "sqlite://finch.db",
		Secret:    secret,
		Version:   "0.4.0",
		Credentials: struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: s.config.Username,
			Password: s.config.Password,
		},
	}

	return s.__helperCopyTemplate(path, "400", "10002:1002", data)
}

func (s *service) __deployCopyGrafanaDashboards() error {
	dest := fmt.Sprintf("%s/grafana/dashboards", s.libDir())

	dashboards := []string{
		"grafana-dashboard-logs-docker.json",
		"grafana-dashboard-logs-journal.json",
		"grafana-dashboard-logs-file.json",
		"grafana-dashboard-metrics.json",
		"grafana-dashboard-profiles-finch.json",
	}

	for _, dashboard := range dashboards {
		path := fmt.Sprintf("%s/%s", dest, dashboard)
		if err := s.__helperCopyConfig(path, "400", "472:472"); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) __deployCopyMimirConfig() error {
	path := fmt.Sprintf("%s/mimir/etc/mimir.yaml", s.libDir())
	return s.__helperCopyConfig(path, "400", "10001:10001")
}

func (s *service) __helperCopyConfig(path, mode, owner string) error {
	fileName := filepath.Base(path)

	content, err := fs.ReadFile(Assets, fileName)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", fileName)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.Write(content); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	if err := s.target.Copy(f.Name(), path, mode, owner); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (s *service) __helperCopyTemplate(path, mode, owner string, data any) error {
	fileName := filepath.Base(path)

	tmpl, err := template.New(fileName+".tmpl").ParseFS(Assets, fileName+".tmpl")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", fileName)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.WriteString(buf.String()); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	if err := s.target.Copy(f.Name(), path, mode, owner); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (s *service) __deployConfigs() error {
	if err := s.__deployCopyLokiConfig(); err != nil {
		return err
	}

	if err := s.__deployCopyLokiUserAuthFile(); err != nil {
		return err
	}

	if err := s.__deployCopyTraefikConfig(); err != nil {
		return err
	}

	if err := s.__deployCopyTraefikHttpConfig(); err != nil {
		return err
	}

	if err := s.__deployCopyTraefikHttpTlsConfig(); err != nil {
		return err
	}

	if err := s.__deployCopyAlloyConfig(); err != nil {
		return err
	}

	if err := s.__deployCopyGrafanaDashboards(); err != nil {
		return err
	}

	if err := s.__deployCopyFinchConfig(); err != nil {
		return err
	}

	if err := s.__deployCopyMimirConfig(); err != nil {
		return err
	}

	return nil
}

func (s *service) __deployCopyComposeFile() error {
	path := fmt.Sprintf("%s/docker-compose.yaml", s.libDir())

	data := struct {
		RootUrl  string
		Username string
		Password string
	}{
		RootUrl:  fmt.Sprintf("https://%s", s.config.Hostname),
		Username: s.config.Username,
		Password: s.config.Password,
	}

	return s.__helperCopyTemplate(path, "400", "0:0", data)
}

func (s *service) __deployComposeUp() error {
	out, err := s.target.Run(fmt.Sprintf("sudo docker compose --file %s/docker-compose.yaml up --detach", s.libDir()))
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (s *service) __deployComposeReady() error {
	cmd := `timeout 180 bash -c 'until curl -fs -o /dev/null -w "%{http_code}" http://localhost | grep -qE "^[234][0-9]{2}$"; do sleep 2; done'`

	out, err := s.target.Run(cmd)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (s *service) __deployCompose() error {
	if err := s.__deployCopyComposeFile(); err != nil {
		return err
	}

	if err := s.__deployComposeUp(); err != nil {
		return err
	}

	if err := s.__deployComposeReady(); err != nil {
		return err
	}

	return nil
}

func (s *service) deployService() error {
	if err := s.__deployDirHierachy(); err != nil {
		return err
	}

	if err := s.__deployConfigs(); err != nil {
		return err
	}

	if err := s.__deployCompose(); err != nil {
		return err
	}

	return nil
}
