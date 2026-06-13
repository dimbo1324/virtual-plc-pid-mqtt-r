# Pre-storage Notes

Stage 07 will add local history without changing PID mathematics, process
simulation, PLC ownership, or MQTT parsing. This document records the current
extension points and constraints; no storage implementation exists yet.

## Current Data Flow

`pkg/plc.Runtime` produces two bounded streams:

- `Snapshots()`: point-in-time PLC and loop telemetry after scans;
- `Events()`: lifecycle, command, tuning, disturbance, reset, and overrun events.

The latest snapshot is also available through `Snapshot()` as a deep copy.
MQTT commands are parsed in `pkg/mqttx`, converted to `plc.Command`, and passed
through the handler assembled in `internal/app`.

## Storage Hook Points

Storage should be assembled in `internal/app` and implemented under
`internal/storage`.

- Runtime events should pass through the existing application-owned event
  fan-out before reaching MQTT or storage.
- Snapshot persistence may consume `Runtime.Snapshots()` through one
  application-owned reader, then fan out to storage and future UI consumers.
- Command persistence should wrap the typed command handler in `internal/app`
  so the request and result are recorded without coupling MQTT to storage.

Go channels are queues, not broadcast buses. MQTT, storage, and future web code
must not independently compete for values from the same runtime channel.

## Planned Data Categories

- telemetry samples derived from snapshots;
- PLC and application events;
- typed commands with source, timestamp, outcome, and error;
- PID setpoint, gain, and mode changes.

## Constraints for Stage 07

1. Storage failure must not stop or delay the PLC scan loop.
2. Persistence queues must be bounded with an explicit overflow policy.
3. Database writes must run outside the PLC mutex and scan goroutine.
4. Stored snapshots and event detail maps must be copied before asynchronous use.
5. Shutdown must drain or explicitly abandon queued records within a timeout.
6. Retention and batching must be configurable and testable.
7. MQTT, PID, simulator, and PLC packages must not import storage.
8. Secrets and MQTT credentials must never enter command/event payload history.

The current runtime channels intentionally prefer PLC availability over
guaranteed delivery when consumers are slow. Stage 07 must define whether
storage requires loss accounting, sampling, or a dedicated bounded queue.
