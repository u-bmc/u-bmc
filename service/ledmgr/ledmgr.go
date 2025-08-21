// SPDX-License-Identifier: BSD-3-Clause

package ledmgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that LEDMgr implements service.Service.
var _ service.Service = (*LEDMgr)(nil)

type LEDMgr struct {
	config
}

func New(opts ...Option) *LEDMgr {
	cfg := &config{
		name: "ledmgr",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &LEDMgr{
		config: *cfg,
	}
}

func (s *LEDMgr) Name() string {
	return s.name
}

func (s *LEDMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting LED manager", "service", s.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping LED manager", "service", s.name, "reason", ctx.Err())

	return ctx.Err()
}
