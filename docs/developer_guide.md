# Developer Guide

Before every development stage, read
`docs/virtual-plc-pid-mqtt-r_project_specification.md` completely and keep the
stage scope explicit.

## Repository Orientation

- CLI changes belong in `cmd/vplc`.
- Application assembly and service lifecycle belong in `internal/app`.
- JSON shape and validation belong in `internal/config`.
- Reusable control behavior belongs in `pkg/pid`, `pkg/simulator`, or `pkg/plc`.
- MQTT transport behavior belongs in `pkg/mqttx`.
- Future persistence belongs under `internal/storage`; it must not be imported
  by reusable packages.

Avoid catch-all helper packages and interfaces without a concrete boundary or
testability need.

## Run

Validate and perform short initialization:

```bash
go run ./cmd/vplc --validate-config --config configs/default.json
go run ./cmd/vplc --config configs/default.json
```

Run the PLC until Ctrl+C:

```bash
go run ./cmd/vplc --run --config configs/default.json
```

The `scripts/run.ps1` and `scripts/run.sh` helpers run this long-lived mode.

## Test and Build

```bash
gofmt -w .
go mod tidy
go mod verify
go test ./...
go vet ./...
go build -o dist/vplc ./cmd/vplc
```

MQTT integration remains opt-in:

```bash
VPLC_RUN_MQTT_TESTS=1 go test ./tests/integration/... -v
```

## Adding Functionality

1. Identify the owning package from the responsibility map.
2. Add or update behavioral tests before changing shared contracts.
3. Keep external data validation at the boundary.
4. Use context cancellation for every long-running goroutine.
5. Keep channels bounded and define backpressure or drop behavior.
6. Update contracts and current-status documentation in the same change.

For storage specifically, follow [Pre-storage Notes](pre_storage_notes.md).

## Release Safety

Before merging a stage:

1. Run all package and repository checks.
2. Confirm optional checks are honestly reported when unavailable.
3. Inspect staged files for secrets and generated artifacts.
4. Ensure `dist/`, databases, logs, and local environment files are not staged.
5. Merge only from a clean stage branch after checks pass.

The current project has no automated CI pipeline; these checks are manual until
a later stage introduces CI.
