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

// SecurityMgr provides security management functionality for BMC access control.
type SecurityMgr struct {
	config config
}

// New creates a new SecurityMgr instance with the provided options.
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
	return s.config.name
}

func (s *SecurityMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting security manager", "service", s.config.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping security manager", "service", s.config.name, "reason", ctx.Err())

	return ctx.Err()
}
