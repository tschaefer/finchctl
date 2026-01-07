/*
Copyright (c) Tobias Schäfer. All rights reservem.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"context"
	"time"

	"github.com/tschaefer/finchctl/internal/api"
	"github.com/tschaefer/finchctl/internal/grpc"
)

type InfoData struct {
	ID        string `json:"id"`
	Hostname  string `json:"hostname"`
	CreatedAt string `json:"created_at"`
	Release   string `json:"release"`
	Commit    string `json:"commit"`
}

func (s *Service) infoService() (*InfoData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ctx, client, err := grpc.NewClient(ctx, s.config.Hostname, api.NewInfoServiceClient)
	if err != nil {
		return nil, &InfoServiceError{Message: err.Error(), Reason: ""}
	}
	defer func() {
		_ = client.Close()
	}()

	info, err := client.Handler().GetServiceInfo(ctx, &api.GetServiceInfoRequest{})
	if err != nil {
		return nil, &InfoServiceError{Message: err.Error(), Reason: ""}
	}

	return &InfoData{
		ID:        info.Id,
		Hostname:  info.Hostname,
		CreatedAt: info.CreatedAt,
		Release:   info.Release,
		Commit:    info.Commit,
	}, nil
}
