// SPDX-License-Identifier: BSD-3-Clause

package thermalmgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that ThermalMgr implements service.Service.
var _ service.Service = (*ThermalMgr)(nil)

type ThermalMgr struct {
	config
}

func New(opts ...Option) *ThermalMgr {
	cfg := &config{
		name: "thermalmgr",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &ThermalMgr{
		config: *cfg,
	}
}

func (s *ThermalMgr) Name() string {
	return s.name
}

func (s *ThermalMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting thermal manager", "service", s.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping thermal manager", "service", s.name, "reason", ctx.Err())

	return ctx.Err()
}
