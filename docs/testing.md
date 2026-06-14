# Testing Guide

## Overview

All tests run without external services. No Docker, MQTT broker, or browser is required for the standard test suite.

## Running Tests

### Full test suite with race detector

```bash
go test -race ./...
```

### With coverage report

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Smoke tests only

```bash
go test -run TestSmoke -v ./tests/smoke/...
```

Or via the helper script:

```bash
./scripts/smoke.sh      # Linux / macOS
.\scripts\smoke.ps1     # Windows
```

### MQTT broker integration tests

Requires a running MQTT broker on `localhost:1883`:

```bash
docker compose -f docker/docker-compose.yml up -d
VPLC_RUN_MQTT_TESTS=1 go test -v ./tests/integration/...
```

## Test Packages

| Package | Coverage | Notes |
|---------|----------|-------|
| `pkg/pid` | ~93% | PID controller: modes, limits, anti-windup |
| `pkg/simulator` | ~92% | Process dynamics, noise, disturbances |
| `pkg/plc` | ~82% | Runtime lifecycle, commands, snapshots |
| `pkg/mqttx` | ~52% | Topics, payloads, command parsing |
| `internal/config` | ~84% | Loader, validation, defaults |
| `internal/storage` | ~76% | SQLite, JSONL, recorder, migrations |
| `internal/app` | ~44% | Lifecycle, config mapping, command handler |
| `internal/web` | ~66% | API handlers, SSE, validation helpers |
| `tests/smoke` | — | End-to-end app start/stop without services |
| `tests/integration` | — | MQTT broker integration (gated) |

## Test Architecture

### No external services

Unit and integration tests in each package use only in-process components. The `pkg/mqttx` tests include a `TestMQTTClientNew_Disabled` test that verifies the client can be created without connecting.

### Race detector

All CI runs use `-race`. Run locally before committing:

```bash
go test -race ./...
```

### Smoke tests (`tests/smoke/`)

Smoke tests create an `internal/app.App` directly, disable all external services (MQTT, storage, web), and verify:

1. `config.Default().Validate()` passes
2. `App.Run()` foundation check completes without error
3. `App.RunRuntime()` starts and stops cleanly within a 60ms timeout

### MQTT integration tests (`tests/integration/`)

Gated by environment variable `VPLC_RUN_MQTT_TESTS=1`. Test a real connect/publish/subscribe cycle against a broker.

## Writing New Tests

- Place package tests in the same directory as the source (`package foo`).
- Place external-package tests in `package foo_test` for black-box coverage.
- Use `httptest.NewServer` for HTTP handler tests; never bind real ports.
- Avoid `time.Sleep` in tests; prefer channels or short context timeouts.
- Do not mock the PLC runtime in web handler tests — use `testRuntime(t)` which creates a real, stopped runtime.
