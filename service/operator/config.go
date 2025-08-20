// SPDX-License-Identifier: BSD-3-Clause

package operator

import (
	"log/slog"
	"time"

	"github.com/u-bmc/u-bmc/service"
	"github.com/u-bmc/u-bmc/service/consolesrv"
	"github.com/u-bmc/u-bmc/service/inventorymgr"
	"github.com/u-bmc/u-bmc/service/ipc"
	"github.com/u-bmc/u-bmc/service/ipmisrv"
	"github.com/u-bmc/u-bmc/service/kvmsrv"
	"github.com/u-bmc/u-bmc/service/ledmgr"
	"github.com/u-bmc/u-bmc/service/powermgr"
	"github.com/u-bmc/u-bmc/service/securitymgr"
	"github.com/u-bmc/u-bmc/service/sensormon"
	"github.com/u-bmc/u-bmc/service/statemgr"
	"github.com/u-bmc/u-bmc/service/telemetry"
	"github.com/u-bmc/u-bmc/service/thermalmgr"
	"github.com/u-bmc/u-bmc/service/updatemgr"
	"github.com/u-bmc/u-bmc/service/usermgr"
	"github.com/u-bmc/u-bmc/service/websrv"
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

// WithName sets the name for the operator configuration.
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

// WithID sets the unique identifier for the operator configuration.
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

// WithDisableLogo controls whether the logo display is disabled.
// When set to true, the logo will not be shown during startup.
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

// WithCustomLogo sets a custom logo to be displayed instead of the default logo.
// The customLogo parameter should be the path to the logo file or logo content.
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

// WithOtelSetup sets up OpenTelemetry configuration by providing a setup function.
// The function will be called during operator initialization to configure telemetry.
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

// WithLogger sets a custom structured logger for the operator.
// If not provided, a default logger will be used.
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

// WithTimeout sets the timeout duration for operator operations.
// This controls how long the operator will wait for operations to complete.
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

// WithIPC configures the Inter-Process Communication service with the provided options.
// This service handles communication between different BMC processes.
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

// WithTelemetry configures the telemetry service with the provided options.
// This service collects and reports metrics and observability data.
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

// WithConsolesrv configures the console service with the provided options.
// This service provides serial console access to the managed system.
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

// WithInventorymgr configures the inventory manager service with the provided options.
// This service manages hardware inventory information and component discovery.
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

// WithIpmisrv configures the IPMI service with the provided options.
// This service implements the Intelligent Platform Management Interface protocol.
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

// WithKvmsrv configures the KVM service with the provided options.
// This service provides keyboard, video, and mouse redirection capabilities.
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

// WithLedmgr configures the LED manager service with the provided options.
// This service controls system status and identification LEDs.
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

// WithPowermgr configures the power manager service with the provided options.
// This service handles system power control operations like power on/off and reset.
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

// WithSecuritymgr configures the security manager service with the provided options.
// This service handles authentication, authorization, and security policies.
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

// WithSensormon configures the sensor monitoring service with the provided options.
// This service monitors hardware sensors for temperature, voltage, fan speed, etc.
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

// WithStatemgr configures the state manager service with the provided options.
// This service manages system state transitions and maintains state consistency.
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

// WithThermalmgr configures the thermal manager service with the provided options.
// This service manages system cooling and thermal protection policies.
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

// WithUpdatemgr configures the update manager service with the provided options.
// This service handles firmware and software updates for system components.
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

// WithUsermgr configures the user manager service with the provided options.
// This service manages user accounts, passwords, and access permissions.
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

// WithWebsrv configures the web server service with the provided options.
// This service provides the web-based management interface and REST APIs.
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

// WithExtraServices adds additional custom services to the operator configuration.
// These services will be managed alongside the standard BMC services.
func WithExtraServices(services ...service.Service) Option {
	return &servicesOption{
		services: services,
	}
}
