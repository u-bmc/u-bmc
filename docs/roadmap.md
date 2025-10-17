# u-bmc Roadmap

This document is the single source of truth for planned work and open items. It complements the codebase and the package documentation (doc.go files) by outlining deliverables, priorities, and sequencing. If you plan new features or find gaps, add them here rather than scattering TODOs in other files.

The roadmap is grouped into three horizons: Now, Next, and Later. Items may move between horizons as we validate scope and gain confidence.

~~~mermaid
flowchart LR
    A[Now] --> B[Next] --> C[Later]

    subgraph Now
    N1[Web UI v1<br/>power & state, sensors, users]
    N2[Sensor monitoring<br/>thresholds & events]
    N3[Web API hardening<br/>authn/z, TLS, CORS]
    N4[Platform enablement<br/>initial targets]
    end

    subgraph Next
    X1[KVM service<br/>video + HID]
    X2[SOL service<br/>serial over LAN]
    X3[Update manager<br/>A/B, signed images]
    X4[Redfish v1<br/>core resources]
    end

    subgraph Later
    L1[IPMI server<br/>mandatory command set]
    L2[Rootfs builder<br/>reproducible images]
    L3[Security posture<br/>2FA, secrets mgmt]
    L4[Observability++<br/>dashboards, alerts]
    end
~~~

## Guiding principles

We optimize for correctness, clear service boundaries, and operational robustness. A feature lands when it has end-to-end tests, instrumentation, and user-facing documentation. The authoritative runtime behavior is the implementation; the protobuf schema under schema/v1alpha1 is the authoritative API contract. The ConnectRPC API is served over HTTPS and transcoded to REST; Redfish and IPMI are compatibility layers on top of the same service graph.

## Now

Web UI v1
The first version of the Web UI should cover daily operator tasks through the existing ConnectRPC API (transcoded to REST) without requiring SSH access.

- Scope: authentication, power and state operations, live sensor views, basic inventory, user management, and a minimal event log view.
- Delivery: the web UI talks only to websrv; no direct access to other services.
- Packaging: development build for local testing; production build served by websrv.

Service documentation completion
Comprehensive service documentation has been completed for all major implemented services.

- Core services documented: operator, ipc, websrv, statemgr, powermgr, thermalmgr, sensormon, ledmgr
- API documentation: complete ConnectRPC and REST API reference with integration examples
- Documentation standards: centralized in docs/ directory with consistent formatting and cross-references
- Package documentation: available on pkg.go.dev for detailed implementation guidance

Web API hardening
Strengthen the default posture of websrv.

- Transport: TLS 1.3 default; self-signed for development and ACME for production.
- Authn/z: session and token handling, role-aware handlers, and request validation.
- Cross-cutting: CORS policy, request limits, structured errors, and tracing.

Platform enablement
Bring up initial targets in targets/, following the porting guidance in docs/porting.md.

- Deliverables: GPIO line mapping for power and LEDs, I2C bus and device discovery, thermal zones, and inventory basics.
- Non-userspace items are tracked in this roadmap's platform section (see Later → Rootfs builder, kernel and device tree notes).

## Next

KVM service (kvmsrv)
Provide remote keyboard, video, and mouse access that integrates with the Web UI.

- Video: capture and encode pipeline suitable for constrained BMC hardware.
- Input: keyboard and mouse injection with proper focus and safety controls.
- Transport: secure channel via websrv with backpressure and rate control.
- UX: basic browser viewer in the Web UI.

Serial-over-LAN (consolesrv)
Expose host serial console over the network and integrate a browser terminal.

- Protocols: web-based terminal via websrv, and compatibility with IPMI v2.0 SOL framing later.
- Features: replay window, flow control, and optional audit logging with redaction.

Update manager (updatemgr)
Implement safe firmware and software updates.

- A/B layout: active/backup images with rollback.
- Integrity: signed images, provenance (SBOMs), and progress reporting.
- Coordination: quiesce services, update, verify, and return to service, with telemetry and events.

Redfish API (v1)
Introduce a minimal Redfish layer mapped onto our services.

- Scope: Systems, Chassis, Managers, and basic power/thermal endpoints.
- Mapping: Redfish models translated to the schema/v1alpha1 service graph; websrv routes requests.
- Strategy: deliver a small, well-tested core first; expand breadth incrementally.

## Later

IPMI server (ipmisrv)
Support required commands for compatibility with existing tooling.

- Command sets: mandatory chassis, sensor, and session commands first.
- Strategy: bridge IPMI operations to the same internal services (statemgr, sensormon, powermgr) to avoid divergent logic.

Rootfs builder (dagger/)
Create a reproducible builder for images and artifacts.

- Build: containerized, hermetic pipelines with pinned inputs; distroless or minimal base.
- Output: images for supported SoCs, persistent data layout, and signing.
- TODOs (non-userspace): kernel configuration, device tree overlays, udev rules, hwmon configuration, GPIO pinmux, I2C muxing, and watchdog policies. These items are platform-specific and will be tracked centrally here until the builder exists.

Security posture
Improve account and secret management.

- Accounts: password policy, rotation, and 2FA options (TOTP first).
- Secrets: secure storage and process access patterns; certificate lifecycle management in websrv.
- Policy: audit logs for privileged actions with sampling to telemetry.

Observability and operations
Make the system easy to run and debug at scale.

- Telemetry: metrics, traces, and events with consistent attributes; exemplars for key paths.
- Dashboards and alerts: reference dashboards and SLO-aligned alert rules.
- Diagnostics: service health endpoints and on-demand bundle (config snapshot, logs, and recent events).

Performance and scalability
Keep latency low and resource use predictable on BMC-class hardware.

- IPC: NATS JetStream tuning for retention and backpressure.
- Web: QUIC tuning and efficient streaming for KVM and SOL.
- Services: memory caps and CPU budgets; steady-state audits.

Release and supply chain
Ship artifacts that are trustworthy and easy to consume.

- Versioning: follow VERSIONING.md; keep schema evolution compatible within v1alpha1.
- Artifacts: signed images, checksums, and SBOMs; container images where appropriate.
- Upgrades: documented rollback and recovery procedures.

## Platform enablement (central checklist)

- GPIO: power/reset/identify lines mapped with clear active states and debounce characteristics.
- LEDs: power, status, and identify mapping with supported patterns (on/off/blink).
- I2C: bus topology, device addresses, and muxes; PMBus devices and scaling coefficients.
- Thermal: zones, sensors, and PID defaults per platform (see docs/thermalmgr.md for configuration).
- Inventory: asset information available via schema/v1alpha1.
- Sensors: hwmon device mapping and threshold configuration (see docs/sensormon.md for setup).
- Non-userspace: kernel, device tree, udev, hwmon, pinmux, watchdog (tracked until the rootfs builder is in place).

## API direction

The ConnectRPC API is the primary entry point and is available now. It supports JSON and protobuf over HTTP/1.1, HTTP/2, and HTTP/3. REST is available through transcoding from protobuf annotations in schema/v1alpha1. Redfish and IPMI are compatibility layers that will route into the same operators and managers. The API surface evolves through schema changes with explicit, documented transitions.

## How to propose changes

- Propose: open an issue that describes the user-facing problem and the service boundaries it touches.
- Discuss: align on scope and sequencing; add the item here if it crosses a release boundary.
- Land: ship with tests, telemetry, and docs. If the change adds configuration, document defaults and migration notes.
- Clean up: ensure there are no stray TODOs in code or docs—keep this file as the only central backlog.

## Related documents

- docs/overview.md — high-level system overview
- docs/architecture.md — service graph and operator model
- docs/state.md — state machine wrapper and patterns
- docs/gpio.md — GPIO abstractions and usage
- docs/api.md — API usage and request examples
- docs/porting.md — how to configure and port a new platform
- docs/operator.md — service supervision and lifecycle management
- docs/ipc.md — embedded NATS messaging and service communication
- docs/websrv.md — web server and API gateway service
- docs/statemgr.md — system state machines and coordination
- docs/powermgr.md — power management and sequencing service
- docs/ledmgr.md — LED control and visual indication service
- docs/api.md — comprehensive API reference and integration guide
- docs/sensormon.md — sensor monitoring service documentation
- docs/thermalmgr.md — thermal management service documentation

Status and priorities evolve as we integrate more platforms and validate the runtime. This roadmap keeps us aligned on the next most valuable increments while protecting the system’s reliability.
