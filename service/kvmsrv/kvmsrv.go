// SPDX-License-Identifier: BSD-3-Clause

package kvmsrv

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that KVMSrv implements service.Service.
var _ service.Service = (*KVMSrv)(nil)

type KVMSrv struct {
	config
}

func New(opts ...Option) *KVMSrv {
	cfg := &config{
		name: "kvmsrv",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &KVMSrv{
		config: *cfg,
	}
}

func (s *KVMSrv) Name() string {
	return s.name
}

func (s *KVMSrv) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting KVM server", "service", s.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping KVM server", "service", s.name, "reason", ctx.Err())

	return ctx.Err()
}
