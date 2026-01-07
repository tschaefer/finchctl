/*
Copyright (c) Tobias Schäfer. All rights reserved.
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

type DescribeLogsJournal struct {
	Enable bool `json:"enable"`
}

type DescribeLogsDocker struct {
	Enable bool `json:"enable"`
}

type DescribeLogs struct {
	Files   []string            `json:"files"`
	Journal DescribeLogsJournal `json:"journal"`
	Docker  DescribeLogsDocker  `json:"docker"`
}

type DescribeMetrics struct {
	Enable  bool     `json:"enable"`
	Targets []string `json:"targets"`
}

type DescribeProfiles struct {
	Enable bool `json:"enable"`
}

type DescribeData struct {
	ResourceID string           `json:"rid"`
	Hostname   string           `json:"hostname"`
	Labels     []string         `json:"labels"`
	Logs       DescribeLogs     `json:"logs"`
	Metrics    DescribeMetrics  `json:"metrics"`
	Profiles   DescribeProfiles `json:"profiles"`
}

func (a *Agent) describeAgent(service, rid string) (*DescribeData, error) {
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
	files := []string{}
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

		files = append(files, url.Path)
	}

	labels := data.Labels
	if labels == nil {
		labels = []string{}
	}

	metricsTargets := data.MetricsTargets
	if metricsTargets == nil {
		metricsTargets = []string{}
	}

	return &DescribeData{
		ResourceID: data.ResourceId,
		Hostname:   data.Hostname,
		Labels:     labels,
		Logs: DescribeLogs{
			Files: files,
			Journal: DescribeLogsJournal{
				Enable: journal,
			},
			Docker: DescribeLogsDocker{
				Enable: docker,
			},
		},
		Metrics: DescribeMetrics{
			Enable:  data.Metrics,
			Targets: metricsTargets,
		},
		Profiles: DescribeProfiles{
			Enable: data.Profiles,
		},
	}, nil
}
