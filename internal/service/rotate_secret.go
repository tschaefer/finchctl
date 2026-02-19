/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

func (s *Service) rotateServiceSecret() error {
	if err := s.__updateSetTargetConfiguration(); err != nil {
		return convertError(err, &RotateCertificateError{})
	}

	cfgPath := fmt.Sprintf("%s/finch.json", s.libDir())
	out, err := s.target.Run(fmt.Sprintf("sudo cat %s", cfgPath))
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

	out, err = s.target.Run(fmt.Sprintf("sudo docker compose --file %s/docker-compose.yaml restart finch", s.libDir()))
	if err != nil {
		return &RotateServiceSecretError{Message: err.Error(), Reason: string(out)}
	}

	return nil
}
