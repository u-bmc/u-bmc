// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"

	"github.com/u-bmc/u-bmc/service/operator"
)

func main() {
	if err := operator.New(
		operator.WithName("asrock-turin2d24tm3-2l-operator"),
		// Init on this platform handles mounts; keep operator startup resilient.
		operator.WithMountCheck(false),
	).Run(context.Background(), nil); err != nil {
		panic(err)
	}
}
