# Changelog

## [0.1.0] - Unreleased

### Added

- Stage 01 project foundation.
- Go module skeleton.
- Minimal CLI entry point.
- JSON configuration skeleton.
- Basic config validation.
- Basic logger setup.
- Repository structure for future PID, PLC, simulator, MQTT, storage and web UI stages.
- Reusable `pkg/pid` controller package.
- PID modes: auto, manual, hold and disabled.
- Output limits and anti-windup behavior.
- Unit tests for PID behavior and validation.
- Process simulator package with first-order process dynamics.
- Deterministic noise support using a per-process random source.
- Manual disturbance injection and expiry.
- Simulator validation and unit tests.
- Virtual PLC runtime with scan cycle, loop registry, snapshots, events and command handling.
- MQTT interface for telemetry publishing, status publishing and command subscription.
- Local Mosquitto Docker Compose configuration.
- PLC and MQTT unit tests.
