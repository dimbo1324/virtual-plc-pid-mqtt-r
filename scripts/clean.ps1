$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

Remove-Item -LiteralPath (Join-Path $repoRoot "dist") -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath (Join-Path $repoRoot "bin") -Recurse -Force -ErrorAction SilentlyContinue
