# PowerMgr and StateMgr Integration Example

This document demonstrates how the `powermgr` and `statemgr` services work together to provide coordinated power and state management for BMC systems.

## Architecture Overview

```
API Client
    │
    ▼
┌─────────────┐    NATS IPC     ┌─────────────┐
│  StateMgr   │ ◄────────────► │  PowerMgr   │
│             │                │             │
│ - Validates │                │ - Executes  │
│   states    │                │   physical  │
│ - Manages   │                │   power ops │
│   FSMs      │                │ - GPIO/HW   │
│ - Persists  │                │   control   │
│   events    │                │             │
└─────────────┘                └─────────────┘
```

## Message Flow Example: Host Power On

### 1. API Request
Client sends a power on request to StateMgr:
```
POST /api/v1/hosts/0/actions
{
  "action": "HOST_ACTION_ON"
}
```

### 2. StateMgr Processing
StateMgr receives the request and:
- Validates current state (OFF → ON transition allowed)
- Updates FSM to TRANSITIONING state
- Sends action request to PowerMgr via NATS

### 3. NATS IPC Message
StateMgr → PowerMgr:
```
Subject: powermgr.host.0.action
Message: ChangeHostStateRequest {
  host_name: "host.0"
  action: HOST_ACTION_ON
}
```

### 4. PowerMgr Execution
PowerMgr receives the message and:
- Validates the action
- Executes physical power operation via GPIO backend
- Responds with success/failure

### 5. NATS IPC Response
PowerMgr → StateMgr:
```
Response: ChangeHostStateResponse {
  current_status: HOST_STATUS_TRANSITIONING
}
```

### 6. State Completion
StateMgr receives the response and:
- Updates FSM state to ON (on success) or ERROR (on failure)
- Persists state change event
- Broadcasts state change notification
- Responds to API client

## Configuration Example

### StateMgr Configuration
```go
statemgr := statemgr.New(
    statemgr.WithHostManagement(true),
    statemgr.WithNumHosts(2),
    statemgr.WithBroadcastStateChanges(true),
    statemgr.WithPersistStateChanges(true),
)
```

### PowerMgr Configuration
```go
powermgr := powermgr.New(
    powermgr.WithHostManagement(true),
    powermgr.WithNumHosts(2),
    powermgr.WithGPIOChip("/dev/gpiochip0"),
    powermgr.WithComponents(map[string]powermgr.ComponentConfig{
        "host.0": {
            Name: "host.0",
            Type: "host",
            Enabled: true,
            GPIO: powermgr.GPIOConfig{
                PowerButton: powermgr.GPIOLineConfig{
                    Line: "power-button-0",
                    Direction: gpio.DirectionOutput,
                    ActiveState: gpio.ActiveLow,
                },
                ResetButton: powermgr.GPIOLineConfig{
                    Line: "reset-button-0",
                    Direction: gpio.DirectionOutput,
                    ActiveState: gpio.ActiveLow,
                },
                PowerStatus: powermgr.GPIOLineConfig{
                    Line: "power-good-0",
                    Direction: gpio.DirectionInput,
                    ActiveState: gpio.ActiveHigh,
                },
            },
            PowerOnDelay: 200 * time.Millisecond,
            PowerOffDelay: 200 * time.Millisecond,
            ResetDelay: 100 * time.Millisecond,
            ForceOffDelay: 4 * time.Second,
        },
        "host.1": {
            // Similar configuration for host.1
        },
    }),
)
```

## Complete Action Mappings

### Host Actions
| HostAction | PowerMgr Operation | Physical Action |
|------------|-------------------|-----------------|
| `HOST_ACTION_ON` | `PowerOn()` | 200ms power button press |
| `HOST_ACTION_OFF` | `PowerOff(force=false)` | 200ms power button press |
| `HOST_ACTION_FORCE_OFF` | `PowerOff(force=true)` | 4s power button hold |
| `HOST_ACTION_REBOOT` | `Reset()` | 100ms reset button press |
| `HOST_ACTION_FORCE_RESTART` | `Reset()` | 100ms reset button press |

### Chassis Actions
| ChassisAction | PowerMgr Operation | Physical Action |
|---------------|-------------------|-----------------|
| `CHASSIS_ACTION_ON` | `PowerOn()` | Chassis power enable |
| `CHASSIS_ACTION_OFF` | `PowerOff(force=false)` | Graceful chassis shutdown |
| `CHASSIS_ACTION_EMERGENCY_SHUTDOWN` | `PowerOff(force=true)` | Immediate power cut |
| `CHASSIS_ACTION_POWER_CYCLE` | `PowerOff()` + `PowerOn()` | Off, wait 2s, on |

### Management Controller Actions
| ManagementControllerAction | PowerMgr Operation | Physical Action |
|---------------------------|-------------------|-----------------|
| `MANAGEMENT_CONTROLLER_ACTION_REBOOT` | `Reset()` | BMC reset |
| `MANAGEMENT_CONTROLLER_ACTION_WARM_RESET` | `Reset()` | BMC warm reset |
| `MANAGEMENT_CONTROLLER_ACTION_COLD_RESET` | `Reset()` | BMC cold reset |
| `MANAGEMENT_CONTROLLER_ACTION_HARD_RESET` | `Reset()` | BMC hard reset |
| `MANAGEMENT_CONTROLLER_ACTION_FACTORY_RESET` | `Reset()` | BMC factory reset |

## Error Handling

### PowerMgr Errors
If PowerMgr fails to execute a physical operation:
```
Response: Error {
  code: POWER_OPERATION_FAILED
  message: "GPIO operation failed: permission denied"
}
```

### StateMgr Response
StateMgr handles the error by:
- Transitioning FSM to ERROR state
- Logging the failure
- Responding to API client with error details

### Recovery
- System administrators can diagnose hardware issues
- Components can be reset or reconfigured
- State machines can be manually transitioned to valid states

## Monitoring and Observability

### State Events
StateMgr publishes state change events:
```
Subject: statemgr.event.host.0.transition
Message: HostStateChange {
  host_name: "host.0"
  previous_status: HOST_STATUS_OFF
  current_status: HOST_STATUS_ON
  cause: HOST_ACTION_ON
  changed_at: "2023-12-07T10:30:00Z"
}
```

### Power Events
PowerMgr logs power operations:
```
INFO Host power action completed component=host.0 action=HOST_ACTION_ON
```

### Metrics
Both services expose metrics for:
- Operation counts and durations
- Success/failure rates
- State transition frequencies
- Hardware operation statistics

## Service Coordination Benefits

1. **Separation of Concerns**: StateMgr handles state logic, PowerMgr handles hardware
2. **Reliability**: Physical operations are isolated from state management
3. **Testability**: Services can be tested independently
4. **Extensibility**: Different power backends without changing state logic
5. **Observability**: Clear operation boundaries and event tracking
6. **Error Isolation**: Hardware failures don't corrupt state machines

## Development Workflow

### Adding New Component Types
1. Define protobuf messages for new component actions
2. Add state machines to StateMgr
3. Add power backend support to PowerMgr
4. Configure NATS IPC endpoints
5. Test integration end-to-end

### Custom Power Backends
1. Implement `PowerBackend` interface
2. Configure PowerMgr to use custom backend
3. StateMgr integration remains unchanged
4. Physical operations execute via custom backend

This architecture provides a robust, scalable foundation for BMC power and state management with clear service boundaries and comprehensive error handling.
