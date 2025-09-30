// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/u-bmc/u-bmc/pkg/cert"
	"github.com/u-bmc/u-bmc/service/operator"
	"github.com/u-bmc/u-bmc/service/statemgr"
	"github.com/u-bmc/u-bmc/service/websrv"
)

func main() {
	// The device has only 512MB of RAM; limit memory usage to 256MB
	debug.SetMemoryLimit(256 * 1024 * 1024)

	// Configure state management
	stateConfig := []statemgr.Option{
		statemgr.WithStreamRetention(0), // Keep forever for audit trail
		statemgr.WithHostManagement(true),
		statemgr.WithChassisManagement(true),
		statemgr.WithBMCManagement(true),
		statemgr.WithNumHosts(1),
		statemgr.WithNumChassis(1),
		statemgr.WithStateTimeout(20 * time.Second),
		statemgr.WithBroadcastStateChanges(true),
		statemgr.WithPersistStateChanges(false),
	}

	webConfig := []websrv.Option{
		websrv.WithAddr(":443"),
		websrv.WithWebUI(false),
		websrv.WithWebUIPath("/opt/u-bmc/webui"),
		websrv.WithReadTimeout(30 * time.Second),
		websrv.WithWriteTimeout(30 * time.Second),
		websrv.WithIdleTimeout(120 * time.Second),
		websrv.WithRmemMax("7500000"),
		websrv.WithWmemMax("7500000"),
		websrv.WithCertificateType(cert.CertificateTypeSelfSigned),
		websrv.WithHostname("fmadio-5514-bmc.local"),
		websrv.WithCertPath("/var/cache/cert/ubmc-cert.pem"),
		websrv.WithKeyPath("/var/cache/cert/ubmc-key.pem"),
		websrv.WithAlternativeNames("u-bmc-local"),
	}

	if err := operator.New(
		// Init on this platform handles mounts; keep operator startup resilient.
		operator.WithMountCheck(false),
		// Not implemented
		operator.WithoutConsolesrv(),
		operator.WithoutInventorymgr(),
		operator.WithoutIpmisrv(),
		operator.WithoutTelemetry(),
		operator.WithoutUpdatemgr(),
		operator.WithoutUsermgr(),
		operator.WithoutSecuritymgr(),
		// Implemented
		operator.WithStatemgr(stateConfig...),
		operator.WithWebsrv(webConfig...),
		operator.WithoutPowermgr(),
		operator.WithoutLedmgr(),
		operator.WithoutKvmsrv(),
		operator.WithoutSensormon(),
		operator.WithoutThermalmgr(),
	).Run(context.Background(), nil); err != nil {
		panic(err)
	}
}
