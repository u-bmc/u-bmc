// SPDX-License-Identifier: BSD-3-Clause

package consolesrv

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that ConsoleSrv implements service.Service.
var _ service.Service = (*ConsoleSrv)(nil)

type ConsoleSrv struct {
	config
}

func New(opts ...Option) *ConsoleSrv {
	cfg := &config{
		name: "consolesrv",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &ConsoleSrv{
		config: *cfg,
	}
}

func (s *ConsoleSrv) Name() string {
	return s.name
}

func (s *ConsoleSrv) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting console server", "service", s.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping console server", "service", s.name, "reason", ctx.Err())

	return ctx.Err()
}
