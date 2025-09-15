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
	"go.opentelemetry.io/otel/trace"
)

func (p *PowerMgr) handleChassisPowerAction(ctx context.Context, req micro.Request) {
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
	componentName := fmt.Sprintf("chassis.%s", chassisID)

	var request schemav1alpha1.ChangeChassisStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	if request.Action == schemav1alpha1.ChassisAction_CHASSIS_ACTION_UNSPECIFIED {
		ipc.RespondWithError(ctx, req, ErrInvalidPowerAction, "unspecified action")
		return
	}

	var err error
	switch request.Action {
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_ON:
		err = p.backend.PowerOn(ctx, componentName)
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF:
		err = p.backend.PowerOff(ctx, componentName, false)
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_EMERGENCY_SHUTDOWN:
		err = p.backend.PowerOff(ctx, componentName, true)
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_POWER_CYCLE:
		if powerOffErr := p.backend.PowerOff(ctx, componentName, false); powerOffErr != nil {
			err = powerOffErr
		} else {
			time.Sleep(2 * time.Second)
			err = p.backend.PowerOn(ctx, componentName)
		}
	default:
		ipc.RespondWithError(ctx, req, ErrPowerOperationNotSupported, fmt.Sprintf("unsupported action: %v", request.Action))
		return
	}

	if err != nil {
		ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
		return
	}

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
			"action", request.Action.String())
	}
}
