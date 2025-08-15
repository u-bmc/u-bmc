// SPDX-License-Identifier: BSD-3-Clause

package kvmsrv

import (
	"context"

	"github.com/nats-io/nats.go"
	"u-bmc.org/u-bmc/pkg/log"
)

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

	return nil
}
