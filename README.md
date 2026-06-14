# virtual-plc-pid-mqtt-r

A lightweight virtual PLC simulator in Go with reusable PID controllers,
first-order process models, MQTT telemetry, and remote command handling.

## Safety Notice

This software is a virtual PLC / PID / MQTT simulator.
It is not intended to control real industrial equipment.
Do not connect this simulator directly to real actuators or safety-critical systems.

## Current Status

Stages 01-07 are implemented:

- JSON configuration and CLI foundation
- Reusable PID controller package
- Deterministic synthetic process simulator
- Concurrent PLC runtime with real elapsed `dt`, snapshots, events, and commands
- MQTT status, telemetry, events, command subscription, LWT, and reconnect behavior
- Local Mosquitto Docker Compose configuration
- **Local SQLite history for telemetry samples, events, commands, and PID changes**
- **JSONL event log writer**
- **Async bounded recorder preventing scan-loop blocking**

HTTP/SSE, the web dashboard, CI, authentication, TLS setup, and real equipment
communication are outside the current implementation.

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

## Storage and History

When storage is enabled (the default since Stage 07), the runtime writes:

- **SQLite database** at `data/history.db` — telemetry samples, events,
  commands, and PID changes.
- **JSONL event log** at `logs/events.jsonl` — one JSON object per line for
  every runtime event.

Both files are created automatically on first run. They are excluded from Git
(see `.gitignore`).

To clean local runtime artifacts:

```powershell
# Windows
.\scripts\clean.ps1
```

```bash
# Linux/macOS
./scripts/clean.sh
```

**Warning:** `data/*.db` and `logs/*.jsonl` are not tracked by Git. Back them
up manually if you need to preserve history.

See [Storage History](docs/storage_history.md) for table schemas, retention
policy, inspection commands, and the integration design.

## Configuration

The default JSON file is `configs/default.json`. Unknown JSON fields and extra
JSON values are rejected. Validation covers PLC timing, MQTT connection fields
and topic shape, loop names and modes, PID/process ranges, and conditional
storage paths.

Storage is enabled by default. Set `"enabled": false` in the `storage` section
to run without persistence.

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

1. Foundation, PID, simulator, PLC runtime, MQTT, and local storage: completed.
2. Next: `stage-08-local-web-dashboard` — embedded HTTP/SSE dashboard.
3. Later: portfolio polish, export CLI, and CI pipeline.

Development conventions and release checks are described in the
[Developer Guide](docs/developer_guide.md).
