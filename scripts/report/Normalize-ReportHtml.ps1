param(
    [Parameter(Mandatory = $false)]
    [string] $OutDir = (Resolve-Path ".\out").Path,

    [Parameter(Mandatory = $false)]
    [string] $HtmlName = "recovery-report.html"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Get-FirstMatchValue {
    param(
        [Parameter(Mandatory = $true)][string] $Text,
        [Parameter(Mandatory = $true)][string] $Pattern
    )
    $m = [regex]::Match($Text, $Pattern, [System.Text.RegularExpressions.RegexOptions]::IgnoreCase -bor [System.Text.RegularExpressions.RegexOptions]::Singleline)
    if ($m.Success) { return $m.Groups[1].Value }
    return $null
}

function Ensure-Doctype {
    param([Parameter(Mandatory = $true)][string] $Text)
    if ($Text -match '(?is)<!doctype\s+html') { return $Text }
    return "<!doctype html>`n$Text"
}

function Ensure-HtmlHeadBodySkeleton {
    param([Parameter(Mandatory = $true)][string] $Text)

    # If already has <html>, assume it’s mostly a real document and only patch missing bits.
    $hasHtml = [regex]::IsMatch($Text, '(?is)<html\b')
    $hasHead = [regex]::IsMatch($Text, '(?is)<head\b')
    $hasBody = [regex]::IsMatch($Text, '(?is)<body\b')

    if (-not $hasHtml) {
        # If it's a fragment, wrap it. We'll add head/body.
        $Text = "<html>`n$Text`n</html>"
        $hasHtml = $true
    }

    if (-not $hasHead) {
        # Insert a minimal head right after <html ...>
        $headBlock = "<head><meta charset=""utf-8""><meta name=""viewport"" content=""width=device-width, initial-scale=1""></head>"
        $Text = [regex]::Replace($Text, '(?is)(<html\b[^>]*>)', "`$1$headBlock", 1)
        $hasHead = $true
    }

    if (-not $hasBody) {
        # Best-effort: place <body> after </head> if present; otherwise after <head...>
        if ([regex]::IsMatch($Text, '(?is)</head>')) {
            $Text = [regex]::Replace($Text, '(?is)</head>', '</head><body>', 1)
        }
        else {
            $Text = [regex]::Replace($Text, '(?is)(<head\b[^>]*>)', "`$1<body>", 1)
        }

        # Ensure closing </body> before </html>
        if ([regex]::IsMatch($Text, '(?is)</html>')) {
            $Text = [regex]::Replace($Text, '(?is)</html>', '</body></html>', 1)
        }
        else {
            $Text = "$Text</body>"
        }

        $hasBody = $true
    }

    # Ensure closing tags exist (single pass, minimal assumptions)
    if (-not [regex]::IsMatch($Text, '(?is)</body>')) {
        if ([regex]::IsMatch($Text, '(?is)</html>')) {
            $Text = [regex]::Replace($Text, '(?is)</html>', '</body></html>', 1)
        }
        else {
            $Text = "$Text</body>"
        }
    }

    if (-not [regex]::IsMatch($Text, '(?is)</html>')) {
        $Text = "$Text</html>"
    }

    return $Text
}

function Ensure-HeadMetaOnce {
    param([Parameter(Mandatory = $true)][string] $Text)

    # Do NOT duplicate meta tags if already there.
    $hasCharset = [regex]::IsMatch($Text, '(?is)<meta\b[^>]*charset\s*=')
    $hasViewport = [regex]::IsMatch($Text, '(?is)<meta\b[^>]*name\s*=\s*["'']viewport["'']')

    if ($hasCharset -and $hasViewport) { return $Text }

    # Insert missing meta tags at start of <head>
    $inserts = ""

    if (-not $hasCharset) {
        $inserts += '<meta charset="utf-8">'
    }
    if (-not $hasViewport) {
        $inserts += '<meta name="viewport" content="width=device-width, initial-scale=1">'
    }

    if ($inserts -ne "") {
        $Text = [regex]::Replace($Text, '(?is)(<head\b[^>]*>)', "`$1$inserts", 1)
    }

    return $Text
}

function Ensure-BodyWrapOnce {
    param([Parameter(Mandatory = $true)][string] $Text)

    # Idempotent wrap markers
    if ([regex]::IsMatch($Text, '(?is)<!--\s*KRV_WRAP_START\s*-->')) {
        return $Text
    }

    # Insert wrap start right after <body...>
    $Text = [regex]::Replace(
        $Text,
        '(?is)(<body\b[^>]*>)',
        "`$1`n<!-- KRV_WRAP_START -->`n<div class=""krv-wrap"">",
        1
    )

    # Insert wrap end right before </body>
    $Text = [regex]::Replace(
        $Text,
        '(?is)</body>',
        "`n</div>`n<!-- KRV_WRAP_END -->`n</body>",
        1
    )

    return $Text
}

function Wrap-SectionCardBestEffort {
    param(
        [Parameter(Mandatory = $true)][string] $Text,
        [Parameter(Mandatory = $true)][string] $SectionTitle
    )

    # Avoid double-wrapping if already wrapped
    $alreadyWrappedPattern = '(?is)<h2\b[^>]*>\s*' + [regex]::Escape($SectionTitle) + '\s*</h2>\s*<div\s+class=["'']krv-card["'']'
    if ([regex]::IsMatch($Text, $alreadyWrappedPattern)) {
        return $Text
    }

    # Wrap: <h2>Title</h2> + (next UL) into card, best-effort.
    $pattern = '(?is)(<h2\b[^>]*>\s*' + [regex]::Escape($SectionTitle) + '\s*</h2>\s*)(<ul\b[^>]*>.*?</ul>)'
    if ([regex]::IsMatch($Text, $pattern)) {
        $Text = [regex]::Replace($Text, $pattern, '$1<div class="krv-card">$2</div>', 1)
    }

    return $Text
}

function Wrap-FirstParagraphAfterH1 {
    param([Parameter(Mandatory = $true)][string] $Text)

    if ([regex]::IsMatch($Text, '(?is)<h1\b[^>]*>.*?</h1>\s*<div\s+class=["'']krv-card["'']>')) {
        return $Text
    }

    $pattern = '(?is)(<h1\b[^>]*>.*?</h1>\s*)(<p\b[^>]*>.*?</p>)'
    if ([regex]::IsMatch($Text, $pattern)) {
        $Text = [regex]::Replace($Text, $pattern, '$1<div class="krv-card">$2</div>', 1)
    }

    return $Text
}

$OutDir = (Resolve-Path -LiteralPath $OutDir).Path
$htmlPath = Join-Path $OutDir $HtmlName

if (-not (Test-Path -LiteralPath $htmlPath)) {
    throw "HTML report not found: $htmlPath"
}

$src = Get-Content -LiteralPath $htmlPath -Raw -Encoding UTF8
$src = $src.Trim()

$src = Ensure-Doctype -Text $src
$src = Ensure-HtmlHeadBodySkeleton -Text $src
$src = Ensure-HeadMetaOnce -Text $src

# Wrap body contents (idempotent)
$src = Ensure-BodyWrapOnce -Text $src

# Best-effort layout improvements (idempotent)
$src = Wrap-SectionCardBestEffort -Text $src -SectionTitle "Artifacts"
$src = Wrap-FirstParagraphAfterH1 -Text $src

Set-Content -LiteralPath $htmlPath -Value $src -Encoding UTF8
Write-Host "Normalized HTML structure: $htmlPath"