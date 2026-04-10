[CmdletBinding()]
param(
    [ValidateSet("zip","exe","all")]
    [string]$Target = "all"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Run([string]$name, [string]$file) {
    Write-Host ""
    Write-Host ("==> " + $name) -ForegroundColor Cyan
    if (-not (Test-Path $file)) { throw "Missing: $file" }

    & pwsh -ExecutionPolicy Bypass -File $file
    $code = $LASTEXITCODE
    if ($code -ne 0) { throw "$name failed (exit $code)" }

    Write-Host ("OK: " + $name) -ForegroundColor Green
}

try {
    if ($PSScriptRoot) { Set-Location $PSScriptRoot }

    switch ($Target) {
        "zip" {
            Run "Build ZIP dist bundle" ".\Build-Dist.ps1"
        }
        "exe" {
            Run "Build EXE launcher" ".\Build-Exe.ps1"
        }
        "all" {
            Run "Build ZIP dist bundle" ".\Build-Dist.ps1"
            Run "Build EXE launcher" ".\Build-Exe.ps1"
        }
    }

    Write-Host ""
    Write-Host "Build complete." -ForegroundColor Green
    Write-Host "Check .\dist\" -ForegroundColor Yellow
    exit 0
} catch {
    Write-Host ""
    Write-Host ("Build FAILED: " + $_.Exception.Message) -ForegroundColor Red
    exit 5
}
