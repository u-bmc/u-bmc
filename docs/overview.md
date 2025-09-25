# u-bmc overview

u-bmc is a service-oriented Baseboard Management Controller (BMC) written in Go. Services are supervised by an operator, communicate over an embedded NATS instance, and expose a public API via ConnectRPC with REST transcoding. Package documentation lives in `doc.go` files and renders on pkg.go.dev; project-wide docs live in this `docs/` directory.

## System at a glance

```mermaid
graph TB
  Client[Clients<br/>Web UI, CLI, Tools]
  Web[websrv<br/>HTTPS + ConnectRPC/REST]
  IPC[ipc<br/>Embedded NATS (+JetStream)]
  Op[operator<br/>Supervision & lifecycle]

  SM[statemgr]
  PM[powermgr]
  TM[thermalmgr]
  SN[sensormon]
  LED[ledmgr]
  INV[inventorymgr]
  UM[usermgr]
  SEC[securitymgr]
  UPD[updatemgr]
  TEL[telemetry]

  IPMI[ipmisrv (compat)]
  KVM[kvmsrv (planned)]
  CON[consolesrv (planned)]

  HW[Hardware<br/>GPIO, I2C/PMBus, HWMON]

  Client --> Web
  Web --> IPC
  IPMI --> IPC
  KVM --> IPC
  CON --> IPC

  Op --> IPC
  Op --> SM
  Op --> PM
  Op --> TM
  Op --> SN
  Op --> LED
  Op --> INV
  Op --> UM
  Op --> SEC
  Op --> UPD
  Op --> TEL
  Op --> Web
  Op --> IPMI
  Op --> KVM
  Op --> CON

  SM --> IPC
  PM --> IPC
  TM --> IPC
  SN --> IPC
  LED --> IPC
  INV --> IPC
  UM --> IPC
  SEC --> IPC
  UPD --> IPC
  TEL --> IPC

  PM --> HW
  TM --> HW
  SN --> HW
  LED --> HW
```

## Components

- operator: starts services in order, supervises restarts, and coordinates graceful shutdown.
- ipc: embedded NATS for low-latency request/reply and pub/sub; JetStream may be used for durability.
- websrv: HTTPS entrypoint serving ConnectRPC; REST is provided via protobuf HTTP annotations.
- statemgr: system state machines and transitions, with LED and power action mapping.
- powermgr: host/chassis/BMC power sequencing via GPIO and related backends.
- thermalmgr: fan control and thermal protection (PID profiles).
- sensormon: sensor discovery and sampling (HWMON, I2C/PMBus).
- ledmgr: power/status/identify LED control.
- inventorymgr, usermgr, securitymgr, updatemgr, telemetry: management, accounts/policy, updates, and observability.
- ipmisrv: IPMI compatibility (work in progress).
- kvmsrv, consolesrv: KVM and SOL services (planned).

## API

- Primary: ConnectRPC over HTTPS with JSON or protobuf payloads.
- REST: exposed from protobuf annotations (paths under `/api/v1alpha1/...`).
- Compatibility: Redfish is planned; IPMI is present as a compatibility service and will route into the same managers.
- See docs/api.md for concrete request examples and endpoints.

## Hardware integration

- GPIO: power/reset/identify, power-good, and LEDs via `pkg/gpio`.
- I2C/SMBus/PMBus: sensor and PSU access via `pkg/i2c`.
- HWMON: Linux sensor interfaces.
- Thermal: zones, sensors, cooling devices; PID control in thermalmgr.

## How this matches the code

- Services live under `service/`, with `doc.go` describing each module.
- Cross-cutting packages live under `pkg/` (gpio, i2c, state, ipc, telemetry, etc.).
- The protobuf schema (authoritative API) is under `schema/v1alpha1/` and drives both ConnectRPC and REST.

## What to read next

- docs/architecture.md — deeper service graph and ordering
- docs/api.md — how to access and query the API (ConnectRPC, REST, future Redfish/IPMI)
- docs/state.md — state machine wrapper and patterns
- docs/gpio.md — GPIO usage and conventions
- docs/integration_example.md — power/state orchestration flow
- docs/integration_complete.md — power/LED integration details
- docs/porting.md — how to configure and port a new platform
- docs/roadmap.md — current open items and planned features
