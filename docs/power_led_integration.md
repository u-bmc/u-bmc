# Power and LED Management Integration

This document describes the protobuf message definitions and integration points for the power management and LED control system.

For an implementation summary, see integration_complete.md.

## Protobuf messages

### ✅ LED Control Messages - IMPLEMENTED

The following protobuf messages have been defined in `schema/v1alpha1/led.proto`:

#### LEDControlRequest
```protobuf
message LEDControlRequest {
  string component_name = 1;
  LEDType led_type = 2;
  LEDState led_state = 3;
  optional uint32 brightness = 4;
  optional uint32 blink_interval_ms = 5;
}

enum LEDType {
  LED_TYPE_UNSPECIFIED = 0;
  LED_TYPE_POWER = 1;
  LED_TYPE_STATUS = 2;
  LED_TYPE_ERROR = 3;
  LED_TYPE_IDENTIFY = 4;
}

enum LEDState {
  LED_STATE_UNSPECIFIED = 0;
  LED_STATE_OFF = 1;
  LED_STATE_ON = 2;
  LED_STATE_BLINK = 3;
  LED_STATE_FAST_BLINK = 4;
}
```

#### LEDControlResponse
```protobuf
message LEDControlResponse {
  LEDState current_state = 1;
  string message = 2;
}
```

#### LEDStatusRequest
```protobuf
message LEDStatusRequest {
  string component_name = 1;
  LEDType led_type = 2;
}
```

#### LEDStatusResponse
```protobuf
message LEDStatusResponse {
  LEDState current_state = 1;
  optional uint32 brightness = 2;
  bool is_blinking = 3;
  uint32 blink_interval_ms = 4;
}
```

### ✅ Power Operation Result Messages - IMPLEMENTED

#### PowerOperationResult
```protobuf
message PowerOperationResult {
  string component_name = 1;
  string operation = 2;
  bool success = 3;
  string error_message = 4;
  google.protobuf.Timestamp completed_at = 5;
  uint32 duration_ms = 6;
}
```

#### StateTransitionNotification
```protobuf
message StateTransitionNotification {
  string component_name = 1;
  string component_type = 2;
  string previous_state = 3;
  string current_state = 4;
  string trigger = 5;
  bool success = 6;
  google.protobuf.Timestamp changed_at = 7;
  uint32 transition_duration_ms = 8;
}
```

## ✅ Integration Points to Implement - ALL COMPLETED

### ✅ 1. PowerMgr -> StateMgr Communication - COMPLETED

#### Current Status: FULLY IMPLEMENTED
- Power operations send proper protobuf `PowerOperationResult` messages
- NATS message publishing implemented with full error handling

#### Completed Changes:
- ✅ `powermgr.go`: `reportStateChange()` function sends `PowerOperationResult` messages
- ✅ Subject pattern: `{stateReportingSubjectPrefix}.{component}.power.result`
- ✅ Messages include operation type, success/failure, timing, and state information

### ✅ 2. StateMgr -> LEDMgr Communication - COMPLETED

#### Current Status: FULLY IMPLEMENTED
- State machine callbacks send proper protobuf `LEDControlRequest` messages
- NATS message publishing implemented with comprehensive LED state mapping

#### Completed Changes:
- ✅ `statemgr.go`: `requestLEDAction()` function sends `LEDControlRequest` messages
- ✅ Subject pattern: `{ledControlSubjectPrefix}.{component}.{led_type}.control`
- ✅ Complete state transition to LED mapping implemented:
  - Power ON -> Power LED solid on
  - Power OFF -> Power LED off
  - Error states -> Error LED blinking/fast blinking
  - Warning states -> Status LED blinking

### ✅ 3. LEDMgr Request/Response Handling - COMPLETED

#### Current Status: FULLY IMPLEMENTED
- Endpoints process proper protobuf messages with validation
- Complete request/response cycle implemented with error handling

#### Completed Changes:
- ✅ `ledmgr.go`: `handleLEDControl()` and `handleLEDStatus()` use protobuf messages
- ✅ Full `LEDControlRequest` parsing with component validation
- ✅ `LEDControlResponse` and `LEDStatusResponse` with current LED state
- ✅ Comprehensive error handling with protobuf error responses

### ✅ 4. State Machine Action Functions - COMPLETED

#### Current Status: FULLY IMPLEMENTED
- State machine actions handle power operation results via NATS subscriptions
- Bidirectional communication implemented with proper state coordination

#### Completed Changes:
- ✅ Power operation result handling via NATS subscriptions in `statemgr.go`
- ✅ State machine triggers fired based on power operation completion/failure
- ✅ Error state transitions implemented for power operation failures
- ✅ State transition notifications published for downstream consumers

## ✅ Configuration Enhancements - COMPLETED

### ✅ 1. Subject Prefix Configuration - IMPLEMENTED

Configuration options for NATS subject prefixes are already implemented:

```go
// powermgr config
stateReportingSubjectPrefix string

// statemgr config
powerControlSubjectPrefix string
ledControlSubjectPrefix   string
```

### ✅ 2. Component-LED Mapping - IMPLEMENTED

Component-specific LED mappings implemented via action parsing in `statemgr.go`:

```go
// parseLEDAction() function handles:
// - Power states (power_on/power_off)
// - Error states (error/critical_error/failed)
// - Status states (warning/status_on/status_off)
// - Identify states (identify_on/identify_off)
```

## ✅ Testing Requirements - READY FOR IMPLEMENTATION

### 1. Integration Tests - Ready
- ✅ Complete message flow implemented: API -> StateMgr -> PowerMgr -> LED feedback
- ✅ Power-off flow with LED state changes ready for testing
- ✅ Error handling and recovery scenarios implemented
- ✅ State persistence and recovery mechanisms in place

### 2. Unit Tests - Ready
- ✅ Protobuf message serialization/deserialization implemented
- ✅ State machine transitions with backend integration ready
- ✅ NATS message publishing and subscription implemented
- ✅ Backend switching (GPIO vs I2C) preserved

### 3. Hardware-in-the-Loop Tests - Ready
- ✅ GPIO control implementation ready for hardware testing
- ✅ I2C communication with LED controllers implemented
- ✅ Power control timing and sequencing preserved
- ✅ Concurrent operations on multiple components supported

## Performance Optimizations - FUTURE ENHANCEMENTS

### 1. Message Batching - Future Enhancement

For high-frequency LED updates, consider batching multiple LED control requests:

```protobuf
message LEDControlBatchRequest {
  repeated LEDControlRequest requests = 1;
}
```

### 2. State Caching - Future Enhancement

Implement local state caching to reduce IPC overhead:
- Cache current LED states in LEDMgr
- Cache component states in StateMgr
- Use TTL-based invalidation for consistency

### 3. Async Operation Handling - Future Enhancement

Implement asynchronous operation handling for better performance:
- Use goroutines for parallel LED updates
- Implement operation queuing for sequential power operations
- Add circuit breaker pattern for failing backends

## Security Considerations - CURRENT STATUS

### 1. Access Control - Implemented

- ✅ Component-level validation in all handlers
- ✅ Operation-level message type validation
- ✅ Secure I2C and GPIO device access preserved

### 2. Input Validation - Implemented

- ✅ Protobuf message field validation with buf/validate
- ✅ Component name validation and sanitization
- ✅ Request validation in all service handlers

## ✅ Documentation Updates - COMPLETED

### ✅ 1. API Documentation - Completed

- ✅ All protobuf messages documented in `schema/v1alpha1/led.proto`
- ✅ Integration examples provided in `INTEGRATION_COMPLETE.md`
- ✅ Error handling patterns documented in service implementations

### 2. Deployment Guide - Future Enhancement

- Hardware wiring diagrams for GPIO connections
- I2C device configuration examples
- Service startup sequence and dependencies

## ✅ Implementation Priority - ALL HIGH PRIORITY ITEMS COMPLETED

1. **✅ COMPLETED**: Protobuf message definitions and basic request/response handling
2. **✅ COMPLETED**: PowerMgr -> StateMgr state feedback communication
3. **✅ COMPLETED**: StateMgr -> LEDMgr LED control communication
4. **✅ COMPLETED**: Error handling and recovery mechanisms
5. **✅ COMPLETED**: Configuration enhancements and flexibility
6. **FUTURE**: Performance optimizations and batching
7. **READY**: Hardware-in-the-loop testing and validation

## ✅ Actual Implementation Timeline

- **Completed**: Protobuf messages and IPC communication
- **Completed**: State machine integration and power coordination
- **Completed**: LED control integration and visual feedback
- **Completed**: Error handling, validation, and core documentation

**Total actual effort: 1 session for complete core integration**
**Status: ✅ Ready for testing and deployment**
