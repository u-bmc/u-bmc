// SPDX-License-Identifier: BSD-3-Clause

package sensormon

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that SensorMon implements service.Service.
var _ service.Service = (*SensorMon)(nil)

type SensorMon struct {
	config
}

func New(opts ...Option) *SensorMon {
	cfg := &config{
		name: "sensormon",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &SensorMon{
		config: *cfg,
	}
}

func (s *SensorMon) Name() string {
	return s.name
}

func (s *SensorMon) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting sensor monitor", "service", s.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Stopping sensor monitor", "service", s.name, "reason", ctx.Err())

	return ctx.Err()
}
