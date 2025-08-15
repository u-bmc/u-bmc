// SPDX-License-Identifier: BSD-3-Clause

package statemgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"u-bmc.org/u-bmc/pkg/log"
)

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

	return nil
}
