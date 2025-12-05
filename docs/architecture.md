# Architecture

u-bmc is composed of small Go services supervised by an operator. Services communicate over an embedded NATS instance (IPC). The public API is served by the web server, which exposes ConnectRPC and REST (transcoded from protobuf annotations). Hardware access is isolated in manager services that use focused packages from pkg/.

```mermaid
graph TB
    %% External
    Client[Clients<br/>Web UI, CLI, Tools]

    %% Entry and IPC
    Web[websrv<br/>HTTPS + ConnectRPC/REST]
    IPC[ipc<br/>Embedded NATS (+JetStream)]
    Op[operator<br/>Supervision & lifecycle]

    %% Core managers
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

    %% Protocol/compat
    IPMI[ipmisrv]
    KVM[kvmsrv]
    CON[consolesrv]

    %% Hardware
    HW[Hardware<br/>GPIO, I2C/PMBus, HWMON]

    %% Edges
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

    classDef ext fill:#eef7ff,stroke:#c3defe
    classDef entry fill:#f4ecff,stroke:#d8cafe
    classDef ipc fill:#fff2da,stroke:#ffd18a
    classDef op fill:#e8f6ee,stroke:#a8e6c3
    classDef core fill:#eef7ee,stroke:#c9e6c9
    classDef proto fill:#eef5ff,stroke:#cfe0ff
    classDef hw fill:#ffeceb,stroke:#ffc1bd

    class Client ext
    class Web entry
    class IPC ipc
    class Op op
    class SM,PM,TM,SN,LED,INV,UM,SEC,UPD,TEL core
    class IPMI,KVM,CON proto
    class HW hw
```

Service map (directory → purpose)
- service/operator — orchestrates startup, supervision, graceful shutdown, and shared setup.
- service/ipc — embedded NATS server and connection glue for in‑process messaging.
- service/websrv — HTTPS entrypoint; serves ConnectRPC and REST (transcoded).
- service/statemgr — system state machines and transitions.
- service/powermgr — host/chassis/BMC power sequencing and reporting.
- service/thermalmgr — fan control, PID profiles, and thermal protection.
- service/sensormon — sensor discovery, polling, and threshold events.
- service/ledmgr — status/identify/power LED control.
- service/inventorymgr — asset and component metadata.
- service/usermgr — user accounts and authentication glue.
- service/securitymgr — authorization and security policy.
- service/updatemgr — software/firmware update coordination.
- service/telemetry — metrics and tracing integration.
- service/ipmisrv — IPMI compatibility (planned/partial).
- service/kvmsrv — remote video + HID (planned).
- service/consolesrv — serial-over-LAN (planned).

Supporting packages (selection)
- pkg/gpio — GPIO via gpio‑cdev.
- pkg/i2c — I2C/I3C/SMBus/PMBus access.
- pkg/hwmon — Linux HWMON helpers.
- pkg/state — thin wrapper around stateless FSM.
- pkg/ipc — IPC helpers and response utilities.
- pkg/log, pkg/process, pkg/mount, pkg/file, pkg/telemetry — shared utilities.

API surface
- Schema: schema/v1alpha1/*.proto (authoritative request/response types and REST annotations).
- Access: see [docs/api.md](api.md) for ConnectRPC and REST examples, authentication, and content types.

Lifecycle and ordering
- The operator starts IPC first, then core managers in parallel (power, sensors, thermal, state), followed by management and protocol services, and finally the web server. Shutdown proceeds in reverse while preserving dependencies.
## References

- [docs/overview.md](overview.md) — higher-level narrative and context
- [docs/api.md](api.md) — how to query the API
- [docs/sensormon.md](sensormon.md) — sensor monitoring service configuration and usage
- [docs/thermalmgr.md](thermalmgr.md) — thermal management and PID control setup
- pkg/* and service/* doc.go — package-level details on pkg.go.dev
