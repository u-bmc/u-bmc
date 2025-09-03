// SPDX-License-Identifier: BSD-3-Clause

// Package websrv provides a web server implementation that supports both HTTP/2 and HTTP/3
// protocols with TLS encryption. It serves Connect RPC APIs and optionally a web UI.
package websrv

import (
	"context"
	"fmt"
	"time"

	"github.com/lorenzosaino/go-sysctl"
	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/cert"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that WebSrv implements service.Service.
var _ service.Service = (*WebSrv)(nil)

// WebSrv is a web server that provides HTTP/2 and HTTP/3 endpoints for BMC operations.
type WebSrv struct {
	config
}

// New creates a new WebSrv instance with the provided options.
func New(opts ...Option) *WebSrv {
	cfg := &config{
		name:         "websrv",
		addr:         ":443",
		webui:        false,
		webuiPath:    "/usr/share/webui",
		readTimeout:  5 * time.Second,
		writeTimeout: 5 * time.Second,
		idleTimeout:  120 * time.Second,
		rmemMax:      "7500000",
		wmemMax:      "7500000",
		certConfig:   cert.NewConfig(),
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}

	// Ensure certificate configuration has proper defaults
	cfg.SetCertDefaults()

	return &WebSrv{
		config: *cfg,
	}
}

// Name returns the service name.
func (s *WebSrv) Name() string {
	return s.name
}

// Run starts the web server and blocks until the context is canceled.
func (s *WebSrv) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting web server",
		"service", s.name,
		"addr", s.addr,
		"webui", s.webui,
		"cert_type", s.GetCertConfig().Type,
	)

	// Configure system parameters for optimal QUIC performance
	if err := s.configureSysctl(ctx); err != nil {
		l.WarnContext(ctx, "Failed to configure sysctls for QUIC", "error", err)
	}

	// Setup HTTP router with all endpoints and middleware
	router, err := s.setupRouter()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSetupRouter, err)
	}

	// Setup TLS configuration and certificates
	tlsConfig, httpHandler, err := s.setupTLS()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSetupTLS, err)
	}

	// Perform health check before starting
	if err := s.HealthCheck(router, tlsConfig); err != nil {
		return fmt.Errorf("server health check failed: %w", err)
	}

	l.DebugContext(ctx, "Server configuration",
		"listen_addresses", s.GetListenAddresses(),
		"server_info", s.GetServerInfo(),
	)

	// Start all servers
	return s.StartServers(ctx, router, tlsConfig, httpHandler)
}

// configureSysctl sets kernel parameters needed for optimal QUIC performance.
func (s *WebSrv) configureSysctl(ctx context.Context) error {
	if err := sysctl.Set("net.core.rmem_max", s.rmemMax); err != nil {
		return fmt.Errorf("%w: %w", ErrSetRmemMax, err)
	}
	if err := sysctl.Set("net.core.wmem_max", s.wmemMax); err != nil {
		return fmt.Errorf("%w: %w", ErrSetWmemMax, err)
	}
	return nil
}
