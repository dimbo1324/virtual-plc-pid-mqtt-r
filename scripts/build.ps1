$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

New-Item -ItemType Directory -Force -Path dist | Out-Null
go build -o dist/vplc.exe ./cmd/vplc
