param(
  [string]$MainPath = "cmd\scan\main.go"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function MustExist($p) {
  if (!(Test-Path $p)) { throw "Missing: $p" }
}

# Sanity: ensure enrich package exists
MustExist ".\internal\enrich\enrich.go"
MustExist ".\internal\enrich\report.go"
MustExist ".\internal\risk\risk.go"
MustExist ".\internal\trend\trend.go"
Write-Host "Enrich packages present."

# Detect module path from go.mod
$goMod = "go.mod"
MustExist $goMod
$modLine = (Get-Content $goMod | Select-String -Pattern '^module\s+').Line
if (!$modLine) { throw "Could not detect module path from go.mod" }
$modulePath = ($modLine -replace '^module\s+', '').Trim()
Write-Host "Module: $modulePath"

# Ensure internal/enrich/enrich.go imports are correct
$enrichPath = "internal\enrich\enrich.go"
$enrich = Get-Content $enrichPath -Raw
if ($enrich -match 'github.com/your/module') {
  $enrich = $enrich -replace 'github.com/your/module', $modulePath
  Set-Content -Path $enrichPath -Value $enrich -Encoding utf8
  Write-Host "Patched import placeholders in $enrichPath"
}

# Patch main.go to import enrich + run it with loud logging
MustExist $MainPath
$main = Get-Content $MainPath -Raw

$importLine = "`t`"$modulePath/internal/enrich`""

# 1) Ensure import exists
if ($main -notmatch [regex]::Escape($importLine)) {
  if ($main -match '(?s)import\s*\(\s*') {
    $main = $main -replace '(?s)import\s*\(\s*', "import (`n$importLine`n"
    Write-Host "Inserted enrich import into import() block."
  } else {
    throw "main.go has no import() block. Can't auto-insert enrich import safely."
  }
} else {
  Write-Host "Enrich import already present."
}

# 2) Remove any old Phase2 hook blocks we previously injected (avoid duplicates)
$main = [regex]::Replace($main, '(?s)\n\t// Phase2: enrich reports \(trend \+ risk\).*?\n', "`n")

# 3) Find outDir variable name (best-effort)
# We look for the line that prints "Output:" and capture the variable used in fmt.Printf/Printf-like calls.
# If not found, we default to outDir.
$outVar = "outDir"

$matches = [regex]::Matches($main, 'Output:\s*%s.*?\)\s*,\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)', [System.Text.RegularExpressions.RegexOptions]::Singleline)
if ($matches.Count -gt 0) {
  $outVar = $matches[0].Groups[1].Value
}

Write-Host "Using output dir variable: $outVar"

# 4) Inject new hook AFTER the line that prints "Scan complete." or at end of main().
$hook = @"
// Phase2: enrich reports (trend + risk) - loud mode
if en, err := enrich.Run(enrich.Options{OutDir: $outVar, LastNCount: 10}); err != nil {
fmt.Printf("Enrich: FAILED (%v)\n", err)
} else if err := enrich.WriteArtifacts($outVar, en); err != nil {
fmt.Printf("Enrich: FAILED writing artifacts (%v)\n", err)
} else {
fmt.Printf("Enriched: %s\\recovery-enriched.json\n", $outVar)
}
"@

if ($main -match 'Scan complete\.') {
  # Insert right after the Scan complete print call line (best effort)
  $main = [regex]::Replace(
    $main,
    '(Scan complete\.\s*"\)\s*)',
    "`$1`n$hook`n",
    1,
    [System.Text.RegularExpressions.RegexOptions]::Singleline
  )
  Write-Host "Injected hook after 'Scan complete.' print."
} else {
  # Fallback: insert before the last closing brace of main (best effort)
  $idx = $main.LastIndexOf("}")
  if ($idx -lt 0) { throw "Could not find a closing brace in main.go" }
  $main = $main.Substring(0, $idx) + "`n" + $hook + "`n" + $main.Substring($idx)
  Write-Host "Injected hook near end of file (fallback)."
}

Set-Content -Path $MainPath -Value $main -Encoding utf8
Write-Host "Patched $MainPath"

# gofmt + rebuild
Write-Host "Running gofmt..."
gofmt -w . | Out-Null

Write-Host "Building scan.exe..."
go build -o scan.exe .\cmd\scan

Write-Host "Patch applied successfully."
