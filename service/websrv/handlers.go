// SPDX-License-Identifier: BSD-3-Clause

package websrv

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	schemav1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1/schemav1alpha1connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ProtoServer implements all the Connect RPC service handlers for the BMC API.
type ProtoServer struct {
	schemav1alpha1connect.UnimplementedBMCServiceHandler
	nc     *nats.Conn
	logger *slog.Logger
	tracer trace.Tracer
}

// NewProtoServer creates a new ProtoServer instance.
func NewProtoServer(nc *nats.Conn, logger *slog.Logger) *ProtoServer {
	return &ProtoServer{
		nc:     nc,
		logger: logger,
		tracer: otel.Tracer("protoserver"),
	}
}

// GetSystemInfo handles the GetSystemInfo RPC call.
func (s *ProtoServer) GetSystemInfo(ctx context.Context, req *schemav1alpha1.GetSystemInfoRequest) (*schemav1alpha1.GetSystemInfoResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetSystemInfo")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetSystemInfo"),
	)

	s.logger.DebugContext(ctx, "Processing GetSystemInfo request")

	var sysResp schemav1alpha1.GetSystemInfoResponse
	if err := s.requestNATS(ctx, "system.info", req, &sysResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetSystemInfo request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetSystemInfo request")
	return &sysResp, nil
}

// GetHealth handles the GetHealth RPC call.
func (s *ProtoServer) GetHealth(ctx context.Context, req *schemav1alpha1.GetHealthRequest) (*schemav1alpha1.GetHealthResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetHealth")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetHealth"),
	)

	s.logger.DebugContext(ctx, "Processing GetHealth request")

	var healthResp schemav1alpha1.GetHealthResponse
	if err := s.requestNATS(ctx, "system.health", req, &healthResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetHealth request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetHealth request")
	return &healthResp, nil
}

// GetHost handles the GetHost RPC call.
func (s *ProtoServer) GetHost(ctx context.Context, req *schemav1alpha1.GetHostRequest) (*schemav1alpha1.GetHostResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetHost")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetHost"),
		attribute.String("host.name", req.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing GetHost request",
		slog.String("host_name", req.GetName()))

	hostName, err := sanitizeSubjectToken(req.GetName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid host name in GetHost request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var hostResp schemav1alpha1.GetHostResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("statemgr.host.%s.state", hostName), req, &hostResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetHost request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetHost request",
		slog.String("host_name", req.GetName()))
	return &hostResp, nil
}

// ChangeHostState handles the ChangeHostState RPC call.
func (s *ProtoServer) ChangeHostState(ctx context.Context, req *schemav1alpha1.ChangeHostStateRequest) (*schemav1alpha1.ChangeHostStateResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ChangeHostState")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ChangeHostState"),
		attribute.String("host.name", req.GetHostName()),
		attribute.String("host.action", req.GetAction().String()),
	)

	s.logger.DebugContext(ctx, "Processing ChangeHostState request",
		slog.String("host_name", req.GetHostName()),
		slog.String("action", req.GetAction().String()))

	hostName, err := sanitizeSubjectToken(req.GetHostName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid host name in ChangeHostState request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var stateResp schemav1alpha1.ChangeHostStateResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("statemgr.host.%s.control", hostName), req, &stateResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ChangeHostState request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ChangeHostState request",
		slog.String("host_name", req.GetHostName()),
		slog.String("action", req.GetAction().String()),
		slog.String("new_status", stateResp.GetCurrentStatus().String()))
	return &stateResp, nil
}

// GetChassis handles the GetChassis RPC call.
func (s *ProtoServer) GetChassis(ctx context.Context, req *schemav1alpha1.GetChassisRequest) (*schemav1alpha1.GetChassisResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetChassis")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetChassis"),
		attribute.String("chassis.name", req.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing GetChassis request",
		slog.String("chassis_name", req.GetName()))

	chassisName, err := sanitizeSubjectToken(req.GetName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid chassis name in GetChassis request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var chassisResp schemav1alpha1.GetChassisResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("statemgr.chassis.%s.state", chassisName), req, &chassisResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetChassis request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetChassis request",
		slog.String("chassis_name", req.GetName()))
	return &chassisResp, nil
}

// ChangeChassisState handles the ChangeChassisState RPC call.
func (s *ProtoServer) ChangeChassisState(ctx context.Context, req *schemav1alpha1.ChangeChassisStateRequest) (*schemav1alpha1.ChangeChassisStateResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ChangeChassisState")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ChangeChassisState"),
		attribute.String("chassis.name", req.GetChassisName()),
		attribute.String("chassis.action", req.GetAction().String()),
	)

	s.logger.DebugContext(ctx, "Processing ChangeChassisState request",
		slog.String("chassis_name", req.GetChassisName()),
		slog.String("action", req.GetAction().String()))

	chassisName, err := sanitizeSubjectToken(req.GetChassisName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid chassis name in ChangeChassisState request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var stateResp schemav1alpha1.ChangeChassisStateResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("statemgr.chassis.%s.control", chassisName), req, &stateResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ChangeChassisState request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ChangeChassisState request",
		slog.String("chassis_name", req.GetChassisName()),
		slog.String("action", req.GetAction().String()),
		slog.String("new_status", stateResp.GetCurrentStatus().String()))
	return &stateResp, nil
}

// GetManagementController handles the GetManagementController RPC call.
func (s *ProtoServer) GetManagementController(ctx context.Context, req *schemav1alpha1.GetManagementControllerRequest) (*schemav1alpha1.GetManagementControllerResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetManagementController")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetManagementController"),
		attribute.String("controller.name", req.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing GetManagementController request",
		slog.String("controller_name", req.GetName()))

	controllerName, err := sanitizeSubjectToken(req.GetName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid controller name in GetManagementController request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var controllerResp schemav1alpha1.GetManagementControllerResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("statemgr.bmc.%s.state", controllerName), req, &controllerResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetManagementController request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetManagementController request",
		slog.String("controller_name", req.GetName()))
	return &controllerResp, nil
}

// ChangeManagementControllerState handles the ChangeManagementControllerState RPC call.
func (s *ProtoServer) ChangeManagementControllerState(ctx context.Context, req *schemav1alpha1.ChangeManagementControllerStateRequest) (*schemav1alpha1.ChangeManagementControllerStateResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ChangeManagementControllerState")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ChangeManagementControllerState"),
		attribute.String("controller.name", req.GetControllerName()),
		attribute.String("controller.action", req.GetAction().String()),
	)

	s.logger.DebugContext(ctx, "Processing ChangeManagementControllerState request",
		slog.String("controller_name", req.GetControllerName()),
		slog.String("action", req.GetAction().String()))

	controllerName, err := sanitizeSubjectToken(req.GetControllerName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid controller name in ChangeManagementControllerState request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var stateResp schemav1alpha1.ChangeManagementControllerStateResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("statemgr.bmc.%s.control", controllerName), req, &stateResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ChangeManagementControllerState request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ChangeManagementControllerState request",
		slog.String("controller_name", req.GetControllerName()),
		slog.String("action", req.GetAction().String()),
		slog.String("new_status", stateResp.GetCurrentStatus().String()))
	return &stateResp, nil
}

// GetAssetInfo handles the GetAssetInfo RPC call.
func (s *ProtoServer) GetAssetInfo(ctx context.Context, req *schemav1alpha1.GetAssetInfoRequest) (*schemav1alpha1.GetAssetInfoResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetAssetInfo")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetAssetInfo"),
	)

	s.logger.DebugContext(ctx, "Processing GetAssetInfo request")

	var assetResp schemav1alpha1.GetAssetInfoResponse
	if err := s.requestNATS(ctx, "inventorymgr.asset.info", req, &assetResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetAssetInfo request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetAssetInfo request")
	return &assetResp, nil
}

// SetAssetInfo handles the SetAssetInfo RPC call.
func (s *ProtoServer) SetAssetInfo(ctx context.Context, req *schemav1alpha1.SetAssetInfoRequest) (*schemav1alpha1.SetAssetInfoResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.SetAssetInfo")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "SetAssetInfo"),
	)

	s.logger.DebugContext(ctx, "Processing SetAssetInfo request")

	var assetResp schemav1alpha1.SetAssetInfoResponse
	if err := s.requestNATS(ctx, "inventorymgr.asset.update", req, &assetResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process SetAssetInfo request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed SetAssetInfo request")
	return &assetResp, nil
}

// ListChassis handles the ListChassis RPC call.
func (s *ProtoServer) ListChassis(ctx context.Context, req *schemav1alpha1.ListChassisRequest) (*schemav1alpha1.ListChassisResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListChassis")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListChassis"),
	)

	s.logger.DebugContext(ctx, "Processing ListChassis request")

	var chassisResp schemav1alpha1.ListChassisResponse
	if err := s.requestNATS(ctx, "statemgr.chassis.list", req, &chassisResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListChassis request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListChassis request")
	return &chassisResp, nil
}

// UpdateChassis handles the UpdateChassis RPC call.
func (s *ProtoServer) UpdateChassis(ctx context.Context, req *schemav1alpha1.UpdateChassisRequest) (*schemav1alpha1.UpdateChassisResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.UpdateChassis")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "UpdateChassis"),
		attribute.String("chassis.name", req.GetChassis().GetName()),
	)

	s.logger.DebugContext(ctx, "Processing UpdateChassis request",
		slog.String("chassis_name", req.GetChassis().GetName()))

	chassisName, err := sanitizeSubjectToken(req.GetChassis().GetName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid chassis name in UpdateChassis request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var chassisResp schemav1alpha1.UpdateChassisResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("statemgr.chassis.%s.update", chassisName), req, &chassisResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process UpdateChassis request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed UpdateChassis request",
		slog.String("chassis_name", req.GetChassis().GetName()))
	return &chassisResp, nil
}

// ListHosts handles the ListHosts RPC call.
func (s *ProtoServer) ListHosts(ctx context.Context, req *schemav1alpha1.ListHostsRequest) (*schemav1alpha1.ListHostsResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListHosts")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListHosts"),
	)

	s.logger.DebugContext(ctx, "Processing ListHosts request")

	var hostResp schemav1alpha1.ListHostsResponse
	if err := s.requestNATS(ctx, "statemgr.host.list", req, &hostResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListHosts request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListHosts request")
	return &hostResp, nil
}

// UpdateHost handles the UpdateHost RPC call.
func (s *ProtoServer) UpdateHost(ctx context.Context, req *schemav1alpha1.UpdateHostRequest) (*schemav1alpha1.UpdateHostResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.UpdateHost")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "UpdateHost"),
		attribute.String("host.name", req.GetHost().GetName()),
	)

	s.logger.DebugContext(ctx, "Processing UpdateHost request",
		slog.String("host_name", req.GetHost().GetName()))

	hostName, err := sanitizeSubjectToken(req.GetHost().GetName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid host name in UpdateHost request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var hostResp schemav1alpha1.UpdateHostResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("statemgr.host.%s.update", hostName), req, &hostResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process UpdateHost request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed UpdateHost request",
		slog.String("host_name", req.GetHost().GetName()))
	return &hostResp, nil
}

// ListManagementControllers handles the ListManagementControllers RPC call.
func (s *ProtoServer) ListManagementControllers(ctx context.Context, req *schemav1alpha1.ListManagementControllersRequest) (*schemav1alpha1.ListManagementControllersResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListManagementControllers")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListManagementControllers"),
	)

	s.logger.DebugContext(ctx, "Processing ListManagementControllers request")

	var controllerResp schemav1alpha1.ListManagementControllersResponse
	if err := s.requestNATS(ctx, "statemgr.bmc.list", req, &controllerResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListManagementControllers request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListManagementControllers request")
	return &controllerResp, nil
}

// UpdateManagementController handles the UpdateManagementController RPC call.
func (s *ProtoServer) UpdateManagementController(ctx context.Context, req *schemav1alpha1.UpdateManagementControllerRequest) (*schemav1alpha1.UpdateManagementControllerResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.UpdateManagementController")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "UpdateManagementController"),
		attribute.String("controller.name", req.GetController().GetName()),
	)

	s.logger.DebugContext(ctx, "Processing UpdateManagementController request",
		slog.String("controller_name", req.GetController().GetName()))

	controllerName, err := sanitizeSubjectToken(req.GetController().GetName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid controller name in UpdateManagementController request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var controllerResp schemav1alpha1.UpdateManagementControllerResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("statemgr.bmc.%s.update", controllerName), req, &controllerResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process UpdateManagementController request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed UpdateManagementController request",
		slog.String("controller_name", req.GetController().GetName()))
	return &controllerResp, nil
}

// ListSensors handles the ListSensors RPC call.
func (s *ProtoServer) ListSensors(ctx context.Context, req *schemav1alpha1.ListSensorsRequest) (*schemav1alpha1.ListSensorsResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListSensors")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListSensors"),
	)

	s.logger.DebugContext(ctx, "Processing ListSensors request")

	var sensorResp schemav1alpha1.ListSensorsResponse
	if err := s.requestNATS(ctx, "sensormon.sensors.list", req, &sensorResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListSensors request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListSensors request")
	return &sensorResp, nil
}

// GetSensor handles the GetSensor RPC call.
func (s *ProtoServer) GetSensor(ctx context.Context, req *schemav1alpha1.GetSensorRequest) (*schemav1alpha1.GetSensorResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetSensor")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetSensor"),
		attribute.String("sensor.name", req.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing GetSensor request",
		slog.String("sensor_name", req.GetName()))

	sensorName, err := sanitizeSubjectToken(req.GetName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid sensor name in GetSensor request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var sensorResp schemav1alpha1.GetSensorResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("sensormon.sensor.%s.state", sensorName), req, &sensorResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetSensor request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetSensor request",
		slog.String("sensor_name", req.GetName()))
	return &sensorResp, nil
}

// GetThermalZone handles the GetThermalZone RPC call.
func (s *ProtoServer) GetThermalZone(ctx context.Context, req *schemav1alpha1.GetThermalZoneRequest) (*schemav1alpha1.GetThermalZoneResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetThermalZone")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetThermalZone"),
		attribute.String("zone.name", req.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing GetThermalZone request",
		slog.String("zone_name", req.GetName()))

	zoneName, err := sanitizeSubjectToken(req.GetName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid zone name in GetThermalZone request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var thermalResp schemav1alpha1.GetThermalZoneResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("thermalmgr.zone.%s.state", zoneName), req, &thermalResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetThermalZone request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetThermalZone request",
		slog.String("zone_name", req.GetName()))
	return &thermalResp, nil
}

// SetThermalZone handles the SetThermalZone RPC call.
func (s *ProtoServer) SetThermalZone(ctx context.Context, req *schemav1alpha1.SetThermalZoneRequest) (*schemav1alpha1.SetThermalZoneResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.SetThermalZone")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "SetThermalZone"),
		attribute.String("zone.name", req.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing SetThermalZone request",
		slog.String("zone_name", req.GetName()))

	zoneName, err := sanitizeSubjectToken(req.GetName())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid zone name in SetThermalZone request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var thermalResp schemav1alpha1.SetThermalZoneResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("thermalmgr.zone.%s.update", zoneName), req, &thermalResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process SetThermalZone request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed SetThermalZone request",
		slog.String("zone_name", req.GetName()))
	return &thermalResp, nil
}

// ListThermalZones handles the ListThermalZones RPC call.
func (s *ProtoServer) ListThermalZones(ctx context.Context, req *schemav1alpha1.ListThermalZonesRequest) (*schemav1alpha1.ListThermalZonesResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListThermalZones")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListThermalZones"),
	)

	s.logger.DebugContext(ctx, "Processing ListThermalZones request")

	var thermalResp schemav1alpha1.ListThermalZonesResponse
	if err := s.requestNATS(ctx, "thermalmgr.zones.list", req, &thermalResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListThermalZones request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListThermalZones request")
	return &thermalResp, nil
}

// CreateUser handles the CreateUser RPC call.
func (s *ProtoServer) CreateUser(ctx context.Context, req *schemav1alpha1.CreateUserRequest) (*schemav1alpha1.CreateUserResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "CreateUser"),
		attribute.String("user.name", req.GetUser().GetUsername()),
	)

	s.logger.DebugContext(ctx, "Processing CreateUser request",
		slog.String("user_name", req.GetUser().GetUsername()))

	var userResp schemav1alpha1.CreateUserResponse
	if err := s.requestNATS(ctx, "usermgr.user.create", req, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process CreateUser request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed CreateUser request",
		slog.String("user_name", req.GetUser().GetUsername()))
	return &userResp, nil
}

// GetUser handles the GetUser RPC call.
func (s *ProtoServer) GetUser(ctx context.Context, req *schemav1alpha1.GetUserRequest) (*schemav1alpha1.GetUserResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetUser"),
		attribute.String("user.name", req.GetUsername()),
	)

	s.logger.DebugContext(ctx, "Processing GetUser request",
		slog.String("user_name", req.GetUsername()))

	username, err := sanitizeSubjectToken(req.GetUsername())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid username in GetUser request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var userResp schemav1alpha1.GetUserResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("usermgr.user.%s.info", username), req, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetUser request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetUser request",
		slog.String("user_name", req.GetUsername()))
	return &userResp, nil
}

// UpdateUser handles the UpdateUser RPC call.
func (s *ProtoServer) UpdateUser(ctx context.Context, req *schemav1alpha1.UpdateUserRequest) (*schemav1alpha1.UpdateUserResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.UpdateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "UpdateUser"),
		attribute.String("user.name", req.GetUser().GetUsername()),
	)

	s.logger.DebugContext(ctx, "Processing UpdateUser request",
		slog.String("user_name", req.GetUser().GetUsername()))

	username, err := sanitizeSubjectToken(req.GetUser().GetUsername())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid username in UpdateUser request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var userResp schemav1alpha1.UpdateUserResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("usermgr.user.%s.update", username), req, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process UpdateUser request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed UpdateUser request",
		slog.String("user_name", req.GetUser().GetUsername()))
	return &userResp, nil
}

// DeleteUser handles the DeleteUser RPC call.
func (s *ProtoServer) DeleteUser(ctx context.Context, req *schemav1alpha1.DeleteUserRequest) (*schemav1alpha1.DeleteUserResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.DeleteUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "DeleteUser"),
		attribute.String("user.name", req.GetId()),
	)

	s.logger.DebugContext(ctx, "Processing DeleteUser request",
		slog.String("user_id", req.GetId()))

	userID, err := sanitizeSubjectToken(req.GetId())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid user ID in DeleteUser request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var userResp schemav1alpha1.DeleteUserResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("usermgr.user.%s.delete", userID), req, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process DeleteUser request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed DeleteUser request",
		slog.String("user_id", req.GetId()))
	return &userResp, nil
}

// ListUsers handles the ListUsers RPC call.
func (s *ProtoServer) ListUsers(ctx context.Context, req *schemav1alpha1.ListUsersRequest) (*schemav1alpha1.ListUsersResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListUsers")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListUsers"),
	)

	s.logger.DebugContext(ctx, "Processing ListUsers request")

	var userResp schemav1alpha1.ListUsersResponse
	if err := s.requestNATS(ctx, "usermgr.users.list", req, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListUsers request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListUsers request")
	return &userResp, nil
}

// ChangePassword handles the ChangePassword RPC call.
func (s *ProtoServer) ChangePassword(ctx context.Context, req *schemav1alpha1.ChangePasswordRequest) (*schemav1alpha1.ChangePasswordResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ChangePassword")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ChangePassword"),
		attribute.String("user.id", req.GetId()),
	)

	s.logger.DebugContext(ctx, "Processing ChangePassword request",
		slog.String("user_id", req.GetId()))

	userID, err := sanitizeSubjectToken(req.GetId())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid user ID in ChangePassword request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var passwordResp schemav1alpha1.ChangePasswordResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("usermgr.user.%s.password.change", userID), req, &passwordResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ChangePassword request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ChangePassword request",
		slog.String("user_id", req.GetId()))
	return &passwordResp, nil
}

// ResetPassword handles the ResetPassword RPC call.
func (s *ProtoServer) ResetPassword(ctx context.Context, req *schemav1alpha1.ResetPasswordRequest) (*schemav1alpha1.ResetPasswordResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ResetPassword")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ResetPassword"),
		attribute.String("user.id", req.GetId()),
	)

	s.logger.DebugContext(ctx, "Processing ResetPassword request",
		slog.String("user_id", req.GetId()))

	userID, err := sanitizeSubjectToken(req.GetId())
	if err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid user ID in ResetPassword request", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var passwordResp schemav1alpha1.ResetPasswordResponse
	if err := s.requestNATS(ctx, fmt.Sprintf("usermgr.user.%s.password.reset", userID), req, &passwordResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ResetPassword request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ResetPassword request",
		slog.String("user_id", req.GetId()))
	return &passwordResp, nil
}

// AuthenticateUser handles the AuthenticateUser RPC call.
func (s *ProtoServer) AuthenticateUser(ctx context.Context, req *schemav1alpha1.AuthenticateUserRequest) (*schemav1alpha1.AuthenticateUserResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.AuthenticateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "AuthenticateUser"),
		attribute.String("user.name", req.GetUsername()),
	)

	s.logger.DebugContext(ctx, "Processing AuthenticateUser request",
		slog.String("user_name", req.GetUsername()))

	var authResp schemav1alpha1.AuthenticateUserResponse
	if err := s.requestNATS(ctx, "securitymgr.user.authenticate", req, &authResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process AuthenticateUser request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed AuthenticateUser request",
		slog.String("user_name", req.GetUsername()),
		slog.Bool("success", authResp.GetSuccess()))
	return &authResp, nil
}

type vtMessage interface {
	MarshalVT() ([]byte, error)
}

type vtUnmarshaler interface {
	UnmarshalVT([]byte) error
}

// sanitizeSubjectToken ensures user-provided identifiers cannot inject extra subject tokens or wildcards.
// Only allow [A-Za-z0-9_-].
func sanitizeSubjectToken(tok string) (string, error) {
	if tok == "" {
		return "", fmt.Errorf("empty subject token")
	}
	for _, r := range tok {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return "", fmt.Errorf("invalid subject token %q: only [A-Za-z0-9_-] allowed", tok)
	}
	return tok, nil
}

// requestNATS forwards an RPC over NATS with context propagation and robust error mapping.
func (s *ProtoServer) requestNATS(ctx context.Context, subject string, req vtMessage, resp vtUnmarshaler) error {
	if s.nc == nil || s.nc.Status() != nats.CONNECTED {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("nats not connected"))
	}

	reqBytes, err := req.MarshalVT()
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to marshal request: %w", err))
	}

	// Ensure we have a deadline on the context to avoid hanging requests.
	if _, ok := ctx.Deadline(); !ok {
		nctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		ctx = nctx
	}

	msg, err := s.nc.RequestWithContext(ctx, subject, reqBytes)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, nats.ErrTimeout):
			return connect.NewError(connect.CodeDeadlineExceeded, fmt.Errorf("request timed out: %w", err))
		case errors.Is(err, nats.ErrNoResponders):
			return connect.NewError(connect.CodeUnavailable, fmt.Errorf("no responders for %s", subject))
		case errors.Is(err, nats.ErrConnectionClosed), errors.Is(err, nats.ErrDisconnected), errors.Is(err, nats.ErrConnectionDraining):
			return connect.NewError(connect.CodeUnavailable, fmt.Errorf("nats connection not available: %w", err))
		default:
			return connect.NewError(connect.CodeUnavailable, fmt.Errorf("nats request failed: %w", err))
		}
	}

	if err := resp.UnmarshalVT(msg.Data); err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to unmarshal response: %w", err))
	}

	return nil
}
