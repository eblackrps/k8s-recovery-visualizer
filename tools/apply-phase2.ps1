param(
  [string]$MainPath = "cmd\scan\main.go"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Write-TextFile($Path, $Content) {
  $dir = Split-Path -Parent $Path
  if ($dir -and !(Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
  $Content | Out-File -FilePath $Path -Encoding utf8 -Force
  Write-Host "Wrote $Path"
}

# 1) New internal packages (trend, risk, enrich)
Write-TextFile "internal\trend\trend.go" @"
package trend

import "math"

type Direction string

const (
Up   Direction = "up"
Down Direction = "down"
Flat Direction = "flat"
)

type Trend struct {
DeltaScore   float64   \`json:"deltaScore"\`
DeltaPercent float64   \`json:"deltaPercent"\`
Direction    Direction \`json:"direction"\`
From         float64   \`json:"from"\`
To           float64   \`json:"to"\`
}

func Compute(prev, curr float64) Trend {
d := curr - prev

dir := Flat
if d > 0.00001 {
dir = Up
} else if d < -0.00001 {
dir = Down
}

dp := 0.0
if math.Abs(prev) > 0.00001 {
dp = (d / prev) * 100.0
}

// round lightly for reports
return Trend{
DeltaScore:   round(d, 2),
DeltaPercent: round(dp, 2),
Direction:    dir,
From:         round(prev, 2),
To:           round(curr, 2),
}
}

func round(v float64, places int) float64 {
p := math.Pow(10, float64(places))
return math.Round(v*p) / p
}
"@

Write-TextFile "internal\risk\risk.go" @"
package risk

type Posture string

const (
Low      Posture = "LOW"
Moderate Posture = "MODERATE"
High     Posture = "HIGH"
Critical Posture = "CRITICAL"
)

type Rating struct {
Score    float64 \`json:"score"\`
Maturity string  \`json:"maturity"\`
Posture  Posture \`json:"posture"\`
}

func FromScore(score float64, maturity string) Rating {
// Use score primarily. maturity is passed through for report consistency.
p := Critical
switch {
case score >= 90:
p = Low
case score >= 70:
p = Moderate
case score >= 50:
p = High
default:
p = Critical
}
return Rating{Score: score, Maturity: maturity, Posture: p}
}
"@

Write-TextFile "internal\enrich\enrich.go" @"
package enrich

import (
"encoding/json"
"fmt"
"os"
"path/filepath"
"time"

"github.com/your/module/internal/risk"
"github.com/your/module/internal/trend"
)

// NOTE: Update module import path above once, if needed.
// This file is otherwise hands-free.

type HistoryIndex struct {
Entries []HistoryEntry \`json:"entries"\`
}

type HistoryEntry struct {
TimestampUtc string  \`json:"timestampUtc"\`
Overall      float64 \`json:"overall"\`
Maturity     string  \`json:"maturity"\`
}

type Enriched struct {
GeneratedUtc string       \`json:"generatedUtc"\`
Current      HistoryEntry \`json:"current"\`
Previous     *HistoryEntry \`json:"previous,omitempty"\`
Trend        *trend.Trend \`json:"trend,omitempty"\`
Risk         risk.Rating  \`json:"risk"\`
LastN        []float64    \`json:"lastN"\`
}

type Options struct {
OutDir     string
LastNCount int
}

func Run(opts Options) (*Enriched, error) {
if opts.OutDir == "" {
opts.OutDir = "out"
}
if opts.LastNCount <= 0 {
opts.LastNCount = 10
}

historyPath := filepath.Join(opts.OutDir, "history", "index.json")
b, err := os.ReadFile(historyPath)
if err != nil {
// No history: still produce something minimal
return &Enriched{
GeneratedUtc: time.Now().UTC().Format(time.RFC3339),
Risk:         risk.FromScore(0, ""),
LastN:        []float64{},
}, nil
}

var idx HistoryIndex
if err := json.Unmarshal(b, &idx); err != nil {
return nil, fmt.Errorf("parse history index: %w", err)
}
if len(idx.Entries) == 0 {
return &Enriched{
GeneratedUtc: time.Now().UTC().Format(time.RFC3339),
Risk:         risk.FromScore(0, ""),
LastN:        []float64{},
}, nil
}

curr := idx.Entries[len(idx.Entries)-1]
var prev *HistoryEntry
if len(idx.Entries) >= 2 {
p := idx.Entries[len(idx.Entries)-2]
prev = &p
}

last := make([]float64, 0, min(opts.LastNCount, len(idx.Entries)))
for i := max(0, len(idx.Entries)-opts.LastNCount); i < len(idx.Entries); i++ {
last = append(last, idx.Entries[i].Overall)
}

var tr *trend.Trend
if prev != nil {
t := trend.Compute(prev.Overall, curr.Overall)
tr = &t
}

en := &Enriched{
GeneratedUtc: time.Now().UTC().Format(time.RFC3339),
Current:      curr,
Previous:     prev,
Trend:        tr,
Risk:         risk.FromScore(curr.Overall, curr.Maturity),
LastN:        last,
}
return en, nil
}

func min(a, b int) int {
if a < b { return a }
return b
}

func max(a, b int) int {
if a > b { return a }
return b
}
"@

Write-TextFile "internal\enrich\report.go" @"
package enrich

import (
"encoding/json"
"fmt"
"os"
"path/filepath"
"strings"
)

func WriteArtifacts(outDir string, en *Enriched) error {
if outDir == "" {
outDir = "out"
}
// JSON
jpath := filepath.Join(outDir, "recovery-enriched.json")
jb, _ := json.MarshalIndent(en, "", "  ")
if err := os.WriteFile(jpath, jb, 0644); err != nil {
return fmt.Errorf("write enriched json: %w", err)
}

// Markdown append
mdPath := filepath.Join(outDir, "recovery-report.md")
if b, err := os.ReadFile(mdPath); err == nil {
aug := string(b) + "\n\n" + renderMarkdown(en) + "\n"
if err := os.WriteFile(mdPath, []byte(aug), 0644); err != nil {
return fmt.Errorf("append md report: %w", err)
}
}

// HTML append (best-effort)
htmlPath := filepath.Join(outDir, "recovery-report.html")
if b, err := os.ReadFile(htmlPath); err == nil {
aug := injectHTML(string(b), en)
if err := os.WriteFile(htmlPath, []byte(aug), 0644); err != nil {
return fmt.Errorf("append html report: %w", err)
}
}

return nil
}

func renderMarkdown(en *Enriched) string {
var sb strings.Builder
sb.WriteString("## Trend & Risk\n\n")
sb.WriteString(fmt.Sprintf("- **DR Risk Posture:** %s\n", en.Risk.Posture))
if en.Trend == nil || en.Previous == nil {
sb.WriteString("- **Trend:** FIRST RUN (no previous scan found)\n")
} else {
arrow := "→"
switch en.Trend.Direction {
case "up":
arrow = "↑"
case "down":
arrow = "↓"
}
sb.WriteString(fmt.Sprintf("- **Trend:** %s %+0.2f ( %+0.2f%% )\n", arrow, en.Trend.DeltaScore, en.Trend.DeltaPercent))
sb.WriteString(fmt.Sprintf("- **Previous:** %0.2f (%s)\n", en.Trend.From, en.Previous.Maturity))
sb.WriteString(fmt.Sprintf("- **Current:** %0.2f (%s)\n", en.Trend.To, en.Current.Maturity))
}
if len(en.LastN) > 0 {
sb.WriteString("\n**Last runs:** ")
for i, v := range en.LastN {
if i > 0 { sb.WriteString(" → ") }
sb.WriteString(fmt.Sprintf("%0.2f", v))
}
sb.WriteString("\n")
}
return sb.String()
}

func injectHTML(html string, en *Enriched) string {
// Insert before </body> if present, else append.
block := renderHTML(en)
idx := strings.LastIndex(strings.ToLower(html), "</body>")
if idx == -1 {
return html + "\n" + block
}
return html[:idx] + "\n" + block + "\n" + html[idx:]
}

func renderHTML(en *Enriched) string {
// Inline Chart.js CDN + simple line chart of LastN
labels := make([]string, 0, len(en.LastN))
data := make([]string, 0, len(en.LastN))
for i, v := range en.LastN {
labels = append(labels, fmt.Sprintf("\"Run %d\"", i+1))
data = append(data, fmt.Sprintf("%0.2f", v))
}

trendLine := "FIRST RUN"
if en.Trend != nil {
trendLine = fmt.Sprintf("%s %+0.2f (%+0.2f%%)", strings.ToUpper(string(en.Trend.Direction)), en.Trend.DeltaScore, en.Trend.DeltaPercent)
}

return fmt.Sprintf(\`
<section style="margin-top:24px; padding:16px; border:1px solid #ddd; border-radius:12px;">
  <h2>Trend &amp; Risk</h2>
  <ul>
    <li><b>DR Risk Posture:</b> %s</li>
    <li><b>Trend:</b> %s</li>
  </ul>

  <div style="max-width:900px">
    <canvas id="drTrendChart"></canvas>
  </div>

  <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
  <script>
    (function(){
      const ctx = document.getElementById('drTrendChart');
      if(!ctx) return;
      new Chart(ctx, {
        type: 'line',
        data: {
          labels: [%s],
          datasets: [{
            label: 'Overall Score',
            data: [%s],
            tension: 0.25
          }]
        },
        options: {
          responsive: true,
          plugins: { legend: { display: true } },
          scales: {
            y: { suggestedMin: 0, suggestedMax: 100 }
          }
        }
      });
    })();
  </script>
</section>
\`, en.Risk.Posture, trendLine, strings.Join(labels, ","), strings.Join(data, ","))
}
"@

# 2) Auto-fix module import path in internal/enrich/enrich.go
# We will derive module path from go.mod
$goMod = "go.mod"
if (!(Test-Path $goMod)) { throw "go.mod not found in repo root." }
$modLine = (Get-Content $goMod | Select-String -Pattern '^module\s+').Line
if (!$modLine) { throw "Could not detect module path from go.mod" }
$modulePath = ($modLine -replace '^module\s+', '').Trim()

$enrichPath = "internal\enrich\enrich.go"
$enrich = Get-Content $enrichPath -Raw
$enrich = $enrich -replace 'github.com/your/module', $modulePath
Set-Content -Path $enrichPath -Value $enrich -Encoding utf8
Write-Host "Patched import path in $enrichPath => $modulePath"

# 3) Patch cmd\scan\main.go to call enrich after scan completes.
if (!(Test-Path $MainPath)) { throw "Main not found at $MainPath" }
$main = Get-Content $MainPath -Raw

if ($main -match 'internal/enrich') {
  Write-Host "main.go already references internal/enrich, skipping patch."
} else {
  # Insert import
  $main = $main -replace '(?s)import\s*\(\s*', "import (`n`t`"$modulePath/internal/enrich`"`n"
  
  # Insert call near end: try to place after the string "Scan complete." print or right before return.
  if ($main -match 'Scan complete\.') {
    $main = $main -replace 'Scan complete\.\s*"', 'Scan complete."'
    $main = $main -replace '(?s)(Scan complete\.\s*"\)\s*)', "`$1`n`t// Phase2: enrich reports (trend + risk)`n`tif en, err := enrich.Run(enrich.Options{OutDir: outDir, LastNCount: 10}); err == nil { _ = enrich.WriteArtifacts(outDir, en) }`n"
  } else {
    # Fallback: inject before final return in main
    $main = $main -replace '(?s)\n(\s*return\s*)', "`n`t// Phase2: enrich reports (trend + risk)`n`tif en, err := enrich.Run(enrich.Options{OutDir: outDir, LastNCount: 10}); err == nil { _ = enrich.WriteArtifacts(outDir, en) }`n`n`$1"
  }

  Set-Content -Path $MainPath -Value $main -Encoding utf8
  Write-Host "Patched $MainPath with enrich hook."
}

# 4) gofmt everything and build
Write-Host "Running gofmt..."
gofmt -w . | Out-Null

Write-Host "Building scan.exe..."
go build -o scan.exe .\cmd\scan

Write-Host "Phase 2 applied successfully."
