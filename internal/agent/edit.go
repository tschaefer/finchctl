/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"context"
	"time"

	"github.com/tschaefer/finchctl/internal/api"
	"github.com/tschaefer/finchctl/internal/grpc"
)

type EditData struct {
	ResourceId     string   `json:"resource_id"`
	LogSources     []string `json:"log_sources"`
	Metrics        bool     `json:"metrics"`
	MetricsTargets []string `json:"metrics_targets"`
	Profiles       bool     `json:"profiles"`
	Labels         []string `json:"labels"`
}

func (a *Agent) editAgent(service string, data *EditData) error {
	ctx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()

	ctx, client, err := grpc.NewClient(ctx, service, api.NewAgentServiceClient)
	if err != nil {
		return &EditAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = client.Close()
	}()

	_, err = client.Handler().UpdateAgent(ctx, &api.UpdateAgentRequest{
		Rid:            data.ResourceId,
		LogSources:     data.LogSources,
		Metrics:        data.Metrics,
		MetricsTargets: data.MetricsTargets,
		Profiles:       data.Profiles,
		Labels:         data.Labels,
	})
	if err != nil {
		return &EditAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}
