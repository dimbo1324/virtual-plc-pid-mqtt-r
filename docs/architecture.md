# Architecture

The project is designed as a single Go binary with configuration and local
runtime files. Stage 01 only establishes the executable foundation.

- `cmd/` contains the `vplc` entry point and CLI parsing.
- `internal/` contains application wiring, configuration, and logging.
- `pkg/` reserves reusable packages for PID, PLC, simulation, and MQTT concerns.
- `configs/` contains JSON runtime configuration.
- `web/` will contain assets embedded and served by Go.

PID mathematics, the PLC scan cycle, process simulation, and MQTT integration
remain separate concerns and will be implemented in later stages.
