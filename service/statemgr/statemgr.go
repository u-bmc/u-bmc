// SPDX-License-Identifier: BSD-3-Clause

package statemgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that StateMgr implements service.Service.
var _ service.Service = (*StateMgr)(nil)

type StateMgr struct {
	config
}

func New(opts ...Option) *StateMgr {
	cfg := &config{
		name: "statemgr",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &StateMgr{
		config: *cfg,
	}
}

func (s *StateMgr) Name() string {
	return s.name
}

func (s *StateMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting state manager", "service", s.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping state manager", "service", s.name, "reason", ctx.Err())

	return ctx.Err()
}
