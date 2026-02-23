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

func (a *Agent) configAgent(service, rid string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()

	ctx, client, err := grpc.NewClient(ctx, service, api.NewAgentServiceClient)
	if err != nil {
		return nil, &ConfigAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = client.Close()
	}()

	cfg, err := client.Handler().GetAgentConfig(ctx, &api.GetAgentConfigRequest{
		Rid: rid,
	})
	if err != nil {
		return nil, &ConfigAgentError{Message: err.Error(), Reason: ""}
	}

	return cfg.Config, nil
}
