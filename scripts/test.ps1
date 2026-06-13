$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

$goFiles = Get-ChildItem -Path . -Recurse -Filter *.go -File | ForEach-Object FullName
if ($goFiles) {
  gofmt -w $goFiles
}
go vet ./...
go test ./...
