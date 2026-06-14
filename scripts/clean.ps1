param(
    [switch]$DryRun,
    [switch]$All
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

function Remove-Target {
    param([string]$Path)
    if (Test-Path $Path) {
        if ($DryRun) {
            Write-Host "  [dry-run] would remove: $Path"
        } else {
            Remove-Item -LiteralPath $Path -Recurse -Force
            Write-Host "  removed: $Path"
        }
    }
}

function Remove-Glob {
    param([string]$Pattern)
    Get-ChildItem -Path $repoRoot -Filter $Pattern -File -ErrorAction SilentlyContinue |
        ForEach-Object {
            if ($DryRun) {
                Write-Host "  [dry-run] would remove: $($_.FullName)"
            } else {
                Remove-Item -LiteralPath $_.FullName -Force
                Write-Host "  removed: $($_.FullName)"
            }
        }
}

function Remove-RuntimeFiles {
    param([string]$Dir)
    Get-ChildItem $Dir -File -ErrorAction SilentlyContinue |
        Where-Object { $_.Name -ne ".gitkeep" } |
        ForEach-Object {
            if ($DryRun) {
                Write-Host "  [dry-run] would remove: $($_.FullName)"
            } else {
                Remove-Item -LiteralPath $_.FullName -Force
                Write-Host "  removed: $($_.FullName)"
            }
        }
}

if ($DryRun) { Write-Host "Dry-run mode: no files will be deleted." }

Write-Host "Cleaning build artifacts..."
Remove-Target (Join-Path $repoRoot "dist")
Remove-Target (Join-Path $repoRoot "bin")
Remove-Target (Join-Path $repoRoot "release")
Remove-Glob "vplc.exe"
Remove-Glob "vplc"

Write-Host "Cleaning test artifacts..."
Remove-Glob "coverage.out"
Remove-Glob "cov.out"
Remove-Glob "*.out"
Remove-Glob "*.test"
Remove-Glob "*.test.exe"

Write-Host "Cleaning temp directories..."
Remove-Target (Join-Path $repoRoot "tmp")
Remove-Target (Join-Path $repoRoot "temp")

Write-Host "Cleaning runtime data..."
Remove-RuntimeFiles (Join-Path $repoRoot "data")
Remove-RuntimeFiles (Join-Path $repoRoot "logs")

if ($All) {
    Write-Host "Cleaning Go test cache..."
    if (-not $DryRun) { go clean -testcache }
    else { Write-Host "  [dry-run] would run: go clean -testcache" }

    Write-Host "Cleaning Go build cache..."
    if (-not $DryRun) { go clean -cache }
    else { Write-Host "  [dry-run] would run: go clean -cache" }
}

Write-Host "Clean complete."
