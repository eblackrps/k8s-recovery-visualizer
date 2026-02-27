[CmdletBinding()]
param(
    [string]$OutRoot = ".\out"
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

function Ensure-Dir([string]$p) {
    if (-not (Test-Path $p)) { New-Item -ItemType Directory -Path $p | Out-Null }
}

try {
    # Always run from the folder containing this script/EXE
    if ($PSScriptRoot) { Set-Location $PSScriptRoot }

    Say "K8s Recovery Visualizer - DR Scan Launcher" Cyan
    Say ("Working Dir: " + (Get-Location)) DarkGray

    if (-not (Check-Command "pwsh"))   { Fail "PowerShell 7 (pwsh) is required in PATH." 2 }
    if (-not (Check-Command "kubectl")) { Fail "kubectl is required in PATH." 2 }

    # Basic kubectl health checks
    $ctx = (& kubectl config current-context 2>$null | Out-String).Trim()
    if (-not $ctx) { Fail "kubectl has no current context. Configure kubeconfig first." 3 }
    Say "Context: $ctx" Green

    & kubectl get nodes 1>$null 2>$null
    if ($LASTEXITCODE -ne 0) { Fail "Cannot reach cluster with current context (kubectl get nodes failed)." 3 }

    # Timestamped output directory
    $stamp = (Get-Date).ToString("yyyyMMdd-HHmmss")
    Ensure-Dir $OutRoot

    $outDir = Join-Path $OutRoot ("drscan-" + $stamp)
    Ensure-Dir $outDir

    $runner = Join-Path (Get-Location) "Invoke-DrScan.ps1"
    if (-not (Test-Path $runner)) { Fail "Missing Invoke-DrScan.ps1 next to launcher." 4 }

    Say ("Running scan -> " + $outDir) Cyan

    & pwsh -ExecutionPolicy Bypass -File $runner -OutDir $outDir
    $code = $LASTEXITCODE

    if ($code -ne 0) { Fail "Scan failed with exit code $code." $code }

    Say "DR Scan complete." Green
    Say ("JSON : " + (Join-Path $outDir "drscan-report.json")) Yellow
    Say ("MD   : " + (Join-Path $outDir "drscan-report.md")) Yellow
    Say ("HTML : " + (Join-Path $outDir "drscan-report.html")) Yellow
    exit 0
} catch {
    Fail ("Launcher failed: " + $_.Exception.Message) 5
}
