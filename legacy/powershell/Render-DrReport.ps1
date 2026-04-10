[CmdletBinding()]
param(
    [string]$InputPath  = ".\out\drscan-report.json",
    [string]$MdPath     = ".\out\drscan-report.md",
    [string]$HtmlPath   = ".\out\drscan-report.html",
    [string]$LogPath    = ".\out\drscan-render-log.txt"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Ensure-Dir {
    param([string]$Path)
    $dir = Split-Path -Parent $Path
    if ($dir -and -not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir | Out-Null }
}

Ensure-Dir -Path $MdPath
Ensure-Dir -Path $HtmlPath
Ensure-Dir -Path $LogPath

"=== Render-DrReport.ps1 START $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Encoding UTF8

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

function Read-Json {
    param([string]$Path)
    if (-not (Test-Path $Path)) { throw "Input not found: $Path" }
    $raw = Get-Content -Path $Path -Raw -ErrorAction Stop
    return ($raw | ConvertFrom-Json)
}

function EscHtml {
    param([string]$s)
    if ($null -eq $s) { return "" }
    return ($s -replace '&','&amp;' -replace '<','&lt;' -replace '>','&gt;' -replace '"','&quot;' -replace "'","&#39;")
}

function TierClass {
    param([string]$tier)
    switch ($tier) {
        "PLATINUM" { "tier-platinum" }
        "GOLD"     { "tier-gold" }
        "SILVER"   { "tier-silver" }
        "BRONZE"   { "tier-bronze" }
        default    { "tier-critical" }
    }
}

Write-Host "DR Report Renderer Starting..." -ForegroundColor Cyan
Log "Starting render"

try {
    $r = Read-Json -Path $InputPath

    $overallScore = [int]$r.overall.score
    $tier = [string]$r.overall.tier
    $status = [string]$r.overall.status
    $context = [string]$r.context
    $ts = [string]$r.timestampUtc

    # Section scores (some are nested)
    $storageScore = $null
    if ($r.sections.storage.storage -and $r.sections.storage.storage.score -ne $null) { $storageScore = [int]$r.sections.storage.storage.score }

    $backupScore = $null
    if ($r.sections.backup.readiness -and $r.sections.backup.readiness.score -ne $null) { $backupScore = [int]$r.sections.backup.readiness.score }

    $networkScore = $null
    if ($r.sections.network.readiness -and $r.sections.network.readiness.score -ne $null) { $networkScore = [int]$r.sections.network.readiness.score }

    $portScore = $null
    if ($r.sections.portability.readiness -and $r.sections.portability.readiness.score -ne $null) { $portScore = [int]$r.sections.portability.readiness.score }

    # Markdown
    $md = New-Object System.Text.StringBuilder
    [void]$md.AppendLine("# Kubernetes DR Scan Report")
    [void]$md.AppendLine("")
    [void]$md.AppendLine("**Timestamp (UTC):** $ts  ")
    [void]$md.AppendLine("**Context:** $context  ")
    [void]$md.AppendLine("")
    [void]$md.AppendLine("## Overall")
    [void]$md.AppendLine("")
    [void]$md.AppendLine("| Score | Tier | Status |")
    [void]$md.AppendLine("|---:|---|---|")
    [void]$md.AppendLine("| $overallScore | $tier | $status |")
    [void]$md.AppendLine("")
    [void]$md.AppendLine("## Section Scores")
    [void]$md.AppendLine("")
    [void]$md.AppendLine("| Section | Score | Notes |")
    [void]$md.AppendLine("|---|---:|---|")
    [void]$md.AppendLine("| Storage | $storageScore | PV/PVC/SC risk scoring |")
    [void]$md.AppendLine("| Backup Readiness | $backupScore | Backup operators, snapshot CRDs, etc. |")
    [void]$md.AppendLine("| Network Readiness | $networkScore | CNI/DNS/Ingress/Services |")
    [void]$md.AppendLine("| Portability | $portScore | Scheduling/policy/webhooks |")
    [void]$md.AppendLine("")

    # Add key flags per section
    [void]$md.AppendLine("## Key Flags")
    [void]$md.AppendLine("")

    function Add-FlagBlock([string]$title, $flagsObj) {
        [void]$md.AppendLine("### $title")
        if ($flagsObj -and @($flagsObj).Count -gt 0) {
            foreach ($f in @($flagsObj)) { [void]$md.AppendLine("- $f") }
        } else {
            [void]$md.AppendLine("- None")
        }
        [void]$md.AppendLine("")
    }

    Add-FlagBlock -title "Backup"     -flagsObj $r.sections.backup.readiness.flags
    Add-FlagBlock -title "Network"    -flagsObj $r.sections.network.readiness.flags
    Add-FlagBlock -title "Portability" -flagsObj $r.sections.portability.readiness.flags

    # Workloads summary
    [void]$md.AppendLine("## Workloads Summary")
    $wTotal = 0
    if ($r.sections.workloads.summary -and $r.sections.workloads.summary.totalWorkloads -ne $null) { $wTotal = [int]$r.sections.workloads.summary.totalWorkloads }
    $wRisky = 0
    if ($r.sections.workloads.summary -and $r.sections.workloads.summary.riskyWorkloads -ne $null) { $wRisky = [int]$r.sections.workloads.summary.riskyWorkloads }
    [void]$md.AppendLine("")
    [void]$md.AppendLine("| Total Workloads | Risky Workloads |")
    [void]$md.AppendLine("|---:|---:|")
    [void]$md.AppendLine("| $wTotal | $wRisky |")
    [void]$md.AppendLine("")

    $md.ToString() | Out-File -FilePath $MdPath -Encoding UTF8

    # HTML
    $tierClass = TierClass -tier $tier
    $html = New-Object System.Text.StringBuilder
    [void]$html.AppendLine("<!doctype html>")
    [void]$html.AppendLine("<html><head><meta charset='utf-8'/>")
    [void]$html.AppendLine("<meta name='viewport' content='width=device-width, initial-scale=1'/>")
    [void]$html.AppendLine("<title>Kubernetes DR Scan Report</title>")
    [void]$html.AppendLine("<style>")
    [void]$html.AppendLine("body{font-family:Segoe UI,Arial,sans-serif;margin:24px;background:#0b0f14;color:#e6edf3}")
    [void]$html.AppendLine(".card{background:#111826;border:1px solid #223049;border-radius:12px;padding:16px;margin:12px 0}")
    [void]$html.AppendLine("h1,h2,h3{margin:0 0 10px 0}")
    [void]$html.AppendLine("table{border-collapse:collapse;width:100%;margin-top:10px}")
    [void]$html.AppendLine("th,td{border:1px solid #223049;padding:8px;text-align:left}")
    [void]$html.AppendLine("th{background:#0f1724}")
    [void]$html.AppendLine(".pill{display:inline-block;padding:4px 10px;border-radius:999px;font-weight:600;border:1px solid #223049}")
    [void]$html.AppendLine(".tier-platinum{background:#1a2b3a}")
    [void]$html.AppendLine(".tier-gold{background:#3a2f14}")
    [void]$html.AppendLine(".tier-silver{background:#242a33}")
    [void]$html.AppendLine(".tier-bronze{background:#2f1d14}")
    [void]$html.AppendLine(".tier-critical{background:#3a1414}")
    [void]$html.AppendLine(".muted{color:#9fb0c3}")
    [void]$html.AppendLine("ul{margin:8px 0 0 18px}")
    [void]$html.AppendLine("</style></head><body>")

    [void]$html.AppendLine("<h1>Kubernetes DR Scan Report</h1>")
    [void]$html.AppendLine("<div class='card'>")
    [void]$html.AppendLine("<div class='muted'>Timestamp (UTC): " + (EscHtml $ts) + "<br/>Context: " + (EscHtml $context) + "</div>")
    [void]$html.AppendLine("</div>")

    [void]$html.AppendLine("<div class='card'>")
    [void]$html.AppendLine("<h2>Overall</h2>")
    [void]$html.AppendLine("<div class='pill " + $tierClass + "'>Score: " + $overallScore + " | " + (EscHtml $tier) + " | " + (EscHtml $status) + "</div>")
    [void]$html.AppendLine("</div>")

    [void]$html.AppendLine("<div class='card'>")
    [void]$html.AppendLine("<h2>Section Scores</h2>")
    [void]$html.AppendLine("<table><thead><tr><th>Section</th><th>Score</th><th>Notes</th></tr></thead><tbody>")
    [void]$html.AppendLine("<tr><td>Storage</td><td>$storageScore</td><td>PV/PVC/SC risk scoring</td></tr>")
    [void]$html.AppendLine("<tr><td>Backup Readiness</td><td>$backupScore</td><td>Backup operators, snapshot CRDs, etc.</td></tr>")
    [void]$html.AppendLine("<tr><td>Network Readiness</td><td>$networkScore</td><td>CNI/DNS/Ingress/Services</td></tr>")
    [void]$html.AppendLine("<tr><td>Portability</td><td>$portScore</td><td>Scheduling/policy/webhooks</td></tr>")
    [void]$html.AppendLine("</tbody></table>")
    [void]$html.AppendLine("</div>")

    function HtmlFlags([string]$title, $flagsObj) {
        [void]$html.AppendLine("<div class='card'><h2>" + (EscHtml $title) + " Flags</h2>")
        if ($flagsObj -and @($flagsObj).Count -gt 0) {
            [void]$html.AppendLine("<ul>")
            foreach ($f in @($flagsObj)) { [void]$html.AppendLine("<li>" + (EscHtml ([string]$f)) + "</li>") }
            [void]$html.AppendLine("</ul>")
        } else {
            [void]$html.AppendLine("<div class='muted'>None</div>")
        }
        [void]$html.AppendLine("</div>")
    }

    HtmlFlags -title "Backup" -flagsObj $r.sections.backup.readiness.flags
    HtmlFlags -title "Network" -flagsObj $r.sections.network.readiness.flags
    HtmlFlags -title "Portability" -flagsObj $r.sections.portability.readiness.flags

    [void]$html.AppendLine("</body></html>")

    $html.ToString() | Out-File -FilePath $HtmlPath -Encoding UTF8

    Write-Host "Render complete." -ForegroundColor Green
    Write-Host "MD:   $MdPath" -ForegroundColor Yellow
    Write-Host "HTML: $HtmlPath" -ForegroundColor Yellow
    Write-Host "Log:  $LogPath" -ForegroundColor Yellow

    Log "SUCCESS md=$MdPath html=$HtmlPath"
    "=== Render-DrReport.ps1 END $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Append -Encoding UTF8
    exit 0
} catch {
    Fail -Code 5 -Msg "Render failed. See log: $LogPath" -Err $_
}
