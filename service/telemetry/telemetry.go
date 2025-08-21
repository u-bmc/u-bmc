// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that Telemetry implements service.Service.
var _ service.Service = (*Telemetry)(nil)

type Telemetry struct {
	config
}

func New(opts ...Option) *Telemetry {
	cfg := &config{
		name: "telemetry",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &Telemetry{
		config: *cfg,
	}
}

func (s *Telemetry) Name() string {
	return s.name
}

func (s *Telemetry) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting telemetry service", "service", s.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping telemetry service", "service", s.name, "reason", ctx.Err())

	return ctx.Err()
}
