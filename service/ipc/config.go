// SPDX-License-Identifier: BSD-3-Clause

package ipc

import (
	"fmt"
	"time"

	"github.com/nats-io/nats-server/v2/server"
)

// Default configuration constants.
const (
	DefaultServiceName        = "ipc"
	DefaultServiceDescription = "Inter-process communication service for BMC components"
	DefaultServiceVersion     = "1.0.0"
	DefaultServerName         = "u-bmc-ipc"
	DefaultStoreDir           = "/var/lib/u-bmc/ipc"
	DefaultMaxMemory          = 64 * 1024 * 1024 // 64MB
	DefaultMaxStorage         = 10 * 1024 * 1024 // 10MB
	DefaultStartupTimeout     = 30 * time.Second
	DefaultShutdownTimeout    = 10 * time.Second
)

type config struct {
	serviceName                 string
	serviceDescription          string
	serviceVersion              string
	serverName                  string
	storeDir                    string
	enableJetStream             bool
	dontListen                  bool
	maxMemory                   int64
	maxStorage                  int64
	startupTimeout              time.Duration
	shutdownTimeout             time.Duration
	maxConnections              int
	maxControlLine              int32
	maxPayload                  int32
	writeDeadline               time.Duration
	pingInterval                time.Duration
	maxPingsOut                 int
	enableSlowConsumerDetection bool
	slowConsumerThreshold       time.Duration
}

// Option represents a configuration option for the IPC service.
type Option interface {
	apply(*config)
}

type serviceNameOption struct {
	name string
}

func (o *serviceNameOption) apply(c *config) {
	c.serviceName = o.name
}

// WithServiceName sets the name of the service.
func WithServiceName(name string) Option {
	return &serviceNameOption{name: name}
}

type serviceDescriptionOption struct {
	description string
}

func (o *serviceDescriptionOption) apply(c *config) {
	c.serviceDescription = o.description
}

// WithServiceDescription sets the description of the service.
func WithServiceDescription(description string) Option {
	return &serviceDescriptionOption{description: description}
}

type serviceVersionOption struct {
	version string
}

func (o *serviceVersionOption) apply(c *config) {
	c.serviceVersion = o.version
}

// WithServiceVersion sets the version of the service.
func WithServiceVersion(version string) Option {
	return &serviceVersionOption{version: version}
}

type serverNameOption struct {
	name string
}

func (o *serverNameOption) apply(c *config) {
	c.serverName = o.name
}

// WithServerName sets the NATS server name.
func WithServerName(name string) Option {
	return &serverNameOption{name: name}
}

type storeDirOption struct {
	dir string
}

func (o *storeDirOption) apply(c *config) {
	c.storeDir = o.dir
}

// WithStoreDir sets the JetStream storage directory.
func WithStoreDir(dir string) Option {
	return &storeDirOption{dir: dir}
}

type enableJetStreamOption struct {
	enable bool
}

func (o *enableJetStreamOption) apply(c *config) {
	c.enableJetStream = o.enable
}

// WithJetStream enables or disables JetStream functionality.
func WithJetStream(enable bool) Option {
	return &enableJetStreamOption{enable: enable}
}

type dontListenOption struct {
	dontListen bool
}

func (o *dontListenOption) apply(c *config) {
	c.dontListen = o.dontListen
}

// WithDontListen prevents the server from listening on network interfaces.
func WithDontListen(dontListen bool) Option {
	return &dontListenOption{dontListen: dontListen}
}

type maxMemoryOption struct {
	maxMemory int64
}

func (o *maxMemoryOption) apply(c *config) {
	c.maxMemory = o.maxMemory
}

// WithMaxMemory sets the maximum memory usage for JetStream.
func WithMaxMemory(maxMemory int64) Option {
	return &maxMemoryOption{maxMemory: maxMemory}
}

type maxStorageOption struct {
	maxStorage int64
}

func (o *maxStorageOption) apply(c *config) {
	c.maxStorage = o.maxStorage
}

// WithMaxStorage sets the maximum storage usage for JetStream.
func WithMaxStorage(maxStorage int64) Option {
	return &maxStorageOption{maxStorage: maxStorage}
}

type startupTimeoutOption struct {
	timeout time.Duration
}

func (o *startupTimeoutOption) apply(c *config) {
	c.startupTimeout = o.timeout
}

// WithStartupTimeout sets the maximum time to wait for server startup.
func WithStartupTimeout(timeout time.Duration) Option {
	return &startupTimeoutOption{timeout: timeout}
}

type shutdownTimeoutOption struct {
	timeout time.Duration
}

func (o *shutdownTimeoutOption) apply(c *config) {
	c.shutdownTimeout = o.timeout
}

// WithShutdownTimeout sets the maximum time to wait for graceful shutdown.
func WithShutdownTimeout(timeout time.Duration) Option {
	return &shutdownTimeoutOption{timeout: timeout}
}

type maxConnectionsOption struct {
	maxConnections int
}

func (o *maxConnectionsOption) apply(c *config) {
	c.maxConnections = o.maxConnections
}

// WithMaxConnections sets the maximum number of concurrent connections.
func WithMaxConnections(maxConnections int) Option {
	return &maxConnectionsOption{maxConnections: maxConnections}
}

type maxControlLineOption struct {
	maxControlLine int32
}

func (o *maxControlLineOption) apply(c *config) {
	c.maxControlLine = o.maxControlLine
}

// WithMaxControlLine sets the maximum size of control messages.
func WithMaxControlLine(maxControlLine int32) Option {
	return &maxControlLineOption{maxControlLine: maxControlLine}
}

type maxPayloadOption struct {
	maxPayload int32
}

func (o *maxPayloadOption) apply(c *config) {
	c.maxPayload = o.maxPayload
}

// WithMaxPayload sets the maximum size of message payloads.
func WithMaxPayload(maxPayload int32) Option {
	return &maxPayloadOption{maxPayload: maxPayload}
}

type writeDeadlineOption struct {
	writeDeadline time.Duration
}

func (o *writeDeadlineOption) apply(c *config) {
	c.writeDeadline = o.writeDeadline
}

// WithWriteDeadline sets the write deadline for connections.
func WithWriteDeadline(writeDeadline time.Duration) Option {
	return &writeDeadlineOption{writeDeadline: writeDeadline}
}

type pingIntervalOption struct {
	pingInterval time.Duration
}

func (o *pingIntervalOption) apply(c *config) {
	c.pingInterval = o.pingInterval
}

// WithPingInterval sets the ping interval for client connections.
func WithPingInterval(pingInterval time.Duration) Option {
	return &pingIntervalOption{pingInterval: pingInterval}
}

type maxPingsOutOption struct {
	maxPingsOut int
}

func (o *maxPingsOutOption) apply(c *config) {
	c.maxPingsOut = o.maxPingsOut
}

// WithMaxPingsOut sets the maximum number of outstanding pings.
func WithMaxPingsOut(maxPingsOut int) Option {
	return &maxPingsOutOption{maxPingsOut: maxPingsOut}
}

type enableSlowConsumerDetectionOption struct {
	enable bool
}

func (o *enableSlowConsumerDetectionOption) apply(c *config) {
	c.enableSlowConsumerDetection = o.enable
}

// WithSlowConsumerDetection enables or disables slow consumer detection.
func WithSlowConsumerDetection(enable bool) Option {
	return &enableSlowConsumerDetectionOption{enable: enable}
}

type slowConsumerThresholdOption struct {
	threshold time.Duration
}

func (o *slowConsumerThresholdOption) apply(c *config) {
	c.slowConsumerThreshold = o.threshold
}

// WithSlowConsumerThreshold sets the threshold for slow consumer detection.
func WithSlowConsumerThreshold(threshold time.Duration) Option {
	return &slowConsumerThresholdOption{threshold: threshold}
}

type serverOptionsOption struct {
	opts *server.Options
}

func (o *serverOptionsOption) apply(c *config) {
	if o.opts != nil {
		// Apply server options to config
		if o.opts.ServerName != "" {
			c.serverName = o.opts.ServerName
		}
		if o.opts.StoreDir != "" {
			c.storeDir = o.opts.StoreDir
		}
		c.enableJetStream = o.opts.JetStream
		c.dontListen = o.opts.DontListen
		if o.opts.JetStreamMaxMemory > 0 {
			c.maxMemory = o.opts.JetStreamMaxMemory
		}
		if o.opts.JetStreamMaxStore > 0 {
			c.maxStorage = o.opts.JetStreamMaxStore
		}
		if o.opts.MaxConn > 0 {
			c.maxConnections = o.opts.MaxConn
		}
		if o.opts.MaxControlLine > 0 {
			c.maxControlLine = o.opts.MaxControlLine
		}
		if o.opts.MaxPayload > 0 {
			c.maxPayload = o.opts.MaxPayload
		}
		if o.opts.WriteDeadline > 0 {
			c.writeDeadline = o.opts.WriteDeadline
		}
		if o.opts.PingInterval > 0 {
			c.pingInterval = o.opts.PingInterval
		}
		if o.opts.MaxPingsOut > 0 {
			c.maxPingsOut = o.opts.MaxPingsOut
		}
	}
}

// WithServerOpts applies NATS server options to the configuration.
// This option is provided for backward compatibility and will override
// other configuration options with values from the server.Options.
func WithServerOpts(opts *server.Options) Option {
	return &serverOptionsOption{opts: opts}
}

// Validate checks if the configuration is valid.
func (c *config) Validate() error {
	if c.serviceName == "" {
		return fmt.Errorf("%w: service name cannot be empty", ErrInvalidConfiguration)
	}

	if c.serviceVersion == "" {
		return fmt.Errorf("%w: service version cannot be empty", ErrInvalidConfiguration)
	}

	if c.serverName == "" {
		return fmt.Errorf("%w: server name cannot be empty", ErrInvalidConfiguration)
	}

	if c.enableJetStream && c.storeDir == "" {
		return fmt.Errorf("%w: store directory cannot be empty when JetStream is enabled", ErrInvalidConfiguration)
	}

	if c.maxMemory < 0 {
		return fmt.Errorf("%w: max memory cannot be negative", ErrInvalidConfiguration)
	}

	if c.maxStorage < 0 {
		return fmt.Errorf("%w: max storage cannot be negative", ErrInvalidConfiguration)
	}

	if c.startupTimeout <= 0 {
		return fmt.Errorf("%w: startup timeout must be positive", ErrInvalidConfiguration)
	}

	if c.shutdownTimeout <= 0 {
		return fmt.Errorf("%w: shutdown timeout must be positive", ErrInvalidConfiguration)
	}

	if c.maxConnections < 0 {
		return fmt.Errorf("%w: max connections cannot be negative", ErrInvalidConfiguration)
	}

	if c.maxControlLine <= 0 {
		return fmt.Errorf("%w: max control line must be positive", ErrInvalidConfiguration)
	}

	if c.maxPayload <= 0 {
		return fmt.Errorf("%w: max payload must be positive", ErrInvalidConfiguration)
	}

	if c.writeDeadline <= 0 {
		return fmt.Errorf("%w: write deadline must be positive", ErrInvalidConfiguration)
	}

	if c.pingInterval <= 0 {
		return fmt.Errorf("%w: ping interval must be positive", ErrInvalidConfiguration)
	}

	if c.maxPingsOut <= 0 {
		return fmt.Errorf("%w: max pings out must be positive", ErrInvalidConfiguration)
	}

	if c.enableSlowConsumerDetection && c.slowConsumerThreshold <= 0 {
		return fmt.Errorf("%w: slow consumer threshold must be positive when detection is enabled", ErrInvalidConfiguration)
	}

	return nil
}

// WithName is a backward compatibility alias for WithServiceName.
// Deprecated: Use WithServiceName instead.
func WithName(name string) Option {
	return WithServiceName(name)
}

// ToServerOptions converts the configuration to NATS server.Options.
func (c *config) ToServerOptions() *server.Options {
	opts := &server.Options{
		ServerName:             c.serverName,
		DontListen:             c.dontListen,
		JetStream:              c.enableJetStream,
		DisableJetStreamBanner: true,
		StoreDir:               c.storeDir,
		MaxConn:                c.maxConnections,
		MaxControlLine:         c.maxControlLine,
		MaxPayload:             c.maxPayload,
		WriteDeadline:          c.writeDeadline,
		PingInterval:           c.pingInterval,
		MaxPingsOut:            c.maxPingsOut,
	}

	if c.enableJetStream {
		opts.JetStreamMaxMemory = c.maxMemory
		opts.JetStreamMaxStore = c.maxStorage
	}

	return opts
}
