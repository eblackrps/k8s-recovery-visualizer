[CmdletBinding()]
param(
    [string]$DistRoot = ".\dist"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Ensure-Dir([string]$p) {
    if (-not (Test-Path $p)) { New-Item -ItemType Directory -Path $p | Out-Null }
}

function Get-LatestTag {
    $t = (& git tag --sort=-creatordate) 2>$null
    if ($LASTEXITCODE -ne 0) { return "v0.0.0" }
    $first = @($t)[0]
    if (-not $first) { return "v0.0.0" }
    return ($first | Out-String).Trim()
}

function Copy-IfExists([string]$src, [string]$dst) {
    if (-not (Test-Path $src)) { throw "Missing required file: $src" }
    Copy-Item -Path $src -Destination $dst -Force
}

$tag = Get-LatestTag
$stamp = (Get-Date).ToString("yyyyMMdd-HHmmss")
$bundleName = "k8s-recovery-visualizer-drscan-$tag-$stamp"
$bundleDir = Join-Path $DistRoot $bundleName

Ensure-Dir $DistRoot
Ensure-Dir $bundleDir
Ensure-Dir (Join-Path $bundleDir "out")

# Files required in customer bundle
$files = @(
    "scan.ps1",
    "Invoke-DrScan.ps1",
    "Collect-Workloads.ps1",
    "Score-Storage.ps1",
    "Detect-BackupReadiness.ps1",
    "Assess-NetworkReadiness.ps1",
    "Assess-Portability.ps1",
    "Build-DrReport.ps1",
    "Render-DrReport.ps1",
    "Run-DrScan.ps1",
    "README.md"
)

foreach ($f in $files) {
    Copy-IfExists -src (Join-Path "." $f) -dst (Join-Path $bundleDir $f)
}

@"
# DR Scan Prereqs

## Required
- PowerShell 7 (`pwsh`)
- kubectl in PATH
- kubeconfig configured to the target cluster
- Read-only access to core resources across namespaces (scan can still run partially with limited RBAC)

## Run
Open PowerShell 7 in this folder and run:

pwsh -ExecutionPolicy Bypass -File .\Run-DrScan.ps1 -OutDir .\out

Outputs:
- .\out\drscan-report.json
- .\out\drscan-report.md
- .\out\drscan-report.html
"@ | Set-Content -Path (Join-Path $bundleDir "PREREQS.md") -Encoding UTF8

# Create zip
$zipPath = Join-Path $DistRoot ($bundleName + ".zip")
if (Test-Path $zipPath) { Remove-Item $zipPath -Force }

Compress-Archive -Path (Join-Path $bundleDir "*") -DestinationPath $zipPath -Force

Write-Host "Dist bundle created:" -ForegroundColor Green
Write-Host $zipPath -ForegroundColor Yellow
Write-Host "Folder:" -ForegroundColor Green
Write-Host $bundleDir -ForegroundColor Yellow
