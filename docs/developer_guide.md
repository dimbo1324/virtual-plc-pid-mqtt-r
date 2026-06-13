# Developer Guide

Before every development stage, read
`docs/virtual-plc-pid-mqtt-r_project_specification.md` completely.

## Run

```bash
go run ./cmd/vplc --config configs/default.json
```

## Test

```bash
gofmt -w .
go test ./...
go vet ./...
```

## Build

```bash
go build -o dist/vplc ./cmd/vplc
```

Development is organized into stages. Keep each stage within its defined scope
and do not claim later functionality before it is implemented.
