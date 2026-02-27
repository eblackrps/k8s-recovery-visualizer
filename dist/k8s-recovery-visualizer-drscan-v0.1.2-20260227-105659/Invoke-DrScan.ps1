[CmdletBinding()]
param(
    [string]$OutDir = ".\out"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Ensure-Dir {
    param([string]$Path)
    if (-not (Test-Path $Path)) { New-Item -ItemType Directory -Path $Path | Out-Null }
}

function Run-Step {
    param(
        [Parameter(Mandatory)][string]$Name,
        [Parameter(Mandatory)][string]$File,
        [Parameter(Mandatory)][string[]]$Args
    )

    Write-Host ""
    Write-Host ("==> " + $Name) -ForegroundColor Cyan

    if (-not (Test-Path $File)) {
        throw "Missing script: $File"
    }

    $cmd = @("-ExecutionPolicy","Bypass","-File",$File) + $Args
    & pwsh @cmd
    $code = $LASTEXITCODE

    if ($code -ne 0) {
        throw "$Name failed (exit $code)."
    }

    Write-Host ("OK: " + $Name) -ForegroundColor Green
}

try {
    Ensure-Dir -Path $OutDir

    # Paths
    $workloadJson = Join-Path $OutDir "workload-report.json"
    $workloadLog  = Join-Path $OutDir "workload-log.txt"

    $storageJson = Join-Path $OutDir "storage-score.json"
    $storageLog  = Join-Path $OutDir "storage-score-log.txt"

    $backupJson = Join-Path $OutDir "backup-readiness.json"
    $backupLog  = Join-Path $OutDir "backup-readiness-log.txt"

    $networkJson = Join-Path $OutDir "network-readiness.json"
    $networkLog  = Join-Path $OutDir "network-readiness-log.txt"

    $portJson = Join-Path $OutDir "portability.json"
    $portLog  = Join-Path $OutDir "portability-log.txt"

    $reportJson = Join-Path $OutDir "drscan-report.json"
    $reportLog  = Join-Path $OutDir "drscan-report-log.txt"

    $mdPath   = Join-Path $OutDir "drscan-report.md"
    $htmlPath = Join-Path $OutDir "drscan-report.html"
    $renderLog = Join-Path $OutDir "drscan-render-log.txt"

    # Run collectors
    Run-Step -Name "Collect Workloads" -File ".\Collect-Workloads.ps1" -Args @(
        "-OutputPath",$workloadJson,
        "-LogPath",$workloadLog
    )

    Run-Step -Name "Score Storage" -File ".\Score-Storage.ps1" -Args @(
        "-OutputPath",$storageJson,
        "-LogPath",$storageLog
    )

    Run-Step -Name "Detect Backup Readiness" -File ".\Detect-BackupReadiness.ps1" -Args @(
        "-OutputPath",$backupJson,
        "-LogPath",$backupLog
    )

    Run-Step -Name "Assess Network Readiness" -File ".\Assess-NetworkReadiness.ps1" -Args @(
        "-OutputPath",$networkJson,
        "-LogPath",$networkLog
    )

    Run-Step -Name "Assess Portability" -File ".\Assess-Portability.ps1" -Args @(
        "-OutputPath",$portJson,
        "-LogPath",$portLog
    )

    # Merge
    Run-Step -Name "Build Unified DR Report" -File ".\Build-DrReport.ps1" -Args @(
        "-OutDir",$OutDir,
        "-OutputPath",$reportJson,
        "-LogPath",$reportLog
    )

    # Render
    Run-Step -Name "Render MD/HTML" -File ".\Render-DrReport.ps1" -Args @(
        "-InputPath",$reportJson,
        "-MdPath",$mdPath,
        "-HtmlPath",$htmlPath,
        "-LogPath",$renderLog
    )

    Write-Host ""
    Write-Host "DR Scan complete." -ForegroundColor Green
    Write-Host ("JSON : " + $reportJson) -ForegroundColor Yellow
    Write-Host ("MD   : " + $mdPath) -ForegroundColor Yellow
    Write-Host ("HTML : " + $htmlPath) -ForegroundColor Yellow
    exit 0
} catch {
    Write-Host ""
    Write-Host ("DR Scan FAILED: " + $_.Exception.Message) -ForegroundColor Red
    exit 5
}
