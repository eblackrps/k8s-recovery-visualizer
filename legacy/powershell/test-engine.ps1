# test-engine.ps1
param(
  [string]$ProfilePath = ".\profiles\default.json",
  [string]$Kubeconfig,
  [string]$Context,
  [switch]$DebugCollect
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$cmd = @(".\scan.ps1", "-ProfilePath", $ProfilePath)
if ($Kubeconfig) { $cmd += @("-Kubeconfig", $Kubeconfig) }
if ($Context)    { $cmd += @("-Context", $Context) }
if ($DebugCollect) { $cmd += @("-DebugCollect") }

pwsh @cmd