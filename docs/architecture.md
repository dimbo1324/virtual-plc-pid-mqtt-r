# Architecture

The project is a single Go binary with explicit package boundaries. It models
a local virtual PLC and does not communicate with real industrial equipment.

## Responsibility Map

| Path | Responsibility |
|---|---|
| `cmd/vplc` | CLI parsing, configuration loading, signal-aware entrypoint |
| `internal/app` | application assembly, lifecycle, and runtime output fan-out |
| `internal/config` | strict JSON loading and application-level validation |
| `internal/logging` | standard-library `slog` setup |
| `pkg/pid` | independent PID controller mathematics and modes |
| `pkg/simulator` | independent first-order synthetic process model |
| `pkg/plc` | scan cycle, loop ownership, commands, snapshots, and events |
| `pkg/mqttx` | MQTT topics, payloads, connection lifecycle, and command parsing |
| `docker` | local Mosquitto configuration only |
| `docs` | contracts, operating guidance, and design notes |

The intended dependency direction is:

```text
cmd/vplc -> internal/app -> internal/config + internal/logging + pkg/*
                              pkg/mqttx -> pkg/plc
                              pkg/plc   -> pkg/pid + pkg/simulator
```

Reusable packages do not import application wiring, configuration, web, or
storage packages. No global mutable runtime is used.

## Runtime Flow

```text
JSON configuration
  -> internal/app mapping
  -> PLC runtime
       -> PID update
       -> process simulation
       -> latest snapshot + bounded snapshot stream
       -> bounded event stream
  -> application-owned fan-out
       -> MQTT telemetry/events
       -> future storage sink (not implemented)
```

`pkg/plc` owns mutable loop state behind a mutex. Its output channels are
bounded and non-blocking so a slow consumer cannot halt the scan cycle. The
application owns consumption and fan-out; future services must not compete by
reading the same channel independently.

## Lifecycle

Normal CLI startup without `--run` validates and initializes the foundation,
then exits. `--run` starts the PLC, optionally connects MQTT, and waits for an
OS shutdown signal. MQTT connection failure is non-fatal. Application-owned
goroutines observe the same context and are joined during shutdown.

## Current Boundary

Implemented: foundation, configuration, PID, simulator, PLC runtime, MQTT, and
local Mosquitto Compose.

Not implemented: storage/history, JSONL, SQLite, HTTP, SSE, web dashboard, CI,
authentication, TLS configuration, or real equipment communication.

See [Pre-storage Notes](pre_storage_notes.md) for the next extension boundary.
