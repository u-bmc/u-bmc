// SPDX-License-Identifier: BSD-3-Clause

package state

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// ExampleUsage demonstrates how to use the simplified state package.
func ExampleUsage() {
	// Basic state machine creation
	basicMachine, err := NewStateMachine(
		WithName("example"),
		WithInitialState("idle"),
		WithStates("idle", "working", "done", "error"),
		WithTransition("idle", "working", "start"),
		WithTransition("working", "done", "finish"),
		WithTransition("working", "error", "fail"),
		WithTransition("error", "idle", "reset"),
		WithTransition("done", "idle", "reset"),
		WithStateTimeout(10*time.Second),
		WithPersistence(func(name, state string) error {
			fmt.Printf("Persisting %s state: %s\n", name, state)
			return nil
		}),
		WithBroadcast(func(name, prev, curr, trigger string) error {
			fmt.Printf("State change: %s %s->%s via %s\n", name, prev, curr, trigger)
			return nil
		}),
	)
	if err != nil {
		slog.Error("Failed to create state machine", "error", err)
		return
	}

	ctx := context.Background()
	if err := basicMachine.Start(ctx); err != nil {
		slog.Error("Failed to start state machine", "error", err)
		return
	}

	fmt.Printf("Initial state: %s\n", basicMachine.State())
	fmt.Printf("Can start: %t\n", basicMachine.CanFire("start"))

	// Fire a transition
	if err := basicMachine.Fire(ctx, "start"); err != nil {
		slog.Error("Failed to fire transition", "error", err)
		return
	}

	fmt.Printf("After start: %s\n", basicMachine.State())
	fmt.Printf("Available triggers: %v\n", basicMachine.PermittedTriggers())

	// Power state machine using builder
	powerMachine, err := NewBMCPowerBuilder("chassis-power").
		WithPowerOnAction(func(from, to, trigger string) error {
			fmt.Printf("Power on action: %s->%s\n", from, to)
			return nil
		}).
		WithPowerOffAction(func(from, to, trigger string) error {
			fmt.Printf("Power off action: %s->%s\n", from, to)
			return nil
		}).
		WithPowerOnGuard(func() bool {
			// Check if power-on is safe
			return true
		}).
		WithTimeout(30 * time.Second).
		Build()

	if err != nil {
		slog.Error("Failed to build power state machine", "error", err)
		return
	}

	if err := powerMachine.Start(ctx); err != nil {
		slog.Error("Failed to start power machine", "error", err)
		return
	}

	fmt.Printf("Power machine state: %s\n", powerMachine.State())

	// State manager example
	manager := NewManager()
	if err := manager.Add(basicMachine); err != nil {
		slog.Error("Failed to add machine to manager", "error", err)
		return
	}
	if err := manager.Add(powerMachine); err != nil {
		slog.Error("Failed to add power machine to manager", "error", err)
		return
	}

	machines := manager.List()
	fmt.Printf("Managed machines: %v\n", machines)

	// Get specific machine
	if machine, err := manager.Get("example"); err == nil {
		fmt.Printf("Retrieved machine state: %s\n", machine.State())
	}

	// Cleanup
	if err := manager.StopAll(ctx); err != nil {
		slog.Error("Failed to stop all machines", "error", err)
	}
}

// ExampleThermalManagement shows thermal state machine usage.
func ExampleThermalManagement() {
	thermalMachine, err := NewThermalBuilder("cpu-thermal").
		WithWarningThreshold(75.0).
		WithCriticalThreshold(90.0).
		WithEmergencyAction(func(from, to, trigger string) error {
			fmt.Printf("EMERGENCY: Thermal protection activated!\n")
			// Emergency shutdown logic here
			return nil
		}).
		Build()

	if err != nil {
		slog.Error("Failed to create thermal machine", "error", err)
		return
	}

	ctx := context.Background()
	if err := thermalMachine.Start(ctx); err != nil {
		slog.Error("Failed to start thermal machine", "error", err)
		return
	}

	// Simulate temperature increase
	fmt.Printf("Initial thermal state: %s\n", thermalMachine.State())

	if err := thermalMachine.Fire(ctx, "temp_warning"); err != nil {
		slog.Error("Failed to trigger warning", "error", err)
		return
	}

	fmt.Printf("Warning state: %s\n", thermalMachine.State())

	if err := thermalMachine.Fire(ctx, "temp_critical"); err != nil {
		slog.Error("Failed to trigger critical", "error", err)
		return
	}

	fmt.Printf("Critical state: %s\n", thermalMachine.State())
}

// ExampleCustomTransitions demonstrates advanced transition features.
func ExampleCustomTransitions() {
	var readyCount int

	machine, err := NewStateMachine(
		WithName("custom"),
		WithInitialState("init"),
		WithStates("init", "ready", "running", "error"),

		// Guarded transition - only allow ready after 3 attempts
		WithGuardedTransition("init", "ready", "check_ready", func() bool {
			readyCount++
			return readyCount >= 3
		}),

		// Action transition - log when starting
		WithActionTransition("ready", "running", "start", func(from, to, trigger string) error {
			fmt.Printf("Starting operation: %s->%s\n", from, to)
			return nil
		}),

		// Complete transition with both guard and action
		WithCompleteTransition("running", "error", "fail",
			func() bool {
				// Simulate random failure
				return readyCount%2 == 0
			},
			func(from, to, trigger string) error {
				fmt.Printf("Operation failed: %s->%s\n", from, to)
				return nil
			},
		),

		WithTransition("error", "init", "reset"),
		WithTransition("running", "ready", "stop"),
	)

	if err != nil {
		slog.Error("Failed to create custom machine", "error", err)
		return
	}

	ctx := context.Background()
	if err := machine.Start(ctx); err != nil {
		slog.Error("Failed to start machine", "error", err)
		return
	}

	// Try to go ready multiple times
	for i := 0; i < 5; i++ {
		fmt.Printf("Attempt %d: state=%s\n", i+1, machine.State())
		if machine.CanFire("check_ready") {
			if err := machine.Fire(ctx, "check_ready"); err != nil {
				fmt.Printf("Check ready failed: %v\n", err)
			}
		}
		if machine.State() == "ready" {
			break
		}
	}

	fmt.Printf("Final state: %s\n", machine.State())
}
