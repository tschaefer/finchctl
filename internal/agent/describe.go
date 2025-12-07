/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"context"
	"net/url"
	"time"

	"github.com/tschaefer/finchctl/internal/api"
	"github.com/tschaefer/finchctl/internal/grpc"
)

type DescribeData struct {
	Hostname       string   `json:"hostname"`
	ResourceID     string   `json:"resource_id"`
	Files          []string `json:"files"`
	Journal        bool     `json:"journal"`
	Docker         bool     `json:"docker"`
	Metrics        bool     `json:"metrics"`
	MetricsTargets []string `json:"metrics_targets"`
	Profiles       bool     `json:"profiles"`
	Labels         []string `json:"labels"`
}

func (a *agent) describeAgent(service, rid string) (*DescribeData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ctx, client, err := grpc.NewClient(ctx, service, api.NewAgentServiceClient)
	if err != nil {
		return nil, &DescribeAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = client.Close()
	}()

	data, err := client.Handler().GetAgent(ctx, &api.GetAgentRequest{
		Rid: rid,
	})
	if err != nil {
		return nil, &DescribeAgentError{Message: err.Error(), Reason: ""}
	}

	journal := false
	docker := false
	logSources := []string{}
	for _, src := range data.LogSources {
		url, err := url.Parse(src)
		if err != nil {
			continue
		}

		switch url.Scheme {
		case "journal":
			journal = true
			continue
		case "docker":
			docker = true
			continue
		}

		logSources = append(logSources, url.Path)
	}

	labels := data.Tags
	if labels == nil {
		labels = []string{}
	}

	metricsTargets := data.MetricsTargets
	if metricsTargets == nil {
		metricsTargets = []string{}
	}

	return &DescribeData{
		Hostname:       data.Hostname,
		ResourceID:     data.ResourceId,
		Files:          logSources,
		Journal:        journal,
		Docker:         docker,
		Metrics:        data.Metrics,
		MetricsTargets: metricsTargets,
		Profiles:       data.Profiles,
		Labels:         labels,
	}, nil
}
