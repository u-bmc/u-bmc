// SPDX-License-Identifier: BSD-3-Clause

package ipmisrv

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
)

type IPMISrv struct {
	config
}

func New(opts ...Option) *IPMISrv {
	cfg := &config{
		name: "ipmisrv",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &IPMISrv{
		config: *cfg,
	}
}

func (s *IPMISrv) Name() string {
	return s.name
}

func (s *IPMISrv) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting inventory manager", "service", s.name)

	return nil
}
