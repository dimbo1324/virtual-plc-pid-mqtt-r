# virtual-plc-pid-mqtt-r

A lightweight Go foundation for a reusable virtual PLC, PID, process simulation,
MQTT telemetry, local history, and embedded dashboard project.

## Safety Notice

This software is a virtual PLC / PID / MQTT simulator.
It is not intended to control real industrial equipment.
Do not connect it directly to real actuators or safety-critical systems.

## Current Status

Stage 03/04: PID core verified and process simulator package implemented.
PLC runtime, MQTT, storage and web UI are planned but not implemented yet.

## Architecture Overview

The project targets a single Go binary. The entry point loads and validates JSON
configuration, initializes standard-library logging, and starts the application
foundation. Future stages will add independent PID, PLC, simulator, MQTT,
storage, and web components.

## Repository Structure

```text
cmd/vplc/       CLI entry point
internal/app/   application lifecycle foundation and version
internal/config JSON configuration loading and validation
internal/logging standard-library logger setup
pkg/            documented placeholders for reusable future packages
configs/        default JSON configuration
docs/           project specification and supporting documentation
scripts/        PowerShell and shell helper commands
```

## Requirements

- Go 1.25.5 or later for the current module.
- Go 1.26 is the target baseline when it becomes available in the development environment.
- No external services are required for Stage 03/04.

## Quick Start

```bash
go run ./cmd/vplc --version
go run ./cmd/vplc --validate-config --config configs/default.json
go run ./cmd/vplc --config configs/default.json
```

## Configuration

The default configuration is `configs/default.json`. It defines the future
application, PLC, MQTT, web, storage, and pressure/temperature/level loop shape.
Stage 01 loads and validates this data but does not activate those subsystems.

## Development Commands

```bash
gofmt -w .
go test ./...
go vet ./...
go build -o dist/vplc ./cmd/vplc
```

Windows build:

```powershell
go build -o dist\vplc.exe .\cmd\vplc
```

## Roadmap

1. Stage 01: project foundation completed.
2. Stage 02: reusable PID package implemented.
3. Stage 03/04: PID core verified and process simulator implemented.
4. Next: PLC runtime connecting PID controllers to process models.
5. Later stages: MQTT, storage, embedded web UI, and portfolio polish.

## Portfolio Summary

The finished project will demonstrate Go backend and edge development,
industrial automation concepts, PID control, synthetic process modeling, MQTT,
local telemetry history, and a compact operator dashboard without a complex
deployment stack.
