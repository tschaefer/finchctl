/*
Copyright (c) Tobias Schäfer. All rights reserved.
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

	"github.com/tschaefer/finchctl/internal/config"
	"github.com/tschaefer/finchctl/internal/mtls"
	"github.com/tschaefer/finchctl/internal/version"
)

func (s *Service) __deployMakeDirHierarchy() error {
	directories := []string{
		"grafana/dashboards",
		"loki/{data,etc}",
		"alloy/{data,etc}",
		"traefik/etc/{certs.d,conf.d}",
		"mimir/{data,etc}",
		"pyroscope/data",
	}
	for _, dir := range directories {
		out, err := s.target.Run("sudo mkdir -p " + filepath.Join(s.libDir(), dir))
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: string(out)}
		}
	}

	return nil
}

func (s *Service) __deploySetDirHierarchyPermission() error {
	ownership := map[string]string{
		"grafana":                      "472:472",
		"grafana/dashboards":           "472:472",
		"loki":                         "10001:10001",
		"loki/{data,etc}":              "10001:10001",
		"alloy":                        "0:0",
		"alloy/{data,etc}":             "0:0",
		"traefik":                      "0:0",
		"traefik/etc":                  "0:0",
		"traefik/etc/{certs.d,conf.d}": "0:0",
		"mimir":                        "10001:10001",
		"mimir/{data,etc}":             "10001:10001",
		"pyroscope":                    "10001:10001",
		"pyroscope/data":               "10001:10001",
	}

	for path, owner := range ownership {
		cmd := fmt.Sprintf("sudo chown %s %s", owner, filepath.Join(s.libDir(), path))
		out, err := s.target.Run(cmd)
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: string(out)}
		}
	}

	return nil
}

func (s *Service) __deployCopyLokiConfig() error {
	path := filepath.Join(s.libDir(), "loki/etc/loki.yaml")
	return s.__helperCopyConfig(path, "400", "10001:10001")
}

func (s *Service) __deployCopyTraefikConfig() error {
	path := filepath.Join(s.libDir(), "traefik/etc/traefik.yaml")

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

func (s *Service) __deployCopyTraefikHttpConfig() error {
	path := filepath.Join(s.libDir(), "traefik/etc/conf.d/http.yaml")

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

func (s *Service) __deployCopyTraefikHttpTlsConfig() error {
	if s.config.LetsEncrypt.Enabled {
		path := filepath.Join(s.libDir(), "traefik/etc/conf.d/letsencrypt.yaml")

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
			pem := filepath.Join(s.libDir(), "traefik/etc/certs.d", k+".pem")
			if err := s.target.Copy(v, pem, "400", "0:0"); err != nil {
				return &DeployServiceError{Message: err.Error(), Reason: ""}
			}
		}
	}

	return nil
}

func (s *Service) __deployGenerateMTLSCertificates() error {
	caCertPEM, caKeyPEM, err := mtls.GenerateCA(s.config.Hostname)
	if err != nil {
		return &DeployServiceError{Message: "failed to generate CA certificate", Reason: err.Error()}
	}

	clientCertPEM, clientKeyPEM, err := mtls.GenerateClient(s.config.Hostname, caCertPEM, caKeyPEM)
	if err != nil {
		return &DeployServiceError{Message: "failed to generate client certificate", Reason: err.Error()}
	}

	caCertPath := filepath.Join(s.libDir(), "traefik/etc/certs.d", version.ResourceID()+".pem")
	f, err := os.CreateTemp("", "ca.pem")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.Write(caCertPEM); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	if err := s.target.Copy(f.Name(), caCertPath, "400", "0:0"); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	if s.dryRun {
		return nil
	}

	if err := config.UpdateStack(s.config.Hostname, clientCertPEM, clientKeyPEM); err != nil {
		return &DeployServiceError{Message: "failed to update stack certificates", Reason: err.Error()}
	}

	return nil
}

func (s *Service) __deployCopyAlloyConfig() error {
	path := filepath.Join(s.libDir(), "alloy/etc/alloy.config")

	data := struct {
		Hostname string
	}{
		Hostname: s.config.Hostname,
	}

	return s.__helperCopyTemplate(path, "400", "0:0", data)
}

func (s *Service) __deployCopyFinchConfig() error {
	path := filepath.Join(s.libDir(), "finch.json")

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	secret := base64.StdEncoding.EncodeToString(key)

	hash := sha256.Sum256([]byte(s.config.Hostname))
	data := FinchConfig{
		Id:        hex.EncodeToString(hash[:])[0:16],
		CreatedAt: time.Now().Format(time.RFC3339),
		Hostname:  s.config.Hostname,
		Database:  "sqlite://finch.db",
		Profiler:  "http://pyroscope:4040",
		Secret:    secret,
		Version:   "1.7.0",
	}

	return s.__helperCopyTemplate(path, "400", "0:0", data)
}

func (s *Service) __deployCopyGrafanaDashboards() error {
	dest := filepath.Join(s.libDir(), "grafana/dashboards")

	dashboards := []string{
		"grafana-dashboard-logs-docker.json",
		"grafana-dashboard-logs-journal.json",
		"grafana-dashboard-logs-file.json",
		"grafana-dashboard-metrics.json",
		"grafana-dashboard-profiles-finch.json",
	}

	for _, dashboard := range dashboards {
		path := filepath.Join(dest, dashboard)
		if err := s.__helperCopyConfig(path, "400", "472:472"); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) __deployCopyMimirConfig() error {
	path := filepath.Join(s.libDir(), "mimir/etc/mimir.yaml")
	return s.__helperCopyConfig(path, "400", "10001:10001")
}

func (s *Service) __helperCopyConfig(path, mode, owner string) error {
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

func (s *Service) __helperCopyTemplate(path, mode, owner string, data any) error {
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

func (s *Service) __deployCopyComposeFile() error {
	path := filepath.Join(s.libDir(), "docker-compose.yaml")

	data := struct {
		RootUrl string
	}{
		RootUrl: fmt.Sprintf("https://%s", s.config.Hostname),
	}

	return s.__helperCopyTemplate(path, "400", "0:0", data)
}

func (s *Service) __deployComposeUp() error {
	out, err := s.target.Run("sudo docker compose --file " + filepath.Join(s.libDir(), "docker-compose.yaml") + " up --detach")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (s *Service) __deployComposeReady() error {
	cmd := `timeout 180 bash -c 'until curl -fs -o /dev/null -w "%{http_code}" http://localhost | grep -qE "^[234][0-9]{2}$"; do sleep 2; done'`

	out, err := s.target.Run(cmd)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}

func (s *Service) deployService() error {
	if err := s.__deployMakeDirHierarchy(); err != nil {
		return err
	}

	if err := s.__deploySetDirHierarchyPermission(); err != nil {
		return err
	}

	if err := s.__deployCopyLokiConfig(); err != nil {
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

	if err := s.__deployGenerateMTLSCertificates(); err != nil {
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
