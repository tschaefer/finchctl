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

type ListData struct {
	Hostname   string `json:"hostname"`
	ResourceID string `json:"rid"`
}

func (a *agent) listAgents(service string) (*[]ListData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ctx, client, err := grpc.NewClient(ctx, service, api.NewAgentServiceClient)
	if err != nil {
		return nil, &ListAgentsError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = client.Close()
	}()

	list, err := client.Handler().ListAgents(ctx, &api.ListAgentsRequest{})
	if err != nil {
		return nil, &ListAgentsError{Message: err.Error(), Reason: ""}
	}

	result := make([]ListData, 0, len(list.Agents))
	for _, agent := range list.Agents {
		result = append(result, ListData{
			Hostname:   agent.Hostname,
			ResourceID: agent.Rid,
		})
	}

	return &result, nil
}
