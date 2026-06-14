$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

$unformatted = gofmt -l .
if ($unformatted) {
    Write-Host "Formatting files:"
    $unformatted | ForEach-Object { Write-Host "  $_" }
    gofmt -w .
    Write-Host "Done."
} else {
    Write-Host "All files are already formatted."
}
