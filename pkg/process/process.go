// SPDX-License-Identifier: BSD-3-Clause

package process

import (
	"context"
	"fmt"

	"cirello.io/oversight/v2"
	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/service"
)

// New creates a new oversight.ChildProcess that wraps a service.Service.
// It takes a service and an in-process NATS connection provider, and returns
// a function that can be used as a child process in an oversight tree.
// The returned function will run the service with the provided context and
// connection, and will recover from any panics, converting them to errors
// that include the service name for better debugging.
func New(s service.Service, ipcConn nats.InProcessConnProvider) oversight.ChildProcess {
	return func(ctx context.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%s panicked: %v", s.Name(), r)
			}
		}()

		return s.Run(ctx, ipcConn)
	}
}
