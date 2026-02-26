param(
  [string]$MainPath = "cmd\scan\main.go"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function MustExist($p) { if (!(Test-Path $p)) { throw "Missing: $p" } }

# Sanity checks
MustExist ".\internal\enrich\enrich.go"
MustExist ".\internal\enrich\report.go"
MustExist ".\internal\risk\risk.go"
MustExist ".\internal\trend\trend.go"
Write-Host "Enrich packages present."

MustExist "go.mod"
$modLine = (Get-Content go.mod | Select-String -Pattern '^module\s+').Line
if (!$modLine) { throw "Could not detect module path from go.mod" }
$modulePath = ($modLine -replace '^module\s+', '').Trim()
Write-Host "Module: $modulePath"

# Ensure placeholder imports are patched
$enrichPath = "internal\enrich\enrich.go"
$enrich = Get-Content $enrichPath -Raw
if ($enrich -match 'github.com/your/module') {
  $enrich = $enrich -replace 'github.com/your/module', $modulePath
  Set-Content -Path $enrichPath -Value $enrich -Encoding utf8
  Write-Host "Patched import placeholders in $enrichPath"
}

# Read main.go lines
MustExist $MainPath
$lines = Get-Content $MainPath

# Ensure enrich import exists (simple text check)
$importNeedle = "`"$modulePath/internal/enrich`""
if (-not ($lines -join "`n" | Select-String -SimpleMatch $importNeedle)) {
  # Insert into import ( block after the 'import (' line
  $idxImport = -1
  for ($i=0; $i -lt $lines.Count; $i++) {
    if ($lines[$i].Trim() -eq "import (") { $idxImport = $i; break }
  }
  if ($idxImport -lt 0) { throw "No 'import (' block found. Can't auto-insert enrich import safely." }

  $new = New-Object System.Collections.Generic.List[string]
  for ($i=0; $i -lt $lines.Count; $i++) {
    $new.Add($lines[$i])
    if ($i -eq $idxImport) {
      $new.Add("`t$importNeedle")
    }
  }
  $lines = $new.ToArray()
  Write-Host "Inserted enrich import."
} else {
  Write-Host "Enrich import already present."
}

# Remove any old Phase2 hook blocks (line-based, no regex)
$startMarker = "`t// Phase2: enrich reports (trend + risk)"
$endMarker   = "`t}"

$clean = New-Object System.Collections.Generic.List[string]
$skipping = $false
$skipDepth = 0

for ($i=0; $i -lt $lines.Count; $i++) {
  $line = $lines[$i]

  if (-not $skipping -and $line -eq $startMarker) {
    # Skip until we exit the injected if/else block (we'll detect by braces)
    $skipping = $true
    $skipDepth = 0
    continue
  }

  if ($skipping) {
    if ($line -match "^\s*if\s+en,\s+err\s+:=") { $skipDepth++ }
    if ($line.Trim() -eq "}") {
      if ($skipDepth -gt 0) { $skipDepth-- }
      else {
        # end of outer injected block
        $skipping = $false
      }
    }
    continue
  }

  $clean.Add($line)
}
$lines = $clean.ToArray()

# Determine outDir variable name (best effort)
$outVar = "outDir"

# Find a likely assignment like: outDir := ...
for ($i=0; $i -lt $lines.Count; $i++) {
  if ($lines[$i] -match '^\s*(outDir)\s*:?=\s*') { $outVar = "outDir"; break }
}
Write-Host "Using output dir variable: $outVar"

# Hook to insert (LOUD)
$hookLines = @(
  "`t// Phase2: enrich reports (trend + risk)",
  "`tif en, err := enrich.Run(enrich.Options{OutDir: $outVar, LastNCount: 10}); err != nil {",
  "`t`tfmt.Printf(`"Enrich: FAILED (%v)\n`", err)",
  "`t} else if err := enrich.WriteArtifacts($outVar, en); err != nil {",
  "`t`tfmt.Printf(`"Enrich: FAILED writing artifacts (%v)\n`", err)",
  "`t} else {",
  "`t`tfmt.Printf(`"Enriched: %s\\recovery-enriched.json\n`", $outVar)",
  "`t}"
)

# Insert after first line containing "Scan complete."
$insertAt = -1
for ($i=0; $i -lt $lines.Count; $i++) {
  if ($lines[$i] -like '*Scan complete.*') { $insertAt = $i + 1; break }
}

if ($insertAt -lt 0) {
  # fallback: insert near end of main.go before last line
  $insertAt = [Math]::Max(0, $lines.Count - 1)
  Write-Host "Did not find 'Scan complete.' line. Using fallback insertion near EOF."
} else {
  Write-Host "Found 'Scan complete.' line. Inserting hook after it."
}

$newLines = New-Object System.Collections.Generic.List[string]
for ($i=0; $i -lt $lines.Count; $i++) {
  $newLines.Add($lines[$i])
  if ($i -eq ($insertAt - 1)) {
    foreach ($h in $hookLines) { $newLines.Add($h) }
    $newLines.Add("") # spacer
  }
}

Set-Content -Path $MainPath -Value ($newLines -join "`n") -Encoding utf8
Write-Host "Patched $MainPath"

Write-Host "Running gofmt..."
gofmt -w . | Out-Null

Write-Host "Building scan.exe..."
go build -o scan.exe .\cmd\scan

Write-Host "Patch applied successfully."
