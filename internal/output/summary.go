package output

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"sort"

	"k8s-recovery-visualizer/internal/model"
)

// WriteSummary writes a print-optimised single-page executive summary to path.
func WriteSummary(path string, b *model.Bundle) error {
	var buf bytes.Buffer
	buildSummary(&buf, b)
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func buildSummary(buf *bytes.Buffer, b *model.Bundle) {
	w := func(s string) { buf.WriteString(s) }
	wf := func(f string, a ...any) { buf.WriteString(fmt.Sprintf(f, a...)) }
	e := html.EscapeString

	matColor := map[string]string{
		"PLATINUM": "#79c0ff", "GOLD": "#f2cc60",
		"SILVER": "#c9d1d9", "BRONZE": "#ffa657",
	}[b.Score.Maturity]
	if matColor == "" {
		matColor = "#c9d1d9"
	}

	platform := b.Cluster.Platform.Provider
	if platform == "" {
		platform = "unknown"
	}
	backupTool := b.Inventory.Backup.PrimaryTool
	if backupTool == "" {
		backupTool = "none"
	}

	w(`<!DOCTYPE html><html lang="en"><head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>DR Executive Summary</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#fff;color:#111;font-family:"Segoe UI",Arial,sans-serif;font-size:13px;line-height:1.5;padding:32px 40px}
h1{font-size:1.4em;font-weight:700;margin-bottom:2px}
h2{font-size:1em;font-weight:600;margin:18px 0 6px;border-bottom:1px solid #e0e0e0;padding-bottom:3px}
.meta{color:#555;font-size:.82em;margin-bottom:20px}
.score-hero{display:flex;align-items:center;gap:24px;margin-bottom:20px;padding:16px 20px;border:1px solid #e0e0e0;border-radius:6px;background:#fafafa}
.score-big{font-size:3.2em;font-weight:800;line-height:1}
.badge{display:inline-block;padding:4px 14px;border-radius:14px;font-weight:700;font-size:1em;border:2px solid}
.domains{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin-bottom:18px}
.dom{border:1px solid #e0e0e0;border-radius:5px;padding:10px;text-align:center;background:#fafafa}
.dom .v{font-size:1.8em;font-weight:700}
.dom .l{font-size:.74em;color:#555;margin-top:1px}
.dom .bar{background:#e0e0e0;border-radius:3px;height:4px;margin-top:6px}
.dom .fill{height:4px;border-radius:3px;background:#555}
table{width:100%;border-collapse:collapse;margin-top:4px;font-size:.84em}
th{background:#f0f0f0;text-align:left;padding:5px 8px;border-bottom:1px solid #ccc;font-weight:600}
td{padding:5px 8px;border-bottom:1px solid #e8e8e8;vertical-align:top}
.c-CRITICAL{color:#c0392b;font-weight:600}
.c-HIGH{color:#d35400;font-weight:600}
.c-MEDIUM{color:#7d6608}
.c-LOW,.c-INFO{color:#555}
.ok{color:#1a7a1a}.bad{color:#c0392b}
.chip{display:inline-block;padding:1px 8px;border-radius:10px;font-size:.77em;border:1px solid #ccc;margin:1px}
.footer{margin-top:28px;color:#888;font-size:.78em;border-top:1px solid #e0e0e0;padding-top:10px}
.print-btn{display:inline-block;margin-bottom:20px;padding:7px 18px;background:#1a1a2e;color:#fff;border:none;border-radius:4px;cursor:pointer;font-size:.86em}
@media print{
  .print-btn{display:none}
  body{padding:16px 20px}
  @page{margin:1.5cm}
}
</style></head><body>
`)

	// Print button (hidden when printing)
	w(`<button class="print-btn" onclick="window.print()">Print / Save as PDF</button>`)

	// Title
	wf(`<h1>Kubernetes DR Readiness — Executive Summary</h1>`)
	metaParts := []string{}
	if b.Metadata.CustomerID != "" {
		metaParts = append(metaParts, "Customer: "+b.Metadata.CustomerID)
	}
	if b.Metadata.ClusterName != "" {
		metaParts = append(metaParts, "Cluster: "+b.Metadata.ClusterName)
	}
	if b.Metadata.Site != "" {
		metaParts = append(metaParts, "Site: "+b.Metadata.Site)
	}
	if b.Metadata.Environment != "" {
		metaParts = append(metaParts, "Env: "+b.Metadata.Environment)
	}
	metaParts = append(metaParts, "Scan: "+b.Metadata.GeneratedAt)
	wf(`<div class="meta">`)
	for i, p := range metaParts {
		if i > 0 {
			w(` &nbsp;|&nbsp; `)
		}
		w(e(p))
	}
	w(`</div>`)

	// Score hero
	wf(`<div class="score-hero">
<div class="score-big">%d<span style="font-size:.4em;color:#555">/100</span></div>
<div>
<div class="badge" style="color:%s;border-color:%s">%s</div>
<div style="color:#555;font-size:.82em;margin-top:6px">Platform: %s &nbsp; Backup: %s &nbsp; Target: %s</div>
</div>
</div>`,
		b.Score.Overall.Final,
		matColor, matColor, e(b.Score.Maturity),
		e(platform), e(backupTool), e(b.Target))

	// Domain breakdown
	w(`<div class="domains">`)
	for _, d := range []struct {
		label, weight string
		score         int
	}{
		{"Storage", "35%", b.Score.Storage.Final},
		{"Workload", "20%", b.Score.Workload.Final},
		{"Config", "15%", b.Score.Config.Final},
		{"Backup / Recovery", "30%", b.Score.Backup.Final},
	} {
		wf(`<div class="dom"><div class="v">%d</div><div class="l">%s <span style="color:#888">%s</span></div><div class="bar"><div class="fill" style="width:%d%%"></div></div></div>`,
			d.score, e(d.label), e(d.weight), d.score)
	}
	w(`</div>`)

	// Environment summary
	wf(`<h2>Environment</h2>
<table><tbody>
<tr><td style="width:160px">Provider</td><td>%s</td><td style="width:160px">K8s Version</td><td>%s</td></tr>
<tr><td>Nodes</td><td>%d</td><td>Namespaces</td><td>%d</td></tr>
<tr><td>Backup Tool</td><td class="%s">%s</td><td>Recovery Target</td><td>%s</td></tr>
<tr><td>Helm Releases</td><td>%d</td><td>Certificates</td><td>%d</td></tr>
</tbody></table>`,
		e(platform), e(b.Cluster.Platform.K8sVersion),
		len(b.Inventory.Nodes), len(b.Inventory.Namespaces),
		func() string {
			if backupTool == "none" {
				return "bad"
			}
			return "ok"
		}(), e(backupTool), e(b.Target),
		len(b.Inventory.HelmReleases), len(b.Inventory.Certificates))

	// Top findings — CRITICAL + HIGH only, max 10
	var topFindings []model.Finding
	for _, f := range b.Inventory.Findings {
		if f.Severity == "CRITICAL" || f.Severity == "HIGH" {
			topFindings = append(topFindings, f)
		}
	}
	// Sort CRITICAL before HIGH
	sort.Slice(topFindings, func(i, j int) bool {
		if topFindings[i].Severity == topFindings[j].Severity {
			return topFindings[i].ResourceID < topFindings[j].ResourceID
		}
		return topFindings[i].Severity == "CRITICAL"
	})
	if len(topFindings) > 10 {
		topFindings = topFindings[:10]
	}

	crit, high, med, low := 0, 0, 0, 0
	for _, f := range b.Inventory.Findings {
		switch f.Severity {
		case "CRITICAL":
			crit++
		case "HIGH":
			high++
		case "MEDIUM":
			med++
		default:
			low++
		}
	}

	wf(`<h2>Findings Summary</h2>`)
	wf(`<p style="margin-bottom:8px">
<span class="chip" style="border-color:#c0392b;color:#c0392b">CRITICAL: %d</span>
<span class="chip" style="border-color:#d35400;color:#d35400">HIGH: %d</span>
<span class="chip" style="border-color:#7d6608;color:#7d6608">MEDIUM: %d</span>
<span class="chip">LOW/INFO: %d</span>
</p>`, crit, high, med, low)

	if len(topFindings) > 0 {
		w(`<table><thead><tr><th>Severity</th><th>Resource</th><th>Issue</th><th>Recommendation</th></tr></thead><tbody>`)
		for _, f := range topFindings {
			wf(`<tr><td class="c-%s">%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(f.Severity), e(f.Severity), e(f.ResourceID), e(f.Message), e(f.Recommendation))
		}
		w(`</tbody></table>`)
		if crit+high > 10 {
			wf(`<p style="color:#555;font-size:.82em;margin-top:4px">Showing top 10 of %d critical/high findings. See full report for complete list.</p>`, crit+high)
		}
	} else {
		w(`<p class="ok">No critical or high severity findings.</p>`)
	}

	// Remediation top priorities
	var p1 []model.RemediationStep
	for _, s := range b.Inventory.RemediationSteps {
		if s.Priority == 1 {
			p1 = append(p1, s)
		}
	}
	if len(p1) > 0 {
		if len(p1) > 5 {
			p1 = p1[:5]
		}
		w(`<h2>Priority Actions</h2>`)
		w(`<table><thead><tr><th>Category</th><th>Action</th></tr></thead><tbody>`)
		for _, s := range p1 {
			wf(`<tr><td style="width:120px">%s</td><td>%s</td></tr>`, e(s.Category), e(s.Title))
		}
		w(`</tbody></table>`)
	}

	// Footer
	wf(`<div class="footer">Generated by k8s-recovery-visualizer %s &nbsp;|&nbsp; Scan ID: %s</div>`,
		e(b.Metadata.ToolVersion), e(b.Scan.ScanID))

	w(`</body></html>`)
}
