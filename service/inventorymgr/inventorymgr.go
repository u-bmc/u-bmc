// SPDX-License-Identifier: BSD-3-Clause

package inventorymgr

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that InventoryMgr implements service.Service.
var _ service.Service = (*InventoryMgr)(nil)

type InventoryMgr struct {
	config
}

func New(opts ...Option) *InventoryMgr {
	cfg := &config{
		name: "inventorymgr",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &InventoryMgr{
		config: *cfg,
	}
}

func (s *InventoryMgr) Name() string {
	return s.name
}

func (s *InventoryMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting inventory manager", "service", s.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping inventory manager", "service", s.name, "reason", ctx.Err())

	return ctx.Err()
}
