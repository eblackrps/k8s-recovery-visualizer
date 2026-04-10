[CmdletBinding()]
param(
    [string]$OutputPath = ".\out\portability.json",
    [string]$LogPath    = ".\out\portability-log.txt"
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

"=== Assess-Portability.ps1 START $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Encoding UTF8

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
    } catch { return $null }
}

function Get-Prop {
    param(
        [Parameter(Mandatory)]$Obj,
        [Parameter(Mandatory)][string]$Name
    )
    if ($null -eq $Obj) { return $null }
    try {
        $p = $Obj.PSObject.Properties.Match($Name)
        if ($p -and $p.Count -gt 0) { return $p[0].Value }
    } catch {}
    return $null
}

function Get-ObjectPropertyCount {
    param($Obj)
    if ($null -eq $Obj) { return 0 }
    try { if ($Obj -is [System.Collections.IDictionary]) { return $Obj.Count } } catch {}
    try {
        if ($Obj.PSObject -and $Obj.PSObject.Properties) {
            return @($Obj.PSObject.Properties).Count
        }
    } catch {}
    return 0
}

function Get-ObjectKeys {
    param($Obj)
    if ($null -eq $Obj) { return @() }
    try {
        if ($Obj -is [System.Collections.IDictionary]) { return @($Obj.Keys) }
    } catch {}
    try {
        if ($Obj.PSObject -and $Obj.PSObject.Properties) {
            return @($Obj.PSObject.Properties | Select-Object -ExpandProperty Name)
        }
    } catch {}
    return @()
}

Write-Host "Portability Assessment Starting..." -ForegroundColor Cyan
Log "Starting portability assessment"

try {
    $ctx = (& kubectl config current-context 2>$null)
    if (-not $ctx) { throw "No current context set." }
    Log "Context: $ctx"
} catch {
    Fail -Code 2 -Msg "Cannot determine kubectl context." -Err $_
}

$nodes   = Try-KubectlJson "get nodes -o json"
$pods    = Try-KubectlJson "get pods -A -o json"
$ns      = Try-KubectlJson "get ns -o json"
$pdb     = Try-KubectlJson "get pdb -A -o json"
$pc      = Try-KubectlJson "get priorityclass -o json"
$vh      = Try-KubectlJson "get validatingwebhookconfiguration -o json"
$mh      = Try-KubectlJson "get mutatingwebhookconfiguration -o json"

# --- Node taints/labels ---
$nodeCount = 0
$taintedNodes = @()
$specialLabelNodes = @()

if ($nodes -and (Get-Prop $nodes "items")) {
    foreach ($n in @($nodes.items)) {
        if (-not $n) { continue }
        $nodeCount++
        $name = (Get-Prop (Get-Prop $n "metadata") "name")

        $spec = Get-Prop $n "spec"
        $taintsRaw = $null
        if ($spec) { $taintsRaw = Get-Prop $spec "taints" }
        $taints = @($taintsRaw)

        if (@($taints).Count -gt 0) {
            $taintedNodes += @{ node=$name; taints=$taints }
        }

        $meta = Get-Prop $n "metadata"
        $labels = $null
        if ($meta) { $labels = Get-Prop $meta "labels" }

        $labelKeys = @(Get-ObjectKeys -Obj $labels)
        if ($labelKeys.Count -gt 0) {
            $customKeys = @()
            foreach ($k in $labelKeys) {
                if ($k -match '\.kubernetes\.io/' -or $k -match '^kubernetes\.io/' -or $k -match '^node-role\.kubernetes\.io/' ) { continue }
                if ($k -match '^topology\.kubernetes\.io/' -or $k -match '^failure-domain\.beta\.kubernetes\.io/' ) { continue }
                $customKeys += $k
            }
            if ($customKeys.Count -gt 0) {
                $specialLabelNodes += @{ node=$name; customLabelKeys=$customKeys }
            }
        }
    }
}

# --- Namespace PSA labels ---
$psaLabeled = 0
$psaNamespaces = @()
if ($ns -and (Get-Prop $ns "items")) {
    foreach ($n in @($ns.items)) {
        if (-not $n) { continue }
        $meta = Get-Prop $n "metadata"
        $labels = $null
        if ($meta) { $labels = Get-Prop $meta "labels" }
        if ($labels -and (
            (Get-Prop $labels 'pod-security.kubernetes.io/enforce') -or
            (Get-Prop $labels 'pod-security.kubernetes.io/audit') -or
            (Get-Prop $labels 'pod-security.kubernetes.io/warn')
        )) {
            $psaLabeled++
            $psaNamespaces += (Get-Prop $meta "name")
        }
    }
}

# --- Workload pinning signals ---
$pinnedPods = @()
if ($pods -and (Get-Prop $pods "items")) {
    foreach ($p in @($pods.items)) {
        if (-not $p) { continue }
        $meta = Get-Prop $p "metadata"
        $spec = Get-Prop $p "spec"
        if (-not $meta -or -not $spec) { continue }

        $nsn  = Get-Prop $meta "namespace"
        $name = Get-Prop $meta "name"

        $nodeSelector = Get-Prop $spec "nodeSelector"
        $affinity     = Get-Prop $spec "affinity"
        $tolerations  = @($(Get-Prop $spec "tolerations"))

        $pinned = $false
        $reasons = @()

        if ((Get-ObjectPropertyCount -Obj $nodeSelector) -gt 0) { $pinned = $true; $reasons += "nodeSelector" }

        if ($affinity) {
            if ((Get-Prop $affinity "nodeAffinity") -or (Get-Prop $affinity "podAffinity") -or (Get-Prop $affinity "podAntiAffinity")) {
                $pinned = $true
                $reasons += "affinity"
            }
        }

        if (@($tolerations).Count -gt 0) { $pinned = $true; $reasons += "tolerations" }

        if ($pinned) { $pinnedPods += @{ pod="$nsn/$name"; reasons=$reasons } }
    }
}

# --- Counts ---
$pdbCount = 0
if ($pdb -and (Get-Prop $pdb "items")) { $pdbCount = @($pdb.items).Count }

$pcCount = 0
if ($pc -and (Get-Prop $pc "items")) { $pcCount = @($pc.items).Count }

$validatingWebhookCount = 0
if ($vh -and (Get-Prop $vh "items")) { $validatingWebhookCount = @($vh.items).Count }

$mutatingWebhookCount = 0
if ($mh -and (Get-Prop $mh "items")) { $mutatingWebhookCount = @($mh.items).Count }

# --- Score ---
$flags = New-Object System.Collections.Generic.List[string]
$score = 100

if (@($taintedNodes).Count -gt 0) { $score -= 10; $flags.Add("Tainted nodes detected: $(@($taintedNodes).Count)") }
if (@($specialLabelNodes).Count -gt 0) { $score -= 5; $flags.Add("Custom node labels detected on $(@($specialLabelNodes).Count) nodes") }
if (@($pinnedPods).Count -gt 0) { $score -= 15; $flags.Add("Pods with scheduling constraints detected: $(@($pinnedPods).Count)") }
if ($pdbCount -eq 0 -and ($pods -and (Get-Prop $pods "items"))) { $score -= 5; $flags.Add("No PodDisruptionBudgets detected") }
if ($psaLabeled -gt 0) { $flags.Add("Pod Security Admission labels present on namespaces: $psaLabeled") }
if ($validatingWebhookCount -gt 0 -or $mutatingWebhookCount -gt 0) { $score -= 10; $flags.Add("Admission webhooks present (Validating=$validatingWebhookCount, Mutating=$mutatingWebhookCount)") }

if ($score -lt 0) { $score = 0 }
if ($score -gt 100) { $score = 100 }

$riskTier = "LOW"
if ($score -lt 60) { $riskTier = "HIGH" }
elseif ($score -lt 80) { $riskTier = "MEDIUM" }

$result = @{
    timestampUtc = (Get-Date).ToUniversalTime().ToString("o")
    context      = (& kubectl config current-context 2>$null)
    readiness    = @{ score=$score; riskTier=$riskTier; flags=@($flags) }
    nodes = @{
        total = $nodeCount
        taintedNodes = $taintedNodes
        nodesWithCustomLabels = $specialLabelNodes
    }
    policy = @{
        podSecurityAdmissionNamespaces = $psaNamespaces
        podDisruptionBudgets = @{ count = $pdbCount }
        priorityClasses = @{ count = $pcCount }
        admissionWebhooks = @{ validating=$validatingWebhookCount; mutating=$mutatingWebhookCount }
    }
    workloads = @{
        podsWithSchedulingConstraints = $pinnedPods
    }
}

try {
    $result | ConvertTo-Json -Depth 12 | Out-File $OutputPath -Encoding UTF8

    Write-Host "Portability assessment complete." -ForegroundColor Green
    Write-Host "Output: $OutputPath" -ForegroundColor Yellow
    Write-Host "Log:    $LogPath" -ForegroundColor Yellow

    Log "SUCCESS output=$OutputPath"
    "=== Assess-Portability.ps1 END $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Append -Encoding UTF8
    exit 0
} catch {
    Fail -Code 5 -Msg "Portability assessment failed. See log: $LogPath" -Err $_
}
