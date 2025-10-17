// SPDX-License-Identifier: BSD-3-Clause

package websrv

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	schemav1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1/schemav1alpha1connect"
	"github.com/u-bmc/u-bmc/pkg/ipc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
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
		tracer: otel.Tracer("websrv"),
	}
}

// GetSystemInfo handles the GetSystemInfo RPC call.
func (s *ProtoServer) GetSystemInfo(ctx context.Context, req *connect.Request[schemav1alpha1.GetSystemInfoRequest]) (*connect.Response[schemav1alpha1.GetSystemInfoResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetSystemInfo")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetSystemInfo"),
	)

	s.logger.DebugContext(ctx, "Processing GetSystemInfo request")

	var systemResp schemav1alpha1.GetSystemInfoResponse
	if err := s.requestNATS(ctx, ipc.SubjectSystemInfo, req.Msg, &systemResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetSystemInfo request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetSystemInfo request")
	return connect.NewResponse(&systemResp), nil
}

// GetHealth handles the GetHealth RPC call.
func (s *ProtoServer) GetHealth(ctx context.Context, req *connect.Request[schemav1alpha1.GetHealthRequest]) (*connect.Response[schemav1alpha1.GetHealthResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetHealth")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetHealth"),
	)

	s.logger.DebugContext(ctx, "Processing GetHealth request")

	var healthResp schemav1alpha1.GetHealthResponse
	if err := s.requestNATS(ctx, ipc.SubjectSystemHealth, req.Msg, &healthResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetHealth request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetHealth request")
	return connect.NewResponse(&healthResp), nil
}

// GetHost handles the GetHost RPC call.
func (s *ProtoServer) GetHost(ctx context.Context, req *connect.Request[schemav1alpha1.GetHostRequest]) (*connect.Response[schemav1alpha1.GetHostResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetHost")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetHost"),
		attribute.String("host.name", req.Msg.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing GetHost request",
		slog.String("host_name", req.Msg.GetName()))

	var hostResp schemav1alpha1.GetHostResponse
	if err := s.requestNATS(ctx, ipc.SubjectHostState, req.Msg, &hostResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetHost request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetHost request",
		slog.String("host_name", req.Msg.GetName()))
	return connect.NewResponse(&hostResp), nil
}

// ChangeHostState handles the ChangeHostState RPC call.
func (s *ProtoServer) ChangeHostState(ctx context.Context, req *connect.Request[schemav1alpha1.ChangeHostStateRequest]) (*connect.Response[schemav1alpha1.ChangeHostStateResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ChangeHostState")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ChangeHostState"),
		attribute.String("host.name", req.Msg.GetHostName()),
		attribute.String("host.action", req.Msg.GetAction().String()),
	)

	s.logger.DebugContext(ctx, "Processing ChangeHostState request",
		slog.String("host_name", req.Msg.GetHostName()),
		slog.String("action", req.Msg.GetAction().String()))

	var hostResp schemav1alpha1.ChangeHostStateResponse
	if err := s.requestNATS(ctx, ipc.SubjectHostControl, req.Msg, &hostResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ChangeHostState request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ChangeHostState request",
		slog.String("host_name", req.Msg.GetHostName()))
	return connect.NewResponse(&hostResp), nil
}

// GetChassis handles the GetChassis RPC call.
func (s *ProtoServer) GetChassis(ctx context.Context, req *connect.Request[schemav1alpha1.GetChassisRequest]) (*connect.Response[schemav1alpha1.GetChassisResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetChassis")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetChassis"),
		attribute.String("chassis.name", req.Msg.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing GetChassis request",
		slog.String("chassis_name", req.Msg.GetName()))

	var chassisResp schemav1alpha1.GetChassisResponse
	if err := s.requestNATS(ctx, ipc.SubjectChassisState, req.Msg, &chassisResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetChassis request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetChassis request",
		slog.String("chassis_name", req.Msg.GetName()))
	return connect.NewResponse(&chassisResp), nil
}

// ChangeChassisState handles the ChangeChassisState RPC call.
func (s *ProtoServer) ChangeChassisState(ctx context.Context, req *connect.Request[schemav1alpha1.ChangeChassisStateRequest]) (*connect.Response[schemav1alpha1.ChangeChassisStateResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ChangeChassisState")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ChangeChassisState"),
		attribute.String("chassis.name", req.Msg.GetChassisName()),
		attribute.String("chassis.action", req.Msg.GetAction().String()),
	)

	s.logger.DebugContext(ctx, "Processing ChangeChassisState request",
		slog.String("chassis_name", req.Msg.GetChassisName()),
		slog.String("action", req.Msg.GetAction().String()))

	var chassisResp schemav1alpha1.ChangeChassisStateResponse
	if err := s.requestNATS(ctx, ipc.SubjectChassisControl, req.Msg, &chassisResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ChangeChassisState request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ChangeChassisState request",
		slog.String("chassis_name", req.Msg.GetChassisName()))
	return connect.NewResponse(&chassisResp), nil
}

// GetManagementController handles the GetManagementController RPC call.
func (s *ProtoServer) GetManagementController(ctx context.Context, req *connect.Request[schemav1alpha1.GetManagementControllerRequest]) (*connect.Response[schemav1alpha1.GetManagementControllerResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetManagementController")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetManagementController"),
		attribute.String("controller.name", req.Msg.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing GetManagementController request",
		slog.String("controller_name", req.Msg.GetName()))

	var bmcResp schemav1alpha1.GetManagementControllerResponse
	if err := s.requestNATS(ctx, ipc.SubjectBMCState, req.Msg, &bmcResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetManagementController request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetManagementController request",
		slog.String("controller_name", req.Msg.GetName()))
	return connect.NewResponse(&bmcResp), nil
}

// ChangeManagementControllerState handles the ChangeManagementControllerState RPC call.
func (s *ProtoServer) ChangeManagementControllerState(ctx context.Context, req *connect.Request[schemav1alpha1.ChangeManagementControllerStateRequest]) (*connect.Response[schemav1alpha1.ChangeManagementControllerStateResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ChangeManagementControllerState")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ChangeManagementControllerState"),
		attribute.String("controller.name", req.Msg.GetControllerName()),
		attribute.String("controller.action", req.Msg.GetAction().String()),
	)

	s.logger.DebugContext(ctx, "Processing ChangeManagementControllerState request",
		slog.String("controller_name", req.Msg.GetControllerName()),
		slog.String("action", req.Msg.GetAction().String()))

	var bmcResp schemav1alpha1.ChangeManagementControllerStateResponse
	if err := s.requestNATS(ctx, ipc.SubjectBMCControl, req.Msg, &bmcResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ChangeManagementControllerState request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ChangeManagementControllerState request",
		slog.String("controller_name", req.Msg.GetControllerName()))
	return connect.NewResponse(&bmcResp), nil
}

// GetAssetInfo handles the GetAssetInfo RPC call.
func (s *ProtoServer) GetAssetInfo(ctx context.Context, req *connect.Request[schemav1alpha1.GetAssetInfoRequest]) (*connect.Response[schemav1alpha1.GetAssetInfoResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetAssetInfo")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetAssetInfo"),
	)

	s.logger.DebugContext(ctx, "Processing GetAssetInfo request")

	var assetResp schemav1alpha1.GetAssetInfoResponse
	if err := s.requestNATS(ctx, ipc.SubjectAssetInfo, req.Msg, &assetResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetAssetInfo request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetAssetInfo request")
	return connect.NewResponse(&assetResp), nil
}

// SetAssetInfo handles the SetAssetInfo RPC call.
func (s *ProtoServer) SetAssetInfo(ctx context.Context, req *connect.Request[schemav1alpha1.SetAssetInfoRequest]) (*connect.Response[schemav1alpha1.SetAssetInfoResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.SetAssetInfo")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "SetAssetInfo"),
	)

	s.logger.DebugContext(ctx, "Processing SetAssetInfo request")

	var assetResp schemav1alpha1.SetAssetInfoResponse
	if err := s.requestNATS(ctx, ipc.SubjectAssetUpdate, req.Msg, &assetResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process SetAssetInfo request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed SetAssetInfo request")
	return connect.NewResponse(&assetResp), nil
}

// ListChassis handles the ListChassis RPC call.
func (s *ProtoServer) ListChassis(ctx context.Context, req *connect.Request[schemav1alpha1.ListChassisRequest]) (*connect.Response[schemav1alpha1.ListChassisResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListChassis")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListChassis"),
	)

	s.logger.DebugContext(ctx, "Processing ListChassis request")

	var chassisResp schemav1alpha1.ListChassisResponse
	if err := s.requestNATS(ctx, ipc.SubjectChassisList, req.Msg, &chassisResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListChassis request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListChassis request")
	return connect.NewResponse(&chassisResp), nil
}

// UpdateChassis handles the UpdateChassis RPC call.
func (s *ProtoServer) UpdateChassis(ctx context.Context, req *connect.Request[schemav1alpha1.UpdateChassisRequest]) (*connect.Response[schemav1alpha1.UpdateChassisResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.UpdateChassis")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "UpdateChassis"),
		attribute.String("chassis.name", req.Msg.GetChassis().GetName()),
	)

	s.logger.DebugContext(ctx, "Processing UpdateChassis request",
		slog.String("chassis_name", req.Msg.GetChassis().GetName()))

	var chassisResp schemav1alpha1.UpdateChassisResponse
	if err := s.requestNATS(ctx, ipc.SubjectChassisInfo, req.Msg, &chassisResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process UpdateChassis request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed UpdateChassis request",
		slog.String("chassis_name", req.Msg.GetChassis().GetName()))
	return connect.NewResponse(&chassisResp), nil
}

// ListHosts handles the ListHosts RPC call.
func (s *ProtoServer) ListHosts(ctx context.Context, req *connect.Request[schemav1alpha1.ListHostsRequest]) (*connect.Response[schemav1alpha1.ListHostsResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListHosts")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListHosts"),
	)

	s.logger.DebugContext(ctx, "Processing ListHosts request")

	var hostResp schemav1alpha1.ListHostsResponse
	if err := s.requestNATS(ctx, ipc.SubjectHostList, req.Msg, &hostResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListHosts request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListHosts request")
	return connect.NewResponse(&hostResp), nil
}

// UpdateHost handles the UpdateHost RPC call.
func (s *ProtoServer) UpdateHost(ctx context.Context, req *connect.Request[schemav1alpha1.UpdateHostRequest]) (*connect.Response[schemav1alpha1.UpdateHostResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.UpdateHost")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "UpdateHost"),
		attribute.String("host.name", req.Msg.GetHost().GetName()),
	)

	s.logger.DebugContext(ctx, "Processing UpdateHost request",
		slog.String("host_name", req.Msg.GetHost().GetName()))

	var hostResp schemav1alpha1.UpdateHostResponse
	if err := s.requestNATS(ctx, ipc.SubjectHostInfo, req.Msg, &hostResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process UpdateHost request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed UpdateHost request",
		slog.String("host_name", req.Msg.GetHost().GetName()))
	return connect.NewResponse(&hostResp), nil
}

// ListManagementControllers handles the ListManagementControllers RPC call.
func (s *ProtoServer) ListManagementControllers(ctx context.Context, req *connect.Request[schemav1alpha1.ListManagementControllersRequest]) (*connect.Response[schemav1alpha1.ListManagementControllersResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListManagementControllers")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListManagementControllers"),
	)

	s.logger.DebugContext(ctx, "Processing ListManagementControllers request")

	var bmcResp schemav1alpha1.ListManagementControllersResponse
	if err := s.requestNATS(ctx, ipc.SubjectBMCList, req.Msg, &bmcResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListManagementControllers request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListManagementControllers request")
	return connect.NewResponse(&bmcResp), nil
}

// UpdateManagementController handles the UpdateManagementController RPC call.
func (s *ProtoServer) UpdateManagementController(ctx context.Context, req *connect.Request[schemav1alpha1.UpdateManagementControllerRequest]) (*connect.Response[schemav1alpha1.UpdateManagementControllerResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.UpdateManagementController")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "UpdateManagementController"),
		attribute.String("controller.name", req.Msg.GetController().GetName()),
	)

	s.logger.DebugContext(ctx, "Processing UpdateManagementController request",
		slog.String("controller_name", req.Msg.GetController().GetName()))

	var bmcResp schemav1alpha1.UpdateManagementControllerResponse
	if err := s.requestNATS(ctx, ipc.SubjectBMCInfo, req.Msg, &bmcResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process UpdateManagementController request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed UpdateManagementController request",
		slog.String("controller_name", req.Msg.GetController().GetName()))
	return connect.NewResponse(&bmcResp), nil
}

// ListSensors handles the ListSensors RPC call.
func (s *ProtoServer) ListSensors(ctx context.Context, req *connect.Request[schemav1alpha1.ListSensorsRequest]) (*connect.Response[schemav1alpha1.ListSensorsResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListSensors")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListSensors"),
	)

	s.logger.DebugContext(ctx, "Processing ListSensors request")

	var sensorResp schemav1alpha1.ListSensorsResponse
	if err := s.requestNATS(ctx, ipc.SubjectSensorList, req.Msg, &sensorResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListSensors request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListSensors request")
	return connect.NewResponse(&sensorResp), nil
}

// GetSensor handles the GetSensor RPC call.
func (s *ProtoServer) GetSensor(ctx context.Context, req *connect.Request[schemav1alpha1.GetSensorRequest]) (*connect.Response[schemav1alpha1.GetSensorResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetSensor")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetSensor"),
		attribute.String("sensor.name", req.Msg.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing GetSensor request",
		slog.String("sensor_name", req.Msg.GetName()))

	var sensorResp schemav1alpha1.GetSensorResponse
	if err := s.requestNATS(ctx, ipc.SubjectSensorInfo, req.Msg, &sensorResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetSensor request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetSensor request",
		slog.String("sensor_name", req.Msg.GetName()))
	return connect.NewResponse(&sensorResp), nil
}

// GetThermalZone handles the GetThermalZone RPC call.
func (s *ProtoServer) GetThermalZone(ctx context.Context, req *connect.Request[schemav1alpha1.GetThermalZoneRequest]) (*connect.Response[schemav1alpha1.GetThermalZoneResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetThermalZone")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetThermalZone"),
		attribute.String("zone.name", req.Msg.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing GetThermalZone request",
		slog.String("zone_name", req.Msg.GetName()))

	var thermalResp schemav1alpha1.GetThermalZoneResponse
	if err := s.requestNATS(ctx, ipc.SubjectThermalZoneInfo, req.Msg, &thermalResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetThermalZone request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetThermalZone request",
		slog.String("zone_name", req.Msg.GetName()))
	return connect.NewResponse(&thermalResp), nil
}

// SetThermalZone handles the SetThermalZone RPC call.
func (s *ProtoServer) SetThermalZone(ctx context.Context, req *connect.Request[schemav1alpha1.SetThermalZoneRequest]) (*connect.Response[schemav1alpha1.SetThermalZoneResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.SetThermalZone")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "SetThermalZone"),
		attribute.String("zone.name", req.Msg.GetName()),
	)

	s.logger.DebugContext(ctx, "Processing SetThermalZone request",
		slog.String("zone_name", req.Msg.GetName()))

	var thermalResp schemav1alpha1.SetThermalZoneResponse
	if err := s.requestNATS(ctx, ipc.SubjectThermalZoneSet, req.Msg, &thermalResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process SetThermalZone request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed SetThermalZone request",
		slog.String("zone_name", req.Msg.GetName()))
	return connect.NewResponse(&thermalResp), nil
}

// ListThermalZones handles the ListThermalZones RPC call.
func (s *ProtoServer) ListThermalZones(ctx context.Context, req *connect.Request[schemav1alpha1.ListThermalZonesRequest]) (*connect.Response[schemav1alpha1.ListThermalZonesResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListThermalZones")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListThermalZones"),
	)

	s.logger.DebugContext(ctx, "Processing ListThermalZones request")

	var thermalResp schemav1alpha1.ListThermalZonesResponse
	if err := s.requestNATS(ctx, ipc.SubjectThermalZoneList, req.Msg, &thermalResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListThermalZones request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListThermalZones request")
	return connect.NewResponse(&thermalResp), nil
}

// CreateUser handles the CreateUser RPC call.
func (s *ProtoServer) CreateUser(ctx context.Context, req *connect.Request[schemav1alpha1.CreateUserRequest]) (*connect.Response[schemav1alpha1.CreateUserResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "CreateUser"),
		attribute.String("user.name", req.Msg.GetUser().GetUsername()),
	)

	s.logger.DebugContext(ctx, "Processing CreateUser request",
		slog.String("user_name", req.Msg.GetUser().GetUsername()))

	var userResp schemav1alpha1.CreateUserResponse
	if err := s.requestNATS(ctx, ipc.SubjectUserCreate, req.Msg, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process CreateUser request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed CreateUser request",
		slog.String("user_name", req.Msg.GetUser().GetUsername()))
	return connect.NewResponse(&userResp), nil
}

// GetUser handles the GetUser RPC call.
func (s *ProtoServer) GetUser(ctx context.Context, req *connect.Request[schemav1alpha1.GetUserRequest]) (*connect.Response[schemav1alpha1.GetUserResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.GetUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "GetUser"),
		attribute.String("user.name", req.Msg.GetUsername()),
	)

	s.logger.DebugContext(ctx, "Processing GetUser request",
		slog.String("user_name", req.Msg.GetUsername()))

	var userResp schemav1alpha1.GetUserResponse
	if err := s.requestNATS(ctx, ipc.SubjectUserInfo, req.Msg, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process GetUser request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed GetUser request",
		slog.String("user_name", req.Msg.GetUsername()))
	return connect.NewResponse(&userResp), nil
}

// UpdateUser handles the UpdateUser RPC call.
func (s *ProtoServer) UpdateUser(ctx context.Context, req *connect.Request[schemav1alpha1.UpdateUserRequest]) (*connect.Response[schemav1alpha1.UpdateUserResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.UpdateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "UpdateUser"),
		attribute.String("user.name", req.Msg.GetUser().GetUsername()),
	)

	s.logger.DebugContext(ctx, "Processing UpdateUser request",
		slog.String("user_name", req.Msg.GetUser().GetUsername()))

	var userResp schemav1alpha1.UpdateUserResponse
	if err := s.requestNATS(ctx, ipc.SubjectUserUpdate, req.Msg, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process UpdateUser request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed UpdateUser request",
		slog.String("user_name", req.Msg.GetUser().GetUsername()))
	return connect.NewResponse(&userResp), nil
}

// DeleteUser handles the DeleteUser RPC call.
func (s *ProtoServer) DeleteUser(ctx context.Context, req *connect.Request[schemav1alpha1.DeleteUserRequest]) (*connect.Response[schemav1alpha1.DeleteUserResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.DeleteUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "DeleteUser"),
		attribute.String("user.name", req.Msg.GetId()),
	)

	s.logger.DebugContext(ctx, "Processing DeleteUser request",
		slog.String("user_id", req.Msg.GetId()))

	var userResp schemav1alpha1.DeleteUserResponse
	if err := s.requestNATS(ctx, ipc.SubjectUserDelete, req.Msg, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process DeleteUser request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed DeleteUser request",
		slog.String("user_id", req.Msg.GetId()))
	return connect.NewResponse(&userResp), nil
}

// ListUsers handles the ListUsers RPC call.
func (s *ProtoServer) ListUsers(ctx context.Context, req *connect.Request[schemav1alpha1.ListUsersRequest]) (*connect.Response[schemav1alpha1.ListUsersResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ListUsers")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ListUsers"),
	)

	s.logger.DebugContext(ctx, "Processing ListUsers request")

	var userResp schemav1alpha1.ListUsersResponse
	if err := s.requestNATS(ctx, ipc.SubjectUserList, req.Msg, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ListUsers request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ListUsers request")
	return connect.NewResponse(&userResp), nil
}

// ChangePassword handles the ChangePassword RPC call.
func (s *ProtoServer) ChangePassword(ctx context.Context, req *connect.Request[schemav1alpha1.ChangePasswordRequest]) (*connect.Response[schemav1alpha1.ChangePasswordResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ChangePassword")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ChangePassword"),
		attribute.String("user.name", req.Msg.GetId()),
	)

	s.logger.DebugContext(ctx, "Processing ChangePassword request",
		slog.String("user_id", req.Msg.GetId()))

	var userResp schemav1alpha1.ChangePasswordResponse
	if err := s.requestNATS(ctx, ipc.SubjectUserChangePassword, req.Msg, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ChangePassword request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ChangePassword request",
		slog.String("user_id", req.Msg.GetId()))
	return connect.NewResponse(&userResp), nil
}

// ResetPassword handles the ResetPassword RPC call.
func (s *ProtoServer) ResetPassword(ctx context.Context, req *connect.Request[schemav1alpha1.ResetPasswordRequest]) (*connect.Response[schemav1alpha1.ResetPasswordResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.ResetPassword")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "ResetPassword"),
		attribute.String("user.name", req.Msg.GetId()),
	)

	s.logger.DebugContext(ctx, "Processing ResetPassword request",
		slog.String("user_id", req.Msg.GetId()))

	var userResp schemav1alpha1.ResetPasswordResponse
	if err := s.requestNATS(ctx, ipc.SubjectUserResetPassword, req.Msg, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process ResetPassword request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed ResetPassword request",
		slog.String("user_id", req.Msg.GetId()))
	return connect.NewResponse(&userResp), nil
}

// AuthenticateUser handles the AuthenticateUser RPC call.
func (s *ProtoServer) AuthenticateUser(ctx context.Context, req *connect.Request[schemav1alpha1.AuthenticateUserRequest]) (*connect.Response[schemav1alpha1.AuthenticateUserResponse], error) {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.AuthenticateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.service", "BMCService"),
		attribute.String("rpc.method", "AuthenticateUser"),
		attribute.String("user.name", req.Msg.GetUsername()),
	)

	s.logger.DebugContext(ctx, "Processing AuthenticateUser request",
		slog.String("user_name", req.Msg.GetUsername()))

	var userResp schemav1alpha1.AuthenticateUserResponse
	if err := s.requestNATS(ctx, ipc.SubjectUserAuthenticate, req.Msg, &userResp); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to process AuthenticateUser request", "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Successfully processed AuthenticateUser request",
		slog.String("user_name", req.Msg.GetUsername()))
	return connect.NewResponse(&userResp), nil
}

type vtMessage interface {
	MarshalVT() ([]byte, error)
}

type vtUnmarshaler interface {
	UnmarshalVT([]byte) error
}

func sanitizeSubjectToken(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}
	if strings.ContainsAny(token, " \t\n\r.>*") {
		return "", fmt.Errorf("token contains invalid characters")
	}
	return token, nil
}

func (s *ProtoServer) requestNATS(ctx context.Context, subject string, req, resp proto.Message) error {
	ctx, span := s.tracer.Start(ctx, "ProtoServer.requestNATS")
	defer span.End()

	span.SetAttributes(
		attribute.String("nats.subject", subject),
	)

	var reqData []byte
	var err error

	// Use VTProtobuf marshaling if available for better performance
	if vtReq, ok := req.(vtMessage); ok {
		reqData, err = vtReq.MarshalVT()
	} else {
		reqData, err = proto.Marshal(req)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request with context timeout
	msg, err := s.nc.RequestWithContext(ctx, subject, reqData)
	if err != nil {
		if err == context.DeadlineExceeded {
			return connect.NewError(connect.CodeDeadlineExceeded, fmt.Errorf("request timeout for subject %s", subject))
		}
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("NATS request failed for subject %s: %w", subject, err))
	}

	// Unmarshal response using VTProtobuf if available
	if vtResp, ok := resp.(vtUnmarshaler); ok {
		err = vtResp.UnmarshalVT(msg.Data)
	} else {
		err = proto.Unmarshal(msg.Data, resp)
	}
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
