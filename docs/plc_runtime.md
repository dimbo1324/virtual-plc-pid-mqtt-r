# PLC Runtime

`pkg/plc` is a reusable virtual PLC runtime. It connects each configured
`pkg/pid.Controller` to one `pkg/simulator.Process` and owns their concurrent
scan lifecycle.

## Lifecycle

The runtime supports `stopped`, `starting`, `running`, `stopping`, and
`faulted` states. `Start` is non-blocking. `Stop` cancels the scan goroutine and
waits for it through the caller's context. Repeated start and stop calls are
safe.

## Scan Cycle

Each scan uses elapsed wall-clock time since the previous tick, rather than
assuming an ideal interval:

1. Read the current process value.
2. Update the PID controller with the measured `dt`.
3. Apply PID output (`MV`) to the process.
4. Advance the process model by the same `dt`.
5. Build an immutable snapshot and increment the scan counter.
6. Send the snapshot to a bounded, non-blocking channel.
7. Emit `plc_scan_overrun` when execution reaches the configured warning.

This is a simulator on a general-purpose operating system and is not a
hard-real-time controller.

## Loop Modes

- `auto`: PID calculates output.
- `manual`: the stored manual output is applied.
- `hold`: the previous output is retained.
- `disabled`: the PID minimum output is used as the safe value.

## Snapshots and Concurrency

`Snapshot` contains PLC status and a map of loop engineering values (`SP`,
`PV`, `MV`, error, mode, quality, and gains). The runtime protects mutable
controllers, process models, lifecycle state, and the latest snapshot with an
`RWMutex`. `Snapshot()` returns a deep copy of its map.

Events and scan snapshots use bounded channels. Slow consumers cannot block
the scan indefinitely; messages are dropped when a channel buffer is full.
These channels are queues rather than broadcast subscriptions. Application
wiring must own each read and fan values out to MQTT, storage, or future UI
consumers without making them compete for the same channel.

## Commands

`ApplyCommand` supports:

- `set_setpoint`
- `set_pid_gains`
- `set_mode`
- `set_manual_output`
- `inject_disturbance`
- `reset_loop`
- `start_plc`
- `stop_plc`

Commands validate loop names, finite numeric values, setpoint limits, modes,
PID gains, and disturbance duration. Rejected commands do not panic and emit a
`command_rejected` event. Applied commands emit their domain event and a
`command_applied` event.

## Current Boundary

The runtime has no MQTT, HTTP, web, storage, or logging dependency. Persistent
history and a web dashboard are intentionally not part of this stage.
The planned storage boundary is documented in
[`pre_storage_notes.md`](pre_storage_notes.md).
