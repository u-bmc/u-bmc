# Porting guide: adding a new hardware platform

This document explains how to add a new mainboard or platform to u-bmc. It focuses on practical steps for configuring GPIO, LEDs, sensors, thermal zones, and service wiring. It also lists the non‑userspace and rootfs builder tasks that need to be tracked while the builder is not yet available.

Who this is for
- Platform engineers bringing up new boards or SoCs.
- Contributors adding support for a new target under the `targets` tree.

Key ideas that inform porting
- Services are small and focused. The operator supervision process starts IPC first, then core services (power, thermal, sensors), then protocol and system services (web, telemetry, etc.).
- IPC uses NATS in-process; all cross-service requests flow through it. Keep physical I/O localized to the managing service (for example, power sequencing lives in powermgr).
- The public API is served by websrv using ConnectRPC, with REST transcoding based on protobuf annotations. The protobuf schema in `schema/v1alpha1` is the authoritative API contract.
- Package documentation lives in `doc.go` and item comments. Do not add markdown files inside packages. Platform documentation belongs either here in docs or in the target’s own README (allowed only for hardware targets, the rootfs builder in dagger, and the web UI).

Directory layout
- Add your new platform under `targets`, for example `targets/mainboards/<vendor>/<board>/`.
- A platform may include its own README because targets are explicitly allowed to have one. Keep it brief and link back to `docs/` for general instructions.
- The platform’s entrypoint is a small `main.go` that configures and runs the operator with platform‑specific service configuration.

Prerequisites
- Linux kernel with GPIO character device (gpio-cdev) and the right I2C/HWMON/PMBus drivers for your hardware.
- Device tree (or equivalent mechanism) that exposes stable line names and bus topology.
- Access to platform schematics or verified mappings for power/reset/identify lines, LED drivers, I2C buses, fans, and sensors.
- Go toolchain that matches the module in `go.mod`.

High‑level porting steps
1) Collect hardware facts
- SoC and board family (ASPEED, Nuvoton, etc.).
- GPIO lines for power button, reset button, force off (long press), power good, identify, and status LEDs.
- I2C buses, muxes, and device addresses (PMBus PSUs, temperature sensors, EEPROM, fan controllers).
- Fan control capabilities (PWM, tach).
- Thermal layout (zones, sensors per zone, basic PID defaults).
- Watchdog device and policy.
- Any platform inventory attributes that should be exposed.

2) Define stable names and conventions
- GPIO line names should come from the kernel’s line naming. Prefer stable functional names over numeric offsets (for example, use power-good-0 vs GPIOA7).
- Component names should be predictable: use patterns like host.0, host.1, chassis, bmc, psu.0, psu.1, fan.0, fan.1.
- Sensor IDs and names should be stable and human‑readable (for example, cpu0_temp, inlet_temp, psu0_vout). Map them to schema contexts and units.

3) Create the target entrypoint
- Add a `main.go` under your target directory. The program should:
  - Construct an operator with a reasonable startup timeout.
  - Enable IPC first and then the services you need (powermgr, statemgr, sensormon, thermalmgr, ledmgr, websrv, telemetry, etc.).
  - Provide platform configuration to each service through its options (see each service’s `doc.go` for the supported options).
  - Optionally expose a health log or minimal startup banner.
- Use the ASUS IEC example under `targets/mainboards/asus/iec/` as a reference for wiring the operator.

4) Configure power and LEDs
- powermgr:
  - Set the GPIO chip device path for your platform (for example, `/dev/gpiochip0`).
  - For each controllable component (for example, `host.0`):
    - Map `power-button` (momentary press), `reset-button` (momentary), and `power-good` (input).
    - Provide timings for power on, power off, reset, and force off. Common examples are 200ms press for on/off and 4s hold for force off.
  - If the board has chassis‑level actions, configure those as components too.
- ledmgr:
  - Map logical LED types to hardware control points:
    - Power LED: on when host is on; off when host is off.
    - Status/health LED: on or blinking on warning; fast blink on error.
    - Identify LED: on when identify is requested; off when cleared.
  - The statemgr can request LED changes based on transitions; ensure the action mapping is coherent with your LED hardware.

5) Configure sensors and inventory
- sensormon:
  - Enumerate sensors from HWMON and I2C devices. Assign stable IDs and names.
  - Map each sensor to a schema context and unit:
    - Temperature: Celsius
    - Voltage: Volts
    - Current: Amps
    - Tach/fans: RPM or Percent if abstracted
    - Power: Watts
  - Configure thresholds and debouncing where appropriate so noise does not cause flapping.
- inventorymgr:
  - Populate asset fields when available: product name, manufacturer, serial number, part number, UUID, etc.
  - Use the schema’s asset message to set known manufacturing and installation dates when the platform can provide them.

6) Configure thermal zones and fans
- thermalmgr:
  - Define zones (for example, `chassis`, `cpu`, `psu`) and the set of sensor names that feed each zone.
  - Define cooling devices (fans, blowers, pumps) that are controlled per zone, and their control mode (automatic/PID).
  - Provide a sensible default thermal profile per platform (quiet, balanced, aggressive) and PID defaults (kp, ki, kd, sample time).
  - Ensure fan tach and PWM are discoverable and writable by the process. Stabilize paths via udev if needed.

7) Web, API, and auth
- websrv:
  - Set the bind address and certificate policy:
    - Self‑signed in development.
    - ACME/Let’s Encrypt or your own certificates in production.
  - Confirm CORS, request limits, and timeouts are sensible for your environment.
  - Expose the API at `/api/v1alpha1/…`. The ConnectRPC interface is always available; REST is provided by protobuf annotations.
- usermgr/securitymgr:
  - Bring up a minimal local user flow if the platform requires local login at first boot.
  - Integrate authentication mode for your deployment (tokens, sessions).

8) Test the platform end‑to‑end
- Smoke tests:
  - Start the operator and ensure services initialize in order; check logs.
  - List sensors via `/api/v1alpha1/sensors` and confirm expected IDs and units.
  - Toggle host power via `/api/v1alpha1/hosts/{name}/state` with `HOST_ACTION_ON` and `HOST_ACTION_OFF`.
  - Verify LED behavior changes with state transitions.
  - Exercise a thermal zone PUT to change `targetTemperature` and confirm fan behavior.
- Reliability tests:
  - Reboot and power cycle the host repeatedly; confirm no stuck pins or races.
  - Disconnect or fault sensors (where safe) and confirm graceful degradation.
  - Network loss and recovery for clients; confirm idempotent operations.

9) Documentation
- Keep platform‑specific notes in the target’s README (allowed for hardware targets). Avoid TODOs in code; instead:
  - Add open items to `docs/roadmap.md` under the platform enablement and non‑userspace sections.
  - Link back to `docs/api.md`, `docs/overview.md`, and this file for shared guidance.

Service‑specific guidance (brief)
- powermgr: uses `pkg/gpio` for line control; prefer the convenience toggles for momentary presses. Confirm active‑low vs active‑high and debounce.
- sensormon: combine HWMON polling with I2C/PMBus reads; map sensor readings to the correct schema units; record `last_reading_timestamp`.
- thermalmgr: keep PID bounds realistic for your fan hardware; avoid oscillation by tuning sample time and constraints.
- websrv: stick to TLS 1.3; use QUIC if supported; set reasonable read/write/idle timeouts; prefer serving the Web UI from the same process for simplicity.
- statemgr: ensure action mappings for power and LEDs are complete; verify that chassis, host, and BMC actions drive the expected transitions.

Naming conventions that help
- Component names: `host.0`, `host.1`, `chassis`, `bmc`.
- Sensor IDs: `cpu0_temp`, `inlet_temp`, `psu0_vout`.
- LED types: `power`, `status`, `identify` mapped to schema enums.
- GPIO lines: stable functional names (for example, `power-button-0`, `reset-button-0`, `power-good-0`).

API and validation notes
- The authoritative API surface is defined in `schema/v1alpha1/*.proto` and served by websrv.
- REST endpoints and methods are defined via `google.api.http` annotations in the same schema files.
- Validation uses `buf/validate`; when adding platform‑specific rules in services, match the semantics of the schema and prefer explicit error details.

Non‑userspace and rootfs builder TODOs
Track these centrally in `docs/roadmap.md`. The items below are common across platforms and should not be duplicated as inline code TODOs.

Kernel and device tree (platform bring‑up)
- Enable gpio‑cdev, I2C, PMBus, HWMON, watchdog, and sensor drivers for your platform.
- Provide device tree overlays for:
  - Pinmux and line naming for power/reset/identify and LEDs.
  - I2C buses, muxes, and PMBus devices with proper addresses and compatibles.
  - Fan controllers (tach and PWM), and any GPIO expanders.
  - Thermal sensors and zones for early defaults.
- Ensure watchdog device is exposed consistently.

Userspace integration (stable paths)
- udev rules to stabilize `/dev/gpiochipN` and sensor symlinks.
- hwmon label mapping consistency across reboots.
- Permissions for device files accessed by services.

Rootfs and image composition (builder pending)
- Filesystem layout: read‑only root, writable data, and logs.
- A/B system for safe updates with rollback.
- Signed images, SBOM generation, and attestation metadata.
- Kernel and module packaging per platform.
- Bootloader, partitioning, and secure boot policies where applicable.

Security posture and operations
- Default user policy, password rotation, and optional 2FA.
- Certificate lifecycle for websrv (self‑signed for development, ACME or provisioned for production).
- SELinux or AppArmor policy (defer to builder; keep notes on desired profiles).
- Log retention and privacy policy for console and KVM sessions.

Networking and time
- IPv6, VLAN, and mDNS as needed for your sites.
- NTP or time synchronization suitable for the environment.

Performance and observability
- NATS memory and JetStream retention limits for the device class.
- QUIC socket buffer tuning where HTTP/3 is used.
- Metrics and traces with consistent resource attributes.

Bring‑up checklist (summary)
- GPIO lines named and verified; power and reset actions work with correct timings.
- LEDs mapped and respond to state transitions.
- Sensors enumerated with correct units; thresholds produce sane events.
- Thermal zones defined; fans controlled without oscillation.
- Web API reachable with TLS; auth mode is set; CORS and limits applied.
- Operator boots all required services; clean shutdown and restart verified.
- Platform README exists under the target directory with wiring and notes.
- All open items recorded in `docs/roadmap.md` (non‑userspace and builder tasks included).

See also
- docs/overview.md for the system overview.
- docs/architecture.md for the service graph and sequencing model.
- docs/api.md for API access patterns and examples.
- docs/roadmap.md for central TODOs and planned compatibility layers.
