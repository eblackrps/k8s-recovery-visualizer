[CmdletBinding()]
param(
    [string]$OutputPath = ".\out\workload-report.json",
    [string]$LogPath    = ".\out\workload-log.txt"
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

"=== Collect-Workloads.ps1 START $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Encoding UTF8

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

function Test-Kubectl {
    try {
        $p = Get-Command kubectl -ErrorAction Stop
        Log "kubectl found at: $($p.Source)"
        $true
    } catch { $false }
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

function Invoke-KubectlJson {
    param([Parameter(Mandatory)][string]$Args)

    $r = Invoke-KubectlRaw -Args $Args
    if ($r.ExitCode -ne 0) { throw "kubectl failed (exit $($r.ExitCode)) for: $Args" }
    if (-not $r.Stdout) { return $null }

    try { $r.Stdout | ConvertFrom-Json }
    catch {
        $preview = ($r.Stdout | Out-String)
        if ($preview.Length -gt 500) { $preview = $preview.Substring(0,500) + "..." }
        Log "JSON parse failed for: $Args"
        Log "stdout preview: $preview"
        throw
    }
}

Write-Host "K8s DR Workload Collector Starting..." -ForegroundColor Cyan
Log "Starting"

if (-not (Test-Kubectl)) { Fail -Code 1 -Msg "kubectl not found in PATH." }

# Connectivity check
try {
    $ctxRaw = Invoke-KubectlRaw -Args "config current-context"
    if ($ctxRaw.ExitCode -ne 0 -or -not $ctxRaw.Stdout) { throw "No current context set." }
    $context = ($ctxRaw.Stdout | Out-String).Trim()

    $nodes = Invoke-KubectlRaw -Args "get nodes"
    if ($nodes.ExitCode -ne 0) { throw "kubectl get nodes failed." }

    Write-Host "Connected to context: $context" -ForegroundColor Green
    Log "Connected context: $context"
} catch {
    Fail -Code 2 -Msg "Cannot reach Kubernetes cluster. Check kubeconfig/VPN/RBAC. See log: $LogPath" -Err $_
}

function Collect-Type {
    param([Parameter(Mandatory)][string]$KubectlArgs, [Parameter(Mandatory)][string]$Type)

    $obj = Invoke-KubectlJson -Args $KubectlArgs
    if (-not $obj -or -not $obj.items) { return @() }

    $results = @()

    foreach ($item in @($obj.items)) {
        $ns   = $item.metadata.namespace
        $name = $item.metadata.name

        $replicas = $null
        if ($Type -in @("Deployment","StatefulSet","ReplicaSet")) {
            if ($item.spec -and $item.spec.replicas -ne $null) { $replicas = [int]$item.spec.replicas }
            else { $replicas = 1 }
        }

        $flags = @()
        if ($replicas -ne $null -and $replicas -lt 2) { $flags += "Single replica workload" }

        $results += @{
            namespace = $ns
            name      = $name
            type      = $Type
            replicas  = $replicas
            riskFlags = @($flags)   # force array
        }
    }

    @($results) # force array even if single
}

try {
    $report = @{
        timestampUtc = (Get-Date).ToUniversalTime().ToString("o")
        context      = $context
        workloads    = @()
        summary      = @{
            totalWorkloads = 0
            byType         = @{}
            riskyWorkloads = 0
        }
    }

    $buckets = @(
        @{ args="get deployments --all-namespaces -o json";  type="Deployment"  },
        @{ args="get statefulsets --all-namespaces -o json"; type="StatefulSet" },
        @{ args="get daemonsets --all-namespaces -o json";   type="DaemonSet"   },
        @{ args="get cronjobs --all-namespaces -o json";     type="CronJob"     }
    )

    foreach ($b in $buckets) {
        $items = @(Collect-Type -KubectlArgs $b.args -Type $b.type)  # <--- FIX: ALWAYS ARRAY
        $report.workloads += $items
        $report.summary.byType[$b.type] = $items.Count

        foreach ($w in $items) {
            $rf = @($w.riskFlags) # <--- FIX: ALWAYS ARRAY
            if ($rf.Count -gt 0) { $report.summary.riskyWorkloads++ }
        }
    }

    $report.summary.totalWorkloads = @($report.workloads).Count

    $report | ConvertTo-Json -Depth 10 | Out-File $OutputPath -Encoding UTF8

    Write-Host "Collection complete." -ForegroundColor Green
    Write-Host "Output: $OutputPath" -ForegroundColor Yellow
    Write-Host "Log:    $LogPath" -ForegroundColor Yellow

    Log "SUCCESS output=$OutputPath"
    "=== Collect-Workloads.ps1 END $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Append -Encoding UTF8
    exit 0
} catch {
    Fail -Code 5 -Msg "Collector failed. See log: $LogPath" -Err $_
}
