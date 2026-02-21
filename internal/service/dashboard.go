/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"context"
	"time"

	"github.com/tschaefer/finchctl/internal/api"
	"github.com/tschaefer/finchctl/internal/grpc"
)

type DashboardData struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	Url       string `json:"url"`
}

func (s *Service) dashboardService(sessionTimeout int32, role string, scope []string) (*DashboardData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ctx, client, err := grpc.NewClient(ctx, s.config.Hostname, api.NewDashboardServiceClient)
	if err != nil {
		return nil, &DashboardServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = client.Close()
	}()

	data, err := client.Handler().GetDashboardToken(ctx, &api.GetDashboardTokenRequest{
		SessionTimeout: &sessionTimeout,
		Role:           role,
		Scope:          scope,
	})
	if err != nil {
		return nil, &DashboardServiceError{Message: err.Error(), Reason: ""}
	}

	return &DashboardData{
		Token:     data.Token,
		ExpiresAt: data.ExpiresAt,
		Url:       data.DashboardUrl,
	}, nil
}
