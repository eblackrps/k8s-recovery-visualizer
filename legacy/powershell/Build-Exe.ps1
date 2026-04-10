[CmdletBinding()]
param(
    [string]$SourceScript = ".\DrScan-Launcher.ps1",
    [string]$OutDir = ".\dist"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Ensure-Dir([string]$p) {
    if (-not (Test-Path $p)) { New-Item -ItemType Directory -Path $p | Out-Null }
}

function Get-LatestTag {
    $t = (& git tag --sort=-creatordate) 2>$null
    if ($LASTEXITCODE -ne 0) { return "v0.0.0" }
    $first = @($t)[0]
    if (-not $first) { return "v0.0.0" }
    return ($first | Out-String).Trim()
}

Ensure-Dir $OutDir

if (-not (Test-Path $SourceScript)) {
    throw "Missing source script: $SourceScript"
}

# Check for ps2exe module
$ps2exe = Get-Module -ListAvailable -Name ps2exe
if (-not $ps2exe) {
    Write-Host "ps2exe module not found." -ForegroundColor Yellow
    Write-Host "Install it (run PowerShell as Admin):" -ForegroundColor Yellow
    Write-Host "  Install-Module ps2exe -Scope AllUsers" -ForegroundColor Cyan
    Write-Host "Then rerun:" -ForegroundColor Yellow
    Write-Host "  pwsh -ExecutionPolicy Bypass -File .\Build-Exe.ps1" -ForegroundColor Cyan
    exit 3
}

Import-Module ps2exe -ErrorAction Stop

$ver = Get-LatestTag
$exePath = Join-Path $OutDir ("drscan-launcher-" + $ver + ".exe")

# Build EXE (console ON so users see output)
Invoke-ps2exe -InputFile $SourceScript -OutputFile $exePath -ErrorAction Stop

Write-Host "EXE created:" -ForegroundColor Green
Write-Host $exePath -ForegroundColor Yellow
exit 0
