package ipc

import "net"

// IPCConnProvider defines an interface for providing inter-process communication connections.
// Implementations of this interface should provide a way to establish connections
// for communication between processes or components within the same process.
type IPCConnProvider interface {
	// InProcessConn establishes and returns a network connection for inter-process communication.
	// The returned connection can be used for bidirectional communication between processes
	// or components. Returns an error if the connection cannot be established.
	InProcessConn() (net.Conn, error)
}
