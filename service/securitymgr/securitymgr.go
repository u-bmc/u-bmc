// SPDX-License-Identifier: BSD-3-Clause

package securitymgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"u-bmc.org/u-bmc/pkg/log"
)

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

	return nil
}
