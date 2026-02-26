param(
  [string]$MainPath = "cmd\scan\main.go"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function FailIfLastExit([string]$what) { if ($LASTEXITCODE -ne 0) { throw "$what failed with exit code $LASTEXITCODE" } }
function MustExist($p) { if (!(Test-Path $p)) { throw "Missing: $p" } }

MustExist $MainPath
MustExist ".\go.mod"

$modLine = (Get-Content .\go.mod | Select-String -Pattern '^module\s+').Line
if (!$modLine) { throw "Could not detect module path from go.mod" }
$modulePath = ($modLine -replace '^module\s+', '').Trim()

$lines = Get-Content $MainPath

# 1) Ensure enrich import exists
$importNeedle = "`"$modulePath/internal/enrich`""
$allText = $lines -join "`n"
if ($allText -notmatch [regex]::Escape($importNeedle)) {
  $idxImport = -1
  for ($i=0; $i -lt $lines.Count; $i++) {
    if ($lines[$i].Trim() -eq "import (") { $idxImport = $i; break }
  }
  if ($idxImport -lt 0) { throw "No import ( block found in main.go" }

  $new = New-Object System.Collections.Generic.List[string]
  for ($i=0; $i -lt $lines.Count; $i++) {
    $new.Add($lines[$i])
    if ($i -eq $idxImport) { $new.Add("`t$importNeedle") }
  }
  $lines = $new.ToArray()
  Write-Host "Inserted enrich import."
} else {
  Write-Host "Enrich import already present."
}

# 2) Remove any previous injected Phase2 blocks (simple marker-based delete)
$startMarker = "`t// Phase2: enrich reports (trend + risk)"
$clean = New-Object System.Collections.Generic.List[string]
$skip = $false
$braceDepth = 0

for ($i=0; $i -lt $lines.Count; $i++) {
  $line = $lines[$i]

  if (-not $skip -and $line -eq $startMarker) {
    $skip = $true
    $braceDepth = 0
    continue
  }

  if ($skip) {
    # track braces to find end of injected if/else block
    if ($line -match '{') { $braceDepth++ }
    if ($line -match '}') {
      if ($braceDepth -le 0) {
        $skip = $false
      } else {
        $braceDepth--
      }
    }
    continue
  }

  $clean.Add($line)
}
$lines = $clean.ToArray()

# 3) Insert a NEW minimal hook that references ONLY outDir and fmt (fmt already in use)
#    This avoids mdPath/htmlPath/etc entirely.
$hook = @(
  "`t// Phase2: enrich reports (trend + risk)",
  "`tif en, err := enrich.Run(enrich.Options{OutDir: outDir, LastNCount: 10}); err != nil {",
  "`t`tfmt.Printf(`"Enrich: FAILED (%v)\n`", err)",
  "`t} else if err := enrich.WriteArtifacts(outDir, en); err != nil {",
  "`t`tfmt.Printf(`"Enrich: FAILED writing artifacts (%v)\n`", err)",
  "`t} else {",
  "`t`tfmt.Printf(`"Enriched: %s\\recovery-enriched.json\n`", outDir)",
  "`t}",
  ""
)

# Insert before the last line containing just "}" (end of main func)
# We find the LAST closing brace in the file and insert before it.
$insertAt = -1
for ($i=$lines.Count-1; $i -ge 0; $i--) {
  if ($lines[$i].Trim() -eq "}") { $insertAt = $i; break }
}
if ($insertAt -lt 0) { throw "Could not find a closing brace to insert before." }

$newLines = New-Object System.Collections.Generic.List[string]
for ($i=0; $i -lt $lines.Count; $i++) {
  if ($i -eq $insertAt) {
    foreach ($h in $hook) { $newLines.Add($h) }
  }
  $newLines.Add($lines[$i])
}

Set-Content -Path $MainPath -Value ($newLines -join "`n") -Encoding utf8
Write-Host "Patched main.go hook (minimal)."

Write-Host "Running gofmt..."
gofmt -w .\cmd .\internal | Out-Null
FailIfLastExit "gofmt"

Write-Host "Building scan.exe..."
if (Test-Path .\scan.exe) { Remove-Item .\scan.exe -Force }
go build -o scan.exe .\cmd\scan
FailIfLastExit "go build"

Write-Host "OK: scan.exe built."
