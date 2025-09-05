// SPDX-License-Identifier: BSD-3-Clause

package ipc

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/micro"
	"github.com/u-bmc/u-bmc/pkg/log"
)

// RespondWithError sends an error response to a NATS request with proper logging.
func RespondWithError(ctx context.Context, req micro.Request, err error, details string) {
	l := log.GetGlobalLogger()

	l.ErrorContext(ctx, "Request failed",
		"subject", req.Subject(),
		"error", err,
		"details", details)

	if respErr := req.Error("500", fmt.Sprintf("%v: %s", err, details), nil); respErr != nil {
		l.ErrorContext(ctx, "Failed to send error response",
			"subject", req.Subject(),
			"error", respErr)
	}
}
