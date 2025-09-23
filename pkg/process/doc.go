// SPDX-License-Identifier: BSD-3-Clause

// Package process provides utilities for managing service processes within
// an oversight supervision tree. This package bridges the gap between the
// service interface and the oversight process supervisor, enabling robust
// process management with automatic restart capabilities and panic recovery.
//
// The package is designed to work with the u-bmc service architecture where
// multiple services need to be supervised and managed as child processes
// within an oversight tree. It provides panic recovery, error handling, and
// integration with NATS in-process communication.
//
// # Core Functionality
//
// The package provides a single primary function `New()` that creates an
// oversight.ChildProcess wrapper around a service.Service. This wrapper
// handles:
//
//   - Service lifecycle management (start, stop, restart)
//   - Panic recovery with detailed error reporting
//   - Integration with NATS in-process communication
//   - Context-based cancellation and timeout handling
//   - Graceful shutdown coordination
//
// # Basic Usage
//
// Creating a supervised service process:
//
//	type MyService struct {
//		name string
//	}
//
//	func (s *MyService) Name() string { return s.name }
//
//	func (s *MyService) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
//		// Service implementation
//		select {
//		case <-ctx.Done():
//			return ctx.Err()
//		}
//	}
//
//	func setupService() oversight.ChildProcess {
//		svc := &MyService{name: "my-service"}
//		return process.New(svc, ipcConnProvider)
//	}
//
// # Oversight Tree Integration
//
// Integrating multiple services into an oversight supervision tree:
//
//	func setupSupervisionTree() error {
//		// Create NATS in-process connection provider
//		ipcConn := nats.NewInProcessConnProvider()
//
//		// Create individual services
//		authService := &AuthService{}
//		deviceService := &DeviceService{}
//		telemetryService := &TelemetryService{}
//
//		// Wrap services as child processes
//		authChild := process.New(authService, ipcConn)
//		deviceChild := process.New(deviceService, ipcConn)
//		telemetryChild := process.New(telemetryService, ipcConn)
//
//		// Create oversight tree
//		t := oversight.NewTree(
//			oversight.WithSpecification(oversight.Specification{
//				Restart: oversight.Permanent,
//				Strategy: oversight.OneForOne,
//			}),
//			oversight.WithChildren(
//				authChild,
//				deviceChild,
//				telemetryChild,
//			),
//		)
//
//		// Start supervision tree
//		return t.Start(context.Background())
//	}
//
// # Service Implementation Pattern
//
// Recommended pattern for implementing services that work with this package:
//
//	type BMCService struct {
//		name     string
//		config   *Config
//		logger   *slog.Logger
//		server   *http.Server
//		natsConn *nats.Conn
//	}
//
//	func (s *BMCService) Name() string {
//		return s.name
//	}
//
//	func (s *BMCService) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
//		s.logger.Info("Service starting", "service", s.Name())
//
//		// Get NATS connection for IPC
//		nc, err := ipcConn.InProcessConn()
//		if err != nil {
//			return fmt.Errorf("failed to get IPC connection: %w", err)
//		}
//		s.natsConn = nc
//
//		// Initialize service components
//		if err := s.initialize(); err != nil {
//			return fmt.Errorf("service initialization failed: %w", err)
//		}
//
//		// Start HTTP server in a goroutine
//		serverErr := make(chan error, 1)
//		go func() {
//			if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
//				serverErr <- err
//			}
//		}()
//
//		// Wait for context cancellation or server error
//		select {
//		case <-ctx.Done():
//			s.logger.Info("Service shutting down", "service", s.Name())
//			return s.shutdown()
//		case err := <-serverErr:
//			return fmt.Errorf("server error: %w", err)
//		}
//	}
//
// # Panic Recovery and Error Handling
//
// The package automatically handles panics and converts them to errors:
//
//	type PanicProneService struct {
//		name string
//	}
//
//	func (s *PanicProneService) Name() string { return s.name }
//
//	func (s *PanicProneService) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
//		// This panic will be caught and converted to an error
//		if someCondition {
//			panic("something went wrong")
//		}
//
//		// Normal operation
//		return nil
//	}
//
//	// When used with process.New(), the panic becomes:
//	// Error: "panic-prone-service panicked: something went wrong"
//
// # NATS Integration Example
//
// Using NATS for inter-service communication:
//
//	type MessageService struct {
//		name string
//		nc   *nats.Conn
//	}
//
//	func (s *MessageService) Name() string { return s.name }
//
//	func (s *MessageService) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
//		// Get NATS connection
//		nc, err := ipcConn.InProcessConn()
//		if err != nil {
//			return fmt.Errorf("failed to get NATS connection: %w", err)
//		}
//		s.nc = nc
//
//		// Subscribe to messages
//		sub, err := s.nc.Subscribe("bmc.events", s.handleEvent)
//		if err != nil {
//			return fmt.Errorf("failed to subscribe: %w", err)
//		}
//		defer sub.Unsubscribe()
//
//		// Publish service ready event
//		if err := s.nc.Publish("bmc.service.ready", []byte(s.Name())); err != nil {
//			return fmt.Errorf("failed to publish ready event: %w", err)
//		}
//
//		// Wait for context cancellation
//		<-ctx.Done()
//		return nil
//	}
//
//	func (s *MessageService) handleEvent(msg *nats.Msg) {
//		log.Printf("Service %s received event: %s", s.Name(), string(msg.Data))
//
//		// Process event and optionally reply
//		response := fmt.Sprintf("processed by %s", s.Name())
//		msg.Respond([]byte(response))
//	}
//
// # Graceful Shutdown Pattern
//
// Implementing graceful shutdown in services:
//
//	type GracefulService struct {
//		name      string
//		server    *http.Server
//		workers   sync.WaitGroup
//		shutdown  chan struct{}
//	}
//
//	func (s *GracefulService) Name() string { return s.name }
//
//	func (s *GracefulService) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
//		s.shutdown = make(chan struct{})
//
//		// Start background workers
//		s.startWorkers(ctx)
//
//		// Start HTTP server
//		go func() {
//			if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
//				log.Printf("Server error: %v", err)
//			}
//		}()
//
//		// Wait for shutdown signal
//		select {
//		case <-ctx.Done():
//			return s.gracefulShutdown()
//		case <-s.shutdown:
//			return nil
//		}
//	}
//
//	func (s *GracefulService) gracefulShutdown() error {
//		log.Printf("Service %s starting graceful shutdown", s.Name())
//
//		// Shutdown HTTP server with timeout
//		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//		defer cancel()
//
//		if err := s.server.Shutdown(shutdownCtx); err != nil {
//			log.Printf("Server shutdown error: %v", err)
//		}
//
//		// Stop workers and wait for completion
//		close(s.shutdown)
//		s.workers.Wait()
//
//		log.Printf("Service %s shutdown complete", s.Name())
//		return nil
//	}
//
// # Error Propagation and Monitoring
//
// Handling service errors for monitoring and alerting:
//
//	func createMonitoredService(name string) oversight.ChildProcess {
//		svc := &MonitoredService{name: name}
//
//		return func(ctx context.Context) error {
//			// Wrap the service with additional monitoring
//			err := process.New(svc, ipcConn)(ctx)
//
//			if err != nil {
//				// Log error with context
//				log.Printf("Service %s failed: %v", name, err)
//
//				// Send alert or metrics
//				sendServiceAlert(name, err)
//
//				// Could implement custom restart logic here
//				if isRecoverableError(err) {
//					log.Printf("Service %s error is recoverable, restarting...", name)
//					return err // Let oversight handle restart
//				}
//			}
//
//			return err
//		}
//	}
//
// # Testing Services
//
// Testing services that use this package:
//
//	func TestServiceLifecycle(t *testing.T) {
//		// Create test service
//		svc := &TestService{name: "test-service"}
//		ipcConn := nats.NewInProcessConnProvider()
//
//		// Create child process
//		child := process.New(svc, ipcConn)
//
//		// Test with timeout context
//		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//		defer cancel()
//
//		// Run service (should complete within timeout)
//		err := child(ctx)
//
//		// Verify expected behavior
//		if err != context.DeadlineExceeded {
//			t.Errorf("Expected timeout, got: %v", err)
//		}
//	}
//
// # Best Practices
//
// When using this package:
//
//   - Implement proper context handling in service Run() methods
//   - Use structured logging with service names for better observability
//   - Handle NATS connection errors gracefully
//   - Implement graceful shutdown procedures
//   - Avoid long-running blocking operations without context checks
//   - Use appropriate timeouts for external dependencies
//   - Monitor service health and implement health check endpoints
//   - Document service dependencies and startup order requirements
//
// # Performance Considerations
//
// The process wrapper adds minimal overhead:
//
//   - Panic recovery uses defer which has minimal performance impact
//   - Error wrapping creates new error instances but doesn't affect hot paths
//   - NATS connection sharing reduces resource usage across services
//   - Context propagation enables efficient cancellation
//
// For high-performance services:
//
//   - Minimize allocations in service hot paths
//   - Use connection pooling for external resources
//   - Implement proper backpressure mechanisms
//   - Monitor memory usage and goroutine counts
//   - Profile services under load to identify bottlenecks
package process
