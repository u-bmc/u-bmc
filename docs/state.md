# State Package - Simplified State Machine Wrapper

This package provides a lightweight, simplified wrapper around the [stateless](https://github.com/qmuntal/stateless) finite state machine library, specifically designed for BMC (Baseboard Management Controller) systems and other applications requiring robust state management.

## Key Improvements

### 1. Minimized Locking and State Duplication

**Before**: The package maintained its own state tracking (`currentState` field) and extensive read-write locking around state operations, duplicating functionality already provided by the stateless library.

**After**: The wrapper leverages the stateless library's built-in thread safety and state management, eliminating redundant locks and state tracking.

```go
// OLD: Manual state tracking with locks
type FSM struct {
    currentState string
    mu           sync.RWMutex
    // ... other fields
}

func (sm *FSM) CurrentState() string {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    return sm.currentState
}

// NEW: Direct delegation to underlying library
type Machine struct {
    machine *stateless.StateMachine
    // No currentState field, no locks for state access
}

func (m *Machine) State() string {
    state, err := m.machine.State(context.Background())
    if err != nil {
        return "unknown"
    }
    return fmt.Sprintf("%v", state)
}
```

### 2. Simplified Configuration API

**Before**: Complex configuration structures with nested definitions requiring multiple setup steps.

**After**: Functional options pattern with builder functions for common use cases.

```go
// OLD: Complex configuration
config := state.NewConfig(
    state.WithStates(
        state.StateDefinition{
            Name:        "off",
            Description: "System is powered off",
            OnEntry:     entryAction,
            OnExit:      exitAction,
        },
        // ... more complex definitions
    ),
    state.WithTransitions(
        state.TransitionDefinition{
            From:    "off",
            To:      "on",
            Trigger: "power_on",
            Guard:   guardFunc,
            Action:  actionFunc,
        },
        // ... more complex definitions
    ),
)
sm, err := state.New(config)

// NEW: Simplified functional options
sm, err := state.NewStateMachine(
    state.WithName("power-controller"),
    state.WithInitialState("off"),
    state.WithStates("off", "on", "transitioning"),
    state.WithTransition("off", "on", "power_on"),
    state.WithGuardedTransition("on", "off", "power_off", safeToShutdown),
    state.WithPersistence(persistCallback),
    state.WithBroadcast(broadcastCallback),
)
```

### 3. Builder Patterns for Common BMC Use Cases

The package provides pre-built state machine patterns for common BMC scenarios:

```go
// Power management
powerSM, err := state.NewPowerStateMachine("host-0")

// Thermal management
thermalSM, err := state.NewThermalStateMachine("cpu-thermal")

// Health monitoring
healthSM, err := state.NewHealthStateMachine("component-health")

// Firmware updates
firmwareSM, err := state.NewFirmwareUpdateStateMachine("bmc-firmware")

// Fluent builder for custom power machines
customPowerSM, err := state.NewBMCPowerBuilder("chassis-power").
    WithPowerOnAction(powerOnAction).
    WithPowerOffGuard(safeToShutdown).
    WithTimeout(60 * time.Second).
    Build()
```

### 4. Asynchronous Callbacks

**Before**: Synchronous callback execution that could block state transitions.

**After**: Asynchronous execution of persistence and broadcast callbacks to prevent blocking.

```go
// NEW: Non-blocking callback execution
func (m *Machine) Fire(ctx context.Context, trigger string) error {
    // ... transition logic ...

    // Handle callbacks asynchronously
    go m.handlePostTransition(ctx, previousState, currentState, trigger, span)

    return nil
}
```

### 5. Reduced Error Surface

**Before**: Many wrapper-specific errors that duplicated underlying library errors.

**After**: Focused error set with proper error wrapping from the underlying library.

```go
// Simplified error set focusing on wrapper-specific issues
var (
    ErrInvalidConfig = errors.New("invalid state machine configuration")
    ErrPersistenceFailed = errors.New("failed to persist state")
    ErrBroadcastFailed = errors.New("failed to broadcast state change")
    ErrTransitionTimeout = errors.New("state transition timeout")
    // ... focused error set
)
```

## Usage Examples

### Basic State Machine

```go
sm, err := state.NewStateMachine(
    state.WithName("example"),
    state.WithInitialState("idle"),
    state.WithStates("idle", "working", "done"),
    state.WithTransition("idle", "working", "start"),
    state.WithTransition("working", "done", "finish"),
    state.WithPersistence(func(name, state string) error {
        return saveToDatabase(name, state)
    }),
)

ctx := context.Background()
if err := sm.Start(ctx); err != nil {
    log.Fatal(err)
}

// Fire transitions - thread-safe, minimal overhead
if err := sm.Fire(ctx, "start"); err != nil {
    log.Printf("Transition failed: %v", err)
}

// Check state - no additional locking
currentState := sm.State()
```

### BMC Power Management

```go
powerSM, err := state.NewBMCPowerBuilder("host-power").
    WithPowerOnAction(func(from, to, trigger string) error {
        return executeBootSequence()
    }).
    WithPowerOffGuard(func() bool {
        return safeToShutdown()
    }).
    WithTimeout(60 * time.Second).
    WithPersistence(persistToDB).
    Build()
```

### State Machine Manager

```go
manager := state.NewManager()
manager.Add(powerSM)
manager.Add(thermalSM)

// Thread-safe operations
machines := manager.List()
if sm, err := manager.Get("power-controller"); err == nil {
    currentState := sm.State()
}
```

## Performance Benefits

1. **Reduced Memory Footprint**: Eliminated redundant state tracking and large mutex structures
2. **Lower Latency**: Direct delegation to optimized stateless library operations
3. **Better Concurrency**: Leverages stateless library's lock-free operations where possible
4. **Async Operations**: Non-blocking callbacks prevent state transition delays

## Thread Safety

Thread safety is inherited from the underlying stateless library with minimal additional overhead. The wrapper adds no additional locking for:

- State queries (`State()`, `CanFire()`, `PermittedTriggers()`)
- State transitions (`Fire()`)
- State introspection (`IsInState()`)

Only manager operations and lifecycle management use minimal locking.

## Migration Guide

### Configuration Changes

```go
// OLD
config := state.NewConfig(
    state.WithStates(state.StateDefinition{Name: "off"}),
    state.WithTransitions(state.TransitionDefinition{From: "off", To: "on", Trigger: "power_on"}),
)
sm, err := state.New(config)

// NEW
sm, err := state.NewStateMachine(
    state.WithStates("off", "on"),
    state.WithTransition("off", "on", "power_on"),
)
```

### Method Changes

```go
// OLD
currentState := sm.CurrentState()
canFire, err := sm.CanFire("trigger")

// NEW
currentState := sm.State()
canFire := sm.CanFire("trigger")  // No error return
```

### Callback Setup

```go
// OLD
sm.SetPersistenceCallback(callback)
sm.SetBroadcastCallback(callback)

// NEW
sm, err := state.NewStateMachine(
    state.WithPersistence(callback),
    state.WithBroadcast(callback),
    // ... other options
)
```

## Dependencies

- `github.com/qmuntal/stateless` - Core state machine implementation
- `go.opentelemetry.io/otel` - Optional tracing support

The wrapper is designed to be a thin abstraction that maximizes the benefits of the proven stateless library while providing BMC-specific conveniences.
