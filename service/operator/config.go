// SPDX-License-Identifier: BSD-3-Clause

package operator

import (
	"log/slog"
	"time"

	"u-bmc.org/u-bmc/service"
	"u-bmc.org/u-bmc/service/consolesrv"
	"u-bmc.org/u-bmc/service/inventorymgr"
	"u-bmc.org/u-bmc/service/ipc"
	"u-bmc.org/u-bmc/service/ipmisrv"
	"u-bmc.org/u-bmc/service/kvmsrv"
	"u-bmc.org/u-bmc/service/ledmgr"
	"u-bmc.org/u-bmc/service/powermgr"
	"u-bmc.org/u-bmc/service/securitymgr"
	"u-bmc.org/u-bmc/service/sensormon"
	"u-bmc.org/u-bmc/service/statemgr"
	"u-bmc.org/u-bmc/service/telemetry"
	"u-bmc.org/u-bmc/service/thermalmgr"
	"u-bmc.org/u-bmc/service/updatemgr"
	"u-bmc.org/u-bmc/service/usermgr"
	"u-bmc.org/u-bmc/service/websrv"
)

type config struct {
	name        string
	id          string
	disableLogo bool
	customLogo  string
	otelSetup   func()
	logger      *slog.Logger
	timeout     time.Duration
	// IPC service needs special handling
	ipc *ipc.IPC
	// Everything of type service.Service needs to be exported
	Consolesrv   service.Service
	Inventorymgr service.Service
	Ipmisrv      service.Service
	Kvmsrv       service.Service
	Ledmgr       service.Service
	Powermgr     service.Service
	Securitymgr  service.Service
	Sensormon    service.Service
	Statemgr     service.Service
	Telemetry    service.Service
	Thermalmgr   service.Service
	Updatemgr    service.Service
	Usermgr      service.Service
	Websrv       service.Service

	extraServices []service.Service
}

type Option interface {
	apply(*config)
}

type nameOption struct {
	name string
}

func (o *nameOption) apply(c *config) {
	c.name = o.name
}

func WithName(name string) Option {
	return &nameOption{
		name: name,
	}
}

type idOption struct {
	id string
}

func (o *idOption) apply(c *config) {
	c.id = o.id
}

func WithID(id string) Option {
	return &idOption{
		id: id,
	}
}

type disableLogoOption struct {
	disableLogo bool
}

func (o *disableLogoOption) apply(c *config) {
	c.disableLogo = o.disableLogo
}

func WithDisableLogo(disableLogo bool) Option {
	return &disableLogoOption{
		disableLogo: disableLogo,
	}
}

type customLogoOption struct {
	customLogo string
}

func (o *customLogoOption) apply(c *config) {
	c.customLogo = o.customLogo
}

func WithCustomLogo(customLogo string) Option {
	return &customLogoOption{
		customLogo: customLogo,
	}
}

type otelSetupOption struct {
	otelSetup func()
}

func (o *otelSetupOption) apply(c *config) {
	c.otelSetup = o.otelSetup
}

func WithOtelSetup(otelSetup func()) Option {
	return &otelSetupOption{
		otelSetup: otelSetup,
	}
}

type loggerOption struct {
	logger *slog.Logger
}

func (o *loggerOption) apply(c *config) {
	c.logger = o.logger
}

func WithLogger(logger *slog.Logger) Option {
	return &loggerOption{
		logger: logger,
	}
}

type timeoutOption struct {
	timeout time.Duration
}

func (o *timeoutOption) apply(c *config) {
	c.timeout = o.timeout
}

func WithTimeout(timeout time.Duration) Option {
	return &timeoutOption{
		timeout: timeout,
	}
}

type ipcOption struct {
	ipc *ipc.IPC
}

func (o *ipcOption) apply(c *config) {
	c.ipc = o.ipc
}

func WithIPC(opts ...ipc.Option) Option {
	return &ipcOption{
		ipc: ipc.New(opts...),
	}
}

type telemetryOption struct {
	telemetry service.Service
}

func (o *telemetryOption) apply(c *config) {
	c.Telemetry = o.telemetry
}

func WithTelemetry(opts ...telemetry.Option) Option {
	return &telemetryOption{
		telemetry: telemetry.New(opts...),
	}
}

type consolesrvOption struct {
	consolesrv service.Service
}

func (o *consolesrvOption) apply(c *config) {
	c.Consolesrv = o.consolesrv
}

func WithConsolesrv(opts ...consolesrv.Option) Option {
	return &consolesrvOption{
		consolesrv: consolesrv.New(opts...),
	}
}

type inventorymgrOption struct {
	inventorymgr service.Service
}

func (o *inventorymgrOption) apply(c *config) {
	c.Inventorymgr = o.inventorymgr
}

func WithInventorymgr(opts ...inventorymgr.Option) Option {
	return &inventorymgrOption{
		inventorymgr: inventorymgr.New(opts...),
	}
}

type ipmisrvOption struct {
	ipmisrv service.Service
}

func (o *ipmisrvOption) apply(c *config) {
	c.Ipmisrv = o.ipmisrv
}

func WithIpmisrv(opts ...ipmisrv.Option) Option {
	return &ipmisrvOption{
		ipmisrv: ipmisrv.New(opts...),
	}
}

type kvmsrvOption struct {
	kvmsrv service.Service
}

func (o *kvmsrvOption) apply(c *config) {
	c.Kvmsrv = o.kvmsrv
}

func WithKvmsrv(opts ...kvmsrv.Option) Option {
	return &kvmsrvOption{
		kvmsrv: kvmsrv.New(opts...),
	}
}

type ledmgrOption struct {
	ledmgr service.Service
}

func (o *ledmgrOption) apply(c *config) {
	c.Ledmgr = o.ledmgr
}

func WithLedmgr(opts ...ledmgr.Option) Option {
	return &ledmgrOption{
		ledmgr: ledmgr.New(opts...),
	}
}

type powermgrOption struct {
	powermgr service.Service
}

func (o *powermgrOption) apply(c *config) {
	c.Powermgr = o.powermgr
}

func WithPowermgr(opts ...powermgr.Option) Option {
	return &powermgrOption{
		powermgr: powermgr.New(opts...),
	}
}

type securitymgrOption struct {
	securitymgr service.Service
}

func (o *securitymgrOption) apply(c *config) {
	c.Securitymgr = o.securitymgr
}

func WithSecuritymgr(opts ...securitymgr.Option) Option {
	return &securitymgrOption{
		securitymgr: securitymgr.New(opts...),
	}
}

type sensormonOption struct {
	sensormon service.Service
}

func (o *sensormonOption) apply(c *config) {
	c.Sensormon = o.sensormon
}

func WithSensormon(opts ...sensormon.Option) Option {
	return &sensormonOption{
		sensormon: sensormon.New(opts...),
	}
}

type statemgrOption struct {
	statemgr service.Service
}

func (o *statemgrOption) apply(c *config) {
	c.Statemgr = o.statemgr
}

func WithStatemgr(opts ...statemgr.Option) Option {
	return &statemgrOption{
		statemgr: statemgr.New(opts...),
	}
}

type thermalmgrOption struct {
	thermalmgr service.Service
}

func (o *thermalmgrOption) apply(c *config) {
	c.Thermalmgr = o.thermalmgr
}

func WithThermalmgr(opts ...thermalmgr.Option) Option {
	return &thermalmgrOption{
		thermalmgr: thermalmgr.New(opts...),
	}
}

type updatemgrOption struct {
	updatemgr service.Service
}

func (o *updatemgrOption) apply(c *config) {
	c.Updatemgr = o.updatemgr
}

func WithUpdatemgr(opts ...updatemgr.Option) Option {
	return &updatemgrOption{
		updatemgr: updatemgr.New(opts...),
	}
}

type usermgrOption struct {
	usermgr service.Service
}

func (o *usermgrOption) apply(c *config) {
	c.Usermgr = o.usermgr
}

func WithUsermgr(opts ...usermgr.Option) Option {
	return &usermgrOption{
		usermgr: usermgr.New(opts...),
	}
}

type websrvOption struct {
	websrv service.Service
}

func (o *websrvOption) apply(c *config) {
	c.Websrv = o.websrv
}

func WithWebsrv(opts ...websrv.Option) Option {
	return &websrvOption{
		websrv: websrv.New(opts...),
	}
}

type servicesOption struct {
	services []service.Service
}

func (o *servicesOption) apply(c *config) {
	c.extraServices = o.services
}

func WithExtraServices(services ...service.Service) Option {
	return &servicesOption{
		services: services,
	}
}
