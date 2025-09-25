# u-bmc rootfs builder (preview)

This directory will host the reproducible rootfs and image builder for u-bmc, implemented with Dagger. The builder will produce minimal, signed images with an A/B layout and SBOMs, and is intended to run in CI as well as locally.

- Status and backlog are tracked centrally in docs/roadmap.md.
- Platform enablement details and non-userspace prerequisites (kernel, DT, udev, hwmon, pinmux, watchdog) are also tracked in docs/roadmap.md.
- For porting guidance and how the builder will integrate with targets, see docs/porting.md.

Until the builder lands, this directory intentionally contains no implementation. Please avoid adding TODOs here; use the roadmap instead.
