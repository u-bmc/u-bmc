// SPDX-License-Identifier: BSD-3-Clause

package securitymgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that SecurityMgr implements service.Service.
var _ service.Service = (*SecurityMgr)(nil)

type SecurityMgr struct {
	config
}

func New(opts ...Option) *SecurityMgr {
	cfg := &config{
		name: "securitymgr",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &SecurityMgr{
		config: *cfg,
	}
}

func (s *SecurityMgr) Name() string {
	return s.name
}

func (s *SecurityMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting security manager", "service", s.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping security manager", "service", s.name, "reason", ctx.Err())

	return ctx.Err()
}
