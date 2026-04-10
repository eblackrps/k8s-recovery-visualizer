[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$ReportPath,
    [Parameter(Mandatory)][string]$OutDir,
    [string]$HistoryPath = "",
    [string]$LogPath = ".\out\history-log.txt"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Ensure-Dir([string]$Path) {
    $d = Split-Path -Parent $Path
    if ($d -and -not (Test-Path $d)) { New-Item -ItemType Directory -Path $d | Out-Null }
}

Ensure-Dir $LogPath

"=== Update-History.ps1 START $(Get-Date -Format o) ===" | Out-File $LogPath -Encoding UTF8
function Log([string]$Msg) { ("{0}  {1}" -f (Get-Date -Format o), $Msg) | Out-File $LogPath -Append -Encoding UTF8 }

function Read-Json([string]$Path) {
    if (-not (Test-Path $Path)) { return $null }
    $raw = Get-Content -Path $Path -Raw -ErrorAction Stop
    if (-not $raw -or $raw.Trim().Length -lt 2) { return $null }
    return ($raw | ConvertFrom-Json -ErrorAction Stop)
}

try {
    if ($PSScriptRoot) { Set-Location $PSScriptRoot }

    if (-not $HistoryPath) {
        $HistoryPath = Join-Path $OutDir "history\index.json"
    }

    Ensure-Dir $HistoryPath

    if (-not (Test-Path $ReportPath)) { throw "ReportPath not found: $ReportPath" }

    $report = Read-Json $ReportPath
    if (-not $report) { throw "Report JSON could not be read: $ReportPath" }

    $context = $report.context
    if (-not $context) { $context = "unknown" }

    $overall = $report.overall
    if (-not $overall) { throw "Report missing 'overall' block." }

    $score = $overall.score
    $tier  = $overall.tier
    $status = $overall.status

    $minimal = $false
    if ($report.mode -and ($report.mode.PSObject.Properties.Name -contains "minimal")) {
        $minimal = [bool]$report.mode.minimal
    }

    # Load existing history
    $hist = Read-Json $HistoryPath
    if (-not $hist) {
        $hist = [pscustomobject]@{
            entries = @()
        }
    }

    # Normalize entries to array
    $entries = @()
    if ($hist.PSObject.Properties.Name -contains "entries" -and $hist.entries) {
        $entries = @($hist.entries)
    }

    # Find previous entry for same context
    $prev = $null
    for ($i = $entries.Count - 1; $i -ge 0; $i--) {
        if ($entries[$i].context -eq $context) { $prev = $entries[$i]; break }
    }

    $deltaScore = $null
    $tierChanged = $false
    if ($prev) {
        try { $deltaScore = [math]::Round(([double]$score - [double]$prev.overall.score), 2) } catch { $deltaScore = $null }
        if ($prev.overall.tier -and $tier -and ($prev.overall.tier -ne $tier)) { $tierChanged = $true }
    }

    $entry = [pscustomobject]@{
        timestampUtc = $report.timestampUtc
        context      = $context
        minimal      = $minimal
        overall      = [pscustomobject]@{
            score  = $score
            tier   = $tier
            status = $status
        }
        sectionScores = $overall.sectionScores
        trend = [pscustomobject]@{
            previousFound = [bool]($prev -ne $null)
            deltaScore    = $deltaScore
            tierChanged   = $tierChanged
        }
    }

    $entries = @($entries + $entry)

    # Keep last 200 entries to avoid infinite growth
    if ($entries.Count -gt 200) {
        $entries = $entries[($entries.Count - 200)..($entries.Count - 1)]
    }

    $histOut = [pscustomobject]@{ entries = $entries }

    $histOut | ConvertTo-Json -Depth 30 | Out-File $HistoryPath -Encoding UTF8

    # Also write a small trend file next to report (optional convenience)
    $trendPath = Join-Path $OutDir "history\latest-trend.json"
    Ensure-Dir $trendPath
    $entry.trend | ConvertTo-Json -Depth 10 | Out-File $trendPath -Encoding UTF8

    Log "SUCCESS history=$HistoryPath entries=$($entries.Count) context=$context delta=$deltaScore"
    "=== Update-History.ps1 END $(Get-Date -Format o) ===" | Out-File $LogPath -Append -Encoding UTF8

    Write-Host "History updated." -ForegroundColor Green
    Write-Host ("History: " + $HistoryPath) -ForegroundColor Yellow
    Write-Host ("Trend  : " + $trendPath) -ForegroundColor Yellow
    exit 0
} catch {
    Log ("FAIL: " + $_.Exception.Message)
    Write-Host ("History update failed: " + $_.Exception.Message) -ForegroundColor Red
    exit 5
}
