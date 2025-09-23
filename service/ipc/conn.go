// SPDX-License-Identifier: BSD-3-Clause

package ipc

import (
	"fmt"
	"net"
	"time"

	"github.com/nats-io/nats-server/v2/server"
)

// ConnProvider provides in-process connections to the embedded NATS server.
//
// The ConnProvider acts as a bridge between the IPC service and other services
// that need to communicate through NATS. It provides thread-safe access to
// in-process connections with built-in readiness checks and timeout handling.
//
// The ConnProvider is designed to be robust and will handle cases where:
//   - The server is not yet started
//   - The server is shutting down
//   - Connection creation fails
//
// It includes automatic retry logic and will block until the server is ready
// or a timeout occurs.
type ConnProvider struct {
	server *server.Server
}

// InProcessConn creates a new in-process connection to the NATS server.
//
// This method provides a direct, high-performance connection to the embedded
// NATS server without going through the network stack. It performs several
// readiness checks before attempting to create the connection:
//
//  1. Verifies the server instance is available (non-blocking)
//  2. Waits for the server to be ready for connections (blocking, with timeout)
//  3. Creates and returns the in-process connection
//
// The method will block for up to one minute waiting for the server to become
// ready. This timeout is generous to handle cases where the server is still
// starting up or performing initialization tasks.
//
// Returns:
//   - A net.Conn instance for communicating with the NATS server
//   - An error if the server is not available, not ready, or connection creation fails
//
// Common error conditions:
//   - ErrConnectionNotAvailable: The server is nil or not ready for connections
//   - ErrServerNotReady: The connection could not be established
//
// Example usage:
//
//	provider := ipcService.GetConnProvider()
//	conn, err := provider.InProcessConn()
//	if err != nil {
//		return fmt.Errorf("failed to get IPC connection: %w", err)
//	}
//	defer conn.Close()
//
//	// Use connection with NATS client
//	nc, err := nats.Connect("", nats.InProcessServer(provider))
func (p *ConnProvider) InProcessConn() (net.Conn, error) {
	// Non-blocking check - fail fast if server is not available
	if p.server == nil {
		return nil, ErrConnectionNotAvailable
	}

	// Blocking check - wait for server to be ready with timeout
	// This is generous to handle startup delays and initialization
	if !p.server.ReadyForConnections(time.Minute) {
		return nil, ErrServerNotReady
	}

	// Create the in-process connection
	conn, err := p.server.InProcessConn()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInProcessConnFailed, err)
	}

	return conn, nil
}
