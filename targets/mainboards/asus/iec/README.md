# ASUS IPMI Expansion Card (IEC)

This directory contains the u-bmc configuration for the ASUS IPMI Expansion Card, a PCIe x1 form factor baseboard management controller based on the Aspeed AST2600 SoC.

## Hardware Overview

- **Form Factor**: PCIe x1 card
- **SoC**: Aspeed AST2600
- **Memory Limit**: 256MB (configured in main.go, card has 512MB total)

## Features

### Implemented
- **Telemetry**: OpenTelemetry-based monitoring with metrics, traces, and logs
- **State Management**: Host, chassis, and BMC state management with NATS streaming

### Hardware Capabilities (Pending Implementation)
- **KVM**: Full keyboard, video, and mouse redirection
- **SOL**: Serial Over LAN functionality
- **Fan Control**: 8 fan control outputs
- **Sensors**: 3 analog sensor inputs
- **Power Management**: PMBus control interface

## Configuration Status

### ✅ Active Services
- `telemetry`: OpenTelemetry monitoring service
- `statemgr`: State management service

### 🚧 Pending Services (Currently Commented Out)
- `sensormon`: Sensor monitoring service
- `thermalmgr`: Thermal management service
- `powermgr`: Power management service
- `websrv`: Web server with KVM/SOL support

## TODO Items

### Hardware-Specific Configuration
- [ ] Verify AST2600 hwmon device paths for sensor access
- [ ] Determine correct GPIO chip device path (`/dev/gpiochipX`)
- [ ] Map GPIO pin assignments for power control signals
- [ ] Configure PMBus interface addresses and capabilities
- [ ] Set hardware-appropriate temperature thresholds
- [ ] Tune PID controller parameters for thermal management

### Service Implementation
- [ ] Uncomment and verify import paths for disabled services
- [ ] Configure 3 analog sensors with appropriate thresholds
- [ ] Set up 8 fan control outputs with proper PWM mapping
- [ ] Implement PMBus power management integration
- [ ] Configure KVM video capture and input injection
- [ ] Set up SOL (Serial Over LAN) functionality
- [ ] Verify WebUI path and KVM/SOL web interface integration

### Network and Security
- [ ] Determine appropriate hostname and network configuration
- [ ] Configure TLS certificates for web interface
- [ ] Set up appropriate network interface bindings
- [ ] Configure authentication and authorization policies

### Testing and Validation
- [ ] Test sensor readings and threshold monitoring
- [ ] Validate fan control operation and thermal response
- [ ] Verify power management operations (power on/off/reset)
- [ ] Test KVM functionality (video, keyboard, mouse)
- [ ] Validate SOL console access
- [ ] Performance testing under load conditions

## Build Instructions

TODO: Add specific build instructions for this target

## Deployment

TODO: Add deployment instructions specific to this ASUS IPMI expansion card

## Hardware Documentation

TODO: Add links to hardware documentation, schematics, and pin mappings when available

## Support

TODO: Add support contact information and troubleshooting guidelines
