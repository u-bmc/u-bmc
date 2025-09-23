// SPDX-License-Identifier: BSD-3-Clause

// Package ipc provides an in-process NATS server for inter-process communication
// within the u-bmc system. This service acts as the central message bus for all
// other services in the BMC.
//
// The IPC service creates and manages a NATS server instance that runs embedded
// within the u-bmc process, eliminating the need for external NATS server
// dependencies. It provides JetStream capabilities for persistent messaging
// and state management across BMC components.
//
// # Core Features
//
//   - Embedded NATS server with JetStream support
//   - In-process connection provider for other services
//   - Configurable server options and storage directories
//   - Graceful startup and shutdown handling
//   - Integration with u-bmc service framework
//
// # Usage
//
// The IPC service is typically started as one of the first services in the
// BMC system, as other services depend on it for communication:
//
//	ipcService := ipc.New(
//		ipc.WithServiceName("ipc"),
//		ipc.WithServerOpts(&server.Options{
//			ServerName: "bmc-ipc",
//			JetStream:  true,
//			StoreDir:   "/var/lib/u-bmc/ipc",
//		}),
//	)
//
//	// Start the service
//	err := ipcService.Run(ctx, nil)
//
// Other services can obtain connection providers to communicate through the IPC:
//
//	connProvider := ipcService.GetConnProvider()
//	conn, err := connProvider.InProcessConn()
//	if err != nil {
//		// Handle connection error
//	}
//
// # Configuration
//
// The IPC service can be configured with various options:
//
//   - WithServiceName: Set the service name
//   - WithServerOpts: Configure NATS server options
//   - WithStoreDir: Set JetStream storage directory
//   - WithJetStream: Enable/disable JetStream
//
// # Architecture
//
// The IPC service follows the standard u-bmc service pattern:
//
//   - Implements the service.Service interface
//   - Provides a Run method for lifecycle management
//   - Supports graceful shutdown via context cancellation
//   - Integrates with the global logging system
//
// The service creates an embedded NATS server that other services connect to
// using in-process connections, providing high-performance message passing
// without network overhead.
package ipc
