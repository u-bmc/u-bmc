// SPDX-License-Identifier: BSD-3-Clause

package websrv

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lorenzosaino/go-sysctl"
	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/cert"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Compile-time assertion that WebSrv implements service.Service.
var _ service.Service = (*WebSrv)(nil)

// WebSrv is a web server that provides HTTP/2 and HTTP/3 endpoints for BMC operations.
type WebSrv struct {
	config *config
	logger *slog.Logger
	tracer trace.Tracer
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
		config: cfg,
	}
}

// Name returns the service name.
func (s *WebSrv) Name() string {
	return s.config.name
}

// Run starts the web server and blocks until the context is canceled.
func (s *WebSrv) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	s.tracer = otel.Tracer(s.Name())

	ctx, span := s.tracer.Start(ctx, "Run")
	defer span.End()

	s.logger = log.GetGlobalLogger().With("service", s.Name())

	s.logger.InfoContext(ctx, "Starting web server",
		"addr", s.config.addr,
		"webui", s.config.webui,
		"cert_type", s.config.GetCertConfig().Type,
	)

	span.SetAttributes(
		attribute.String("service.name", s.config.name),
		attribute.String("server.addr", s.config.addr),
		attribute.Bool("webui.enabled", s.config.webui),
		attribute.Int("cert.type", int(s.config.GetCertConfig().Type)),
	)

	// Configure system parameters for optimal QUIC performance
	if err := s.configureSysctl(ctx); err != nil {
		s.logger.WarnContext(ctx, "Failed to configure sysctls for QUIC", "error", err)
	}

	nc, err := nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer nc.Drain() //nolint:errcheck

	router, err := s.setupRouter(nc)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrSetupRouter, err)
	}

	tlsConfig, httpHandler, err := s.setupTLS()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrSetupTLS, err)
	}

	// Perform health check before starting
	if err := s.HealthCheck(router, tlsConfig); err != nil {
		span.RecordError(err)
		return fmt.Errorf("server health check failed: %w", err)
	}

	s.logger.DebugContext(ctx, "Server configuration",
		"listen_addresses", s.GetListenAddresses(),
		"server_info", s.GetServerInfo(),
	)

	s.logger.InfoContext(ctx, "Web server started successfully")

	err = s.StartServers(ctx, router, tlsConfig, httpHandler)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

// configureSysctl sets kernel parameters needed for optimal QUIC performance.
func (s *WebSrv) configureSysctl(ctx context.Context) error {
	if err := sysctl.Set("net.core.rmem_max", s.config.rmemMax); err != nil {
		return fmt.Errorf("%w: %w", ErrSetRmemMax, err)
	}
	if err := sysctl.Set("net.core.wmem_max", s.config.wmemMax); err != nil {
		return fmt.Errorf("%w: %w", ErrSetWmemMax, err)
	}
	return nil
}
