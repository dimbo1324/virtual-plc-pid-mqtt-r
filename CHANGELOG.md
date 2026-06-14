# Changelog

## [Unreleased] — Audit fixes (stage 11 follow-up)

### Fixed
- **`externalPVs` never cleared**: added `Runtime.ClearPV(name)` method; `runInputProvider` now calls it on `QualityBad` readings so the simulator resumes when a provider loses signal.
- **`open_browser` was a no-op**: implemented cross-platform browser launch in `cmd/vplc/main.go` via `os/exec` (Windows `start`, macOS `open`, Linux `xdg-open`).
- **`SyntheticProvider` circular feedback loop**: removed the wiring from `lifecycle.go` (it read back its own injected PV, hiding simulator physics); the `runInputProvider` hook remains for real adapters.
- **`storageMode` frozen at startup**: replaced the static string snapshot with `atomic.Value`; JSONL write failures in degraded mode now update the mode to `"failed"` at runtime.
- **Retention policy global, not per-loop**: SQL query now uses `ROW_NUMBER() OVER (PARTITION BY loop_name)` to keep N rows *per loop*, preventing an active loop from displacing history of quieter loops.
- **`Recorder.Stop()` busy-wait**: replaced `time.Sleep(10ms)` poll loop with a `done chan struct{}` closed by the worker goroutine.
- **Empty `loop` field in web commands**: added `requireLoop` helper returning 400 `"loop is required"` in all six loop-command handlers.
- **`storage.Config.Validate()` missing `FallbackType` check**: validation moved into the storage package itself so it applies regardless of caller.
- **`disabled` mode missing from UI select**: added `<option value="disabled">` with EN (`DISABLED`) / RU (`ОТКЛ.`) translations.
- **PID gains not refreshed via SSE**: `kp`, `ki`, `kd` now updated in `updateLoopCard()` on every snapshot, not only at card creation.
- **`loadStatus()` called once**: now called again after SSE reconnect so `storage_mode` badge and header refresh without page reload.
- **Window listener accumulation** (8 `window` event listeners for 3 loop cards + panel): consolidated into a single global `_activeDrag` state with one `mousemove` / one `mouseup` on `window`.
- **Command errors invisible**: `postCommand` now calls `showToast(msg)` on HTTP error or network failure; a 3.5 s red toast appears in the bottom-right corner.
- **Manual content referenced removed `CASCADE` mode**: updated EN and RU PLC Theory table and usage text to reflect `HOLD` / `DISABLED`.

### Added
- `Runtime.ClearPV(name string)` — explicit PV injection removal.
- `TestReadyz_Running` — covers the 200 path of `/readyz` (previously only 503 was tested).
- `TestInjectPV_OverridesSimulator`, `TestClearPV_RestoresSimulator`, `TestScanCounter_Advances` in `pkg/plc/inject_pv_test.go`.
- `TestSmoke_DegradedStorageFallback` — smoke test for the `fallback_on_error + fallback_type: "jsonl"` code path.
- `.cmd-toast` CSS class + `@keyframes toast-in` animation for command error feedback.

---

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
