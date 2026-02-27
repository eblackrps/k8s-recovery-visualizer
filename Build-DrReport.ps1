[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$OutDir,
    [string]$OutputPath = ".\out\drscan-report.json",
    [string]$LogPath    = ".\out\drscan-report-log.txt",
    [switch]$Minimal
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Ensure-Dir([string]$Path) {
    $d = Split-Path -Parent $Path
    if ($d -and -not (Test-Path $d)) { New-Item -ItemType Directory -Path $d | Out-Null }
}
Ensure-Dir $OutputPath
Ensure-Dir $LogPath

"=== Build-DrReport.ps1 START $(Get-Date -Format o) ===" | Out-File $LogPath -Encoding UTF8
function Log([string]$Msg) { ("{0}  {1}" -f (Get-Date -Format o), $Msg) | Out-File $LogPath -Append -Encoding UTF8 }

function Fail([int]$Code, [string]$Msg, [object]$Err=$null) {
    Write-Host $Msg -ForegroundColor Red
    Log "FAIL($Code): $Msg"
    if ($Err) {
        Log ("ERROR TYPE: " + $Err.GetType().FullName)
        Log ("ERROR MSG : " + ($Err.Exception.Message))
        if ($Err.ScriptStackTrace) { Log ("STACK     : " + $Err.ScriptStackTrace) }
        if ($Err.Exception.InnerException) { Log ("INNER    : " + $Err.Exception.InnerException.Message) }
    }
    "=== END FAIL $(Get-Date -Format o) ===" | Out-File $LogPath -Append -Encoding UTF8
    exit $Code
}

function Read-JsonFile([string]$Path, [switch]$Required) {
    if (-not (Test-Path $Path)) {
        if ($Required) { throw "Missing required input: $Path" }
        Log "Optional input missing: $Path"
        return $null
    }

    $raw = Get-Content -Path $Path -Raw -ErrorAction Stop
    if (-not $raw -or $raw.Trim().Length -lt 2) {
        Log "Input empty: $Path"
        return $null
    }

    $obj = $null
    try {
        $obj = $raw | ConvertFrom-Json -ErrorAction Stop
    } catch {
        Log "JSON parse failed: $Path"
        $preview = $raw.Trim()
        if ($preview.Length -gt 300) { $preview = $preview.Substring(0,300) + "..." }
        Log "RAW PREVIEW: $preview"
        throw
    }

    # Determine if it's effectively empty ("{}" -> PSCustomObject with 0 properties)
    # Never rely on .Count on PSObject.Properties (it can be weird). Use array length.
    if ($obj -is [System.Array]) {
        # Arrays are valid content even if length 0; treat 0-length as empty.
        if ($obj.Length -eq 0) { return $null }
        return $obj
    }

    $props = @($obj.PSObject.Properties)
    if (-not $props -or $props.Length -eq 0) {
        Log "Input appears empty object {}: $Path"
        return $null
    }

    return $obj
}

function Get-Prop($obj, [string]$name) {
    if ($null -eq $obj) { return $null }
    $p = $obj.PSObject.Properties[$name]
    if ($null -eq $p) { return $null }
    return $p.Value
}

function Compute-Maturity([double]$Score) {
    if ($Score -ge 95) { "PLATINUM" }
    elseif ($Score -ge 85) { "GOLD" }
    elseif ($Score -ge 70) { "SILVER" }
    elseif ($Score -ge 50) { "BRONZE" }
    else { "RED" }
}

function Compute-Status([double]$Score, [int]$CriticalCount) {
    if ($CriticalCount -gt 0) { "AT_RISK" }
    elseif ($Score -ge 80) { "PASSED" }
    elseif ($Score -ge 60) { "WARNING" }
    else { "FAILED" }
}

Write-Host "DR Report Builder Starting..." -ForegroundColor Cyan
Log "Starting build OutDir=$OutDir Minimal=$Minimal"

try {
    if ($PSScriptRoot) { Set-Location $PSScriptRoot }

    $ctx = (& kubectl config current-context 2>$null | Out-String).Trim()
    if (-not $ctx) { $ctx = "unknown" }

    $pWorkloads  = Join-Path $OutDir "workload-report.json"
    $pStorage    = Join-Path $OutDir "storage-score.json"
    $pBackup     = Join-Path $OutDir "backup-readiness.json"
    $pNetwork    = Join-Path $OutDir "network-readiness.json"
    $pPort       = Join-Path $OutDir "portability.json"

    $workloads = Read-JsonFile $pWorkloads -Required
    $storage   = Read-JsonFile $pStorage   -Required

    # Optional sections: skip entirely in minimal mode
    $backup  = $null
    $network = $null
    $port    = $null
    if (-not $Minimal) {
        $backup  = Read-JsonFile $pBackup
        $network = Read-JsonFile $pNetwork
        $port    = Read-JsonFile $pPort
    }

    # Section scores (best-effort)
    $scores = [ordered]@{
        workloads   = $null
        storage     = $null
        backup      = $null
        network     = $null
        portability = $null
    }

    # Workloads score from risky/total
    $wSummary = Get-Prop $workloads "summary"
    if ($wSummary) {
        $total = Get-Prop $wSummary "totalWorkloads"
        $risky = Get-Prop $wSummary "riskyWorkloads"
        if ($total -ne $null -and [double]$total -gt 0) {
            $r = 0.0
            if ($risky -ne $null) { $r = [double]$risky }
            $s = 100.0 - (($r / [double]$total) * 100.0)
            if ($s -lt 0) { $s = 0 }
            if ($s -gt 100) { $s = 100 }
            $scores.workloads = [math]::Round($s, 2)
        }
    }

    # Storage score from final/max
    $final = Get-Prop $storage "final"
    $max   = Get-Prop $storage "max"
    if ($final -ne $null -and $max -ne $null -and [double]$max -gt 0) {
        $scores.storage = [math]::Round(([double]$final / [double]$max) * 100.0, 2)
    }

    if ($backup)  { $scores.backup       = (Get-Prop $backup "score") }
    if ($network) { $scores.network      = (Get-Prop $network "score") }
    if ($port)    { $scores.portability  = (Get-Prop $port "score") }

    # Present scores
    $present = @()
    foreach ($k in @("workloads","storage","backup","network","portability")) {
        if ($scores[$k] -ne $null) {
            try { $present += [double]$scores[$k] } catch { }
        }
    }
    if ($present.Count -eq 0) { throw "No scores available for overall calculation." }

    $sum = 0.0
    foreach ($v in $present) { $sum += [double]$v }
    $overall = [math]::Round(($sum / [double]$present.Count), 2)

    # Critical blockers (minimal-safe)
    $crit = 0
    $blockers = @()
    $topFlags = $null
    if ($wSummary) { $topFlags = Get-Prop $wSummary "topFlags" }
    if ($topFlags) {
        foreach ($p in @($topFlags.PSObject.Properties)) {
            if ($p.Name -match "hostPath") {
                $crit++
                $blockers += $p.Name
            }
        }
    }

    $tier = Compute-Maturity $overall
    $status = Compute-Status $overall $crit

    # Ensure consistent absent objects in minimal
    if ($Minimal) {
        if ($null -eq $backup)  { $backup  = [pscustomobject]@{ mode="absent"; score=0; notes=@("Not collected (minimal mode).") } }
        if ($null -eq $network) { $network = [pscustomobject]@{ mode="absent"; score=0; notes=@("Not collected (minimal mode).") } }
        if ($null -eq $port)    { $port    = [pscustomobject]@{ mode="absent"; score=0; notes=@("Not collected (minimal mode).") } }
        if ($scores.backup -eq $null) { $scores.backup = 0 }
        if ($scores.network -eq $null) { $scores.network = 0 }
        if ($scores.portability -eq $null) { $scores.portability = 0 }
    }

    $report = [ordered]@{
        timestampUtc = (Get-Date).ToUniversalTime().ToString("o")
        context      = $ctx
        mode         = @{ minimal = [bool]$Minimal }

        overall = @{
            score  = $overall
            tier   = $tier
            status = $status
            criticalBlockers = @{
                count = $crit
                items = $blockers
            }
            sectionScores = $scores
        }

        # Compatibility aliases
        finalScore = $overall
        maturity   = $tier
        drStatus   = $status

        sections = @{
            workloads   = $workloads
            storage     = $storage
            backup      = $backup
            network     = $network
            portability = $port
        }
    }

    $report | ConvertTo-Json -Depth 50 | Out-File $OutputPath -Encoding UTF8

    Write-Host "DR report build complete." -ForegroundColor Green
    Write-Host "Output: $OutputPath" -ForegroundColor Yellow
    Write-Host "Log:    $LogPath" -ForegroundColor Yellow

    Log "SUCCESS output=$OutputPath overall=$overall tier=$tier status=$status"
    "=== Build-DrReport.ps1 END $(Get-Date -Format o) ===" | Out-File $LogPath -Append -Encoding UTF8
    exit 0
} catch {
    Fail 5 ("DR report build failed. See log: " + $LogPath) $_
}
