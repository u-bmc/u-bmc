// SPDX-License-Identifier: BSD-3-Clause

package powermgr

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go/micro"
	schemav1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/ipc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func (p *PowerMgr) handleChassisPowerAction(ctx context.Context, req micro.Request) {
	start := time.Now()
	var operation string
	var componentName string

	if p.tracer != nil {
		var span trace.Span
		_, span = p.tracer.Start(ctx, "powermgr.handleChassisPowerAction")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) != 4 || parts[0] != "powermgr" || parts[1] != "chassis" {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	chassisID := parts[2]
	componentName = fmt.Sprintf("chassis.%s", chassisID)

	var request schemav1alpha1.ChangeChassisStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	if request.Action == schemav1alpha1.ChassisAction_CHASSIS_ACTION_UNSPECIFIED {
		ipc.RespondWithError(ctx, req, ErrInvalidPowerAction, "unspecified action")
		return
	}

	backend, err := p.getBackendForComponent(componentName)
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrBackendNotConfigured, err.Error())
		return
	}

	switch request.Action {
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_ON:
		operation = "power_on"
		err = backend.PowerOn(ctx, componentName)
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF:
		operation = "power_off"
		err = backend.PowerOff(ctx, componentName, false)
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_EMERGENCY_SHUTDOWN:
		operation = "emergency_shutdown"
		err = backend.PowerOff(ctx, componentName, true)
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_POWER_CYCLE:
		operation = "power_cycle"
		if powerOffErr := backend.PowerOff(ctx, componentName, false); powerOffErr != nil {
			err = powerOffErr
		} else {
			time.Sleep(2 * time.Second)
			err = backend.PowerOn(ctx, componentName)
		}
	default:
		ipc.RespondWithError(ctx, req, ErrPowerOperationNotSupported, fmt.Sprintf("unsupported action: %v", request.Action))
		return
	}

	// Record metrics
	p.recordOperation(ctx, operation, componentName, err)
	if p.config.enableMetrics && p.powerOperationDuration != nil {
		duration := time.Since(start).Seconds()
		p.powerOperationDuration.Record(ctx, duration, metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("component", componentName),
		))
	}

	if err != nil {
		p.reportStateChange(ctx, componentName, operation, false)
		ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
		return
	}

	// Report successful state change
	p.reportStateChange(ctx, componentName, operation, true)

	response := &schemav1alpha1.ChangeChassisStateResponse{
		CurrentStatus: schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING,
	}

	resp, marshalErr := response.MarshalVT()
	if marshalErr != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, marshalErr.Error())
		return
	}

	if err := req.Respond(resp); err != nil && p.logger != nil {
		p.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}

	if p.logger != nil {
		p.logger.InfoContext(ctx, "Chassis power action completed",
			"component", componentName,
			"action", request.Action.String(),
			"duration", time.Since(start))
	}
}
