// SPDX-License-Identifier: BSD-3-Clause

package ledmgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"u-bmc.org/u-bmc/pkg/log"
)

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

	return nil
}
