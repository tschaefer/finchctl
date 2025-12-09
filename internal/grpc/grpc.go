/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/tschaefer/finchctl/internal/config"
	"github.com/tschaefer/finchctl/internal/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	SkipTLSVerifyEnv string = "FINCH_SKIP_TLS_VERIFY"
)

type Client[T any] struct {
	handler T
	conn    *grpc.ClientConn
}

func NewClient[T any](ctx context.Context, service string, newHandler func(grpc.ClientConnInterface) T) (context.Context, *Client[T], error) {
	token, err := config.LookupStackAuth(service)
	if err != nil {
		return ctx, nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Basic "+token)

	skipTLSVerify := false
	if v, ok := os.LookupEnv(SkipTLSVerifyEnv); ok {
		l := strings.ToLower(v)
		if l == "1" || l == "true" {
			skipTLSVerify = true
		}
	}

	var creds credentials.TransportCredentials
	if skipTLSVerify {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         service,
		}
		creds = credentials.NewTLS(tlsCfg)
	} else {
		creds = credentials.NewClientTLSFromCert(nil, service)
	}

	userAgent := fmt.Sprintf("finchctl/%s", version.Release())

	conn, err := grpc.NewClient(service+":443", grpc.WithTransportCredentials(creds), grpc.WithUserAgent(userAgent))
	if err != nil {
		return ctx, nil, err
	}
	handler := newHandler(conn)

	return ctx, &Client[T]{
		handler: handler,
		conn:    conn,
	}, nil
}

func (c *Client[T]) Handler() T {
	return c.handler
}

func (c *Client[T]) Close() error {
	return c.conn.Close()
}
