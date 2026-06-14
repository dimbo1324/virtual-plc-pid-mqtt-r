# Real Data Readiness

The application currently uses a built-in first-order process simulator. This document
describes the architecture that makes replacing it with real field data straightforward.

## pkg/input — the Provider interface

`pkg/input` defines a minimal interface that any data source must implement:

```go
type Provider interface {
    Name()  string
    Read(ctx context.Context) ([]TagValue, error)
    Close() error
}
```

`TagValue` carries the tag name, a float64 value, a `Quality` (Good/Uncertain/Bad), and
a timestamp. The interface is deliberately narrow so adapters for OPC-UA, Modbus RTU/TCP,
REST APIs, or industrial historians can be added without touching the PLC runtime.

## SyntheticProvider (current)

`pkg/input.SyntheticProvider` implements `Provider` by calling a user-supplied snapshot
function. In the current wiring this function reads the PLC runtime's own simulated
output — which is why it is called "synthetic". It is used as a reference implementation
and for tests.

```go
p := input.NewSyntheticProvider("sim", func() map[string]float64 {
    snap := runtime.Snapshot()
    m := make(map[string]float64, len(snap.Loops))
    for name, loop := range snap.Loops {
        m[name+".pv"] = loop.ProcessValue
    }
    return m
})
```

## Adding a real adapter

1. Create a new package under `pkg/input/` (e.g. `pkg/input/opcua/`).
2. Implement `input.Provider` — connect in the constructor, read tags in `Read`, disconnect in `Close`.
3. In `internal/app/lifecycle.go`, construct your provider and pass its `Read` output into the
   PLC loop's setpoint / process-value injection point (to be wired in a future stage).
4. Wire the provider's `Close()` into the shutdown sequence alongside MQTT and storage.

## Quality propagation

`TagValue.Quality` maps naturally to `pkg/simulator.Quality`:

| `input.Quality` | Meaning |
|---|---|
| `QualityGood` | Fresh reading from the field device |
| `QualityUncertain` | Device responded but value may be stale |
| `QualityBad` | Comms lost; use last-known value or safe default |

The PLC runtime already surfaces `quality` in `LoopSnapshot` and broadcasts it via SSE,
so the UI will reflect real quality states without further changes.

## Configuration profiles

`configs/demo.json` — MQTT and storage disabled, fast scan rate, browser auto-open. Ideal
for demonstration without any external dependencies.

`configs/no-storage.json` — MQTT enabled, storage disabled. Useful when SQLite is not
available on the target machine.

`configs/no-mqtt.json` — Storage enabled with fallback, MQTT disabled. For environments
without an MQTT broker.
