param(
    [Parameter(Mandatory = $false)]
    [string] $OutDir = (Resolve-Path ".\out").Path,

    [Parameter(Mandatory = $false)]
    [string] $HtmlName = "recovery-report.html"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$OutDir = (Resolve-Path -LiteralPath $OutDir).Path
$htmlPath = Join-Path $OutDir $HtmlName

if (-not (Test-Path -LiteralPath $htmlPath)) {
    throw "HTML report not found: $htmlPath"
}

$src = Get-Content -LiteralPath $htmlPath -Raw -Encoding UTF8

# Idempotent: if our theme already exists, do nothing.
if ([regex]::IsMatch($src, '(?is)<style\b[^>]*id\s*=\s*["'']krv-dark-theme["'']')) {
    Write-Host "Applied dark theme CSS: $htmlPath"
    Write-Host "Theme action: Already present (no changes)"
    return
}

# Require body so we can inject last and win the cascade.
if (-not [regex]::IsMatch($src, '(?is)<body\b[^>]*>')) {
    throw "Apply-DarkTheme: <body> not found. Normalize-ReportHtml.ps1 must run first. File: $htmlPath"
}

$themeCss = @'
<style id="krv-dark-theme">
/* K8s Recovery Visualizer - Dark Theme
   Injected at end of <body> so it wins against generator CSS.
*/
:root { color-scheme: dark; }

html, body {
  background: #0d1117 !important;
  color: #e6edf3 !important;
  font-family: "Segoe UI", Arial, sans-serif !important;
  line-height: 1.5 !important;
}

body {
  margin: 0 !important;
}

.krv-wrap {
  max-width: 1100px;
  margin: 0 auto;
  padding: 22px 22px 34px 22px;
}

.krv-card {
  background: #161b22 !important;
  border: 1px solid #30363d !important;
  border-radius: 12px !important;
  padding: 14px 16px !important;
  margin: 12px 0 !important;
}

h1 {
  color: #f0f6fc !important;
  font-size: 44px !important;
  margin: 10px 0 14px 0 !important;
  letter-spacing: 0.2px !important;
}

h2 {
  color: #f0f6fc !important;
  font-size: 32px !important;
  margin: 26px 0 10px 0 !important;
}

h3, h4 { color: #c9d1d9 !important; }

p, li, ul, ol {
  color: #e6edf3 !important;
  font-size: 16px !important;
}

ul, ol { margin-top: 8px !important; }
li { margin: 6px 0 !important; }

a, a:visited {
  color: #58a6ff !important;
  text-decoration: none !important;
}
a:hover { text-decoration: underline !important; }

/* Tables */
table {
  border-collapse: collapse !important;
  width: 100% !important;
  margin: 10px 0 !important;
}
th, td {
  border: 1px solid #30363d !important;
  padding: 8px 10px !important;
  vertical-align: top !important;
  color: #e6edf3 !important;
}
th { background: #0b1220 !important; }
tr:nth-child(even) td { background: #0f172a !important; }

/* The “text is ass” section:
   Markdown generators often render values as <code> or similar and give them light backgrounds.
   We hammer those into readable dark pills.
*/
code, kbd, samp, tt {
  background: #161b22 !important;
  color: #f0f6fc !important;
  border: 1px solid #30363d !important;
  border-radius: 6px !important;
  padding: 2px 6px !important;
  font-size: 0.95em !important;
}

/* Some generators wrap inline code inside <li> and then style the <li> or <code> differently */
li code, p code, td code {
  background: #161b22 !important;
  color: #f0f6fc !important;
  border-color: #30363d !important;
}

/* Code blocks */
pre {
  background: #0b1220 !important;
  color: #e6edf3 !important;
  border: 1px solid #30363d !important;
  border-radius: 10px !important;
  padding: 10px 12px !important;
  overflow-x: auto !important;
}

/* Make sure nothing is “white on pale gray” */
* {
  text-shadow: none !important;
}

/* Horizontal rules */
hr {
  border: 0 !important;
  border-top: 1px solid #30363d !important;
}

/* If your report uses red for FAIL, keep it readable */
.fail, .error, .danger {
  color: #ff7b72 !important;
}
</style>
'@

$beforeSize = $src.Length

if ([regex]::IsMatch($src, '(?is)</body>')) {
    $src = [regex]::Replace($src, '(?is)</body>', "`n$themeCss`n</body>", 1)
}
elseif ([regex]::IsMatch($src, '(?is)</html>')) {
    # fallback if body close missing but html close exists
    $src = [regex]::Replace($src, '(?is)</html>', "`n$themeCss`n</html>", 1)
}
else {
    # worst-case append
    $src = "$src`n$themeCss"
}

$afterSize = $src.Length
Set-Content -LiteralPath $htmlPath -Value $src -Encoding UTF8

Write-Host "Applied dark theme CSS: $htmlPath"
Write-Host "Theme action: Injected at end of document (wins cascade)"
Write-Host ("File size: {0} -> {1}" -f $beforeSize, $afterSize)