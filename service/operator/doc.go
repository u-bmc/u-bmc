// SPDX-License-Identifier: BSD-3-Clause

// Package operator provides a service orchestrator that manages and supervises
// multiple BMC services in a fault-tolerant manner. It acts as the central
// coordinator for all BMC subsystems, handling service lifecycle management,
// inter-process communication setup, and providing a supervision tree for
// automatic service recovery.
//
// The operator service is the main entry point for the u-bmc system and is
// responsible for starting, monitoring, and coordinating all other BMC services.
// It implements a robust supervision strategy that automatically restarts
// failed services and maintains system stability.
//
// # Core Features
//
//   - Service lifecycle management and orchestration
//   - Fault-tolerant supervision with automatic restart policies
//   - Inter-process communication coordination via NATS
//   - Configurable service selection and ordering
//   - System initialization and mount point management
//   - OpenTelemetry integration for observability
//   - Graceful shutdown handling
//
// # Architecture
//
// The operator follows a supervision tree pattern where services are organized
// in a hierarchical structure with well-defined restart policies. The operator
// itself acts as the root supervisor, managing child services and handling
// their failures according to configured strategies.
//
// The supervision tree includes:
//   - IPC service (highest priority, started first)
//   - Core BMC services (power, thermal, sensors, etc.)
//   - Management services (web interface, IPMI, etc.)
//   - Additional custom services
//
// # Service Management
//
// The operator manages a comprehensive set of BMC services:
//
//   - IPC: Inter-process communication service (NATS server)
//   - Console Server: Serial console access and redirection
//   - Inventory Manager: Hardware component discovery and tracking
//   - IPMI Server: Intelligent Platform Management Interface
//   - KVM Server: Keyboard, video, mouse redirection
//   - LED Manager: System status and identification LED control
//   - Power Manager: System power control and monitoring
//   - Security Manager: Authentication, authorization, and security policies
//   - Sensor Monitor: Hardware sensor monitoring and alerting
//   - State Manager: System state transitions and persistence
//   - Telemetry: Metrics collection and observability
//   - Thermal Manager: Cooling and thermal protection
//   - Update Manager: Firmware and software update management
//   - User Manager: User account and access control
//   - Web Server: REST API and web interface
//
// # Configuration
//
// The operator supports extensive configuration through the options pattern.
// Services can be selectively enabled, disabled, or customized:
//
//	op := operator.New(
//		operator.WithName("production-bmc"),
//		operator.WithTimeout(30*time.Second),
//		operator.WithIPC(
//			ipc.WithServerName("bmc-ipc"),
//			ipc.WithStoreDir("/var/lib/bmc/ipc"),
//		),
//		operator.WithTelemetry(
//			telemetry.WithMetricsEnabled(true),
//			telemetry.WithTracingEnabled(true),
//		),
//		operator.WithExtraServices(myCustomService),
//	)
//
// # Supervision and Fault Tolerance
//
// The operator implements a robust supervision strategy:
//
//   - Transient restart policy: Services are restarted on failure
//   - Configurable timeouts for service startup and shutdown
//   - Isolation: Service failures don't affect other services
//   - Graceful degradation: System continues with reduced functionality
//   - Logging and monitoring of all service state changes
//
// # Inter-Process Communication
//
// The operator coordinates IPC setup for all services:
//
//   - Starts the IPC service first to provide communication infrastructure
//   - Provides connection providers to all other services
//   - Handles IPC service failures and recovery
//   - Supports both embedded and external IPC configurations
//
// # System Initialization
//
// The operator handles various system initialization tasks:
//
//   - Mount point setup for pseudo-filesystems
//   - OpenTelemetry configuration and setup
//   - Persistent ID generation and management
//   - Logo display and branding
//   - Global logger configuration
//
// # Usage Patterns
//
// ## Basic Usage
//
// The simplest way to use the operator is with default configuration:
//
//	op := operator.New()
//	err := op.Run(ctx, nil)
//
// ## Custom Configuration
//
// For production deployments, services are typically customized:
//
//	op := operator.New(
//		operator.WithName("edge-bmc"),
//		operator.WithCustomLogo(myLogo),
//		operator.WithTimeout(15*time.Second),
//		operator.WithMountCheck(true),
//		operator.WithWebsrv(
//			websrv.WithPort(8443),
//			websrv.WithTLS(true),
//		),
//		operator.WithPowermgr(
//			powermgr.WithGPIOChip("/dev/gpiochip0"),
//		),
//	)
//
// ## External IPC Integration
//
// When integrating with external IPC infrastructure:
//
//	// Use external IPC connection
//	err := op.Run(ctx, externalIPCConn)
//
// ## Adding Custom Services
//
// Custom services can be added to the supervision tree:
//
//	myService := &MyCustomService{}
//	op := operator.New(
//		operator.WithExtraServices(myService),
//	)
//
// # Error Handling
//
// The operator provides comprehensive error handling:
//
//   - Configuration validation before startup
//   - Graceful handling of service startup failures
//   - Detailed error reporting with context
//   - Automatic recovery from transient failures
//   - Clean shutdown on fatal errors
//
// # Observability
//
// The operator integrates with OpenTelemetry for comprehensive observability:
//
//   - Distributed tracing across all services
//   - Structured logging with correlation IDs
//   - Metrics collection and reporting
//   - Health check endpoints
//   - Service dependency mapping
//
// # Best Practices
//
// When using the operator:
//
//   - Always provide a context with timeout for Run()
//   - Use structured logging for better observability
//   - Configure appropriate timeouts for your environment
//   - Test service restart scenarios in development
//   - Monitor service health and performance metrics
//   - Implement proper signal handling for graceful shutdown
//
// # Example Implementation
//
//	package main
//
//	import (
//		"context"
//		"os"
//		"os/signal"
//		"syscall"
//		"time"
//
//		"github.com/u-bmc/u-bmc/service/operator"
//		"github.com/u-bmc/u-bmc/service/ipc"
//		"github.com/u-bmc/u-bmc/service/websrv"
//	)
//
//	func main() {
//		// Create operator with custom configuration
//		op := operator.New(
//			operator.WithName("my-bmc"),
//			operator.WithTimeout(20*time.Second),
//			operator.WithIPC(
//				ipc.WithServerName("my-bmc-ipc"),
//				ipc.WithMaxMemory(128*1024*1024), // 128MB
//			),
//			operator.WithWebsrv(
//				websrv.WithPort(8443),
//				websrv.WithTLS(true),
//			),
//		)
//
//		// Set up graceful shutdown
//		ctx, cancel := context.WithCancel(context.Background())
//		defer cancel()
//
//		sigChan := make(chan os.Signal, 1)
//		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
//
//		go func() {
//			<-sigChan
//			cancel()
//		}()
//
//		// Run the operator
//		if err := op.Run(ctx, nil); err != nil {
//			if err != context.Canceled {
//				log.Fatal("Operator failed", "error", err)
//			}
//		}
//	}
//
// # Service Dependencies
//
// The operator manages service dependencies automatically:
//
//  1. IPC service starts first (communication infrastructure)
//  2. Core services start in parallel (sensors, power, thermal)
//  3. Management services start after core services
//  4. Web interface starts last (depends on all other services)
//
// Services can communicate with each other through the IPC infrastructure
// once all services are running and ready.
package operator
