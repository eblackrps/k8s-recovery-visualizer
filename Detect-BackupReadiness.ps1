[CmdletBinding()]
param(
    [string]$OutputPath = ".\out\backup-readiness.json",
    [string]$LogPath    = ".\out\backup-readiness-log.txt"
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

"=== Detect-BackupReadiness.ps1 START $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Encoding UTF8

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

function Invoke-KubectlRaw {
    param([Parameter(Mandatory)][string]$Args)
    $argList = $Args -split '\s+'
    $tmpErr = New-TemporaryFile
    try {
        $raw = & kubectl @argList 2> $tmpErr.FullName
        $err = Get-Content -Path $tmpErr.FullName -Raw -ErrorAction SilentlyContinue
        $code = $LASTEXITCODE
        Log "kubectl $Args"
        Log "exitcode=$code"
        if ($err) { Log "stderr: $err" }
        [pscustomobject]@{ Stdout=$raw; Stderr=$err; ExitCode=$code }
    } finally {
        Remove-Item -Path $tmpErr.FullName -Force -ErrorAction SilentlyContinue
    }
}

function Try-KubectlJson {
    param([Parameter(Mandatory)][string]$Args)
    try {
        $r = Invoke-KubectlRaw -Args $Args
        if ($r.ExitCode -ne 0) { return $null }
        if (-not $r.Stdout) { return $null }
        return ($r.Stdout | ConvertFrom-Json)
    } catch {
        return $null
    }
}

function Has-ApiResource {
    param([Parameter(Mandatory)][string]$Name)
    $r = Invoke-KubectlRaw -Args "api-resources"
    if ($r.ExitCode -ne 0 -or -not $r.Stdout) { return $false }
    return (($r.Stdout | Out-String) -match "(?m)^\s*$Name\s+")
}

function Namespace-Exists {
    param([Parameter(Mandatory)][string]$Ns)
    $r = Invoke-KubectlRaw -Args "get ns $Ns -o json"
    return ($r.ExitCode -eq 0)
}

function Get-DeploymentsInNamespace {
    param([Parameter(Mandatory)][string]$Ns)
    $obj = Try-KubectlJson -Args "get deploy -n $Ns -o json"
    if (-not $obj -or -not $obj.items) { return @() }
    $names = @()
    foreach ($d in @($obj.items)) { $names += $d.metadata.name }
    return @($names)
}

Write-Host "Backup Readiness Detector Starting..." -ForegroundColor Cyan
Log "Starting backup readiness"

# context check
try {
    $ctx = (& kubectl config current-context 2>$null)
    if (-not $ctx) { throw "No current context set." }
    Log "Context: $ctx"
} catch {
    Fail -Code 2 -Msg "Cannot determine kubectl context." -Err $_
}

# Snapshot CRDs / API resources
$hasVSC  = Has-ApiResource -Name "volumesnapshotclasses"
$hasVS   = Has-ApiResource -Name "volumesnapshots"
$hasVSCt = Has-ApiResource -Name "volumesnapshotcontents"

# VolumeSnapshotClass objects
$vscObj = $null
if ($hasVSC) {
    $vscObj = Try-KubectlJson -Args "get volumesnapshotclass -o json"
}
$vscCount = 0
$vscNames = @()
if ($vscObj -and $vscObj.items) {
    foreach ($v in @($vscObj.items)) {
        $vscCount++
        $vscNames += $v.metadata.name
    }
}

# Velero / Kasten detection (best-effort heuristics)
$veleroNs = "velero"
$kastenNs = "kasten-io"
$veleroInstalled = (Namespace-Exists -Ns $veleroNs)
$kastenInstalled = (Namespace-Exists -Ns $kastenNs)

$veleroDeploys = @()
$kastenDeploys = @()
if ($veleroInstalled) { $veleroDeploys = Get-DeploymentsInNamespace -Ns $veleroNs }
if ($kastenInstalled) { $kastenDeploys = Get-DeploymentsInNamespace -Ns $kastenNs }

# PVC presence
$pvcObj = Try-KubectlJson -Args "get pvc --all-namespaces -o json"
$pvcCount = 0
if ($pvcObj -and $pvcObj.items) { $pvcCount = @($pvcObj.items).Count }

# StorageClasses presence
$scObj = Try-KubectlJson -Args "get storageclass -o json"
$scCount = 0
if ($scObj -and $scObj.items) { $scCount = @($scObj.items).Count }

# --- Score (0-100) ---
$flags = New-Object System.Collections.Generic.List[string]
$score = 100

if (-not $veleroInstalled -and -not $kastenInstalled) {
    $score -= 25
    $flags.Add("No backup platform detected (Velero/Kasten not found)")
} elseif ($veleroInstalled) {
    $flags.Add("Velero namespace found")
} elseif ($kastenInstalled) {
    $flags.Add("Kasten namespace found")
}

if (-not ($hasVSC -and $hasVS -and $hasVSCt)) {
    $score -= 25
    $flags.Add("CSI snapshot API resources missing (VolumeSnapshot*). Snapshots likely not configured.")
} else {
    $flags.Add("CSI snapshot API resources present")
}

if ($hasVSC -and $vscCount -eq 0) {
    $score -= 15
    $flags.Add("No VolumeSnapshotClass objects found")
} elseif ($vscCount -gt 0) {
    $flags.Add("VolumeSnapshotClass found: $vscCount")
}

if ($pvcCount -eq 0) {
    $score -= 5
    $flags.Add("No PVCs found (data protection may be config-only)")
}

if ($scCount -eq 0) {
    $score -= 10
    $flags.Add("No StorageClasses found (unusual cluster storage config)")
}

if ($score -lt 0) { $score = 0 }
if ($score -gt 100) { $score = 100 }

$riskTier = "LOW"
if ($score -lt 60) { $riskTier = "HIGH" }
elseif ($score -lt 80) { $riskTier = "MEDIUM" }

$result = @{
    timestampUtc = (Get-Date).ToUniversalTime().ToString("o")
    context      = (& kubectl config current-context 2>$null)
    readiness    = @{
        score    = $score
        riskTier = $riskTier
        flags    = @($flags)
    }
    snapshot = @{
        apiResources = @{
            volumesnapshotclasses   = $hasVSC
            volumesnapshots         = $hasVS
            volumesnapshotcontents  = $hasVSCt
        }
        volumeSnapshotClassCount = $vscCount
        volumeSnapshotClasses    = $vscNames
    }
    backupPlatforms = @{
        velero = @{
            namespaceFound = $veleroInstalled
            deployments    = $veleroDeploys
        }
        kasten = @{
            namespaceFound = $kastenInstalled
            deployments    = $kastenDeploys
        }
    }
    inventory = @{
        pvcCount = $pvcCount
        storageClassCount = $scCount
    }
}

try {
    $result | ConvertTo-Json -Depth 12 | Out-File $OutputPath -Encoding UTF8

    Write-Host "Backup readiness complete." -ForegroundColor Green
    Write-Host "Output: $OutputPath" -ForegroundColor Yellow
    Write-Host "Log:    $LogPath" -ForegroundColor Yellow

    Log "SUCCESS output=$OutputPath"
    "=== Detect-BackupReadiness.ps1 END $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Append -Encoding UTF8
    exit 0
} catch {
    Fail -Code 5 -Msg "Backup readiness failed. See log: $LogPath" -Err $_
}
