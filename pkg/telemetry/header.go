// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"context"

	"github.com/nats-io/nats.go/micro"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// GetCtxFromReq extracts distributed tracing context from a NATS micro service request.
// It uses OpenTelemetry's text map propagator to extract trace context from the request
// headers and returns a new context containing the propagated trace information.
// If no trace context is found in the headers, it returns a context derived from
// context.Background().
func GetCtxFromReq(req micro.Request) context.Context {
	return otel.GetTextMapPropagator().Extract(context.Background(), propagation.HeaderCarrier(req.Headers()))
}
