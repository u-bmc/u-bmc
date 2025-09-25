# u-bmc

u-bmc is a modern, service-oriented Baseboard Management Controller (BMC) written in Go. It emphasizes reliability, clear separation of concerns, and production-grade APIs.

- Communication: embedded NATS for inter-service IPC
- APIs: ConnectRPC with HTTP/1.1, HTTP/2, and HTTP/3 support, JSON and protobuf
- State: lightweight, supervised services with focused responsibilities
- Hardware: GPIO, I2C/SMBus/PMBus, HWMON, and platform abstractions

This repository contains the services, packages, protobuf API, and documentation for the system. The software and API are currently in alpha (v1alpha1 schema).

## Quick orientation

- System overview and architecture
  - docs/overview.md
  - docs/architecture.md
- Services and state machines
  - docs/state.md
- Hardware integration
  - docs/gpio.md
- Power/LED orchestration examples
  - docs/integration_example.md
  - docs/integration_complete.md
- Roadmap
  - docs/roadmap.md

## Architecture in brief

- Operator orchestrates service lifecycle and supervision
- Embedded NATS provides fast, reliable IPC with request/reply and pub/sub
- Web server (websrv) exposes ConnectRPC APIs, with REST transcoding
- Managers handle domains: power, thermal, sensors, inventory, users, security, updates
- Telemetry integrates tracing and metrics across services

For a deeper dive, see docs/overview.md and docs/architecture.md. Those documents include mermaid diagrams of service relationships and message flow.

## API

- Primary API: ConnectRPC served over HTTPS; supports JSON and protobuf content types
- REST endpoints via HTTP annotations in the protobuf schema
  - Stable base path: /api/v1alpha1/…
  - See schema/v1alpha1/*.proto for the authoritative definitions
- Protocol roadmap:
  - ConnectRPC (available now)
  - REST via transcoding (available now)
  - Redfish (planned)
  - IPMI (legacy compatibility; in progress)

A dedicated guide with request examples and schema pointers is available in docs/api.md.

## Platforms and targets

- Hardware targets live under targets/
  - Each real hardware target may include its own README (allowed)
- To port a new platform:
  - Follow the step-by-step guide in docs/porting.md
  - Provide platform wiring and capabilities (GPIO/I2C sensors, power rails)
  - Implement target bootstrap and service configuration
  - Add non-userspace and rootfs builder items to the TODOs as noted in the guide

Future work: a rootfs builder under dagger/ (this directory may contain its own README).

## Web UI

- A modern Web UI communicates via the same public API
- Development and packaging are tracked in docs/webui.md (with a short README allowed in webui/)

## Documentation policy

- Package documentation is in Go source via doc.go and type/function comments
  - These render on pkg.go.dev and serve as the canonical package references
  - Packages should not contain markdown files
- Project-wide and user-facing guides live under docs/ (file names are lowercase)
- Exceptions for README files:
  - Hardware targets in targets/
  - The future rootfs builder in dagger/
  - The webui/

## Contributing and licensing

- Contribution guide: CONTRIBUTING.md
- Releasing/versioning: RELEASING.md and VERSIONING.md
- License: BSD-3-Clause (see LICENSE)

## Getting started

- Read docs/overview.md to understand the components
- Pick a target under targets/ or follow docs/porting.md to add your own
- Use docs/api.md to explore and test the API
- Track upcoming features in docs/roadmap.md

Links to additional documents:
- docs/overview.md
- docs/architecture.md
- docs/state.md
- docs/gpio.md
- docs/integration_example.md
- docs/integration_complete.md
- docs/api.md
- docs/porting.md
- docs/roadmap.md
