# Build and Release Guide

## Development Build

```bash
go build -o dist/vplc ./cmd/vplc    # Linux / macOS
go build -o dist\vplc.exe .\cmd\vplc  # Windows
```

Or via script:

```bash
./scripts/build.sh      # Linux / macOS
.\scripts\build.ps1     # Windows
```

## Smoke Check Before Release

Run the smoke script to verify build, config validation, and Go smoke tests:

```bash
./scripts/smoke.sh      # Linux / macOS
.\scripts\smoke.ps1     # Windows
```

The smoke script:
1. Runs `go mod verify` and `go vet ./...`
2. Builds the binary
3. Runs `--version` and `--validate-config`
4. Executes `TestSmoke*` tests in `tests/smoke/`

## Release Build

Release binaries are stripped (`-s -w`) and embed the version string via `-ldflags`.

### Linux AMD64

```bash
./scripts/release.sh 0.1.0 linux amd64
# output: release/vplc-0.1.0-linux-amd64
```

### Windows AMD64

```powershell
.\scripts\release.ps1 -Version 0.1.0 -GOOS windows -GOARCH amd64
# output: release/vplc-0.1.0-windows-amd64.exe
```

### macOS ARM64

```bash
./scripts/release.sh 0.1.0 darwin arm64
# output: release/vplc-0.1.0-darwin-arm64
```

### Cross-compile from Windows

```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o release/vplc-linux-amd64 ./cmd/vplc
```

## Output Directories

| Directory | Purpose |
|-----------|---------|
| `dist/` | Development builds (gitignored) |
| `release/` | Release artifacts (gitignored) |

## Version String

The version is defined in `internal/app/version.go`:

```go
const Version = "0.1.0"
```

Release scripts override it at build time:

```
-ldflags "-X github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/app.Version=<version>"
```

Check the embedded version:

```bash
./dist/vplc --version
```

## Module Verification

Before releasing, verify the module graph is untampered:

```bash
go mod verify
go mod tidy
git diff go.sum  # should be clean
```
