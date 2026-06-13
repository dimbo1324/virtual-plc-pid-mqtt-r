# virtual-plc-pid-mqtt-r

A lightweight virtual PLC simulator in Go with reusable PID controllers,
first-order process models, MQTT telemetry, and remote command handling.

## Safety Notice

This software is a virtual PLC / PID / MQTT simulator.
It is not intended to control real industrial equipment.
Do not connect this simulator directly to real actuators or safety-critical systems.

## Current Status

Stages 01-06 are implemented:

- JSON configuration and CLI foundation
- reusable PID controller package
- deterministic synthetic process simulator
- concurrent PLC runtime with real elapsed `dt`, snapshots, events, and commands
- MQTT status, telemetry, events, command subscription, LWT, and reconnect behavior
- local Mosquitto Docker Compose configuration

Storage, HTTP/SSE, and the web dashboard are intentionally not implemented yet.
SQLite, JSONL history, CI, authentication, TLS setup, and real equipment
communication are also outside the current implementation.

## Architecture

```text
JSON config
    |
    v
PLC runtime: PID controllers <-> synthetic processes
    | snapshots/events          ^ typed commands
    v                           |
MQTT telemetry/events <-> MQTT command subscription
```

The application remains a single Go binary. `pkg/plc` depends only on
`pkg/pid` and `pkg/simulator`; `pkg/mqttx` depends on `pkg/plc` and Eclipse
Paho. Application-specific mapping and lifecycle code live in `internal/app`.

Detailed boundaries are documented in [Architecture](docs/architecture.md).
The next persistence extension points and risks are documented in
[Pre-storage Notes](docs/pre_storage_notes.md).

## Requirements

- Go 1.25.5 or later for the current module
- Go 1.26 remains the project target when available locally
- Docker Desktop or another MQTT broker is optional for broker integration

## Quick Start

Short foundation checks do not start a long-running service:

```bash
go run ./cmd/vplc --version
go run ./cmd/vplc --validate-config --config configs/default.json
go run ./cmd/vplc --config configs/default.json
```

Run the full PLC lifecycle until Ctrl+C:

```bash
go run ./cmd/vplc --run --config configs/default.json
```

If MQTT is enabled but the broker is unavailable, the PLC continues scanning
and the application retries the initial connection in the background.

The helper scripts start the same long-running mode:

```powershell
.\scripts\run.ps1
```

```bash
./scripts/run.sh
```

## Configuration

The default JSON file is `configs/default.json`. Unknown JSON fields and extra
JSON values are rejected. Validation covers PLC timing, MQTT connection fields
and topic shape, loop names and modes, PID/process ranges, and conditional web
or storage paths.

Web and storage are disabled in the default configuration because those
subsystems are not implemented yet.

```bash
go run ./cmd/vplc --validate-config --config configs/default.json
```

## MQTT Demo

Start Mosquitto:

```bash
docker compose -f docker/docker-compose.yml up -d
```

Start the runtime:

```bash
go run ./cmd/vplc --run --config configs/default.json
```

Observe all project topics:

```bash
mosquitto_sub -h localhost -t "vplc/#" -v
```

Change the pressure setpoint:

```bash
mosquitto_pub -h localhost -t "vplc/vplc_001/commands" -m '{"command_id":"cmd-1","command":"set_setpoint","loop":"pressure","value":7.5}'
```

Stop Mosquitto:

```bash
docker compose -f docker/docker-compose.yml down
```

The bundled broker allows anonymous local connections and does not configure
TLS. Keep it on a trusted development machine only.

See [PLC runtime](docs/plc_runtime.md) and [MQTT contract](docs/mqtt_contract.md)
for APIs, payloads, commands, and lifecycle details.

## Development

```bash
gofmt -w .
go mod tidy
go mod verify
go test ./...
go vet ./...
go build -o dist/vplc ./cmd/vplc
```

Optional broker integration:

```bash
VPLC_RUN_MQTT_TESTS=1 go test ./tests/integration/... -v
```

Windows build:

```powershell
go build -o dist\vplc.exe .\cmd\vplc
```

## Roadmap

1. Foundation, PID, simulator, PLC runtime, and MQTT: completed.
2. Next: `stage-07-storage-history` for bounded local persistence and logging.
3. Later: embedded local web UI, HTTP API, SSE, and portfolio polish.

Development conventions and release checks are described in the
[Developer Guide](docs/developer_guide.md).
