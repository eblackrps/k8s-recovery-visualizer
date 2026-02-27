param(
    [Parameter(Mandatory = $false)]
    [string] $OutDir = (Resolve-Path ".\out").Path,

    [Parameter(Mandatory = $false)]
    [string] $HtmlName = "recovery-report.html"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Ensure-HasDoctypeHtmlHeadBody {
    param([Parameter(Mandatory=$true)][string]$Html)

    $src = $Html.Trim()

    if ($src -notmatch '(?is)<!doctype\s+html') {
        $src = "<!doctype html>`n" + $src
    }

    if ($src -notmatch '(?is)<html\b') {
        $src = "<html>`n" + $src + "`n</html>"
    }

    if ($src -notmatch '(?is)<head\b') {
        $src = [regex]::Replace(
            $src,
            '(?is)(<html\b[^>]*>)',
            '$1' + "`n<head>`n<meta charset=""utf-8"">`n<meta name=""viewport"" content=""width=device-width, initial-scale=1"">`n<title>K8s Recovery Report</title>`n</head>",
            1
        )
    }

    if ($src -notmatch '(?is)<body\b') {
        if ($src -match '(?is)</head>') {
            $src = [regex]::Replace($src, '(?is)</head>', '</head>' + "`n<body>", 1)
        } else {
            $src = [regex]::Replace($src, '(?is)(<head\b[^>]*>)', '$1' + "`n<body>", 1)
        }

        if ($src -match '(?is)</html>') {
            $src = [regex]::Replace($src, '(?is)</html>', "`n</body>`n</html>", 1)
        } else {
            $src = $src + "`n</body>"
        }
    }

    if ($src -notmatch '(?is)</body>') {
        $src = [regex]::Replace($src, '(?is)</html>', "`n</body>`n</html>", 1)
    }

    if ($src -notmatch '(?is)</html>') {
        $src = $src + "`n</html>"
    }

    return $src
}

function Ensure-KrvWrap {
    param([Parameter(Mandatory=$true)][string]$Html)

    $src = $Html
    if ($src -notmatch '(?is)<!--\s*KRV_WRAP_START\s*-->') {
        $src = [regex]::Replace(
            $src,
            '(?is)(<body\b[^>]*>)',
            '$1' + "`n<!-- KRV_WRAP_START -->`n<div class=""krv-wrap"">",
            1
        )
        $src = [regex]::Replace(
            $src,
            '(?is)</body>',
            "`n</div>`n<!-- KRV_WRAP_END -->`n</body>",
            1
        )
    }

    return $src
}

function Wrap-CommonBlocks {
    param([Parameter(Mandatory=$true)][string]$Html)

    $src = $Html

    $src = [regex]::Replace(
        $src,
        '(?is)(<h2[^>]*>\s*Artifacts\s*</h2>\s*)(<ul\b[^>]*>.*?</ul>)',
        '$1<div class="krv-card">$2</div>',
        1
    )

    if ($src -notmatch '(?is)<h1\b[^>]*>.*?</h1>\s*<div\s+class=["'']krv-card["'']>') {
        $src = [regex]::Replace(
            $src,
            '(?is)(<h1\b[^>]*>.*?</h1>\s*)(<p\b[^>]*>.*?</p>)',
            '$1<div class="krv-card">$2</div>',
            1
        )
    }

    return $src
}

function Upgrade-IssuesSection {
    param([Parameter(Mandatory=$true)][string]$Html)

    $src = $Html

    $m = [regex]::Match($src, '(?is)<h2\b[^>]*>\s*Top\s*issues\s*</h2>')
    if (-not $m.Success) {
        return $src
    }

    $headerStart = $m.Index
    $headerEnd   = $m.Index + $m.Length

    $tail = $src.Substring($headerEnd)

    $mNextRel = [regex]::Match($tail, '(?is)<h2\b[^>]*>')
    $mBodyEndRel = [regex]::Match($tail, '(?is)</body>')

    $sectionEnd = $src.Length
    if ($mNextRel.Success) { $sectionEnd = $headerEnd + $mNextRel.Index }
    elseif ($mBodyEndRel.Success) { $sectionEnd = $headerEnd + $mBodyEndRel.Index }

    $before = $src.Substring(0, $headerStart)
    $header = $src.Substring($headerStart, $headerEnd - $headerStart)
    $sectionBody = $src.Substring($headerEnd, $sectionEnd - $headerEnd)
    $after  = $src.Substring($sectionEnd)

    # Count PASS/FAIL markers inside this section
    $passCount = ([regex]::Matches($sectionBody, '(?is)\bPASS\b\s*:')).Count
    $failCount = ([regex]::Matches($sectionBody, '(?is)\bFAIL\b\s*:')).Count
    $checksExecuted = $passCount + $failCount

    $hasFail = ($failCount -gt 0)

    # Idempotent: remove previously injected blocks
    $sectionBody = $sectionBody -replace '(?is)\s*<!--\s*KRV_ISSUES_LEADIN\s*-->.*?<!--\s*KRV_ISSUES_LEADIN_END\s*-->\s*', ''
    $sectionBody = $sectionBody -replace '(?is)\s*<!--\s*KRV_ISSUES_SUCCESS\s*-->.*?<!--\s*KRV_ISSUES_SUCCESS_END\s*-->\s*', ''

    # Confidence heuristic (simple + understandable)
    $confidence = "LOW"
    if ($checksExecuted -le 0) {
        $confidence = "LOW"
    } elseif ($failCount -eq 0) {
        $confidence = "HIGH"
    } elseif ($failCount -le 2) {
        $confidence = "MEDIUM"
    } else {
        $confidence = "LOW"
    }

    if ($hasFail) {
        $header = [regex]::Replace($header, '(?is)Top\s*issues', 'Top Issues', 1)

        $leadIn = @"
<!-- KRV_ISSUES_LEADIN -->
<p class="krv-leadin krv-attn">The following items require attention:</p>
<p class="krv-meta">$checksExecuted checks executed &middot; $failCount failures &middot; Confidence: $confidence</p>
<!-- KRV_ISSUES_LEADIN_END -->
"@
        $sectionBody = "`n$leadIn`n" + $sectionBody
    }
    else {
        $header = [regex]::Replace($header, '(?is)Top\s*issues', 'Validation Summary', 1)

        $success = @"
<!-- KRV_ISSUES_SUCCESS -->
<p class="krv-leadin krv-success">All DR readiness checks passed successfully.</p>
<p class="krv-meta">$checksExecuted checks executed &middot; $failCount failures &middot; Confidence: $confidence</p>
<!-- KRV_ISSUES_SUCCESS_END -->
"@
        $sectionBody = "`n$success`n" + $sectionBody
    }

    return $before + $header + $sectionBody + $after
}

# --- Main ---
$OutDir  = (Resolve-Path -LiteralPath $OutDir).Path
$htmlPath = Join-Path $OutDir $HtmlName

if (-not (Test-Path -LiteralPath $htmlPath)) {
    throw "HTML report not found: $htmlPath"
}

$src = Get-Content -LiteralPath $htmlPath -Raw -Encoding UTF8

$src = Ensure-HasDoctypeHtmlHeadBody -Html $src
$src = Ensure-KrvWrap -Html $src
$src = Wrap-CommonBlocks -Html $src
$src = Upgrade-IssuesSection -Html $src

Set-Content -LiteralPath $htmlPath -Value $src -Encoding UTF8
Write-Host "Normalized HTML structure: $htmlPath"