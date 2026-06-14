# CI Pipeline

## Workflow

The CI workflow lives at `.github/workflows/ci.yml` and runs on every push to `main` or `stage/**` branches, and on every pull request targeting `main`.

## Jobs

### `ci` — Test & Build (ubuntu-latest)

| Step | Command | Fails on |
|------|---------|----------|
| Checkout | `actions/checkout@v4` | — |
| Set up Go | `actions/setup-go@v5` (from `go.mod`) | missing installer |
| Verify modules | `go mod verify` | tampered module |
| Check formatting | `gofmt -l .` | unformatted files |
| go vet | `go vet ./...` | static analysis errors |
| Run tests | `go test -race -count=1 ./...` | test failure or race |
| Build binary | `go build -o /dev/null ./cmd/vplc` | compile error |
| Validate config | `go run ./cmd/vplc --validate-config ...` | config schema error |
| Check vulnerabilities | `govulncheck ./...` | **non-blocking** (`continue-on-error: true`) |

## Go Version

The workflow uses `go-version-file: go.mod` so CI always matches the module's declared minimum Go version. No manual version maintenance required.

## What CI Does Not Test

- MQTT broker integration (requires `VPLC_RUN_MQTT_TESTS=1` and a broker)
- Web dashboard UI (requires a browser)
- Storage with a populated database

These are covered by local development workflows documented in [testing.md](testing.md).

## Adding a New CI Check

Add a new step to `.github/workflows/ci.yml` after the existing test step. Keep the workflow linear — a single job is sufficient for this project size.

## Local CI Simulation

Run the same checks locally before pushing:

```bash
go mod verify
gofmt -l .
go vet ./...
go test -race -count=1 ./...
go build -o /dev/null ./cmd/vplc
go run ./cmd/vplc --validate-config --config configs/default.json
```

Or use the smoke script:

```bash
./scripts/smoke.sh      # Linux / macOS
.\scripts\smoke.ps1     # Windows
```
