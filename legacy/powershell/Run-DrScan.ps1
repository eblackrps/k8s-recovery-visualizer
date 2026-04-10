[CmdletBinding()]
param(
    [string]$OutDir = ".\out"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Say([string]$msg, [ConsoleColor]$c = [ConsoleColor]::Gray) {
    $old = $Host.UI.RawUI.ForegroundColor
    $Host.UI.RawUI.ForegroundColor = $c
    Write-Host $msg
    $Host.UI.RawUI.ForegroundColor = $old
}

function Fail([string]$msg, [int]$code = 5) {
    Say $msg Red
    exit $code
}

function Check-Command([string]$name) {
    try { Get-Command $name -ErrorAction Stop | Out-Null; return $true } catch { return $false }
}

function Try-Kubectl([string]$args) {
    $tmp = New-TemporaryFile
    try {
        $out = & kubectl ($args -split '\s+') 2> $tmp.FullName
        $err = Get-Content $tmp.FullName -Raw -ErrorAction SilentlyContinue
        [pscustomobject]@{ Out=$out; Err=$err; Code=$LASTEXITCODE }
    } finally {
        Remove-Item $tmp.FullName -Force -ErrorAction SilentlyContinue
    }
}

# Run from script directory
if ($PSScriptRoot) { Set-Location $PSScriptRoot }

Say "K8s Recovery Visualizer (DR Scan) - Customer Runner" Cyan
Say ("Working Dir: " + (Get-Location)) DarkGray

if (-not (Check-Command "pwsh")) { Fail "PowerShell 7 (pwsh) is required." 2 }
if (-not (Check-Command "kubectl")) { Fail "kubectl is required (must be in PATH)." 2 }

$r1 = Try-Kubectl "config current-context"
if ($r1.Code -ne 0 -or -not $r1.Out) {
    Fail "kubectl has no current context. Configure kubeconfig first. stderr: $($r1.Err)" 3
}
$ctx = ($r1.Out | Out-String).Trim()
Say "Context: $ctx" Green

$r2 = Try-Kubectl "get nodes"
if ($r2.Code -ne 0) {
    Fail "Cannot reach cluster with current context. stderr: $($r2.Err)" 3
}

$r3 = Try-Kubectl "auth can-i get pods -A"
if ($r3.Code -eq 0) {
    $ans = (($r3.Out | Out-String).Trim())
    if ($ans -ne "yes") {
        Say "Warning: RBAC may be limited (kubectl auth can-i get pods -A = $ans). Scan may be incomplete." Yellow
    }
}

$scan = Join-Path (Get-Location) "scan.ps1"
if (-not (Test-Path $scan)) { Fail "Missing scan.ps1 in bundle." 4 }

Say "Running scan..." Cyan
& pwsh -ExecutionPolicy Bypass -File $scan -OutDir $OutDir
$code = $LASTEXITCODE

if ($code -ne 0) { Fail "Scan failed with exit code $code." $code }

Say "Scan completed successfully." Green
Say ("Outputs in: " + (Resolve-Path $OutDir)) Yellow
exit 0
