// SPDX-License-Identifier: BSD-3-Clause

package websrv

import (
	"github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1/schemav1alpha1connect"
)

// ProtoServer implements all the Connect RPC service handlers for the BMC API.
type ProtoServer struct {
	schemav1alpha1connect.UnimplementedBMCServiceHandler
}
