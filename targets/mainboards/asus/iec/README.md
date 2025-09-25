# ASUS IPMI Expansion Card (IEC)

This target integrates the ASUS IPMI Expansion Card (IEC)—a PCIe x1 form factor BMC based on the ASPEED AST2600—into u-bmc. It provides a clean entry point for platform-specific configuration (GPIO, I2C/PMBus, sensors, thermal zones) and wires the board into the common service graph orchestrated by the operator.

## What this target is

- A focused, platform-specific bootstrap that configures and launches the operator with services relevant to this board.
- A place to define stable names and mappings for GPIO lines (power/reset/identify, power-good), LEDs, sensors, fans, and PMBus devices.
- A thin layer that uses shared packages and services; business logic lives in `service/` and `pkg/`, and the public API is defined in `schema/v1alpha1`.

## Scope and expectations

- Power and state management, sensors, thermal control, and inventory are driven by shared services and exposed via the same API as all other targets.
- The Web UI, Redfish/IPMI compatibility layers, KVM, and SOL are supported by the central service roadmap and are not implemented ad hoc per target.
- Platform bring-up focuses on reliable mappings, safe defaults, and repeatable behavior. The authoritative API is served by `websrv` and documented in `docs/api.md`.

## Hardware assumptions

- SoC: ASPEED AST2600
- Kernel support: gpio-cdev, I2C/PMBus, HWMON, watchdog
- Stable line naming for GPIO and discoverable I2C topology are expected (via device tree or equivalent)

## Using this target

1. Review the porting guidance in `docs/porting.md` for how to provide platform mappings and service options.
2. Bring up the target with the operator configuration for this board, validating power, LEDs, sensors, and thermal behavior.
3. Access and test the API using the examples in `docs/api.md`. The API is available via ConnectRPC and REST (transcoded), under `/api/v1alpha1/...`.

## Status and roadmap

Development status and planned features for this target follow the central project roadmap. For KVM, SOL, update flows, and compatibility layers (Redfish, IPMI), see:

- docs/roadmap.md

## Related documentation

- docs/overview.md — high-level system overview
- docs/architecture.md — service graph and operator model
- docs/api.md — how to access and query the API
- docs/porting.md — how to configure and port a new platform
- docs/webui.md — Web UI scope and integration

This README intentionally avoids per-file TODO lists. Please track open items and planned work in the central roadmap.
