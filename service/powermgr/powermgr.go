// SPDX-License-Identifier: BSD-3-Clause

package powermgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"u-bmc.org/u-bmc/pkg/log"
)

type PowerMgr struct {
	config
}

func New(opts ...Option) *PowerMgr {
	cfg := &config{
		name: "powermgr",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &PowerMgr{
		config: *cfg,
	}
}

func (s *PowerMgr) Name() string {
	return s.name
}

func (s *PowerMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting power manager", "service", s.name)

	return nil
}
