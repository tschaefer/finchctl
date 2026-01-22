/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/tschaefer/finchctl/internal/config"
	"github.com/tschaefer/finchctl/internal/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	SkipTLSVerifyEnv string = "FINCH_SKIP_TLS_VERIFY"
)

type Client[T any] struct {
	handler T
	conn    *grpc.ClientConn
}

func NewClient[T any](ctx context.Context, service string, newHandler func(grpc.ClientConnInterface) T) (context.Context, *Client[T], error) {
	certPEM, keyPEM, err := config.LookupStack(service)
	if err != nil {
		return ctx, nil, err
	}

	// Parse the certificate and key from PEM
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to parse client certificate: %w", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   service,
	}
	if skipTLSVerify() {
		tlsCfg.InsecureSkipVerify = true
	}
	creds := credentials.NewTLS(tlsCfg)

	userAgent := fmt.Sprintf("finchctl/%s", version.Release())

	ip := net.ParseIP(service)
	if ip != nil && ip.To4() == nil {
		service = fmt.Sprintf("[%s]", service)
	}

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

func skipTLSVerify() bool {
	if v, ok := os.LookupEnv(SkipTLSVerifyEnv); ok {
		l := strings.ToLower(v)
		if l == "1" || l == "true" {
			return true
		}
	}
	return false
}
