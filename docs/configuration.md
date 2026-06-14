# Configuration Reference

The application is configured with a single JSON file. Pass it with `--config path/to/file.json`.

## Profiles

| File | Purpose |
|---|---|
| `configs/default.json` | Local development — SQLite + MQTT + web on 127.0.0.1 |
| `configs/docker.json` | Docker Compose — `web.host=0.0.0.0`, `mqtt.broker_url=tcp://mosquitto:1883`, volume paths |
| `configs/demo.json` | Demo — no MQTT, no storage, fast scan, browser auto-open |
| `configs/no-storage.json` | MQTT enabled, storage disabled |
| `configs/no-mqtt.json` | Storage with fallback, MQTT disabled |

## Schema

### `app`

| Field | Type | Description |
|---|---|---|
| `name` | string | Application name (non-empty) |
| `device_id` | string | Device identifier used in MQTT topics and SSE snapshots |
| `auto_start` | bool | Start the PLC runtime immediately on launch |
| `open_browser` | bool | Open the web dashboard in the default browser on start |

### `plc`

| Field | Type | Description |
|---|---|---|
| `scan_interval_ms` | int | PLC scan cycle period in milliseconds (> 0) |
| `publish_interval_ms` | int | MQTT telemetry publish interval (> 0) |
| `ui_update_interval_ms` | int | SSE snapshot broadcast interval (> 0) |
| `scan_overrun_warning_ms` | int | Log a warning if a scan takes longer than this (> 0) |

### `mqtt`

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Enable MQTT client |
| `broker_url` | string | Broker address, e.g. `tcp://localhost:1883` |
| `client_id` | string | MQTT client identifier (unique per broker) |
| `username` / `password` | string | Optional credentials |
| `base_topic` | string | Root topic, e.g. `vplc/device_001` |
| `qos` | int | QoS level: 0, 1, or 2 |
| `connect_timeout_seconds` | int | Connection attempt timeout |
| `reconnect_interval_seconds` | int | Reconnect poll interval |

### `web`

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Enable the embedded HTTP dashboard |
| `host` | string | Listen address (`127.0.0.1` for local, `0.0.0.0` for Docker) |
| `port` | int | Listen port (1–65535) |

### `storage`

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Enable persistence |
| `type` | string | `"sqlite"` (only supported primary type) |
| `sqlite_path` | string | Path to SQLite database file |
| `events_jsonl_path` | string | Path to JSONL event log |
| `app_log_path` | string | Path to application log file |
| `retention_max_samples` | int | Maximum telemetry rows per loop (> 0) |
| `write_queue_size` | int | Async write queue depth (> 0) |
| `fallback_on_error` | bool | Continue without SQLite if it fails to open |
| `fallback_type` | string | `"jsonl"` or `"noop"` — required when `fallback_on_error` is `true` |

See [storage_fallback.md](storage_fallback.md) for full fallback details.

### `loops[]`

At least one loop is required. Each loop:

| Field | Type | Description |
|---|---|---|
| `name` | string | Unique identifier, used in API paths and MQTT topics |
| `display_name` | string | Human-readable label shown in the UI |
| `unit` | string | Engineering unit string (e.g. `bar`, `C`, `%`) |
| `enabled` | bool | Include this loop in the scan |
| `mode` | string | `"auto"`, `"manual"`, `"hold"`, or `"disabled"` |
| `setpoint` | float64 | Initial setpoint value |
| `setpoint_min/max` | float64 | Setpoint clamp range |
| `pid.kp/ki/kd` | float64 | PID gains (≥ 0) |
| `pid.bias` | float64 | PID output bias |
| `pid.output_min/max` | float64 | MV clamp range |
| `process.initial_pv` | float64 | Simulation starting value |
| `process.min/max` | float64 | Simulated process range |
| `process.gain` | float64 | Static process gain |
| `process.tau_seconds` | float64 | First-order lag time constant (> 0) |
| `process.noise_stddev` | float64 | Gaussian noise standard deviation (≥ 0) |
| `process.random_disturbances` | bool | Enable random step disturbances |

## Validate config without starting

```bash
./vplc --validate-config --config configs/docker.json
```

Exit 0 means the config is valid; exit 1 means validation failed with an error message.
