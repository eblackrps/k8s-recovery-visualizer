$ErrorActionPreference = "Stop"

Write-Host "== Build =="
go build -o .\scan.exe .\cmd\scan

Write-Host "== Text mode (dry-run) =="
.\scan.exe -dry-run -out .\out | Out-Host
if ($LASTEXITCODE -ne 0) { throw "Text mode failed with exit code $LASTEXITCODE" }

Write-Host "== JSON mode (dry-run) =="
$json = .\scan.exe -dry-run -ci -out .\out
if ($LASTEXITCODE -ne 0) { throw "JSON mode failed with exit code $LASTEXITCODE" }
$null = $json | ConvertFrom-Json

Write-Host "== Policy fail exit code (min-score 101) =="
$jsonFail = .\scan.exe -dry-run -ci -min-score 101 -out .\out
if ($LASTEXITCODE -ne 2) { throw "Expected exit code 2, got $LASTEXITCODE" }
$null = $jsonFail | ConvertFrom-Json

Write-Host "== PASS =="
