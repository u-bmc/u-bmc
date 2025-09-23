// SPDX-License-Identifier: BSD-3-Clause

// Package ipc provides utilities and interfaces for inter-process communication
// within the u-bmc system. This package serves as a bridge between BMC services
// and the underlying IPC infrastructure, offering abstractions and helpers that
// simplify service-to-service communication.
//
// The IPC package provides the fundamental building blocks for communication
// between BMC components, including connection providers, response helpers,
// and stub implementations for testing and development.
//
// # Core Components
//
// The package defines several key interfaces and utilities:
//
//   - ConnProvider: Interface for obtaining IPC connections
//   - Response helpers: Utilities for standardized error responses
//   - Stub implementations: No-op services for testing
//
// # Connection Management
//
// The ConnProvider interface abstracts the creation of network connections
// for inter-process communication. This allows services to obtain connections
// without needing to know the underlying transport details:
//
//	type ConnProvider interface {
//		InProcessConn() (net.Conn, error)
//	}
//
// Services can use this interface to establish connections with the IPC
// infrastructure, typically an embedded NATS server.
//
// # Response Utilities
//
// The package provides utilities for sending standardized error responses
// in NATS-based communication patterns. These helpers ensure consistent
// error handling across all BMC services:
//
//	// Send an error response with proper logging
//	ipc.RespondWithError(ctx, req, err, "operation failed")
//
// This function automatically logs the error and sends a properly formatted
// error response to the requesting client.
//
// # Integration with Services
//
// The IPC package is designed to work seamlessly with the u-bmc service
// framework. Services typically receive an InProcessConnProvider through
// their Run method and use it to establish communication channels:
//
//	func (s *MyService) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
//		// Connect to IPC infrastructure
//		conn, err := ipcConn.InProcessConn()
//		if err != nil {
//			return err
//		}
//		defer conn.Close()
//
//		// Use connection for service communication
//		// ...
//	}
//
// # Error Handling
//
// The package follows standard Go error handling practices and provides
// centralized error definitions for consistent error reporting across
// the BMC system. All errors are properly wrapped and include contextual
// information for debugging.
//
// # Thread Safety
//
// All public interfaces and functions in this package are designed to be
// thread-safe and can be used concurrently from multiple goroutines.
// Connection providers handle synchronization internally and provide
// safe access to underlying resources.
//
// # Performance Considerations
//
// The IPC utilities are optimized for the BMC use case, where low latency
// and high reliability are more important than maximum throughput. The
// package uses in-process connections when possible to minimize overhead
// and provide the best performance for local communication.
//
// # Best Practices
//
// When using this package:
//
//   - Always check for errors when obtaining connections
//   - Close connections when no longer needed
//   - Use the response helpers for consistent error handling
//   - Leverage stub implementations for testing
//   - Follow the established patterns for service integration
//
// Example usage in a BMC service:
//
//	package myservice
//
//	import (
//		"context"
//		"github.com/nats-io/nats.go"
//		"github.com/u-bmc/u-bmc/pkg/ipc"
//	)
//
//	type Service struct {
//		conn nats.InProcessConnProvider
//	}
//
//	func (s *Service) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
//		s.conn = ipcConn
//
//		// Establish connection
//		conn, err := ipcConn.InProcessConn()
//		if err != nil {
//			return err
//		}
//		defer conn.Close()
//
//		// Set up NATS client
//		nc, err := nats.Connect("", nats.InProcessServer(ipcConn))
//		if err != nil {
//			return err
//		}
//		defer nc.Close()
//
//		// Handle requests
//		_, err = nc.Subscribe("myservice.request", func(m *nats.Msg) {
//			// Process request and respond
//			if err := processRequest(m); err != nil {
//				// Use helper for error response
//				req := micro.NewRequest(m)
//				ipc.RespondWithError(ctx, req, err, "request processing failed")
//				return
//			}
//			// Send success response
//			m.Respond([]byte("success"))
//		})
//
//		// Wait for shutdown
//		<-ctx.Done()
//		return ctx.Err()
//	}
package ipc
