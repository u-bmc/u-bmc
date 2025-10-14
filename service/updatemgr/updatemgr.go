// SPDX-License-Identifier: BSD-3-Clause

package updatemgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that UpdateMgr implements service.Service.
var _ service.Service = (*Updatemgr)(nil)

// Updatemgr provides firmware and software update management for BMC components.
type Updatemgr struct {
	config config
}

// New creates a new Updatemgr instance with the provided options.
func New(opts ...Option) *Updatemgr {
	cfg := &config{
		name: "updatemgr",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &Updatemgr{
		config: *cfg,
	}
}

func (s *Updatemgr) Name() string {
	return s.config.name
}

func (s *Updatemgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting update manager", "service", s.config.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping update manager", "service", s.config.name, "reason", ctx.Err())

	return ctx.Err()
}
