# Changelog

## [0.1.0] - 2026-06-14

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
- Embedded local web dashboard served by `net/http` with `go:embed` static assets.
- REST API: `GET /api/status`, `GET /api/snapshot`, `GET /api/events/recent`, `GET /api/telemetry/recent`.
- Command API: `POST /api/commands/{start,stop,setpoint,pid-gains,mode,manual-output,inject-disturbance,reset-loop}`.
- Server-Sent Events stream at `GET /api/stream` with `snapshot`, `plc_event`, and `heartbeat` event types.
- Vanilla JS dashboard with Canvas trend charts (PV/SP/MV, 300-point rolling buffer), dark engineering palette.
- SSE broker with per-client buffered channels; slow clients silently dropped to avoid blocking the scan loop.
- Fan-out extended in `internal/app` to deliver events and snapshots to both the storage recorder and the web SSE broker.
- `web.enabled` defaults to `true`; binds to `127.0.0.1:8080`.
- `internal/web` package with 24 tests covering API handlers, SSE, validation helpers, and config.
- Stage 09: tests, CI, build and release hardening.
- GitHub Actions CI workflow (`.github/workflows/ci.yml`): format, vet, race-detector tests, build, config validation, optional vulnerability scan.
- Smoke test package (`tests/smoke/`) exercising the full app stack without external services.
- Additional unit tests: `internal/app` helper functions, `internal/web` manual-output and inject-disturbance handlers, `pkg/mqttx` command parsing edge cases.
- `scripts/smoke.ps1` and `scripts/smoke.sh`: build, validate, and run smoke tests.
- `scripts/release.ps1` and `scripts/release.sh`: cross-compile release binaries with version ldflags.
- Documentation: `docs/testing.md`, `docs/build_release.md`, `docs/ci.md`.
- `.gitattributes` enforcing LF line endings for all text files (CRLF for PowerShell/batch).
- `.gitignore` updated to exclude `/release/` directory.
- `internal/storage.Store.Close()` now checkpoints WAL before closing to prevent orphaned `-wal`/`-shm` files on Windows.
