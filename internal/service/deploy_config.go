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
	"text/template"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	CreatedAt   string `json:"created_at"`
	Id          string `json:"id"`
	Database    string `json:"database"`
	Secret      string `json:"secret"`
	Hostname    string `json:"hostname"`
	Version     string `json:"version"`
	Credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"credentials"`
}

func (s *service) configLoki() error {
	dest := "/var/lib/finch/loki/etc/loki.yaml"

	content, err := fs.ReadFile(Assets, "loki.yaml")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", "loki.yaml")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.Write(content); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	if err := s.target.Copy(f.Name(), dest, "400", "10001:10001"); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (s *service) configLokiUser() error {
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

	dest := "/var/lib/finch/traefik/etc/conf.d/loki-users.yaml"

	content, err := fs.ReadFile(Assets, "loki-users.yaml.tmpl")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	tmpl, err := template.New("loki-users").Parse(string(content))
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	out := new(bytes.Buffer)
	err = tmpl.Execute(out, data)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", "loki-users")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.WriteString(out.String()); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	if err := s.target.Copy(f.Name(), dest, "400", "root:root"); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (s *service) configTraefik() error {
	dest := "/var/lib/finch/traefik/etc/traefik.yaml"

	content, err := fs.ReadFile(Assets, "traefik.yaml.tmpl")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	tmpl, err := template.New("traefik").Parse(string(content))
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	letsencrypt := s.config.LetsEncrypt.Email
	if letsencrypt == "" {
		letsencrypt = "acme@example.com"
	}
	data := struct {
		Email string
	}{
		Email: letsencrypt,
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", "traefik.yaml")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.WriteString(buf.String()); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	if err := s.target.Copy(f.Name(), dest, "400", "root:root"); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (s *service) configTraefikHttp() error {
	dest := "/var/lib/finch/traefik/etc/conf.d/http.yaml"

	data := struct {
		HostRule string
	}{
		HostRule: "",
	}
	if s.config.LetsEncrypt.Enabled {
		data.HostRule = fmt.Sprintf("&& Host(`%s`)", s.config.Hostname)
	}

	content, err := fs.ReadFile(Assets, "http.yaml.tmpl")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	tmpl, err := template.New("http").Parse(string(content))
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", "http.yaml")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.WriteString(buf.String()); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	if err := s.target.Copy(f.Name(), dest, "400", "root:root"); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (s *service) configTraefikHttpTls() error {
	if s.config.LetsEncrypt.Enabled {
		data := struct {
			Host string
		}{
			Host: s.config.Hostname,
		}

		dest := "/var/lib/finch/traefik/etc/conf.d/letsencrypt.yaml"

		content, err := fs.ReadFile(Assets, "letsencrypt.yaml.tmpl")
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: ""}
		}

		tmpl, err := template.New("letsencrypt").Parse(string(content))
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: ""}
		}

		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, data)
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: ""}
		}

		f, err := os.CreateTemp("", "letsencrypt.yaml")
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: ""}
		}
		defer func() {
			_ = os.Remove(f.Name())
		}()

		if _, err := f.WriteString(buf.String()); err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: ""}
		}

		if err := s.target.Copy(f.Name(), dest, "400", "root:root"); err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: ""}
		}
	}

	if s.config.CustomTLS.Enabled {
		assets := map[string]string{
			"cert": s.config.CustomTLS.CertFilePath,
			"key":  s.config.CustomTLS.KeyFilePath,
		}
		for k, v := range assets {
			if err := s.target.Copy(v, "/var/lib/finch/traefik/etc/certs.d/"+k+".pem", "400", "root:root"); err != nil {
				return &DeployServiceError{Message: err.Error(), Reason: ""}
			}
		}
	}

	return nil
}

func (s *service) configAlloy() error {
	dest := "/var/lib/finch/alloy/etc/alloy.config"

	data := struct {
		Hostname string
	}{
		Hostname: s.config.Hostname,
	}

	content, err := fs.ReadFile(Assets, "alloy.config.tmpl")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	tmpl, err := template.New("alloy").Parse(string(content))
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", "alloy.config")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.WriteString(buf.String()); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	if err := s.target.Copy(f.Name(), dest, "400", "root:root"); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (s *service) configFinch() error {
	dest := "/var/lib/finch/finch.json"

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
		Version:   "0.1.0",
		Credentials: struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: s.config.Username,
			Password: s.config.Password,
		},
	}

	content, err := fs.ReadFile(Assets, "finch.json.tmpl")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	tmpl, err := template.New("finch").Parse(string(content))
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	f, err := os.CreateTemp("", "finch.json")
	if err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.WriteString(buf.String()); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	if err := s.target.Copy(f.Name(), dest, "400", "root:root"); err != nil {
		return &DeployServiceError{Message: err.Error(), Reason: ""}
	}

	return nil
}

func (s *service) configGrafanaDashboards() error {
	dest := "/var/lib/finch/grafana/dashboards"

	dashboards := []string{
		"grafana-dashboard-logs-docker.json",
		"grafana-dashboard-logs-journal.json",
		"grafana-dashboard-logs-file.json",
	}

	for _, dashboard := range dashboards {
		content, err := fs.ReadFile(Assets, dashboard)
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: ""}
		}
		f, err := os.CreateTemp("", dashboard)
		if err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: ""}
		}
		defer func() {
			_ = os.Remove(f.Name())
		}()
		if _, err := f.Write(content); err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: ""}
		}
		if err := s.target.Copy(f.Name(), dest+"/"+dashboard, "400", "472:472"); err != nil {
			return &DeployServiceError{Message: err.Error(), Reason: ""}
		}
	}

	return nil
}

func (s *service) configSetup() error {
	if err := s.configLoki(); err != nil {
		return err
	}

	if err := s.configLokiUser(); err != nil {
		return err
	}

	if err := s.configTraefik(); err != nil {
		return err
	}

	if err := s.configTraefikHttp(); err != nil {
		return err
	}

	if err := s.configTraefikHttpTls(); err != nil {
		return err
	}

	if err := s.configAlloy(); err != nil {
		return err
	}

	if err := s.configGrafanaDashboards(); err != nil {
		return err
	}

	if err := s.configFinch(); err != nil {
		return err
	}

	return nil
}
