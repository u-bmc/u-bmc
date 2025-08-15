// SPDX-License-Identifier: BSD-3-Clause

package service

import (
	"context"

	"github.com/nats-io/nats.go"
)

// Service is an interface for long running processes or daemons.
// A service might be restarted if it returns an error.
// If a service returns nil, it is regarded to be done, also known as a oneshot service.
// The name should be unique per system. A nats NUID will be created on startup to identify it between connected systems.
type Service interface {
	// Name returns the unique name of the service.
	// This should be unique per system.
	Name() string

	// Run starts the service with the provided context.
	// It returns an error if the service needs to be restarted.
	Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error
}
