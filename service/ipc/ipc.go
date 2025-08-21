// SPDX-License-Identifier: BSD-3-Clause

package ipc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
)

// Compile-time assertion that IPC implements service.Service.
var _ service.Service = (*IPC)(nil)

type IPC struct {
	config

	server *server.Server
}

type ConnProvider struct {
	server *server.Server
}

func (p *ConnProvider) InProcessConn() (net.Conn, error) {
	// Non-blocking check
	if p.server == nil {
		return nil, server.ErrServerNotRunning
	}

	// Blocking check
	if !p.server.ReadyForConnections(time.Minute) {
		return nil, server.ErrServerNotRunning
	}

	return p.server.InProcessConn()
}

func New(opts ...Option) *IPC {
	cfg := &config{
		name: "ipc",
		serverOpts: &server.Options{
			ServerName:             "ipc",
			Debug:                  true,
			DontListen:             true,
			JetStream:              true,
			DisableJetStreamBanner: true,
			StoreDir:               "/var/ipc.data",
		},
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &IPC{
		config: *cfg,
	}
}

func (s *IPC) Name() string {
	return s.name
}

func (s *IPC) GetConnProvider() *ConnProvider {
	return &ConnProvider{
		server: s.server,
	}
}

func (s *IPC) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	// We might be able to handle this in the future, for now bail out
	if ipcConn != nil {
		return fmt.Errorf("existing IPC found, bailing out")
	}

	l := log.GetGlobalLogger()

	ns, err := server.NewServer(s.serverOpts)
	if err != nil {
		return fmt.Errorf("could not create server: %w", err)
	}
	s.server = ns
	s.server.SetLoggerV2(log.NewNATSLogger(l), false, false, false)

	l.InfoContext(ctx, "Starting IPC server", "service", s.name)
	s.server.Start()

	if !s.server.ReadyForConnections(time.Minute) {
		s.server.Shutdown()
		return fmt.Errorf("ipc server timed out: %w", server.ErrServerNotRunning)
	}
	l.InfoContext(ctx, "IPC server started successfully", "service", s.name)

	<-ctx.Done()
	l.InfoContext(ctx, "Received shutdown signal, calling lame duck shutdown", "service", s.name)
	s.server.LameDuckShutdown()

	return ctx.Err()
}
