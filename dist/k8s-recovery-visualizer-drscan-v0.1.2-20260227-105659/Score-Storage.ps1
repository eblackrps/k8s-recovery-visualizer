[CmdletBinding()]
param(
    [string]$OutputPath = ".\out\storage-score.json",
    [string]$LogPath    = ".\out\storage-score-log.txt"
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

"=== Score-Storage.ps1 START $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Encoding UTF8

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

function Invoke-KubectlJson {
    param([Parameter(Mandatory)][string]$Args)
    $r = Invoke-KubectlRaw -Args $Args
    if ($r.ExitCode -ne 0) { throw "kubectl failed (exit $($r.ExitCode)) for: $Args" }
    if (-not $r.Stdout) { return $null }
    $r.Stdout | ConvertFrom-Json
}

function Get-DefaultStorageClasses {
    param($ScItems)
    $defaults = @()
    foreach ($sc in @($ScItems)) {
        if (-not $sc) { continue }
        $ann = $sc.metadata.annotations
        if ($ann) {
            if ($ann.'storageclass.kubernetes.io/is-default-class' -eq 'true' -or
                $ann.'storageclass.beta.kubernetes.io/is-default-class' -eq 'true') {
                $defaults += $sc.metadata.name
            }
        }
    }
    return @($defaults) # force array
}

function Guess-StorageTypeFromSc {
    param($Sc)
    $prov = ($Sc.provisioner | Out-String).Trim().ToLowerInvariant()
    $name = ($Sc.metadata.name | Out-String).Trim().ToLowerInvariant()

    if ($prov -match 'local' -or $name -match 'local') { return "LOCAL" }
    if ($prov -match 'hostpath' -or $name -match 'hostpath') { return "HOSTPATH" }
    if ($prov -match 'nfs' -or $name -match 'nfs') { return "NFS" }
    if ($prov -match 'longhorn' -or $name -match 'longhorn') { return "LONGHORN" }
    if ($prov -match 'rbd|ceph' -or $name -match 'ceph|rbd') { return "CEPH" }
    if ($prov -match 'csi' -and ($name -match 'efs|azurefile|filestore|fsx|netapp|ontap|trident')) { return "RWX_MANAGED" }
    if ($prov -match 'csi') { return "CSI_GENERIC" }
    return "UNKNOWN"
}

function Score-Storage {
    param($StorageClasses, $Pvs, $Pvcs)

    $flags = New-Object System.Collections.Generic.List[string]
    $score = 100
    $breakdown = @{}

    $scItems  = @($StorageClasses.items)
    $pvItems  = @($Pvs.items)
    $pvcItems = @($Pvcs.items)

    $defaultSc = @(Get-DefaultStorageClasses -ScItems $scItems)

    if (@($defaultSc).Count -eq 0) {
        $score -= 15
        $flags.Add("No default StorageClass set")
        $breakdown["defaultStorageClass"] = -15
    } elseif (@($defaultSc).Count -gt 1) {
        $score -= 5
        $flags.Add("Multiple default StorageClasses set: $($defaultSc -join ', ')")
        $breakdown["multipleDefaultSC"] = -5
    } else {
        $breakdown["defaultStorageClass"] = 0
    }

    # PV analysis
    $localPvCount = 0
    $nfsPvCount   = 0
    $hostPathPvCount = 0
    $unknownPvCount = 0

    foreach ($pv in @($pvItems)) {
        if (-not $pv) { continue }
        if ($pv.spec.local) { $localPvCount++ }
        if ($pv.spec.nfs)   { $nfsPvCount++ }
        if ($pv.spec.hostPath) { $hostPathPvCount++ }
        if (-not $pv.spec.csi -and -not $pv.spec.nfs -and -not $pv.spec.local -and -not $pv.spec.hostPath) {
            $unknownPvCount++
        }
    }

    if ($hostPathPvCount -gt 0) {
        $score -= 30
        $flags.Add("hostPath PVs detected: $hostPathPvCount")
        $breakdown["hostPathPV"] = -30
    }

    if ($localPvCount -gt 0) {
        $score -= 25
        $flags.Add("Local PVs detected: $localPvCount (node-tied storage)")
        $breakdown["localPV"] = -25
    }

    if ($nfsPvCount -gt 0) {
        $score -= 10
        $flags.Add("NFS PVs detected: $nfsPvCount (depends on external filer availability)")
        $breakdown["nfsPV"] = -10
    }

    if ($unknownPvCount -gt 0) {
        $score -= 5
        $flags.Add("Unclassified PV types detected: $unknownPvCount")
        $breakdown["unknownPV"] = -5
    }

    # PVC analysis
    $rwo = 0; $rwx = 0; $rom = 0; $noSc = 0
    $scUsage = @{}

    foreach ($pvc in @($pvcItems)) {
        if (-not $pvc) { continue }
        $modes = @($pvc.spec.accessModes)
        if ($modes -contains "ReadWriteOnce") { $rwo++ }
        if ($modes -contains "ReadWriteMany") { $rwx++ }
        if ($modes -contains "ReadOnlyMany")  { $rom++ }

        $sc = $pvc.spec.storageClassName
        if (-not $sc) { $noSc++ }
        else {
            if (-not $scUsage.ContainsKey($sc)) { $scUsage[$sc] = 0 }
            $scUsage[$sc]++
        }
    }

    if ($noSc -gt 0) {
        $score -= 5
        $flags.Add("PVCs without storageClassName: $noSc (default SC dependency)")
        $breakdown["pvcNoSC"] = -5
    }

    if ($rwx -eq 0 -and @($pvcItems).Count -gt 0) {
        $score -= 8
        $flags.Add("No RWX (ReadWriteMany) PVCs detected (shared storage may be limited)")
        $breakdown["noRWX"] = -8
    }

    # SC meta
    $scMeta = @{}
    foreach ($sc in @($scItems)) {
        if (-not $sc) { continue }
        $t = Guess-StorageTypeFromSc -Sc $sc
        $scMeta[$sc.metadata.name] = @{
            type = $t
            provisioner = $sc.provisioner
            volumeBindingMode = $sc.volumeBindingMode
        }
    }

    # Top SC by PVC usage
    $topSc = $null
    $topCount = 0
    foreach ($k in $scUsage.Keys) {
        if ($scUsage[$k] -gt $topCount) { $topSc = $k; $topCount = $scUsage[$k] }
    }

    if ($topSc -and $scMeta.ContainsKey($topSc)) {
        $topType = $scMeta[$topSc].type
        if ($topType -in @("LOCAL","HOSTPATH")) {
            $score -= 20
            $flags.Add("Dominant StorageClass '$topSc' appears to be $topType")
            $breakdown["dominantSCBad"] = -20
        } elseif ($topType -in @("NFS","UNKNOWN","CSI_GENERIC")) {
            $score -= 5
            $flags.Add("Dominant StorageClass '$topSc' appears to be $topType")
            $breakdown["dominantSCWeak"] = -5
        } else {
            $breakdown["dominantSC"] = 0
        }
    }

    if ($score -lt 0) { $score = 0 }
    if ($score -gt 100) { $score = 100 }

    $tier = "LOW"
    if ($score -lt 60) { $tier = "HIGH" }
    elseif ($score -lt 80) { $tier = "MEDIUM" }

    return @{
        score = $score
        riskTier = $tier
        flags = @($flags)
        breakdown = $breakdown
        stats = @{
            pvcTotal = @($pvcItems).Count
            pvTotal  = @($pvItems).Count
            rwoPVCs  = $rwo
            rwxPVCs  = $rwx
            romPVCs  = $rom
            pvcNoStorageClassName = $noSc
            localPVs = $localPvCount
            nfsPVs   = $nfsPvCount
            hostPathPVs = $hostPathPvCount
            unknownPVs  = $unknownPvCount
        }
        storageClassUsage = $scUsage
        storageClassDetails = $scMeta
        defaultStorageClasses = $defaultSc
    }
}

Write-Host "Storage Risk Scoring Starting..." -ForegroundColor Cyan
Log "Starting storage scoring"

try {
    $ctx = (& kubectl config current-context 2>$null)
    if (-not $ctx) { throw "No current context set." }
    Log "Context: $ctx"
} catch {
    Fail -Code 2 -Msg "Cannot determine kubectl context." -Err $_
}

try {
    $sc  = Invoke-KubectlJson "get storageclass -o json"
    $pv  = Invoke-KubectlJson "get pv -o json"
    $pvc = Invoke-KubectlJson "get pvc --all-namespaces -o json"

    $result = @{
        timestampUtc = (Get-Date).ToUniversalTime().ToString("o")
        context      = (& kubectl config current-context 2>$null)
        storage      = (Score-Storage -StorageClasses $sc -Pvs $pv -Pvcs $pvc)
    }

    $result | ConvertTo-Json -Depth 12 | Out-File $OutputPath -Encoding UTF8

    Write-Host "Storage scoring complete." -ForegroundColor Green
    Write-Host "Output: $OutputPath" -ForegroundColor Yellow
    Write-Host "Log:    $LogPath" -ForegroundColor Yellow

    Log "SUCCESS output=$OutputPath"
    "=== Score-Storage.ps1 END $(Get-Date -Format o) ===" | Out-File -FilePath $LogPath -Append -Encoding UTF8
    exit 0
} catch {
    Fail -Code 5 -Msg "Storage scoring failed. See log: $LogPath" -Err $_
}
