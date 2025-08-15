// SPDX-License-Identifier: BSD-3-Clause

package sensormon

import (
	"context"

	"github.com/nats-io/nats.go"
	"u-bmc.org/u-bmc/pkg/log"
)

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

	return nil
}
