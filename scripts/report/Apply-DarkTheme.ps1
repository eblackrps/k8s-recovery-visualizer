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

$content = Get-Content -LiteralPath $htmlPath -Raw -Encoding UTF8

# Idempotent: if already present, do nothing
if ($content -match '(?is)<style\b[^>]*id\s*=\s*["'']krv-dark-theme["'']') {
    Write-Host "Applied dark theme CSS: $htmlPath"
    Write-Host "Theme action: Already present (no changes)"
    exit 0
}

$themeBlock = @"
<style id="krv-dark-theme">
:root{
  color-scheme: dark;
}

body{
  background-color: #0d1117;
  color: #c9d1d9;
  font-family: Segoe UI, Arial, sans-serif;
  margin: 20px;
  line-height: 1.45;
}

h1, h2, h3, h4{
  color: #f0f6fc;
}

h2{
  margin-top: 28px;
}

a{ color: #58a6ff; }
a:hover{ color: #79c0ff; }

table{
  border-collapse: collapse;
  width: 100%;
  margin: 10px 0;
}

th, td{
  border: 1px solid #30363d;
  padding: 8px;
  vertical-align: top;
}

th{
  background-color: #161b22;
}

tr:nth-child(even){
  background-color: #0f1520;
}

code, pre{
  background-color: #161b22;
  color: #c9d1d9;
  padding: 4px;
  border-radius: 6px;
}

pre{
  padding: 12px;
  overflow: auto;
}

.krv-wrap{
  max-width: 1200px;
}

.krv-card{
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 12px;
  padding: 14px 16px;
  margin: 10px 0 18px 0;
}

/* Validation Summary / Top Issues lead-in styling */
.krv-leadin{
  margin: 10px 0 4px 0;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid rgba(255,255,255,0.10);
  background: rgba(255,255,255,0.03);
  font-weight: 700;
}

.krv-leadin.krv-success{
  border-color: rgba(46,160,67,0.45);
  background: rgba(46,160,67,0.14);
  color: #7ee787;
}

.krv-leadin.krv-attn{
  border-color: rgba(210,153,34,0.45);
  background: rgba(210,153,34,0.14);
  color: #f2cc60;
}

.krv-meta{
  margin: 0 0 12px 0;
  color: rgba(201,209,217,0.80);
  font-size: 12.5px;
}
</style>
"@

# Insert theme inside <head>
if ($content -match '(?is)<head\b[^>]*>') {
    $content = [regex]::Replace(
        $content,
        '(?is)(<head\b[^>]*>)',
        '$1' + "`n$themeBlock",
        1
    )
    Set-Content -LiteralPath $htmlPath -Value $content -Encoding UTF8
    Write-Host "Applied dark theme CSS: $htmlPath"
    Write-Host "Theme action: Injected into existing <head>"
    exit 0
}

throw "Apply-DarkTheme: <head> section not found. Normalize first."