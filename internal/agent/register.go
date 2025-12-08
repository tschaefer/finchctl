/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"context"
	"time"

	"github.com/tschaefer/finchctl/internal/api"
	"github.com/tschaefer/finchctl/internal/grpc"
)

type RegisterData struct {
	Hostname       string   `json:"hostname"`
	LogSources     []string `json:"log_sources"`
	Metrics        bool     `json:"metrics"`
	MetricsTargets []string `json:"metrics_targets"`
	Profiles       bool     `json:"profiles"`
	Labels         []string `json:"labels"`
}

func (a *agent) registerAgent(service string, data *RegisterData) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ctx, client, err := grpc.NewClient(ctx, service, api.NewAgentServiceClient)
	if err != nil {
		return nil, &RegisterAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = client.Close()
	}()

	register, err := client.Handler().RegisterAgent(ctx, &api.RegisterAgentRequest{
		Hostname:       data.Hostname,
		LogSources:     data.LogSources,
		Metrics:        data.Metrics,
		MetricsTargets: data.MetricsTargets,
		Profiles:       data.Profiles,
		Labels:         data.Labels,
	})
	if err != nil {
		return nil, &RegisterAgentError{Message: err.Error(), Reason: ""}
	}

	cfg, err := client.Handler().GetAgentConfig(ctx, &api.GetAgentConfigRequest{
		Rid: register.Rid,
	})
	if err != nil {
		return nil, &ConfigAgentError{Message: err.Error(), Reason: ""}
	}

	return cfg.Config, nil
}
