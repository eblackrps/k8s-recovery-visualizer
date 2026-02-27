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
    if (-not $p) { throw "Ensure-Dir called with null/empty path." }
    if (-not (Test-Path $p)) { New-Item -ItemType Directory -Path $p | Out-Null }
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

try {
    # Always run from the folder containing this script (or compiled EXE directory when wrapped)
    if ($PSScriptRoot) { Set-Location $PSScriptRoot }

    Say "K8s Recovery Visualizer - DR Scan Launcher" Cyan
    Say ("Working Dir: " + (Get-Location)) DarkGray

    if (-not (Check-Command "pwsh"))   { Fail "PowerShell 7 (pwsh) is required in PATH." 2 }
    if (-not (Check-Command "kubectl")) { Fail "kubectl is required in PATH." 2 }

    # Cluster connectivity
    $ctx = (& kubectl config current-context 2>$null | Out-String).Trim()
    if (-not $ctx) { Fail "kubectl has no current context. Configure kubeconfig first." 3 }
    Say "Context: $ctx" Green

    $nodes = Try-Kubectl "get nodes"
    if ($nodes.Code -ne 0) { Fail ("Cannot reach cluster (kubectl get nodes failed). " + $nodes.Err) 3 }

    # Detect cluster name (prefer current-context)
    $clusterName = $ctx
    if (-not $clusterName) { $clusterName = "unknown-cluster" }

    # Sanitize name for filesystem
    $clusterName = $clusterName -replace '[^a-zA-Z0-9\-]', '_'

    # Output directory: .\out\<cluster>-YYYYMMDD-HHMMSS
    $stamp = (Get-Date).ToString("yyyyMMdd-HHmmss")
    Ensure-Dir $OutRoot

    $outDir = Join-Path $OutRoot ($clusterName + "-" + $stamp)
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
