# Storage and History — Stage 07

This document describes the local storage layer added in Stage 07.

## What is stored

| Category | Destination |
|---|---|
| Telemetry samples (SP/PV/MV per loop per scan) | SQLite `telemetry_samples` table |
| Application and runtime events | SQLite `events` table + `logs/events.jsonl` |
| Commands (applied and rejected) | SQLite `commands` table |
| PID tuning changes | SQLite `pid_changes` table |

## File paths

| File | Default path | Purpose |
|---|---|---|
| SQLite database | `data/history.db` | All historical records |
| JSONL event log | `logs/events.jsonl` | Human-readable event stream |

Both paths are configurable in `configs/default.json` under the `storage` section.

## Tables

### `schema_migrations`

Tracks which migration versions have been applied. Migrations are idempotent.

### `telemetry_samples`

One row per control loop per PLC scan. Fields: `timestamp`, `device_id`,
`scan_counter`, `loop_name`, `sp`, `pv`, `mv`, `error`, `mode`, `quality`,
`unit`.

### `events`

One row per runtime or application event. Fields: `timestamp`, `level`,
`event_type`, `message`, `details_json`.

### `commands`

One row per received command (whether applied or rejected). Fields:
`timestamp`, `command_id`, `source`, `command_type`, `loop_name`,
`payload_json`, `status`, `error_message`.

### `pid_changes`

One row per `set_pid_gains` command that was successfully applied. Fields:
`timestamp`, `loop_name`, `old_kp`, `old_ki`, `old_kd`, `new_kp`, `new_ki`,
`new_kd`, `source`, `command_id`.

Note: `old_kp`, `old_ki`, `old_kd` are currently stored as `NULL` because the
PLC runtime does not expose the previous gains in the command result event.
This is tracked as future work.

## Retention

Telemetry retention is configured via `retention_max_samples`. When the row
count exceeds this limit, the oldest rows are deleted:

```sql
DELETE FROM telemetry_samples
WHERE id NOT IN (
    SELECT id FROM telemetry_samples ORDER BY id DESC LIMIT <N>
);
```

Retention runs in the background recorder worker every 5 minutes. It applies
only to `telemetry_samples`; events, commands and PID changes are not pruned
by default.

## How storage integrates with the app

```
internal/app.RunRuntime
  ├─ storage.Open        → opens DB, runs migrations
  ├─ storage.NewJSONLWriter → opens events.jsonl
  ├─ storage.NewRecorder → bounded async queue (write_queue_size)
  ├─ rec.Start(ctx)      → launches worker goroutine
  │
  ├─ consumeSnapshots goroutine
  │    └─ reads runtime.Snapshots() → rec.RecordSnapshot()
  │
  ├─ forwardRuntimeEvents goroutine
  │    └─ reads runtime.Events() → rec.RecordEvent() + MQTT publish
  │
  └─ MQTT command handler wrapper
       └─ runtime.ApplyCommand() → rec.RecordCommand() + rec.RecordPIDChange()
```

The recorder uses a bounded channel (default 256 entries). If the queue is
full when a submission arrives, the submission is silently dropped and a
warning is logged. This ensures the PLC scan loop is never blocked by slow
storage writes.

## Why storage is in `internal/storage`

Storage is application-specific infrastructure — it depends on the application
configuration and couples to the SQLite driver. Placing it in `internal/`
prevents accidental import by other projects and keeps `pkg/*` reusable.

## Why `pkg/plc` does not depend on storage

`pkg/plc` must remain a pure runtime module with no knowledge of databases,
files, or application infrastructure. The separation ensures:

- `pkg/plc` tests do not need a database.
- The PLC runtime can be reused in projects without SQLite.
- Storage errors cannot propagate into the scan loop.

Integration is done through `internal/app`, which owns the fan-out between
PLC channels and downstream consumers (MQTT, storage, future web UI).

## Inspecting the database manually

With `sqlite3` installed:

```bash
sqlite3 data/history.db

# List tables
.tables

# Recent telemetry
SELECT timestamp, loop_name, sp, pv, mv FROM telemetry_samples ORDER BY id DESC LIMIT 20;

# Recent events
SELECT timestamp, level, event_type, message FROM events ORDER BY id DESC LIMIT 20;

# Commands
SELECT timestamp, source, command_type, loop_name, status FROM commands ORDER BY id DESC LIMIT 10;

# PID changes
SELECT timestamp, loop_name, new_kp, new_ki, new_kd, source FROM pid_changes ORDER BY id DESC LIMIT 10;

# Migration history
SELECT version, name, applied_at FROM schema_migrations;

.quit
```

JSONL events can be inspected with any text editor or `jq`:

```bash
tail -20 logs/events.jsonl | jq .
```

## Storage failure policy

| Failure point | Behavior |
|---|---|
| `storage.Open` or migration fails at startup | App exits with error |
| Write failure after startup | Warning logged; PLC continues |
| JSONL write failure | Warning logged; SQLite write still attempted |
| Queue full (backpressure) | Sample/event dropped; warning logged |
| Shutdown | Recorder drains queue for up to 3 seconds |

## Not implemented in Stage 07

- Log rotation for `events.jsonl` (future work).
- Per-loop retention limits.
- Old gains in PID change records (recorded as NULL for now).
- Export CLI (`vplc --export`).
- Web dashboard reading from storage directly.
- Retention for events, commands and pid_changes tables.
