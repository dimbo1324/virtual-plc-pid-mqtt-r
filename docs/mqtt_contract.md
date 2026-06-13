# MQTT Contract

`pkg/mqttx` exposes PLC snapshots, status, and events through MQTT and converts
JSON command messages into typed `plc.Command` values. It uses Eclipse Paho and
does not depend on application config, storage, or web packages.

## Topics

For base topic `vplc/vplc_001`:

| Purpose | Topic |
|---|---|
| Retained status and LWT | `vplc/vplc_001/status` |
| PLC snapshots | `vplc/vplc_001/telemetry` |
| Runtime and command events | `vplc/vplc_001/events` |
| Incoming commands | `vplc/vplc_001/commands` |
| Reserved configuration | `vplc/vplc_001/config` |

## Status

Successful connections publish retained `online` status. The Last Will is a
retained `offline` payload with reason `unexpected_disconnect`. Graceful
shutdown publishes `offline` with reason `graceful_shutdown` before disconnect.

```json
{
  "device_id": "vplc_001",
  "status": "online",
  "timestamp": "2026-06-13T14:10:00Z"
}
```

## Telemetry

Telemetry is the JSON representation of `plc.Snapshot` and includes the UTC
timestamp, device ID, scan status, and all configured loops.

```json
{
  "timestamp": "2026-06-13T14:10:00Z",
  "device_id": "vplc_001",
  "plc": {
    "state": "running",
    "scan_interval_ms": 500,
    "last_scan_duration_ms": 1.2,
    "scan_counter": 42
  },
  "loops": {
    "pressure": {
      "name": "pressure",
      "display_name": "Pressure",
      "unit": "bar",
      "sp": 6,
      "pv": 5.8,
      "mv": 61,
      "error": 0.2,
      "mode": "auto",
      "quality": "good",
      "enabled": true,
      "kp": 3,
      "ki": 0.25,
      "kd": 0.05
    }
  }
}
```

## Commands

Setpoint:

```json
{"command_id":"cmd-1","command":"set_setpoint","loop":"pressure","value":7.5}
```

PID gains:

```json
{"command_id":"cmd-2","command":"set_pid_gains","loop":"pressure","kp":2.0,"ki":0.4,"kd":0.05}
```

Manual mode:

```json
{"command_id":"cmd-3","command":"set_mode","loop":"temperature","mode":"manual","manual_output":40}
```

Disturbance:

```json
{"command_id":"cmd-4","command":"inject_disturbance","loop":"pressure","value":-1.5,"duration_seconds":30}
```

Start and stop:

```json
{"command_id":"cmd-5","command":"stop_plc"}
```

```json
{"command_id":"cmd-6","command":"start_plc"}
```

Unknown fields, malformed JSON, unsupported commands, missing required values,
non-finite values, and payloads larger than 64 KiB are rejected. MQTT-originated
commands have source `mqtt`; applied and rejected outcomes are published to the
events topic.

`command_id` is optional at the parser boundary for simple manual demos, but it
is strongly recommended for auditability and future command persistence.

## Local Broker Demo

```bash
docker compose -f docker/docker-compose.yml up -d
go run ./cmd/vplc --run --config configs/default.json
mosquitto_sub -h localhost -t "vplc/#" -v
mosquitto_pub -h localhost -t "vplc/vplc_001/commands" -m '{"command_id":"cmd-1","command":"set_setpoint","loop":"pressure","value":7.5}'
docker compose -f docker/docker-compose.yml down
```

Normal unit tests do not require a broker. The optional integration test runs
only when `VPLC_RUN_MQTT_TESTS=1` is set.

The bundled Mosquitto configuration permits anonymous local access and has no
TLS. It is for local demonstration only and must not be exposed as a production
broker configuration.
