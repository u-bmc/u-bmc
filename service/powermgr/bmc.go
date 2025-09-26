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

func (p *PowerMgr) handleBMCPowerAction(ctx context.Context, req micro.Request) {
	start := time.Now()
	var operation string
	var componentName string

	if p.tracer != nil {
		var span trace.Span
		_, span = p.tracer.Start(ctx, "powermgr.handleBMCPowerAction")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) != 4 || parts[0] != DefaultServiceName || parts[1] != "bmc" {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	bmcID := parts[2]
	componentName = fmt.Sprintf("bmc.%s", bmcID)

	var request schemav1alpha1.ChangeManagementControllerStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	if request.Action == schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_UNSPECIFIED {
		ipc.RespondWithError(ctx, req, ErrInvalidPowerAction, "unspecified action")
		return
	}

	backend, err := p.getBackendForComponent(componentName)
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrBackendNotConfigured, err.Error())
		return
	}

	switch request.Action {
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_REBOOT:
		operation = "reboot"
		err = backend.Reset(ctx, componentName)
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_WARM_RESET:
		operation = "warm_reset"
		err = backend.Reset(ctx, componentName)
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_COLD_RESET:
		operation = "cold_reset"
		err = backend.Reset(ctx, componentName)
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_HARD_RESET:
		operation = "hard_reset"
		err = backend.Reset(ctx, componentName)
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_FACTORY_RESET:
		operation = "factory_reset"
		err = backend.Reset(ctx, componentName)
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

	response := &schemav1alpha1.ChangeManagementControllerStateResponse{
		CurrentStatus: schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY,
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
		p.logger.InfoContext(ctx, "BMC power action completed",
			"component", componentName,
			"action", request.Action.String(),
			"duration", time.Since(start))
	}
}
