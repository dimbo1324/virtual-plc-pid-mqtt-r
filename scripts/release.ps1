param (
    [string]$Version = "0.1.0",
    [string]$GOOS    = "windows",
    [string]$GOARCH  = "amd64"
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

$outputDir = "release"
New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

$ext = if ($GOOS -eq "windows") { ".exe" } else { "" }
$outFile = "$outputDir/vplc-$Version-$GOOS-$GOARCH$ext"
Write-Host "==> Building release: v$Version for $GOOS/$GOARCH"

$env:GOOS   = $GOOS
$env:GOARCH = $GOARCH
go build `
    -ldflags "-X github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/app.Version=$Version -s -w" `
    -o $outFile `
    ./cmd/vplc

Write-Host "==> Release binary: $outFile"
