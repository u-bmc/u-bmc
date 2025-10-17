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

func (p *PowerMgr) handleHostPowerAction(ctx context.Context, req micro.Request) {
	start := time.Now()
	var operation string
	var componentName string

	if p.tracer != nil {
		var span trace.Span
		_, span = p.tracer.Start(ctx, "powermgr.handleHostPowerAction")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	// Validate subject format: powermgr.host.{id}.action
	parts := strings.Split(req.Subject(), ".")
	if len(parts) != 4 || parts[0] != "powermgr" || parts[1] != "host" || parts[3] != "action" {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	hostID := parts[2]
	componentName = fmt.Sprintf("host.%s", hostID)

	var request schemav1alpha1.ChangeHostStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	if request.Action == schemav1alpha1.HostAction_HOST_ACTION_UNSPECIFIED {
		ipc.RespondWithError(ctx, req, ErrInvalidPowerAction, "unspecified action")
		return
	}

	backend, err := p.getBackendForComponent(componentName)
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrBackendNotConfigured, err.Error())
		return
	}

	switch request.Action {
	case schemav1alpha1.HostAction_HOST_ACTION_ON:
		operation = "power_on"
		err = backend.PowerOn(ctx, componentName)
	case schemav1alpha1.HostAction_HOST_ACTION_OFF:
		operation = "power_off"
		err = backend.PowerOff(ctx, componentName, false)
	case schemav1alpha1.HostAction_HOST_ACTION_FORCE_OFF:
		operation = "force_off"
		err = backend.PowerOff(ctx, componentName, true)
	case schemav1alpha1.HostAction_HOST_ACTION_REBOOT:
		operation = "reboot"
		err = backend.Reset(ctx, componentName)
	case schemav1alpha1.HostAction_HOST_ACTION_FORCE_RESTART:
		operation = "force_restart"
		err = backend.Reset(ctx, componentName)
	default:
		ipc.RespondWithError(ctx, req, ErrPowerOperationNotSupported, fmt.Sprintf("unsupported action: %v", request.Action))
		return
	}

	// Record metrics
	p.recordOperation(ctx, operation, componentName, err)
	if p.powerOperationDuration != nil {
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

	response := &schemav1alpha1.ChangeHostStateResponse{
		CurrentStatus: schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING,
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
		p.logger.InfoContext(ctx, "Host power action completed",
			"component", componentName,
			"action", request.Action.String(),
			"duration", time.Since(start))
	}
}
