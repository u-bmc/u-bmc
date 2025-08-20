// SPDX-License-Identifier: BSD-3-Clause

package ipc

import (
	"context"

	"github.com/nats-io/nats.go"
)

// Stub is a no-op implementation of an IPC service that does nothing.
// It can be used as a placeholder or for testing purposes.
type Stub struct{}

// Name returns the identifier name for this stub IPC service.
func (s *Stub) Name() string {
	return "ipc-stub"
}

// Run executes the stub IPC service, which immediately returns without error.
// This is a no-op implementation that accepts a context and connection provider
// but performs no actual IPC operations.
func (s *Stub) Run(_ context.Context, _ nats.InProcessConnProvider) error {
	return nil
}

// NewStub creates and returns a new instance of the stub IPC service.
func NewStub() *Stub {
	return &Stub{}
}
