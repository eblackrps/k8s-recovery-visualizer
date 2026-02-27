param(
    [string]$Kubeconfig,
    [string]$Context,
    [string]$OutDir = ".\out",
    [string]$ExePath = ".\scan.exe",
    [switch]$Ci,
    [switch]$DryRun,
    [switch]$DebugCollect
)

$ErrorActionPreference = "Stop"

# Resolve paths
$exeFull  = (Resolve-Path $ExePath).Path
$outFull  = (Resolve-Path $OutDir -ErrorAction SilentlyContinue)

if (-not $outFull) {
    New-Item -ItemType Directory -Path $OutDir -Force | Out-Null
    $outFull = (Resolve-Path $OutDir).Path
}

Write-Host ""
Write-Host "== Scanner Run =="
Write-Host "Exe:       $exeFull"
Write-Host "OutDir:    $outFull"
if ($Kubeconfig) { Write-Host "Kubeconf:  $Kubeconfig" }

# Build arguments
$args = @("-out", $outFull)

if ($Kubeconfig) { $args += @("-kubeconfig", $Kubeconfig) }
if ($Context)    { $args += @("-context", $Context) }
if ($Ci)         { $args += "-ci" }
if ($DryRun)     { $args += "-dry-run" }

Write-Host "Args:      $($args -join ' ')"

# Start process
$proc = Start-Process -FilePath $exeFull `
                      -ArgumentList $args `
                      -NoNewWindow `
                      -PassThru `
                      -Wait

# Handle exit codes
switch ($proc.ExitCode) {
    0 {
        # success
    }
    2 {
        Write-Warning "scan.exe completed but DR score is below threshold."
    }
    default {
        throw "scan.exe failed with exit code $($proc.ExitCode)"
    }
}

Write-Host ""
Write-Host "== Done =="
Write-Host "Report:    $outFull\recovery-report.html"
