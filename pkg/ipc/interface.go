// SPDX-License-Identifier: BSD-3-Clause

package ipc

import "net"

// ConnProvider defines an interface for providing inter-process communication connections.
// Implementations of this interface should provide a way to establish connections
// for communication between processes or components within the same process.
type ConnProvider interface {
	// InProcessConn establishes and returns a network connection for inter-process communication.
	// The returned connection can be used for bidirectional communication between processes
	// or components. Returns an error if the connection cannot be established.
	InProcessConn() (net.Conn, error)
}
