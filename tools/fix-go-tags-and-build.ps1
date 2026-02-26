Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function WriteUtf8NoBom([string]$path, [string[]]$lines) {
  $dir = Split-Path -Parent $path
  if ($dir -and !(Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
  $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
  [System.IO.File]::WriteAllText($path, (($lines -join "`n") + "`n"), $utf8NoBom)
  Write-Host "Wrote $path"
}

$bt = [char]96  # literal backtick for Go struct tags

WriteUtf8NoBom ".\internal\risk\risk.go" @(
  "package risk",
  "",
  "type Posture string",
  "",
  "const (",
  "    Low      Posture = ""LOW""",
  "    Moderate Posture = ""MODERATE""",
  "    High     Posture = ""HIGH""",
  "    Critical Posture = ""CRITICAL""",
  ")",
  "",
  "type Rating struct {",
  ("    Score    float64 " + $bt + "json:""score""" + $bt),
  ("    Maturity string  " + $bt + "json:""maturity""" + $bt),
  ("    Posture  Posture " + $bt + "json:""posture""" + $bt),
  "}",
  "",
  "func FromScore(score float64, maturity string) Rating {",
  "    p := Critical",
  "    switch {",
  "    case score >= 90:",
  "        p = Low",
  "    case score >= 70:",
  "        p = Moderate",
  "    case score >= 50:",
  "        p = High",
  "    default:",
  "        p = Critical",
  "    }",
  "    return Rating{Score: score, Maturity: maturity, Posture: p}",
  "}"
)

WriteUtf8NoBom ".\internal\trend\trend.go" @(
  "package trend",
  "",
  "import ""math""",
  "",
  "type Direction string",
  "",
  "const (",
  "    Up   Direction = ""up""",
  "    Down Direction = ""down""",
  "    Flat Direction = ""flat""",
  ")",
  "",
  "type Trend struct {",
  ("    DeltaScore   float64   " + $bt + "json:""deltaScore""" + $bt),
  ("    DeltaPercent float64   " + $bt + "json:""deltaPercent""" + $bt),
  ("    Direction    Direction " + $bt + "json:""direction""" + $bt),
  ("    From         float64   " + $bt + "json:""from""" + $bt),
  ("    To           float64   " + $bt + "json:""to""" + $bt),
  "}",
  "",
  "func Compute(prev, curr float64) Trend {",
  "    d := curr - prev",
  "",
  "    dir := Flat",
  "    if d > 0.00001 {",
  "        dir = Up",
  "    } else if d < -0.00001 {",
  "        dir = Down",
  "    }",
  "",
  "    dp := 0.0",
  "    if math.Abs(prev) > 0.00001 {",
  "        dp = (d / prev) * 100.0",
  "    }",
  "",
  "    return Trend{",
  "        DeltaScore:   round(d, 2),",
  "        DeltaPercent: round(dp, 2),",
  "        Direction:    dir,",
  "        From:         round(prev, 2),",
  "        To:           round(curr, 2),",
  "    }",
  "}",
  "",
  "func round(v float64, places int) float64 {",
  "    p := math.Pow(10, float64(places))",
  "    return math.Round(v*p) / p",
  "}"
)

Write-Host "Running gofmt on source dirs only..."
gofmt -w .\cmd .\internal | Out-Null

Write-Host "Rebuilding scan.exe (forced)..."
if (Test-Path .\scan.exe) { Remove-Item .\scan.exe -Force }
go build -o scan.exe .\cmd\scan

Write-Host "OK: build completed."
