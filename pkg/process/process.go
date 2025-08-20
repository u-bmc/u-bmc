// SPDX-License-Identifier: BSD-3-Clause

package process

import (
	"context"
	"fmt"

	"cirello.io/oversight/v2"
	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/service"
)

func New(s service.Service, ipcConn nats.InProcessConnProvider) oversight.ChildProcess {
	return func(ctx context.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%s panicked: %v", s.Name(), r)
			}
		}()

		return s.Run(ctx, ipcConn)
	}
}
