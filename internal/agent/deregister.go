package agent

import (
	"context"
	"time"

	"github.com/tschaefer/finchctl/internal/api"
	"github.com/tschaefer/finchctl/internal/grpc"
)

func (a *agent) deregisterAgent(service, resourceID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ctx, client, err := grpc.NewClient(ctx, service, api.NewAgentServiceClient)
	if err != nil {
		return &DeregisterAgentError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = client.Close()
	}()

	_, err = client.Handler().DeregisterAgent(ctx, &api.DeregisterAgentRequest{
		Rid: resourceID,
	})
	if err != nil {
		return &DeregisterAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}
