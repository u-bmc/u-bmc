// SPDX-License-Identifier: BSD-3-Clause

package statemgr

import (
	"fmt"
	"strings"
	"time"
)

// Default configuration constants.
const (
	DefaultServiceName        = "statemgr"
	DefaultServiceDescription = "State management service for BMC components"
	DefaultServiceVersion     = "1.0.0"
	DefaultStreamName         = "STATEMGR"
)

// Config holds the configuration for the state manager service.
type Config struct {
	// ServiceName is the name of the service in the NATS micro framework
	ServiceName string
	// ServiceDescription provides a human-readable description of the service
	ServiceDescription string
	// ServiceVersion is the semantic version of the service
	ServiceVersion string
	// StreamName is the name of the JetStream stream for state persistence
	StreamName string
	// StreamSubjects are the subjects the stream will listen on
	StreamSubjects []string
	// StreamRetention defines how long to retain state events
	StreamRetention time.Duration
	// EnableHostManagement enables state management for host components
	EnableHostManagement bool
	// EnableChassisManagement enables state management for chassis components
	EnableChassisManagement bool
	// EnableBMCManagement enables state management for BMC components
	EnableBMCManagement bool
	// NumHosts is the number of hosts to manage
	NumHosts int
	// NumChassis is the number of chassis to manage
	NumChassis int
	// StateTimeout is the maximum duration for state transitions
	StateTimeout time.Duration
	// EnableMetrics enables metrics collection for state transitions
	EnableMetrics bool
	// EnableTracing enables distributed tracing for state transitions
	EnableTracing bool
	// BroadcastStateChanges enables broadcasting state changes via NATS
	BroadcastStateChanges bool
	// PersistStateChanges enables persisting state changes to JetStream
	PersistStateChanges bool
}

// Option represents a configuration option for the state manager.
type Option interface {
	apply(*Config)
}

type serviceNameOption struct {
	name string
}

func (o *serviceNameOption) apply(c *Config) {
	c.ServiceName = o.name
}

// WithServiceName sets the name of the service.
func WithServiceName(name string) Option {
	return &serviceNameOption{name: name}
}

type serviceDescriptionOption struct {
	description string
}

func (o *serviceDescriptionOption) apply(c *Config) {
	c.ServiceDescription = o.description
}

// WithServiceDescription sets the description of the service.
func WithServiceDescription(description string) Option {
	return &serviceDescriptionOption{description: description}
}

type serviceVersionOption struct {
	version string
}

func (o *serviceVersionOption) apply(c *Config) {
	c.ServiceVersion = o.version
}

// WithServiceVersion sets the version of the service.
func WithServiceVersion(version string) Option {
	return &serviceVersionOption{version: version}
}

type streamNameOption struct {
	name string
}

func (o *streamNameOption) apply(c *Config) {
	c.StreamName = o.name
}

// WithStreamName sets the JetStream stream name for state persistence.
func WithStreamName(name string) Option {
	return &streamNameOption{name: name}
}

type streamSubjectsOption struct {
	subjects []string
}

func (o *streamSubjectsOption) apply(c *Config) {
	c.StreamSubjects = o.subjects
}

// WithStreamSubjects sets the subjects the JetStream stream will listen on.
func WithStreamSubjects(subjects ...string) Option {
	// sanitize
	set := make(map[string]struct{}, len(subjects))
	out := make([]string, 0, len(subjects))
	for _, s := range subjects {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := set[s]; ok {
			continue
		}
		set[s] = struct{}{}
		out = append(out, s)
	}
	return &streamSubjectsOption{subjects: out}
}

type streamRetentionOption struct {
	retention time.Duration
}

func (o *streamRetentionOption) apply(c *Config) {
	c.StreamRetention = o.retention
}

// WithStreamRetention sets how long to retain state events in JetStream.
func WithStreamRetention(retention time.Duration) Option {
	return &streamRetentionOption{retention: retention}
}

type enableHostManagementOption struct {
	enable bool
}

func (o *enableHostManagementOption) apply(c *Config) {
	c.EnableHostManagement = o.enable
}

// WithHostManagement enables or disables host state management.
func WithHostManagement(enable bool) Option {
	return &enableHostManagementOption{enable: enable}
}

type enableChassisManagementOption struct {
	enable bool
}

func (o *enableChassisManagementOption) apply(c *Config) {
	c.EnableChassisManagement = o.enable
}

// WithChassisManagement enables or disables chassis state management.
func WithChassisManagement(enable bool) Option {
	return &enableChassisManagementOption{enable: enable}
}

type enableBMCManagementOption struct {
	enable bool
}

func (o *enableBMCManagementOption) apply(c *Config) {
	c.EnableBMCManagement = o.enable
}

// WithBMCManagement enables or disables BMC state management.
func WithBMCManagement(enable bool) Option {
	return &enableBMCManagementOption{enable: enable}
}

type numHostsOption struct {
	num int
}

func (o *numHostsOption) apply(c *Config) {
	c.NumHosts = o.num
}

// WithNumHosts sets the number of hosts to manage.
func WithNumHosts(num int) Option {
	return &numHostsOption{num: num}
}

type numChassisOption struct {
	num int
}

func (o *numChassisOption) apply(c *Config) {
	c.NumChassis = o.num
}

// WithNumChassis sets the number of chassis to manage.
func WithNumChassis(num int) Option {
	return &numChassisOption{num: num}
}

type stateTimeoutOption struct {
	timeout time.Duration
}

func (o *stateTimeoutOption) apply(c *Config) {
	c.StateTimeout = o.timeout
}

// WithStateTimeout sets the maximum duration for state transitions.
func WithStateTimeout(timeout time.Duration) Option {
	return &stateTimeoutOption{timeout: timeout}
}

type enableMetricsOption struct {
	enable bool
}

func (o *enableMetricsOption) apply(c *Config) {
	c.EnableMetrics = o.enable
}

// WithMetrics enables or disables metrics collection.
func WithMetrics(enable bool) Option {
	return &enableMetricsOption{enable: enable}
}

type enableTracingOption struct {
	enable bool
}

func (o *enableTracingOption) apply(c *Config) {
	c.EnableTracing = o.enable
}

// WithTracing enables or disables distributed tracing.
func WithTracing(enable bool) Option {
	return &enableTracingOption{enable: enable}
}

type broadcastStateChangesOption struct {
	enable bool
}

func (o *broadcastStateChangesOption) apply(c *Config) {
	c.BroadcastStateChanges = o.enable
}

// WithBroadcastStateChanges enables or disables broadcasting state changes via NATS.
func WithBroadcastStateChanges(enable bool) Option {
	return &broadcastStateChangesOption{enable: enable}
}

type persistStateChangesOption struct {
	enable bool
}

func (o *persistStateChangesOption) apply(c *Config) {
	c.PersistStateChanges = o.enable
}

// WithPersistStateChanges enables or disables persisting state changes to JetStream.
func WithPersistStateChanges(enable bool) Option {
	return &persistStateChangesOption{enable: enable}
}

// NewConfig creates a new state manager configuration with default values.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		ServiceName:             DefaultServiceName,
		ServiceDescription:      DefaultServiceDescription,
		ServiceVersion:          DefaultServiceVersion,
		StreamName:              DefaultStreamName,
		StreamSubjects:          []string{"statemgr.state.>", "statemgr.event.>"},
		StreamRetention:         0, // Keep forever
		EnableHostManagement:    true,
		EnableChassisManagement: true,
		EnableBMCManagement:     true,
		NumHosts:                1,
		NumChassis:              1,
		StateTimeout:            30 * time.Second,
		EnableMetrics:           true,
		EnableTracing:           true,
		BroadcastStateChanges:   true,
		PersistStateChanges:     true,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.StreamRetention < 0 {
		return fmt.Errorf("stream retention cannot be negative")
	}
	if c.ServiceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if c.ServiceVersion == "" {
		return fmt.Errorf("service version cannot be empty")
	}

	if c.StreamName == "" {
		return fmt.Errorf("stream name cannot be empty")
	}

	if len(c.StreamSubjects) == 0 {
		return fmt.Errorf("at least one stream subject must be configured")
	}
	for _, s := range c.StreamSubjects {
		if len(s) == 0 {
			return fmt.Errorf("stream subject cannot be empty")
		}
	}

	if !c.EnableHostManagement && !c.EnableChassisManagement && !c.EnableBMCManagement {
		return fmt.Errorf("at least one component type must be enabled for management")
	}

	if c.EnableHostManagement && c.NumHosts <= 0 {
		return fmt.Errorf("number of hosts must be positive when host management is enabled")
	}

	if c.EnableChassisManagement && c.NumChassis <= 0 {
		return fmt.Errorf("number of chassis must be positive when chassis management is enabled")
	}

	if c.StateTimeout <= 0 {
		return fmt.Errorf("state timeout must be positive")
	}

	return nil
}
