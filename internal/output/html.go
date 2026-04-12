package output

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"sort"

	"k8s-recovery-visualizer/internal/model"
)

// WriteHTML writes the compact legacy-compatible HTML report.
// It is preserved for downstream consumers that still depend on the legacy
// surface, but it now shares the centralized report theme tokens from design.go.
func WriteHTML(path string, b *model.Bundle) error {
	var buf bytes.Buffer
	buildLegacyHTML(&buf, b)
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func buildLegacyHTML(buf *bytes.Buffer, b *model.Bundle) {
	w := func(s string) { buf.WriteString(s) }
	wf := func(f string, a ...any) { buf.WriteString(fmt.Sprintf(f, a...)) }
	e := html.EscapeString

	matColor := maturityAccent(b.Score.Maturity)
	overallTone := scoreAccent(b.Score.Overall.Final)

	platform := b.Cluster.Platform.Provider
	if platform == "" {
		platform = "unknown"
	}
	clusterName := b.Metadata.ClusterName
	if clusterName == "" {
		clusterName = "unknown"
	}
	backupTool := b.Inventory.Backup.PrimaryTool
	if backupTool == "" {
		backupTool = "none"
	}

	writeMetaChip := func(label, value string) {
		if value == "" {
			return
		}
		wf(`<span class="meta-chip">%s <strong>%s</strong></span>`, e(label), e(value))
	}

	w(reportDocumentStart("DR Assessment Report", "legacy-page", legacyPageCSS()))
	wf(`<header class="hero">
<div class="hero-copy">
<span class="eyebrow">Legacy surface</span>
<h1>DR Assessment Report</h1>
<p>Compact offline HTML output for consumers that still rely on the legacy writer, restyled to match the modern report family.</p>
<div class="hero-meta-grid">`)
	writeMetaChip("Cluster", clusterName)
	writeMetaChip("Platform", platform)
	writeMetaChip("Generated", b.Metadata.GeneratedAt)
	writeMetaChip("Schema", b.SchemaVersion)
	w(`</div></div>`)
	wf(`<div class="hero-panel">
<div class="hero-pill">Primary backup tool: %s</div>
<div class="hero-score">
<div>
<div class="hero-score-label">Overall DR Score</div>
<div class="hero-score-value" style="color:%s">%d</div>
</div>
<div class="badge" style="color:%s;border-color:%s">%s</div>
</div>
<div class="badge-row">
<span class="badge badge-subtle">Coverage: %s</span>
<span class="badge badge-subtle">Findings: %d</span>
</div>
<div class="hero-stat-grid">
<div class="hero-stat"><span>Storage</span><strong>%d</strong></div>
<div class="hero-stat"><span>Workload</span><strong>%d</strong></div>
<div class="hero-stat"><span>Config</span><strong>%d</strong></div>
<div class="hero-stat"><span>Backup</span><strong>%d</strong></div>
</div>
</div></header>`,
		e(backupTool), overallTone, b.Score.Overall.Final,
		matColor, matColor, e(b.Score.Maturity),
		e(backupCoverageStatusText(b.Inventory.Backup)), len(b.Inventory.Findings),
		b.Score.Storage.Final, b.Score.Workload.Final, b.Score.Config.Final, b.Score.Backup.Final)

	w(`<div class="grid">`)
	for _, d := range []struct {
		label string
		score int
		max   int
	}{
		{"Storage", b.Score.Storage.Final, b.Score.Storage.Max},
		{"Workload", b.Score.Workload.Final, b.Score.Workload.Max},
		{"Config", b.Score.Config.Final, b.Score.Config.Max},
		{"Backup / Recovery", b.Score.Backup.Final, b.Score.Backup.Max},
	} {
		fill := d.score
		if d.max > 0 && d.max != 100 {
			fill = d.score * 100 / d.max
		}
		wf(`<div class="sbox"><div class="v">%d</div><div class="l">%s</div><div class="bar"><div class="fill" style="width:%d%%"></div></div></div>`,
			d.score, e(d.label), fill)
	}
	w(`</div>`)

	wf(`<div class="summary-grid"><div class="stack">
<section class="card"><span class="section-tag">Overview</span><h2 style="margin-top:0.4rem">Assessment summary</h2>
<div class="kv-grid">
<div class="kv-card">
<h3>Cluster</h3>
<div class="kv-list">
<div class="kv-row"><span class="kv-key">Cluster</span><span class="kv-value">%s</span></div>
<div class="kv-row"><span class="kv-key">Platform</span><span class="kv-value">%s</span></div>
<div class="kv-row"><span class="kv-key">Nodes</span><span class="kv-value">%d</span></div>
<div class="kv-row"><span class="kv-key">Namespaces</span><span class="kv-value">%d</span></div>
</div>
</div>
<div class="kv-card">
<h3>Recovery posture</h3>
<div class="kv-list">
<div class="kv-row"><span class="kv-key">Backup Tool</span><span class="kv-value">%s</span></div>
<div class="kv-row"><span class="kv-key">Coverage</span><span class="kv-value">%s</span></div>
<div class="kv-row"><span class="kv-key">Target</span><span class="kv-value">%s</span></div>
<div class="kv-row"><span class="kv-key">Schema</span><span class="kv-value">%s</span></div>
</div>
</div>
</div>
</section></div><div class="stack">`,
		e(clusterName), e(platform), len(b.Inventory.Nodes), len(b.Inventory.Namespaces),
		e(backupTool), e(backupCoverageStatusText(b.Inventory.Backup)), e(b.Target), e(b.SchemaVersion))

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
	wf(`<section class="card"><span class="section-tag">Findings</span><h2 style="margin-top:0.4rem">Severity distribution</h2>
<div class="badge-row" style="margin-top:0.9rem">
<span class="chip" style="border-color:var(--danger);color:var(--danger)">CRITICAL: %d</span>
<span class="chip" style="border-color:var(--warning-high);color:var(--warning-high)">HIGH: %d</span>
<span class="chip" style="border-color:var(--warning-medium);color:var(--warning-medium)">MEDIUM: %d</span>
<span class="chip">LOW/INFO: %d</span>
</div>
<p class="subtle" style="margin-top:0.9rem">This compact report preserves the core score, maturity, and findings surfaces without relying on any external assets.</p>
</section></div></div>`, crit, high, med, low)

	w(`<section class="card"><span class="section-tag">Details</span><h2 style="margin-top:0.4rem">Findings</h2>`)
	if len(b.Inventory.Findings) == 0 {
		w(`<p class="ok">No critical DR findings detected.</p>`)
	} else {
		sorted := make([]model.Finding, len(b.Inventory.Findings))
		copy(sorted, b.Inventory.Findings)
		sevOrder := map[string]int{"CRITICAL": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3, "INFO": 4}
		sort.Slice(sorted, func(i, j int) bool {
			oi := sevOrder[sorted[i].Severity]
			oj := sevOrder[sorted[j].Severity]
			if oi != oj {
				return oi < oj
			}
			return sorted[i].ResourceID < sorted[j].ResourceID
		})
		w(`<table><thead><tr><th>Severity</th><th>Resource</th><th>Issue</th><th>Recommendation</th></tr></thead><tbody>`)
		for _, f := range sorted {
			wf(`<tr><td class="c-%s">%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(f.Severity), e(f.Severity), e(f.ResourceID), e(f.Message), e(f.Recommendation))
		}
		w(`</tbody></table>`)
	}
	w(`</section>`)

	wf(`<div class="footer">Generated by k8s-recovery-visualizer %s &nbsp;|&nbsp; Scan ID: %s</div>`,
		e(b.Metadata.ToolVersion), e(b.Scan.ScanID))
	w(`</main>`)
	w(reportDocumentEnd())
}
