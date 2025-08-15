// SPDX-License-Identifier: BSD-3-Clause

package operator

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"cirello.io/oversight/v2"
	"github.com/arunsworld/nursery"
	"github.com/nats-io/nats.go"

	"u-bmc.org/u-bmc/pkg/id"
	ipcPkg "u-bmc.org/u-bmc/pkg/ipc"
	"u-bmc.org/u-bmc/pkg/log"
	"u-bmc.org/u-bmc/pkg/mount"
	"u-bmc.org/u-bmc/pkg/process"
	"u-bmc.org/u-bmc/pkg/telemetry"
	"u-bmc.org/u-bmc/service"
	"u-bmc.org/u-bmc/service/consolesrv"
	"u-bmc.org/u-bmc/service/inventorymgr"
	"u-bmc.org/u-bmc/service/ipc"
	"u-bmc.org/u-bmc/service/securitymgr"
	"u-bmc.org/u-bmc/service/statemgr"
	"u-bmc.org/u-bmc/service/updatemgr"
	"u-bmc.org/u-bmc/service/usermgr"
	"u-bmc.org/u-bmc/service/websrv"
)

const defaultLogo = `
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣤⣤⣤⣤⣤⣀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⣿⠟⠉⠁⠀⠈⠉⠛⢿⣦⡀
⠀⠀⠀⠀⠀⠀⠀⣠⣿⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⣿⣆
⠀⠀⠀⠀⠀⠀⢠⣿⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⡆
⠀⠀⠀⠀⠀⠀⣿⡇⠀⠀⣶⣶⣦⠀⠀⠀⣠⣾⣶⡀⠀⠸⣿
⠀⠀⠀⠀⠀⠀⣿⠀⠀⠘⠛⠀⠟⠀⠀⠀⠻⠁⠙⠃⠀⠀⣿
⠀⠀⠀⠀⠀⠀⣿⡇⠀⠀⠀⠀⣴⡆⠀⢠⣦⠀⠀⠀⠀⢠⣿
⠀⠀⠀⠀⠀⠀⠘⣿⠀⠀⠀⠀⠘⣿⣶⣿⠃⠀⠀⠀⠀⣿⠇
⠀⠀⠀⠀⠀⠀⠀⠙⣿⣾⠿⠀⠀⠀⠀⠀⠀⠀⠿⣿⣾⠟
⠀⠀⠀⠀⠀⠀⢀⣿⠛⠀⠀⠀⢀⣤⣤⣤⡀⠀⠀⠀⠙⣿⡄
⠀⠀⠀⠀⠀⢀⣿⠃⠀⠀⠀⣴⡿⠋⠉⠉⢿⣷⠀⠀⠀⠈⣿⡄
⠀⠀⠀⠀⠀⣼⡏⠀⠀⠀⢠⣿⠀⠀⠀⠀⠀⣿⡇⠀⠀⠀⢸⣿
⠀⠀⠀⠀⠀⣿⡇⠀⠀⠀⢸⣿⠀⠀⠀⠀⠀⣿⡇⠀⠀⠀⢀⣿
⠀⠀⠀⢰⣿⠛⠻⣿⠀⠀⢸⣿⠀⠀⠀⠀⠀⣿⡇⠀⠀⣾⠟⠛⢿⣦
⠀⠀⠀⢿⣇⠀⠀⣿⠇⠀⢸⣿⠀⠀⠀⠀⠀⣿⡇⠀⠀⣿⠀⠀⢠⣿
⠀⠀⠀⠈⠻⣿⡿⠋⠀⠀⢸⣿⠀⠀⠀⠀⠀⣿⡇⠀⠀⠙⠿⣿⠿⠁
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣼⣿⣤⠀⠀⠀⣠⣿⣧⣄
⠀⠀⠀⠀⠀⠀⠀⠀⢠⣿⠁⠀⢻⣷⠀⣸⡟⠀⠈⣿⡆
⠀⠀⠀⠀⠀⠀⠀⠀⠈⣿⣄⣀⣾⠏⠀⠸⣷⣄⣠⣿⠃
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠛⠁⠀⠀⠀⠈⠛⠛

⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣶⡄
⠀⣠⣄⠀⢠⣦⠀⠀⠀⠀⣿⣁⣶⣤⡀⠀⢀⣤⣦⣀⣴⣦⡀⠀⠀⣤⣶⣤⡀
⠀⣿⡷⠀⢸⣿⢀⣤⣤⠀⣿⡟⠉⢻⣿⠀⣿⠏⠉⣿⠋⠙⣿⠀⣿⡟⠉⠛⠃
⠀⢻⣷⣀⣾⡟⠀⠉⠉⠀⣿⣧⣀⣼⣿⠀⣿⠀⠀⣿⠀⠀⣿⠀⣿⣧⣀⣴⡆
⠀⠀⠉⠛⠋⠀⠀⠀⠀⠀⠛⠉⠛⠋⠀⠀⠛⠀⠀⠛⠀⠀⠛⠀⠀⠙⠛⠋
`

type Operator struct {
	config
}

func New(opts ...Option) *Operator {
	cfg := &config{
		name:         "operator",
		id:           "",
		disableLogo:  false,
		otelSetup:    telemetry.DefaultSetup,
		logger:       log.NewDefaultLogger(),
		timeout:      10 * time.Second,
		ipc:          ipc.New(),
		Consolesrv:   consolesrv.New(),
		Inventorymgr: inventorymgr.New(),
		Securitymgr:  securitymgr.New(),
		Statemgr:     statemgr.New(),
		Updatemgr:    updatemgr.New(),
		Usermgr:      usermgr.New(),
		Websrv:       websrv.New(),
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &Operator{
		config: *cfg,
	}
}

func (s *Operator) Name() string {
	return s.name
}

func (s *Operator) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) (err error) {
	if s.name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s panicked: %v", s.Name(), r)
		}
	}()

	if s.id == "" {
		idStr, err := id.GetOrCreatePersistentID(s.Name(), "/var/operator/id")
		if err != nil {
			s.id = id.NewID()
		} else {
			s.id = idStr
		}
	}

	// Several services rely on the telemetry setup to be done because of our custom logger.
	// We do the setup here while any non-noop telemetry configuration handling is done
	// in the telemetry service that is optionally started.
	s.otelSetup()

	// This needs to be called after s.otelSetup to make sure any OTEL Log implementation is registered first
	l := log.GetGlobalLogger()

	if !s.disableLogo {
		if s.customLogo != "" {
			l.Info(s.customLogo)
		} else {
			l.Info(defaultLogo)
		}
	}

	// All mount points should have been set up by init
	// but we do not want to rely on it so we mount everything needed
	// that isn't there yet (mostly pseudofilesystems)
	// l.InfoContext(ctx, "Checking filesystem mounts", "service", s.name)
	if err := mount.SetupMounts(); err != nil {
		return fmt.Errorf("failed to setup mounts: %w", err)
	}

	supervisionTree := oversight.New(
		oversight.NeverHalt(),
		oversight.DefaultRestartStrategy(),
		oversight.WithLogger(log.NewOversightLogger(l)),
	)

	// A user needs to either provide a valid ipcConn when starting the operator
	// or let us create an IPC service ourselves from the configuration.
	// If both are provided we will NOT start another IPC service but re-use the provided ipcConn!
	if s.ipc == nil && ipcConn == nil {
		return fmt.Errorf("IPC cannot be nil, provide either ipcConn or WithIPC")
	}

	if s.ipc != nil && ipcConn == nil {
		supervisionTree.Add(
			process.New(s.ipc, nil),
			oversight.Transient(),
			oversight.Timeout(s.timeout),
			s.ipc.Name(),
		)
	} else {
		supervisionTree.Add(
			process.New(ipcPkg.NewStub(), nil),
			oversight.Transient(),
			oversight.Timeout(s.timeout),
			"ipc-stub",
		)
	}

	supervise := func(ctx context.Context, c chan error) {
		c <- supervisionTree.Start(ctx)
	}

	spawnProcs := func(ctx context.Context, c chan error) {
		var conn nats.InProcessConnProvider
		if ipcConn != nil {
			conn = ipcConn
		} else {
			conn = s.ipc.GetConnProvider()
		}

		// Dynamically add all service.Service fields to supervision tree
		configValue := reflect.ValueOf(s.config)
		for i := range configValue.NumField() {
			field := configValue.Field(i)

			// Check if field implements service.Service interface
			if field.IsValid() && field.CanInterface() {
				v := field.Interface()
				if v == nil {
					continue
				}
				if svc, ok := v.(service.Service); ok {
					if err := supervisionTree.Add(
						process.New(svc, conn),
						oversight.Transient(),
						oversight.Timeout(s.timeout),
						svc.Name(),
					); err != nil {
						c <- fmt.Errorf("failed to add process %s to tree: %w", svc.Name(), err)
						return
					}
				}
			}
		}

		for _, svc := range s.extraServices {
			if err := supervisionTree.Add(
				process.New(svc, ipcConn),
				oversight.Transient(),
				oversight.Timeout(s.timeout),
				svc.Name(),
			); err != nil {
				c <- fmt.Errorf("failed to add extra service %s to tree: %w", svc.Name(), err)
				return
			}
		}
	}

	l.InfoContext(ctx, "Starting child routines", "service", s.name)
	return nursery.RunConcurrentlyWithContext(ctx, supervise, spawnProcs)
}
