// SPDX-License-Identifier: BSD-3-Clause

package operator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"sync"

	"cirello.io/oversight/v2"
	"github.com/arunsworld/nursery"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
	"github.com/u-bmc/u-bmc/service/ipmisrv"
	"github.com/u-bmc/u-bmc/service/kvmsrv"
	"github.com/u-bmc/u-bmc/service/ledmgr"
	"github.com/u-bmc/u-bmc/service/powermgr"
	"github.com/u-bmc/u-bmc/service/securitymgr"
	"github.com/u-bmc/u-bmc/service/sensormon"
	"github.com/u-bmc/u-bmc/service/statemgr"
	telemetrySrv "github.com/u-bmc/u-bmc/service/telemetry"
	"github.com/u-bmc/u-bmc/service/thermalmgr"
	"github.com/u-bmc/u-bmc/service/updatemgr"
	"github.com/u-bmc/u-bmc/service/usermgr"
	"github.com/u-bmc/u-bmc/service/websrv"
)

// Compile-time assertion that Operator implements service.Service.
var _ service.Service = (*Operator)(nil)

// Operator manages the lifecycle of BMC services in a supervised environment.
// It provides service orchestration, fault tolerance, and inter-process communication
// coordination for all BMC subsystems.
type Operator struct {
	config *config
	logger *slog.Logger
	tracer trace.Tracer
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
		name:         DefaultOperatorName,
		id:           "",
		disableLogo:  DefaultDisableLogo,
		mountCheck:   DefaultMountCheck,
		otelSetup:    telemetry.DefaultSetup,
		logger:       log.NewDefaultLogger(),
		timeout:      DefaultOperatorTimeout,
		ipc:          ipc.New(),
		Consolesrv:   consolesrv.New(),
		Inventorymgr: inventorymgr.New(),
		Ipmisrv:      ipmisrv.New(),
		Kvmsrv:       kvmsrv.New(),
		Ledmgr:       ledmgr.New(),
		Powermgr:     powermgr.New(),
		Securitymgr:  securitymgr.New(),
		Sensormon:    sensormon.New(),
		Statemgr:     statemgr.New(),
		Telemetry:    telemetrySrv.New(),
		Thermalmgr:   thermalmgr.New(),
		Updatemgr:    updatemgr.New(),
		Usermgr:      usermgr.New(),
		Websrv:       websrv.New(),
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &Operator{
		config: cfg,
	}
}

// Name returns the configured name of the operator service.
func (s *Operator) Name() string {
	return s.config.name
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
	s.tracer = otel.Tracer(s.Name())

	ctx, span := s.tracer.Start(ctx, "Run")
	defer span.End()

	if err := s.config.Validate(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrInvalidConfiguration, err)
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrPanicked, r)
			span.RecordError(err)
		}
	}()

	// Several services rely on the telemetry setup to be done because of our custom logger.
	// We do the setup here while any non-noop telemetry configuration handling is done
	// in the telemetry service that is optionally started.
	// TODO: Implement proper check if this needs to be run at all
	setup := sync.OnceFunc(s.config.otelSetup)
	setup()

	// This needs to be called after s.otelSetup to make sure any OTEL Log implementation is registered first
	if s.config.logger != nil {
		s.logger = s.config.logger.With("service", s.Name())
	} else {
		s.logger = log.GetGlobalLogger().With("service", s.Name())
	}
	l := s.logger

	if s.config.id == "" {
		idStr, err := id.GetOrCreatePersistentID(s.Name(), "/var/operator/id")
		if err != nil {
			l.ErrorContext(ctx, "Failed to get/create persistent ID, using ephemeral ID", "error", err)
			s.config.id = id.NewID()
		} else {
			s.config.id = idStr
		}
	}

	if !s.config.disableLogo {
		if s.config.customLogo != "" {
			l.Info(s.config.customLogo)
		} else {
			l.Info(DefaultLogo)
		}
	}

	// All mount points should have been set up by init
	// but we do not want to rely on it so we mount everything needed
	// that isn't there yet (mostly pseudofilesystems). Controlled by WithMountCheck().
	if s.config.mountCheck {
		l.InfoContext(ctx, "Checking filesystem mounts")
		if err := mount.SetupMounts(); err != nil {
			l.WarnContext(ctx, "Failed to setup mounts correctly, continuing anyway", "state", "degraded", "error", err)
			span.RecordError(fmt.Errorf("%w: %w", ErrSetupMounts, err))
		}
	}

	supervisionTree := oversight.New(
		oversight.NeverHalt(),
		oversight.DefaultRestartStrategy(),
		oversight.WithLogger(log.NewOversightLogger(l)),
	)

	// A user needs to either provide a valid ipcConn when starting the operator
	// or let us create an IPC service ourselves from the configuration.
	// If both are provided we will NOT start another IPC service but re-use the provided ipcConn!
	if s.config.ipc == nil && ipcConn == nil {
		err := ErrIPCNil
		span.RecordError(err)
		return err
	}

	if s.config.ipc != nil && ipcConn == nil {
		if err := supervisionTree.Add(
			process.New(s.config.ipc, nil),
			oversight.Transient(),
			oversight.Timeout(s.config.timeout),
			s.config.ipc.Name(),
		); err != nil {
			err = fmt.Errorf("%w %s to supervision tree: %w", ErrAddProcess, s.config.ipc.Name(), err)
			span.RecordError(err)
			return err
		}
	} else {
		if err := supervisionTree.Add(
			process.New(ipcPkg.NewStub(), nil),
			oversight.Transient(),
			oversight.Timeout(s.config.timeout),
			"ipc-stub",
		); err != nil {
			err = fmt.Errorf("%w %s to supervision tree: %w", ErrAddProcess, "ipc-stub", err)
			span.RecordError(err)
			return err
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
			conn = s.config.ipc.GetConnProvider()
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
						oversight.Timeout(s.config.timeout),
						svc.Name(),
					); err != nil {
						err = fmt.Errorf("%w %s to supervision tree: %w", ErrAddProcess, svc.Name(), err)
						span.RecordError(err)
						c <- err
						return
					}
				}
			}
		}

		for _, svc := range s.config.extraServices {
			if err := supervisionTree.Add(
				process.New(svc, conn),
				oversight.Transient(),
				oversight.Timeout(s.config.timeout),
				svc.Name(),
			); err != nil {
				err = fmt.Errorf("%w %s to supervision tree: %w", ErrAddExtraService, svc.Name(), err)
				span.RecordError(err)
				c <- err
				return
			}
		}
	}

	l.InfoContext(ctx, "Starting BMC services under supervision")

	span.SetAttributes(
		attribute.String("operator.name", s.Name()),
		attribute.String("operator.id", s.config.id),
		attribute.Bool("mount_check", s.config.mountCheck),
		attribute.Bool("disable_logo", s.config.disableLogo),
		attribute.String("timeout", s.config.timeout.String()),
	)

	err = nursery.RunConcurrentlyWithContext(ctx, supervise, spawnProcs)
	if err != nil && errors.Is(err, context.Canceled) {
		span.RecordError(err)
	}

	return err
}
