/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"path"

	"github.com/tschaefer/finchctl/internal/config"
)

func (s *Service) rotateServiceSecret() error {
	if err := s.__updateSetTargetConfiguration(); err != nil {
		return convertError(err, &RotateServiceSecretError{})
	}

	if !s.dryRun {
		if _, err := config.LookupStack(s.config.Hostname); err != nil {
			return &RotateServiceSecretError{Message: err.Error(), Reason: ""}
		}
	}

	cfgPath := path.Join(s.libDir(), "finch.json")
	out, err := s.target.Run("sudo cat " + cfgPath)
	if err != nil {
		return &RotateServiceSecretError{Message: err.Error(), Reason: string(out)}
	}
	if s.dryRun {
		out = []byte(`{}`)
	}

	cfg := FinchConfig{}
	err = json.Unmarshal(out, &cfg)
	if err != nil {
		return &RotateServiceSecretError{Message: err.Error(), Reason: string(out)}
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return &RotateServiceSecretError{Message: err.Error(), Reason: ""}
	}
	secret := base64.StdEncoding.EncodeToString(key)
	cfg.Secret = secret

	if err := s.__helperCopyTemplate(cfgPath, "400", "0:0", cfg); err != nil {
		return convertError(err, &RotateServiceSecretError{})
	}

	out, err = s.target.Run("sudo docker compose --file " + path.Join(s.libDir(), "docker-compose.yaml") + " restart finch")
	if err != nil {
		return &RotateServiceSecretError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
