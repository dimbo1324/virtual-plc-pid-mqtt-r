# Storage Fallback

By default the application fails to start if the SQLite database cannot be opened.
The fallback mechanism lets it continue running when storage is unavailable, at the cost of
losing historical telemetry.

## Configuration

Add to `storage` in your config JSON:

```json
"storage": {
    "enabled": true,
    "type": "sqlite",
    "sqlite_path": "/app/data/history.db",
    "events_jsonl_path": "/app/logs/events.jsonl",
    "fallback_on_error": true,
    "fallback_type": "jsonl"
}
```

### Fields

| Field | Type | Default | Description |
|---|---|---|---|
| `fallback_on_error` | bool | `false` | When `true`, a SQLite open error does not abort startup |
| `fallback_type` | string | `""` | `"jsonl"` — write events to JSONL only; `"noop"` — no storage at all |

`fallback_type` is required (and validated) only when `fallback_on_error` is `true`.

## Modes

| `storage_mode` | Meaning |
|---|---|
| `ok` | SQLite + JSONL both active; full telemetry recorded |
| `degraded` | SQLite failed; events written to JSONL only; telemetry lost |
| `disabled` | Storage not enabled in config |

The current mode is reported by `GET /api/status` in the `storage_mode` field.
The web UI header shows a blinking **STORAGE DEGRADED** badge in degraded mode.

## What is lost in degraded mode

- Historical telemetry samples (PV/SP/MV time series)
- Command records
- PID tuning change history
- The `/api/telemetry/recent` and `/api/events/recent` endpoints will return empty results

What is **not** lost:

- Real-time SSE snapshots (always served from in-memory PLC state)
- MQTT publishing (continues independently)
- Event log lines written to the JSONL file

## Disabling fallback

Set `fallback_on_error: false` (or omit it). The application will refuse to start if SQLite
cannot be opened, which is the safest default for production.
