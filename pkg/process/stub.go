// SPDX-License-Identifier: BSD-3-Clause

package process

import (
	"context"

	"github.com/nats-io/nats.go"
)

// Stub is a no-op implementation of a service that does nothing.
// It can be used as a placeholder, for testing purposes, or to disable
// services by replacing them with a stub implementation.
type Stub struct {
	name string
}

// Name returns the identifier name for this stub service.
func (s *Stub) Name() string {
	return s.name
}

// Run executes the stub service, which immediately returns without error.
// This is a no-op implementation that accepts a context and connection provider
// but performs no actual operations.
func (s *Stub) Run(_ context.Context, _ nats.InProcessConnProvider) error {
	return nil
}

// NewStub creates and returns a new instance of the stub service with the given name.
func NewStub(name string) *Stub {
	return &Stub{
		name: name,
	}
}
