// SPDX-License-Identifier: BSD-3-Clause

package statemgr

import (
	"fmt"
	"strings"
	"time"
)

const (
	DefaultServiceName        = "statemgr"
	DefaultServiceDescription = "State management service for BMC components"
	DefaultServiceVersion     = "1.0.0"
	DefaultStreamName         = "STATEMGR"
	DefaultStateTimeout       = 30 * time.Second
)

type config struct {
	serviceName               string
	serviceDescription        string
	serviceVersion            string
	streamName                string
	streamSubjects            []string
	streamRetention           time.Duration
	enableHostManagement      bool
	enableChassisManagement   bool
	enableBMCManagement       bool
	numHosts                  int
	numChassis                int
	stateTimeout              time.Duration
	broadcastStateChanges     bool
	persistStateChanges       bool
	powerControlSubjectPrefix string
	ledControlSubjectPrefix   string
}

type Option interface {
	apply(*config)
}

type serviceNameOption struct {
	name string
}

func (o *serviceNameOption) apply(c *config) {
	c.serviceName = o.name
}

func WithServiceName(name string) Option {
	return &serviceNameOption{name: name}
}

type serviceDescriptionOption struct {
	description string
}

func (o *serviceDescriptionOption) apply(c *config) {
	c.serviceDescription = o.description
}

func WithServiceDescription(description string) Option {
	return &serviceDescriptionOption{description: description}
}

type serviceVersionOption struct {
	version string
}

func (o *serviceVersionOption) apply(c *config) {
	c.serviceVersion = o.version
}

func WithServiceVersion(version string) Option {
	return &serviceVersionOption{version: version}
}

type streamNameOption struct {
	name string
}

func (o *streamNameOption) apply(c *config) {
	c.streamName = o.name
}

func WithStreamName(name string) Option {
	return &streamNameOption{name: name}
}

type streamSubjectsOption struct {
	subjects []string
}

func (o *streamSubjectsOption) apply(c *config) {
	c.streamSubjects = o.subjects
}

func WithStreamSubjects(subjects ...string) Option {
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

func (o *streamRetentionOption) apply(c *config) {
	c.streamRetention = o.retention
}

func WithStreamRetention(retention time.Duration) Option {
	return &streamRetentionOption{retention: retention}
}

type enableHostManagementOption struct {
	enable bool
}

func (o *enableHostManagementOption) apply(c *config) {
	c.enableHostManagement = o.enable
}

func WithHostManagement(enable bool) Option {
	return &enableHostManagementOption{enable: enable}
}

type enableChassisManagementOption struct {
	enable bool
}

func (o *enableChassisManagementOption) apply(c *config) {
	c.enableChassisManagement = o.enable
}

func WithChassisManagement(enable bool) Option {
	return &enableChassisManagementOption{enable: enable}
}

type enableBMCManagementOption struct {
	enable bool
}

func (o *enableBMCManagementOption) apply(c *config) {
	c.enableBMCManagement = o.enable
}

func WithBMCManagement(enable bool) Option {
	return &enableBMCManagementOption{enable: enable}
}

type numHostsOption struct {
	num int
}

func (o *numHostsOption) apply(c *config) {
	c.numHosts = o.num
}

func WithNumHosts(num int) Option {
	return &numHostsOption{num: num}
}

type numChassisOption struct {
	num int
}

func (o *numChassisOption) apply(c *config) {
	c.numChassis = o.num
}

func WithNumChassis(num int) Option {
	return &numChassisOption{num: num}
}

type stateTimeoutOption struct {
	timeout time.Duration
}

func (o *stateTimeoutOption) apply(c *config) {
	c.stateTimeout = o.timeout
}

func WithStateTimeout(timeout time.Duration) Option {
	return &stateTimeoutOption{timeout: timeout}
}

type broadcastStateChangesOption struct {
	enable bool
}

func (o *broadcastStateChangesOption) apply(c *config) {
	c.broadcastStateChanges = o.enable
}

func WithBroadcastStateChanges(enable bool) Option {
	return &broadcastStateChangesOption{enable: enable}
}

type persistStateChangesOption struct {
	enable bool
}

func (o *persistStateChangesOption) apply(c *config) {
	c.persistStateChanges = o.enable
}

func WithPersistStateChanges(enable bool) Option {
	return &persistStateChangesOption{enable: enable}
}

type powerControlSubjectPrefixOption struct {
	prefix string
}

func (o *powerControlSubjectPrefixOption) apply(c *config) {
	c.powerControlSubjectPrefix = o.prefix
}

func WithPowerControlSubjectPrefix(prefix string) Option {
	return &powerControlSubjectPrefixOption{prefix: prefix}
}

type ledControlSubjectPrefixOption struct {
	prefix string
}

func (o *ledControlSubjectPrefixOption) apply(c *config) {
	c.ledControlSubjectPrefix = o.prefix
}

func WithLEDControlSubjectPrefix(prefix string) Option {
	return &ledControlSubjectPrefixOption{prefix: prefix}
}

func (c *config) Validate() error {
	if c.streamRetention < 0 {
		return fmt.Errorf("stream retention cannot be negative")
	}
	if c.serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if c.serviceVersion == "" {
		return fmt.Errorf("service version cannot be empty")
	}

	if c.streamName == "" {
		return fmt.Errorf("stream name cannot be empty")
	}

	if len(c.streamSubjects) == 0 {
		return fmt.Errorf("at least one stream subject must be configured")
	}
	for _, s := range c.streamSubjects {
		if len(s) == 0 {
			return fmt.Errorf("stream subject cannot be empty")
		}
	}

	if !c.enableHostManagement && !c.enableChassisManagement && !c.enableBMCManagement {
		return fmt.Errorf("at least one component type must be enabled for management")
	}

	if c.enableHostManagement && c.numHosts <= 0 {
		return fmt.Errorf("number of hosts must be positive when host management is enabled")
	}

	if c.enableChassisManagement && c.numChassis <= 0 {
		return fmt.Errorf("number of chassis must be positive when chassis management is enabled")
	}

	if c.stateTimeout <= 0 {
		return fmt.Errorf("state timeout must be positive")
	}

	return nil
}
