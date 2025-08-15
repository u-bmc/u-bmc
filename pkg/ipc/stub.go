// SPDX-License-Identifier: BSD-3-Clause

package ipc

import (
	"context"

	"github.com/nats-io/nats.go"
)

type Stub struct{}

func (s *Stub) Name() string {
	return "ipc-stub"
}

func (s *Stub) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	return nil
}

func NewStub() *Stub {
	return &Stub{}
}
