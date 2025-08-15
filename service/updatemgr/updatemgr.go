// SPDX-License-Identifier: BSD-3-Clause

package updatemgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"u-bmc.org/u-bmc/pkg/log"
)

type Updatemgr struct {
	config
}

func New(opts ...Option) *Updatemgr {
	cfg := &config{
		name: "updatemgr",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &Updatemgr{
		config: *cfg,
	}
}

func (s *Updatemgr) Name() string {
	return s.name
}

func (s *Updatemgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting update manager", "service", s.name)

	return nil
}
