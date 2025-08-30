// SPDX-License-Identifier: BSD-3-Clause

// Package operator provides a service orchestrator that manages and supervises
// multiple BMC services in a fault-tolerant manner. It handles service lifecycle,
// inter-process communication setup, and provides a supervision tree for automatic
// service recovery.
package operator

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"cirello.io/oversight/v2"
	"github.com/arunsworld/nursery"
	"github.com/nats-io/nats.go"

	"github.com/u-bmc/u-bmc/pkg/id"
	ipcPkg "github.com/u-bmc/u-bmc/pkg/ipc"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/mount"
	"github.com/u-bmc/u-bmc/pkg/process"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service"
	"github.com/u-bmc/u-bmc/service/consolesrv"
	"github.com/u-bmc/u-bmc/service/inventorymgr"
	"github.com/u-bmc/u-bmc/service/ipc"
	"github.com/u-bmc/u-bmc/service/securitymgr"
	"github.com/u-bmc/u-bmc/service/statemgr"
	"github.com/u-bmc/u-bmc/service/updatemgr"
	"github.com/u-bmc/u-bmc/service/usermgr"
	"github.com/u-bmc/u-bmc/service/websrv"
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

// Compile-time assertion that Operator implements service.Service.
var _ service.Service = (*Operator)(nil)

// Operator manages the lifecycle of BMC services in a supervised environment.
// It provides service orchestration, fault tolerance, and inter-process communication
// coordination for all BMC subsystems.
type Operator struct {
	config
}

// New creates a new Operator instance with the provided configuration options.
// The operator will be initialized with default services including console server,
// inventory manager, security manager, state manager, update manager, user manager,
// and web server. Additional services can be configured using the provided options.
//
// Example usage:
//
//	op := operator.New(
//		operator.WithName("my-bmc"),
//		operator.WithTimeout(15*time.Second),
//		operator.DisableLogo(),
//	)
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

// Name returns the configured name of the operator service.
func (s *Operator) Name() string {
	return s.name
}

// Run starts the operator and all configured services under supervision.
// It sets up the supervision tree, configures inter-process communication,
// and manages the lifecycle of all BMC services. The operator will run until
// the provided context is canceled or a fatal error occurs.
//
// The ipcConn parameter can be nil if an IPC service is configured via options.
// If both ipcConn and IPC service are provided, the external ipcConn takes precedence.
//
// Returns an error if initialization fails or if any critical service cannot be started.
func (s *Operator) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) (err error) {
	if s.name == "" {
		return ErrNameEmpty
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s %w: %v", s.Name(), ErrPanicked, r)
		}
	}()

	// Several services rely on the telemetry setup to be done because of our custom logger.
	// We do the setup here while any non-noop telemetry configuration handling is done
	// in the telemetry service that is optionally started.
	s.otelSetup()

	// This needs to be called after s.otelSetup to make sure any OTEL Log implementation is registered first
	l := log.GetGlobalLogger()

	if s.id == "" {
		idStr, err := id.GetOrCreatePersistentID(s.Name(), "/var/operator/id")
		if err != nil {
			l.ErrorContext(ctx, "Failed to get/create persistent ID, using ephemeral ID", "error", err)
			s.id = id.NewID()
		} else {
			s.id = idStr
		}
	}

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
	l.InfoContext(ctx, "Checking filesystem mounts", "service", s.name)
	if err := mount.SetupMounts(); err != nil {
		l.WarnContext(ctx, "Failed to setup mounts correctly, continuing anyways", "service", s.name, "error", err)
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
		return ErrIPCNil
	}

	if s.ipc != nil && ipcConn == nil {
		if err := supervisionTree.Add(
			process.New(s.ipc, nil),
			oversight.Transient(),
			oversight.Timeout(s.timeout),
			s.ipc.Name(),
		); err != nil {
			return fmt.Errorf("%w %s to tree: %w", ErrAddProcess, s.ipc.Name(), err)
		}
	} else {
		if err := supervisionTree.Add(
			process.New(ipcPkg.NewStub(), nil),
			oversight.Transient(),
			oversight.Timeout(s.timeout),
			"ipc-stub",
		); err != nil {
			return fmt.Errorf("%w %s to tree: %w", ErrAddProcess, "ipc-stub", err)
		}
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
						c <- fmt.Errorf("%w %s to tree: %w", ErrAddProcess, svc.Name(), err)
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
				c <- fmt.Errorf("%w %s to tree: %w", ErrAddExtraService, svc.Name(), err)
				return
			}
		}
	}

	l.InfoContext(ctx, "Starting child routines", "service", s.name)
	return nursery.RunConcurrentlyWithContext(ctx, supervise, spawnProcs)
}
