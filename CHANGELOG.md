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
- Pre-storage architecture, lifecycle, validation and documentation hardening.
- Explicit application-owned event fan-out and joined background goroutines.
- Additional config, PLC channel, MQTT topic and app lifecycle tests.
- Local SQLite storage for telemetry samples, events, commands and PID changes.
- JSONL event log writer with append mode and concurrent-safe mutex.
- SQLite schema migrations (idempotent, version-tracked in `schema_migrations`).
- Retention cleanup: keeps newest N telemetry rows, configurable via `retention_max_samples`.
- Async bounded recorder that bridges PLC runtime channels to storage without blocking the scan loop.
- `internal/storage` package: `Store`, `JSONLWriter`, `Recorder` with full test coverage.
- App wiring in `internal/app`: storage open/close on startup/shutdown, snapshot consumer goroutine, event recording, command recording, PID change recording.
- Storage failure at startup fails the app clearly; write failures after startup log a warning and keep PLC running.
- `write_queue_size` config field added to `storage` section.
- `docs/storage_history.md` explaining storage design, tables, paths and inspection.
