[CmdletBinding()]
param(
    [string]$OutDir = ".\out",
    [switch]$Minimal,
    [switch]$NoRender,
    [switch]$JsonOnly
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Ensure-Dir {
    param([Parameter(Mandatory)][string]$Path)
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

    & pwsh -ExecutionPolicy Bypass -File $File @Args
    $code = $LASTEXITCODE

    if ($code -ne 0) {
        throw "$Name failed (exit $code)."
    }

    Write-Host ("OK: " + $Name) -ForegroundColor Green
}

function Fail([string]$msg, [int]$code = 5) {
    Write-Host ""
    Write-Host ("DR Scan FAILED: " + $msg) -ForegroundColor Red
    exit $code
}

try {
    if ($PSScriptRoot) { Set-Location $PSScriptRoot }

    if ($JsonOnly) { $NoRender = $true }

    Ensure-Dir -Path $OutDir

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

    $mdPath     = Join-Path $OutDir "drscan-report.md"
    $htmlPath   = Join-Path $OutDir "drscan-report.html"
    $renderLog  = Join-Path $OutDir "drscan-render-log.txt"

    $histLog    = Join-Path $OutDir "history-log.txt"

    Write-Host "Invoke-DrScan Starting..." -ForegroundColor Cyan
    Write-Host ("OutDir: " + (Resolve-Path $OutDir)) -ForegroundColor DarkGray
    if ($Minimal) { Write-Host "Mode: Minimal" -ForegroundColor Yellow }
    elseif ($NoRender) { Write-Host "Mode: NoRender" -ForegroundColor Yellow }
    else { Write-Host "Mode: Full" -ForegroundColor Yellow }

    Run-Step -Name "Collect Workloads" -File ".\Collect-Workloads.ps1" -Args @(
        "-OutputPath",$workloadJson,
        "-LogPath",$workloadLog
    )

    Run-Step -Name "Score Storage" -File ".\Score-Storage.ps1" -Args @(
        "-OutputPath",$storageJson,
        "-LogPath",$storageLog
    )

    if (-not $Minimal) {
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
    } else {
        "{}" | Out-File $backupJson -Encoding UTF8
        "{}" | Out-File $networkJson -Encoding UTF8
        "{}" | Out-File $portJson -Encoding UTF8
    }

    $buildArgs = @("-OutDir",$OutDir,"-OutputPath",$reportJson,"-LogPath",$reportLog)
    if ($Minimal) { $buildArgs += "-Minimal" }
    Run-Step -Name "Build Unified DR Report" -File ".\Build-DrReport.ps1" -Args $buildArgs

    # NEW: History/Trend
    Run-Step -Name "Update History/Trend" -File ".\Update-History.ps1" -Args @(
        "-ReportPath",$reportJson,
        "-OutDir",$OutDir,
        "-LogPath",$histLog
    )

    if (-not $NoRender) {
        Run-Step -Name "Render MD/HTML" -File ".\Render-DrReport.ps1" -Args @(
            "-InputPath",$reportJson,
            "-MdPath",$mdPath,
            "-HtmlPath",$htmlPath,
            "-LogPath",$renderLog
        )
    } else {
        Write-Host ""
        Write-Host "Skipping render (NoRender/JsonOnly enabled)." -ForegroundColor Yellow
    }

    Write-Host ""
    Write-Host "DR Scan complete." -ForegroundColor Green
    Write-Host ("JSON : " + $reportJson) -ForegroundColor Yellow
    if (-not $NoRender) {
        Write-Host ("MD   : " + $mdPath) -ForegroundColor Yellow
        Write-Host ("HTML : " + $htmlPath) -ForegroundColor Yellow
    }

    exit 0
} catch {
    $msg = $_.Exception.Message
    if ($msg -match "forbidden" -or $msg -match "Unauthorized") { Fail "RBAC/Auth issue: $msg" 6 }
    if ($msg -match "Missing script") { Fail "Packaging error: $msg" 7 }
    Fail $msg 5
}
