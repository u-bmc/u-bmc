// SPDX-License-Identifier: BSD-3-Clause

package usermgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that UserMgr implements service.Service.
var _ service.Service = (*UserMgr)(nil)

type UserMgr struct {
	config
}

func New(opts ...Option) *UserMgr {
	cfg := &config{
		name: "usermgr",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &UserMgr{
		config: *cfg,
	}
}

func (s *UserMgr) Name() string {
	return s.name
}

func (s *UserMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting user manager", "service", s.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping user manager", "service", s.name, "reason", ctx.Err())

	return ctx.Err()
}
