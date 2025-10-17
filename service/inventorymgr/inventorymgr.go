// SPDX-License-Identifier: BSD-3-Clause

package inventorymgr

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	schemav1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/ipc"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Compile-time assertion that InventoryMgr implements service.Service.
var _ service.Service = (*InventoryMgr)(nil)

// InventoryMgr manages hardware inventory information for BMC components.
type InventoryMgr struct {
	config       config
	nc           *nats.Conn
	microService micro.Service
	logger       *slog.Logger
	tracer       trace.Tracer
}

// New creates a new InventoryMgr instance with the provided options.
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
	return s.config.name
}

func (s *InventoryMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	s.tracer = otel.Tracer(s.config.name)
	s.logger = log.GetGlobalLogger().With("service", s.config.name)

	s.logger.InfoContext(ctx, "Starting inventory manager", "service", s.config.name)

	var err error
	s.nc, err = nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer s.nc.Drain()

	s.microService, err = micro.AddService(s.nc, micro.Config{
		Name:        s.config.name,
		Version:     "1.0.0",
		Description: "Inventory management service",
	})
	if err != nil {
		return fmt.Errorf("failed to create micro service: %w", err)
	}

	if err := s.registerEndpoints(ctx); err != nil {
		return fmt.Errorf("failed to register endpoints: %w", err)
	}

	s.logger.InfoContext(ctx, "Inventory manager started successfully")

	<-ctx.Done()
	s.logger.InfoContext(ctx, "Stopping inventory manager", "service", s.config.name, "reason", ctx.Err())

	return ctx.Err()
}

func (s *InventoryMgr) registerEndpoints(ctx context.Context) error {
	groups := make(map[string]micro.Group)

	if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectAssetInfo,
		micro.HandlerFunc(s.createRequestHandler(ctx, s.handleAssetInfo)), groups); err != nil {
		return fmt.Errorf("failed to register asset info endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectAssetList,
		micro.HandlerFunc(s.createRequestHandler(ctx, s.handleAssetList)), groups); err != nil {
		return fmt.Errorf("failed to register asset list endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectAssetUpdate,
		micro.HandlerFunc(s.createRequestHandler(ctx, s.handleAssetUpdate)), groups); err != nil {
		return fmt.Errorf("failed to register asset update endpoint: %w", err)
	}

	return nil
}

func (s *InventoryMgr) createRequestHandler(parentCtx context.Context, handler func(context.Context, micro.Request)) micro.HandlerFunc {
	return func(req micro.Request) {
		ctx := telemetry.GetCtxFromReq(req)
		ctx = context.WithoutCancel(ctx)

		if s.tracer != nil {
			_, span := s.tracer.Start(ctx, "inventorymgr.handleRequest")
			span.SetAttributes(
				attribute.String("subject", req.Subject()),
				attribute.String("service", s.Name()),
			)
			defer span.End()
		}

		handler(ctx, req)
	}
}

func (s *InventoryMgr) handleAssetInfo(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "inventorymgr.handleAssetInfo")
		defer span.End()
	}

	var request schemav1alpha1.GetAssetInfoRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrUnmarshalingFailed, err.Error())
		return
	}

	// Mock response for now
	manufacturer := "Mock Manufacturer"
	serialNumber := "MOCK-PN-001"
	partNumber := "MOCK-PN-001"
	response := &schemav1alpha1.GetAssetInfoResponse{
		AssetInfo: []*schemav1alpha1.AssetInfo{
			{
				Manufacturer: &manufacturer,
				SerialNumber: &serialNumber,
				PartNumber:   &partNumber,
			},
		},
	}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func (s *InventoryMgr) handleAssetList(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "inventorymgr.handleAssetList")
		defer span.End()
	}

	// Mock response for now
	manufacturer := "Mock Manufacturer"
	serialNumber := "MOCK-PN-001"
	partNumber := "MOCK-PN-001"
	response := &schemav1alpha1.GetAssetInfoResponse{
		AssetInfo: []*schemav1alpha1.AssetInfo{
			{
				Manufacturer: &manufacturer,
				SerialNumber: &serialNumber,
				PartNumber:   &partNumber,
			},
		},
	}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func (s *InventoryMgr) handleAssetUpdate(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "inventorymgr.handleAssetUpdate")
		defer span.End()
	}

	var request schemav1alpha1.SetAssetInfoRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrUnmarshalingFailed, err.Error())
		return
	}

	// Mock response for now
	response := &schemav1alpha1.SetAssetInfoResponse{}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}
