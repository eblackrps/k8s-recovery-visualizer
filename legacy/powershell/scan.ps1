[CmdletBinding()]
param(
    [string]$OutDir = ".\out"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

try {
    # Always run from repo root (script folder)
    if ($PSScriptRoot) { Set-Location $PSScriptRoot }

    $runner = Join-Path (Get-Location) "Invoke-DrScan.ps1"
    if (-not (Test-Path $runner)) {
        throw "Missing Invoke-DrScan.ps1 in repo root. Expected: $runner"
    }

    & pwsh -ExecutionPolicy Bypass -File $runner -OutDir $OutDir
    exit $LASTEXITCODE
} catch {
    Write-Host ("scan.ps1 FAILED: " + $_.Exception.Message) -ForegroundColor Red
    exit 5
}
