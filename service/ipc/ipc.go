// SPDX-License-Identifier: BSD-3-Clause

package ipc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Compile-time assertion that IPC implements service.Service.
var _ service.Service = (*IPC)(nil)

// IPC provides an embedded NATS server for inter-process communication
// within the u-bmc system. It acts as the central message bus for all
// other services in the BMC.
//
// The IPC service creates and manages a NATS server instance that runs
// embedded within the u-bmc process, eliminating the need for external
// NATS server dependencies. It provides JetStream capabilities for
// persistent messaging and state management across BMC components.
type IPC struct {
	config *config
	server *server.Server
	logger *slog.Logger
	tracer trace.Tracer
}

// New creates a new IPC service instance with the provided configuration options.
//
// The IPC service is configured with sensible defaults but can be customized
// using the provided Option functions. Common configurations include:
//
//   - Setting custom storage directories for JetStream
//   - Enabling debug or trace logging
//   - Configuring memory and storage limits
//   - Setting connection limits and timeouts
//
// Example usage:
//
//	ipcService := ipc.New(
//		ipc.WithServiceName("bmc-ipc"),
//		ipc.WithStoreDir("/var/lib/u-bmc/ipc"),
//		ipc.WithMaxMemory(128 * 1024 * 1024), // 128MB
//	)
func New(opts ...Option) *IPC {
	cfg := &config{
		serviceName:                 DefaultServiceName,
		serviceDescription:          DefaultServiceDescription,
		serviceVersion:              DefaultServiceVersion,
		serverName:                  DefaultServerName,
		storeDir:                    DefaultStoreDir,
		enableJetStream:             true,
		dontListen:                  true,
		maxMemory:                   DefaultMaxMemory,
		maxStorage:                  DefaultMaxStorage,
		startupTimeout:              DefaultStartupTimeout,
		shutdownTimeout:             DefaultShutdownTimeout,
		maxConnections:              0, // 0 means unlimited
		maxControlLine:              1024,
		maxPayload:                  1048576, // 1MB
		writeDeadline:               2 * time.Second,
		pingInterval:                2 * time.Minute,
		maxPingsOut:                 2,
		enableSlowConsumerDetection: true,
		slowConsumerThreshold:       5 * time.Second,
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &IPC{
		config: cfg,
	}
}

// Name returns the service name as configured.
// This implements the service.Service interface.
func (s *IPC) Name() string {
	return s.config.serviceName
}

// Run starts the IPC service and runs the embedded NATS server.
//
// This method implements the service.Service interface and handles the
// complete lifecycle of the NATS server:
//
//  1. Validates the service configuration
//  2. Creates and configures the NATS server
//  3. Starts the server and waits for it to be ready
//  4. Runs until the context is canceled
//  5. Performs graceful shutdown
//
// The method will return an error if:
//   - The configuration is invalid
//   - The NATS server cannot be created or started
//   - The server fails to become ready within the startup timeout
//   - An existing IPC connection is provided (not supported)
//
// The ipcConn parameter should be nil for the IPC service, as it provides
// the IPC infrastructure rather than consuming it.
func (s *IPC) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	s.tracer = otel.Tracer(s.config.serviceName)

	ctx, span := s.tracer.Start(ctx, "Run")
	defer span.End()

	s.logger = log.GetGlobalLogger().With("service", s.config.serviceName)
	s.logger.InfoContext(ctx, "Starting IPC service",
		"version", s.config.serviceVersion,
		"server_name", s.config.serverName,
		"jetstream_enabled", s.config.enableJetStream,
		"store_dir", s.config.storeDir)

	// We might be able to handle this in the future, for now bail out
	if ipcConn != nil {
		err := fmt.Errorf("existing IPC found, bailing out")
		span.RecordError(err)
		return err
	}

	if err := s.config.Validate(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrInvalidConfiguration, err)
	}

	// Create the embedded NATS server with configured options
	serverOpts := s.config.ToServerOptions()
	ns, err := server.NewServer(serverOpts)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrServerCreationFailed, err)
	}
	s.server = ns
	s.server.SetLoggerV2(log.NewNATSLogger(s.logger), true, false, false)

	// Start the server
	s.logger.InfoContext(ctx, "Starting NATS server", "server_name", s.config.serverName)
	s.server.Start()

	// Wait for server to be ready with timeout
	if !s.server.ReadyForConnections(s.config.startupTimeout) {
		s.server.Shutdown()
		err := fmt.Errorf("%w: server not ready within %v", ErrServerTimeout, s.config.startupTimeout)
		span.RecordError(err)
		return err
	}

	s.logger.InfoContext(ctx, "IPC server started successfully",
		"server_name", s.config.serverName,
		"server_id", s.server.ID(),
		"jetstream_enabled", s.config.enableJetStream)

	span.SetAttributes(
		attribute.String("service.name", s.config.serviceName),
		attribute.String("service.version", s.config.serviceVersion),
		attribute.String("server.name", s.config.serverName),
		attribute.String("server.id", s.server.ID()),
		attribute.Bool("jetstream.enabled", s.config.enableJetStream),
	)

	// Wait for shutdown signal
	<-ctx.Done()

	// Perform graceful shutdown
	return s.shutdown(ctx)
}

// GetConnProvider returns a connection provider that can be used by other
// services to obtain in-process connections to the NATS server.
//
// The returned ConnProvider will wait for the server to be available
// before providing connections. It includes built-in retry logic and
// timeout handling to ensure robust connection establishment.
//
// This method can be called before the server is fully started, as the
// ConnProvider will block and poll until the server becomes available
// or a timeout occurs.
//
// Example usage:
//
//	provider := ipcService.GetConnProvider()
//	conn, err := provider.InProcessConn()
//	if err != nil {
//		// Handle connection error
//	}
//	defer conn.Close()
func (s *IPC) GetConnProvider() *ConnProvider {
	// Block and poll until s.server is not nil for a maximum of the startup timeout.
	// Usually it takes between 2 and 10 milliseconds to start the server,
	// so polling once every millisecond should be fine and the startup timeout
	// is more than enough. If it takes longer than that, something is probably wrong.
	// In that case, we return a ConnProvider with a nil server, which will
	// cause any connection attempts to fail.
	timeout := time.Now().Add(s.config.startupTimeout)
	for s.server == nil && time.Now().Before(timeout) {
		time.Sleep(1 * time.Millisecond)
	}

	return &ConnProvider{
		server: s.server,
	}
}

// shutdown performs graceful shutdown of the NATS server.
//
// This method attempts to perform a lame duck shutdown first, which allows
// existing connections to drain gracefully. If the shutdown does not complete
// within the configured timeout, it will force shutdown the server.
//
// The method always returns the original context error to indicate why
// the shutdown was initiated.
func (s *IPC) shutdown(ctx context.Context) error {
	// Capture the context error before creating shutdown context
	err := ctx.Err()

	// Create a new context for shutdown operations since the original is canceled
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.config.shutdownTimeout)
	defer cancel()

	s.logger.InfoContext(shutdownCtx, "Shutting down IPC server",
		"shutdown_timeout", s.config.shutdownTimeout)

	if s.server != nil {
		// Attempt graceful shutdown with lame duck mode
		s.logger.InfoContext(shutdownCtx, "Initiating lame duck shutdown")
		s.server.LameDuckShutdown()

		// Wait for shutdown to complete or timeout
		done := make(chan struct{})
		go func() {
			defer close(done)
			s.server.Shutdown()
		}()

		select {
		case <-done:
			s.logger.InfoContext(shutdownCtx, "IPC server shutdown completed")
		case <-shutdownCtx.Done():
			s.logger.WarnContext(shutdownCtx, "IPC server shutdown timed out, forcing shutdown")
			// Server.Shutdown() should handle this gracefully even if called multiple times
		}
	}

	return err
}
