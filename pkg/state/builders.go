// SPDX-License-Identifier: BSD-3-Clause

package state

import (
	"fmt"
	"time"
)

// NewStateMachine creates a basic state machine with the provided configuration.
func NewStateMachine(opts ...Option) (*Machine, error) {
	config := NewConfig(opts...)
	return New(config)
}

// NewPowerStateMachine creates a state machine for power management with common power states.
func NewPowerStateMachine(name string, opts ...Option) (*Machine, error) {
	baseOpts := []Option{
		WithName(name),
		WithDescription("Power management state machine"),
		WithInitialState("off"),
		WithStates("off", "on", "transitioning"),
		WithTransition("off", "transitioning", "power_on"),
		WithTransition("on", "transitioning", "power_off"),
		WithTransition("on", "transitioning", "reboot"),
		WithTransition("transitioning", "on", "transition_complete_on"),
		WithTransition("transitioning", "off", "transition_complete_off"),
		WithStateTimeout(30 * time.Second),
	}

	// Merge with provided options (provided options take precedence)
	allOpts := append(baseOpts, opts...)
	return NewStateMachine(allOpts...)
}

// NewThermalStateMachine creates a state machine for thermal management.
func NewThermalStateMachine(name string, opts ...Option) (*Machine, error) {
	baseOpts := []Option{
		WithName(name),
		WithDescription("Thermal management state machine"),
		WithInitialState("normal"),
		WithStates("normal", "warning", "critical", "emergency"),
		WithTransition("normal", "warning", "temp_warning"),
		WithTransition("warning", "normal", "temp_normal"),
		WithTransition("warning", "critical", "temp_critical"),
		WithTransition("critical", "warning", "temp_warning"),
		WithTransition("critical", "emergency", "temp_emergency"),
		WithTransition("emergency", "critical", "temp_recovered"),
		WithStateTimeout(10 * time.Second),
	}

	allOpts := append(baseOpts, opts...)
	return NewStateMachine(allOpts...)
}

// NewBootStateMachine creates a state machine for boot sequence management.
func NewBootStateMachine(name string, opts ...Option) (*Machine, error) {
	baseOpts := []Option{
		WithName(name),
		WithDescription("Boot sequence state machine"),
		WithInitialState("init"),
		WithStates("init", "post", "boot", "ready", "error"),
		WithTransition("init", "post", "start_post"),
		WithTransition("post", "boot", "post_complete"),
		WithTransition("boot", "ready", "boot_complete"),
		WithTransition("post", "error", "post_failed"),
		WithTransition("boot", "error", "boot_failed"),
		WithTransition("error", "init", "reset"),
		WithStateTimeout(60 * time.Second),
	}

	allOpts := append(baseOpts, opts...)
	return NewStateMachine(allOpts...)
}

// NewHealthStateMachine creates a state machine for component health monitoring.
func NewHealthStateMachine(name string, opts ...Option) (*Machine, error) {
	baseOpts := []Option{
		WithName(name),
		WithDescription("Component health state machine"),
		WithInitialState("unknown"),
		WithStates("unknown", "healthy", "degraded", "failed", "maintenance"),
		WithTransition("unknown", "healthy", "health_check_pass"),
		WithTransition("unknown", "degraded", "health_check_degraded"),
		WithTransition("unknown", "failed", "health_check_fail"),
		WithTransition("healthy", "degraded", "health_degraded"),
		WithTransition("healthy", "failed", "health_failed"),
		WithTransition("degraded", "healthy", "health_recovered"),
		WithTransition("degraded", "failed", "health_failed"),
		WithTransition("failed", "maintenance", "enter_maintenance"),
		WithTransition("maintenance", "unknown", "exit_maintenance"),
		WithStateTimeout(15 * time.Second),
	}

	allOpts := append(baseOpts, opts...)
	return NewStateMachine(allOpts...)
}

// NewFirmwareUpdateStateMachine creates a state machine for firmware update processes.
func NewFirmwareUpdateStateMachine(name string, opts ...Option) (*Machine, error) {
	baseOpts := []Option{
		WithName(name),
		WithDescription("Firmware update state machine"),
		WithInitialState("idle"),
		WithStates("idle", "downloading", "validating", "updating", "verifying", "complete", "failed"),
		WithTransition("idle", "downloading", "start_download"),
		WithTransition("downloading", "validating", "download_complete"),
		WithTransition("downloading", "failed", "download_failed"),
		WithTransition("validating", "updating", "validation_complete"),
		WithTransition("validating", "failed", "validation_failed"),
		WithTransition("updating", "verifying", "update_complete"),
		WithTransition("updating", "failed", "update_failed"),
		WithTransition("verifying", "complete", "verification_complete"),
		WithTransition("verifying", "failed", "verification_failed"),
		WithTransition("failed", "idle", "reset"),
		WithTransition("complete", "idle", "reset"),
		WithStateTimeout(300 * time.Second), // Longer timeout for firmware operations
	}

	allOpts := append(baseOpts, opts...)
	return NewStateMachine(allOpts...)
}

// NewNetworkInterfaceStateMachine creates a state machine for network interface management.
func NewNetworkInterfaceStateMachine(name string, opts ...Option) (*Machine, error) {
	baseOpts := []Option{
		WithName(name),
		WithDescription("Network interface state machine"),
		WithInitialState("down"),
		WithStates("down", "configuring", "up", "error"),
		WithTransition("down", "configuring", "configure"),
		WithTransition("configuring", "up", "config_complete"),
		WithTransition("configuring", "error", "config_failed"),
		WithTransition("up", "down", "shutdown"),
		WithTransition("up", "configuring", "reconfigure"),
		WithTransition("error", "down", "reset"),
		WithStateTimeout(20 * time.Second),
	}

	allOpts := append(baseOpts, opts...)
	return NewStateMachine(allOpts...)
}

// BMCPowerBuilder provides a fluent interface for building BMC power state machines.
type BMCPowerBuilder struct {
	name        string
	opts        []Option
	onPowerOn   ActionFunc
	onPowerOff  ActionFunc
	canPowerOn  GuardFunc
	canPowerOff GuardFunc
}

// NewBMCPowerBuilder creates a new BMC power state machine builder.
func NewBMCPowerBuilder(name string) *BMCPowerBuilder {
	return &BMCPowerBuilder{
		name: name,
		opts: []Option{},
	}
}

// WithPowerOnAction sets the action to execute when powering on.
func (b *BMCPowerBuilder) WithPowerOnAction(action ActionFunc) *BMCPowerBuilder {
	b.onPowerOn = action
	return b
}

// WithPowerOffAction sets the action to execute when powering off.
func (b *BMCPowerBuilder) WithPowerOffAction(action ActionFunc) *BMCPowerBuilder {
	b.onPowerOff = action
	return b
}

// WithPowerOnGuard sets a guard condition for power on transitions.
func (b *BMCPowerBuilder) WithPowerOnGuard(guard GuardFunc) *BMCPowerBuilder {
	b.canPowerOn = guard
	return b
}

// WithPowerOffGuard sets a guard condition for power off transitions.
func (b *BMCPowerBuilder) WithPowerOffGuard(guard GuardFunc) *BMCPowerBuilder {
	b.canPowerOff = guard
	return b
}

// WithPersistence adds persistence callback to the state machine.
func (b *BMCPowerBuilder) WithPersistence(callback PersistenceCallback) *BMCPowerBuilder {
	b.opts = append(b.opts, WithPersistence(callback))
	return b
}

// WithBroadcast adds broadcast callback to the state machine.
func (b *BMCPowerBuilder) WithBroadcast(callback BroadcastCallback) *BMCPowerBuilder {
	b.opts = append(b.opts, WithBroadcast(callback))
	return b
}

// WithTimeout sets the state transition timeout.
func (b *BMCPowerBuilder) WithTimeout(timeout time.Duration) *BMCPowerBuilder {
	b.opts = append(b.opts, WithStateTimeout(timeout))
	return b
}

// Build creates the configured BMC power state machine.
func (b *BMCPowerBuilder) Build() (*Machine, error) {
	opts := []Option{
		WithName(b.name),
		WithDescription(fmt.Sprintf("BMC power management for %s", b.name)),
		WithInitialState("off"),
		WithStates("off", "on", "transitioning"),
	}

	// Add transitions with optional guards and actions
	if b.canPowerOn != nil {
		opts = append(opts, WithGuardedTransition("off", "transitioning", "power_on", b.canPowerOn))
	} else {
		opts = append(opts, WithTransition("off", "transitioning", "power_on"))
	}

	if b.canPowerOff != nil {
		opts = append(opts, WithGuardedTransition("on", "transitioning", "power_off", b.canPowerOff))
	} else {
		opts = append(opts, WithTransition("on", "transitioning", "power_off"))
	}

	// Add reboot transition
	opts = append(opts, WithTransition("on", "transitioning", "reboot"))

	// Add completion transitions
	if b.onPowerOn != nil {
		opts = append(opts, WithActionTransition("transitioning", "on", "transition_complete_on", b.onPowerOn))
	} else {
		opts = append(opts, WithTransition("transitioning", "on", "transition_complete_on"))
	}

	if b.onPowerOff != nil {
		opts = append(opts, WithActionTransition("transitioning", "off", "transition_complete_off", b.onPowerOff))
	} else {
		opts = append(opts, WithTransition("transitioning", "off", "transition_complete_off"))
	}

	// Add builder-specific options
	opts = append(opts, b.opts...)

	return NewStateMachine(opts...)
}

// ThermalBuilder provides a fluent interface for building thermal management state machines.
type ThermalBuilder struct {
	name              string
	opts              []Option
	warningThreshold  float64
	criticalThreshold float64
	emergencyAction   ActionFunc
}

// NewThermalBuilder creates a new thermal state machine builder.
func NewThermalBuilder(name string) *ThermalBuilder {
	return &ThermalBuilder{
		name:              name,
		opts:              []Option{},
		warningThreshold:  70.0, // Default warning at 70°C
		criticalThreshold: 85.0, // Default critical at 85°C
	}
}

// WithWarningThreshold sets the temperature threshold for warning state.
func (b *ThermalBuilder) WithWarningThreshold(temp float64) *ThermalBuilder {
	b.warningThreshold = temp
	return b
}

// WithCriticalThreshold sets the temperature threshold for critical state.
func (b *ThermalBuilder) WithCriticalThreshold(temp float64) *ThermalBuilder {
	b.criticalThreshold = temp
	return b
}

// WithEmergencyAction sets the action to execute in emergency thermal conditions.
func (b *ThermalBuilder) WithEmergencyAction(action ActionFunc) *ThermalBuilder {
	b.emergencyAction = action
	return b
}

// WithPersistence adds persistence callback to the state machine.
func (b *ThermalBuilder) WithPersistence(callback PersistenceCallback) *ThermalBuilder {
	b.opts = append(b.opts, WithPersistence(callback))
	return b
}

// WithBroadcast adds broadcast callback to the state machine.
func (b *ThermalBuilder) WithBroadcast(callback BroadcastCallback) *ThermalBuilder {
	b.opts = append(b.opts, WithBroadcast(callback))
	return b
}

// Build creates the configured thermal management state machine.
func (b *ThermalBuilder) Build() (*Machine, error) {
	opts := []Option{
		WithName(b.name),
		WithDescription(fmt.Sprintf("Thermal management for %s", b.name)),
		WithInitialState("normal"),
		WithStates("normal", "warning", "critical", "emergency"),
		WithTransition("normal", "warning", "temp_warning"),
		WithTransition("warning", "normal", "temp_normal"),
		WithTransition("warning", "critical", "temp_critical"),
		WithTransition("critical", "warning", "temp_warning"),
		WithTransition("critical", "emergency", "temp_emergency"),
		WithTransition("emergency", "critical", "temp_recovered"),
		WithStateTimeout(10 * time.Second),
	}

	// Add emergency action if specified
	if b.emergencyAction != nil {
		opts = append(opts, WithActionTransition("critical", "emergency", "temp_emergency", b.emergencyAction))
	}

	// Add builder-specific options
	opts = append(opts, b.opts...)

	return NewStateMachine(opts...)
}
