$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

Write-Host "==> Verifying modules..."
go mod verify

Write-Host "==> Running vet..."
go vet ./...

Write-Host "==> Building binary..."
New-Item -ItemType Directory -Force -Path dist | Out-Null
go build -o dist/vplc.exe ./cmd/vplc

Write-Host "==> Smoke: version check..."
./dist/vplc.exe --version

Write-Host "==> Smoke: config validation..."
./dist/vplc.exe --validate-config --config configs/default.json

Write-Host "==> Running smoke tests..."
go test -run TestSmoke -v ./tests/smoke/...

Write-Host "All smoke checks passed."
