[CmdletBinding()]
param(
    [string]$OutputPath = ".\out\network-readiness.json",
    [string]$LogPath    = ".\out\network-readiness-log.txt"
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

"=== Assess-NetworkReadiness.ps1 START $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Encoding UTF8

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

Write-Host "Network Readiness Assessment Starting..." -ForegroundColor Cyan
Log "Starting network readiness"

try {
    $ctx = (& kubectl config current-context 2>$null)
    if (-not $ctx) { throw "No current context set." }
    Log "Context: $ctx"
} catch {
    Fail -Code 2 -Msg "Cannot determine kubectl context." -Err $_
}

# --- Collect ---
$podsAll = Try-KubectlJson "get pods -A -o json"
$svcsAll = Try-KubectlJson "get svc -A -o json"
$epsAll  = Try-KubectlJson "get endpoints -A -o json"
$ingAll  = Try-KubectlJson "get ingress -A -o json"
$ingCls  = Try-KubectlJson "get ingressclass -o json"
$netpol  = Try-KubectlJson "get networkpolicy -A -o json"

# CoreDNS health
$corednsPods = @()
if ($podsAll -and $podsAll.items) {
    foreach ($p in @($podsAll.items)) {
        if ($p.metadata.namespace -eq "kube-system" -and $p.metadata.name -match "coredns") {
            $corednsPods += $p
        }
    }
}
$corednsReady = 0
foreach ($p in @($corednsPods)) {
    $conds = @($p.status.conditions)
    foreach ($c in $conds) {
        if ($c.type -eq "Ready" -and $c.status -eq "True") { $corednsReady++ }
    }
}
$corednsTotal = @($corednsPods).Count

# Detect CNI-ish pods in kube-system
$cniSignals = New-Object System.Collections.Generic.List[string]
if ($podsAll -and $podsAll.items) {
    foreach ($p in @($podsAll.items)) {
        if ($p.metadata.namespace -ne "kube-system") { continue }
        $n = ($p.metadata.name | Out-String).Trim().ToLowerInvariant()
        if ($n -match "calico")   { $cniSignals.Add("calico") }
        if ($n -match "flannel")  { $cniSignals.Add("flannel") }
        if ($n -match "cilium")   { $cniSignals.Add("cilium") }
        if ($n -match "weave")    { $cniSignals.Add("weave") }
        if ($n -match "canal")    { $cniSignals.Add("canal") }
        if ($n -match "kube-router") { $cniSignals.Add("kube-router") }
        if ($n -match "antrea")   { $cniSignals.Add("antrea") }
    }
}
$cniDetected = @($cniSignals | Select-Object -Unique)

# Ingress controllers heuristics
$ingressControllers = New-Object System.Collections.Generic.List[string]
if ($podsAll -and $podsAll.items) {
    foreach ($p in @($podsAll.items)) {
        $n = ($p.metadata.name | Out-String).Trim().ToLowerInvariant()
        if ($n -match "ingress-nginx" -or $n -match "nginx-ingress") { $ingressControllers.Add("nginx") }
        if ($n -match "traefik") { $ingressControllers.Add("traefik") }
        if ($n -match "haproxy") { $ingressControllers.Add("haproxy") }
        if ($n -match "contour") { $ingressControllers.Add("contour") }
        if ($n -match "istiod" -or $n -match "istio") { $ingressControllers.Add("istio") }
    }
}
$ingressControllers = @($ingressControllers | Select-Object -Unique)

# IngressClass inventory
$ingClassNames = @()
if ($ingCls -and $ingCls.items) {
    foreach ($ic in @($ingCls.items)) { $ingClassNames += $ic.metadata.name }
}

# NetworkPolicies
$netpolCount = 0
if ($netpol -and $netpol.items) { $netpolCount = @($netpol.items).Count }

# Services by type + broken endpoints
$svcCount = 0
$lb = 0; $np = 0; $ext = 0; $cip = 0
$servicesNoEndpoints = @()

$endpointMap = @{}
if ($epsAll -and $epsAll.items) {
    foreach ($ep in @($epsAll.items)) {
        $key = "$($ep.metadata.namespace)/$($ep.metadata.name)"
        $endpointMap[$key] = $ep
    }
}

if ($svcsAll -and $svcsAll.items) {
    foreach ($s in @($svcsAll.items)) {
        $svcCount++
        $t = ($s.spec.type | Out-String).Trim()
        if (-not $t) { $t = "ClusterIP" }

        switch ($t) {
            "LoadBalancer" { $lb++ }
            "NodePort"     { $np++ }
            "ExternalName" { $ext++ }
            default        { $cip++ }
        }

        # endpoints sanity: ignore headless + externalname
        $isHeadless = ($s.spec.clusterIP -eq "None")
        if ($t -eq "ExternalName" -or $isHeadless) { continue }

        $k = "$($s.metadata.namespace)/$($s.metadata.name)"
        if ($endpointMap.ContainsKey($k)) {
            $ep = $endpointMap[$k]
            $hasSubsets = ($ep.subsets -ne $null -and @($ep.subsets).Count -gt 0)
            if (-not $hasSubsets) { $servicesNoEndpoints += $k }
        } else {
            $servicesNoEndpoints += $k
        }
    }
}

# MetalLB detection
$metallbFound = $false
$metallbNs = Try-KubectlJson "get ns metallb-system -o json"
if ($metallbNs) { $metallbFound = $true }

# Ingress objects count
$ingCount = 0
if ($ingAll -and $ingAll.items) { $ingCount = @($ingAll.items).Count }

# --- Score ---
$flags = New-Object System.Collections.Generic.List[string]
$score = 100

if (@($cniDetected).Count -eq 0) {
    $score -= 15
    $flags.Add("CNI plugin not detected via kube-system pod names (heuristic)")
} else {
    $flags.Add("CNI detected: $($cniDetected -join ', ')")
}

if ($corednsTotal -eq 0) {
    $score -= 20
    $flags.Add("CoreDNS not found (kube-system coredns pods missing)")
} elseif ($corednsReady -lt $corednsTotal) {
    $score -= 15
    $flags.Add("CoreDNS not fully Ready ($corednsReady/$corednsTotal)")
} else {
    $flags.Add("CoreDNS Ready ($corednsReady/$corednsTotal)")
}

if ($netpolCount -eq 0) {
    $score -= 5
    $flags.Add("No NetworkPolicies found (may be acceptable, but reduces security posture predictability)")
}

if ($ingCount -gt 0 -and @($ingressControllers).Count -eq 0) {
    $score -= 15
    $flags.Add("Ingress objects exist but no ingress controller detected (heuristic)")
} elseif ($ingCount -gt 0) {
    $flags.Add("Ingress controllers detected: $($ingressControllers -join ', ')")
}

if ($lb -gt 0 -and -not $metallbFound) {
    $score -= 10
    $flags.Add("LoadBalancer services present but MetalLB namespace not found (LB provider may be cloud-managed)")
} elseif ($lb -gt 0 -and $metallbFound) {
    $flags.Add("MetalLB detected and LoadBalancer services present")
}

$noEpCount = @($servicesNoEndpoints).Count
if ($noEpCount -gt 0) {
    $score -= 10
    $flags.Add("Services with no endpoints detected: $noEpCount (routing/selector/readiness issues)")
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
    cni = @{
        detected = $cniDetected
    }
    dns = @{
        corednsTotal = $corednsTotal
        corednsReady = $corednsReady
    }
    ingress = @{
        ingressCount = $ingCount
        ingressClasses = $ingClassNames
        controllersDetected = $ingressControllers
    }
    services = @{
        total = $svcCount
        byType = @{
            clusterIP    = $cip
            nodePort     = $np
            loadBalancer = $lb
            externalName = $ext
        }
        servicesWithNoEndpoints = $servicesNoEndpoints
    }
    addons = @{
        metallbDetected = $metallbFound
    }
    networkPolicy = @{
        count = $netpolCount
    }
}

try {
    $result | ConvertTo-Json -Depth 12 | Out-File $OutputPath -Encoding UTF8

    Write-Host "Network readiness complete." -ForegroundColor Green
    Write-Host "Output: $OutputPath" -ForegroundColor Yellow
    Write-Host "Log:    $LogPath" -ForegroundColor Yellow

    Log "SUCCESS output=$OutputPath"
    "=== Assess-NetworkReadiness.ps1 END $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Append -Encoding UTF8
    exit 0
} catch {
    Fail -Code 5 -Msg "Network readiness failed. See log: $LogPath" -Err $_
}
