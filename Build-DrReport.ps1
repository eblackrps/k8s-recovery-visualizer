[CmdletBinding()]
param(
    [string]$OutDir = ".\out",
    [string]$OutputPath = ".\out\drscan-report.json",
    [string]$LogPath = ".\out\drscan-report-log.txt"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Ensure-Dir {
    param([string]$Path)
    $dir = Split-Path -Parent $Path
    if ($dir -and -not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir | Out-Null }
}

Ensure-Dir -Path $OutputPath
Ensure-Dir -Path $LogPath

"=== Build-DrReport.ps1 START $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Encoding UTF8

function Log {
    param([string]$Msg)
    "$(Get-Date -Format o)  $Msg" | Out-File -FilePath $LogPath -Append -Encoding UTF8
}

function Fail {
    param([int]$Code, [string]$Msg, $Err = $null)
    Write-Host $Msg -ForegroundColor Red
    Log "FAIL($Code): $Msg"
    if ($null -ne $Err) {
        try { Log ("ERROR TYPE: " + $Err.GetType().FullName) } catch {}
        try {
            if ($Err.Exception -and $Err.Exception.Message) { Log ("ERROR MSG : " + $Err.Exception.Message) }
            else { Log ("ERROR MSG : " + ($Err | Out-String)) }
        } catch {}
        try { if ($Err.ScriptStackTrace) { Log ("STACK     : " + $Err.ScriptStackTrace) } } catch {}
    }
    "=== END FAIL $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Append -Encoding UTF8
    exit $Code
}

function Read-JsonFile {
    param([Parameter(Mandatory)][string]$Path)
    if (-not (Test-Path $Path)) { return $null }
    try {
        $raw = Get-Content -Path $Path -Raw -ErrorAction Stop
        if (-not $raw) { return $null }
        return ($raw | ConvertFrom-Json)
    } catch {
        Log "Failed reading JSON: $Path"
        throw
    }
}

function Score-ToTier {
    param([int]$Score)
    if ($Score -ge 95) { return "PLATINUM" }
    if ($Score -ge 85) { return "GOLD" }
    if ($Score -ge 70) { return "SILVER" }
    if ($Score -ge 55) { return "BRONZE" }
    return "CRITICAL"
}

function Pick-FirstContext {
    param($Workloads, $Storage, $Backup, $Network, $Portability)
    foreach ($o in @($Workloads, $Storage, $Backup, $Network, $Portability)) {
        if ($o -and $o.context) { return $o.context }
    }
    return $null
}

Write-Host "DR Report Builder Starting..." -ForegroundColor Cyan
Log "Starting DR report build"

try {
    $workloads = Read-JsonFile -Path (Join-Path $OutDir "workload-report.json")
    $storage   = Read-JsonFile -Path (Join-Path $OutDir "storage-score.json")
    $backup    = Read-JsonFile -Path (Join-Path $OutDir "backup-readiness.json")
    $network   = Read-JsonFile -Path (Join-Path $OutDir "network-readiness.json")
    $port      = Read-JsonFile -Path (Join-Path $OutDir "portability.json")

    $missing = @()
    if (-not $workloads) { $missing += "workload-report.json" }
    if (-not $storage)   { $missing += "storage-score.json" }
    if (-not $backup)    { $missing += "backup-readiness.json" }
    if (-not $network)   { $missing += "network-readiness.json" }
    if (-not $port)      { $missing += "portability.json" }

    if ($missing.Count -gt 0) {
        Fail -Code 3 -Msg ("Missing required inputs in ${OutDir}: " + ($missing -join ", "))
    }

    # Pull section scores safely
    $storageScore = 0
    if ($storage.storage -and $storage.storage.score -ne $null) { $storageScore = [int]$storage.storage.score }

    $backupScore = [int]$backup.readiness.score
    $networkScore = [int]$network.readiness.score
    $portScore = [int]$port.readiness.score

    # workloads has no scoring yet: neutral 100 (light weight)
    $workloadScore = 100

    # Weighted average
    $weighted =
        (0.35 * $storageScore) +
        (0.20 * $backupScore)  +
        (0.20 * $networkScore) +
        (0.20 * $portScore)    +
        (0.05 * $workloadScore)

    $overall = [int][Math]::Round($weighted)

    $tier = (Score-ToTier -Score $overall)

    $status = "PASSED"
    if ($overall -lt 70) { $status = "AT_RISK" }
    if ($overall -lt 55) { $status = "FAILED" }

    $context = Pick-FirstContext -Workloads $workloads -Storage $storage -Backup $backup -Network $network -Portability $port

    $report = @{
        timestampUtc = (Get-Date).ToUniversalTime().ToString("o")
        context = $context
        overall = @{
            score  = $overall
            tier   = $tier
            status = $status
            weights = @{
                storage = 0.35
                backup  = 0.20
                network = 0.20
                portability = 0.20
                workloads = 0.05
            }
            components = @{
                storage     = $storageScore
                backup      = $backupScore
                network     = $networkScore
                portability = $portScore
                workloads   = $workloadScore
            }
        }
        sections = @{
            workloads   = $workloads
            storage     = $storage
            backup      = $backup
            network     = $network
            portability = $port
        }
    }

    $report | ConvertTo-Json -Depth 20 | Out-File $OutputPath -Encoding UTF8

    Write-Host "DR report build complete." -ForegroundColor Green
    Write-Host "Output: $OutputPath" -ForegroundColor Yellow
    Write-Host "Log:    $LogPath" -ForegroundColor Yellow

    Log "SUCCESS output=$OutputPath overall=$overall tier=$tier status=$status"
    "=== Build-DrReport.ps1 END $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Append -Encoding UTF8
    exit 0
} catch {
    Fail -Code 5 -Msg "DR report build failed. See log: $LogPath" -Err $_
}
