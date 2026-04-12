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

	matColor := maturityAccent(b.Score.Maturity)
	overallTone := scoreAccent(b.Score.Overall.Final)

	platform := b.Cluster.Platform.Provider
	if platform == "" {
		platform = "unknown"
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

	w(reportDocumentStart("DR Executive Summary", "summary-page", summaryPageCSS()))
	w(`<div class="page-toolbar"><button type="button" class="print-btn" onclick="window.print()">Print / Save as PDF</button></div>`)
	wf(`<header class="hero">
<div class="hero-copy">
<span class="eyebrow">Executive brief</span>
<h1>Kubernetes DR Readiness - Executive Summary</h1>
<p>High-level readiness, backup trust, and priority remediation guidance in a single offline-friendly assessment summary.</p>
<div class="hero-meta-grid">`)
	writeMetaChip("Customer", b.Metadata.CustomerID)
	writeMetaChip("Cluster", b.Metadata.ClusterName)
	writeMetaChip("Site", b.Metadata.Site)
	writeMetaChip("Env", b.Metadata.Environment)
	writeMetaChip("Scan", b.Metadata.GeneratedAt)
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
<span class="badge badge-subtle">Platform: %s</span>
<span class="badge badge-subtle">Coverage: %s</span>
<span class="badge badge-subtle">Assurance: %s</span>
</div>
<div class="hero-stat-grid">
<div class="hero-stat"><span>Nodes</span><strong>%d</strong></div>
<div class="hero-stat"><span>Namespaces</span><strong>%d</strong></div>
<div class="hero-stat"><span>Helm</span><strong>%d</strong></div>
<div class="hero-stat"><span>Certificates</span><strong>%d</strong></div>
</div>
</div></header>`,
		e(backupTool), overallTone, b.Score.Overall.Final,
		matColor, matColor, e(b.Score.Maturity),
		e(platform), e(backupCoverageStatusText(b.Inventory.Backup)), e(backupAssuranceConclusionText(b.Inventory.Backup.Assurance)),
		len(b.Inventory.Nodes), len(b.Inventory.Namespaces), len(b.Inventory.HelmReleases), len(b.Inventory.Certificates))

	w(`<div class="grid">`)
	for _, d := range []struct {
		label, weight string
		score         int
	}{
		{"Storage", domainWeightLabel("storage"), b.Score.Storage.Final},
		{"Workload", domainWeightLabel("workload"), b.Score.Workload.Final},
		{"Config", domainWeightLabel("config"), b.Score.Config.Final},
		{"Backup / Recovery", domainWeightLabel("backup"), b.Score.Backup.Final},
	} {
		wf(`<div class="sbox"><div class="v">%d</div><div class="l">%s <span style="color:#c4b5fd">%s</span></div><div class="bar"><div class="fill" style="width:%d%%"></div></div></div>`,
			d.score, e(d.label), e(d.weight), d.score)
	}
	w(`</div>`)

	w(`<div class="summary-grid"><div class="stack">`)
	toolChipClass := "f"
	if backupTool != "none" {
		toolChipClass = "p"
	}
	wf(`<section class="card"><span class="section-tag">Environment</span><h2 style="margin-top:0.4rem">Platform and trust posture</h2>
<div class="kv-grid">
<div class="kv-card">
<h3>Cluster context</h3>
<div class="kv-list">
<div class="kv-row"><span class="kv-key">Provider</span><span class="kv-value">%s</span></div>
<div class="kv-row"><span class="kv-key">K8s Version</span><span class="kv-value">%s</span></div>
<div class="kv-row"><span class="kv-key">Nodes</span><span class="kv-value">%d</span></div>
<div class="kv-row"><span class="kv-key">Namespaces</span><span class="kv-value">%d</span></div>
<div class="kv-row"><span class="kv-key">Recovery Target</span><span class="kv-value">%s</span></div>
</div>
</div>
<div class="kv-card">
<h3>Backup trust</h3>
<div class="kv-list">
<div class="kv-row"><span class="kv-key">Backup Tool</span><span class="kv-value"><span class="chip %s">%s</span></span></div>
<div class="kv-row"><span class="kv-key">Backup Coverage</span><span class="kv-value">%s</span></div>
<div class="kv-row"><span class="kv-key">Coverage Detail</span><span class="kv-value">%s</span></div>
<div class="kv-row"><span class="kv-key">Backup Assurance</span><span class="kv-value">%s</span></div>
<div class="kv-row"><span class="kv-key">Assurance Summary</span><span class="kv-value">%s</span></div>
<div class="kv-row"><span class="kv-key">Offsite Detail</span><span class="kv-value">%s</span></div>
</div>
</div>
<div class="kv-card">
<h3>Inventory signals</h3>
<div class="kv-list">
<div class="kv-row"><span class="kv-key">Helm Releases</span><span class="kv-value">%d</span></div>
<div class="kv-row"><span class="kv-key">Certificates</span><span class="kv-value">%d</span></div>
<div class="kv-row"><span class="kv-key">Schema Version</span><span class="kv-value">%s</span></div>
</div>
</div>
</div></section>`,
		e(platform), e(b.Cluster.Platform.K8sVersion),
		len(b.Inventory.Nodes), len(b.Inventory.Namespaces), e(b.Target),
		toolChipClass, e(backupTool),
		e(backupCoverageStatusText(b.Inventory.Backup)), e(backupCoverageReasonText(b.Inventory.Backup)),
		e(backupAssuranceConclusionText(b.Inventory.Backup.Assurance)), e(func() string {
			if b.Inventory.Backup.Assurance == nil {
				return "Backup assurance was not assessed."
			}
			return b.Inventory.Backup.Assurance.Summary
		}()),
		e(backupOffsiteDetailText(b.Inventory.Backup)),
		len(b.Inventory.HelmReleases), len(b.Inventory.Certificates), e(b.SchemaVersion))

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

	w(`<section class="card"><span class="section-tag">Top Issues</span><h2 style="margin-top:0.4rem">Critical and high findings</h2>`)

	if len(topFindings) > 0 {
		w(`<table><thead><tr><th>Severity</th><th>Resource</th><th>Issue</th><th>Recommendation</th></tr></thead><tbody>`)
		for _, f := range topFindings {
			wf(`<tr><td class="c-%s">%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(f.Severity), e(f.Severity), e(f.ResourceID), e(f.Message), e(f.Recommendation))
		}
		w(`</tbody></table>`)
		if crit+high > 10 {
			wf(`<p class="subtle" style="font-size:.82em;margin-top:4px">Showing top 10 of %d critical/high findings. See full report for complete list.</p>`, crit+high)
		}
	} else {
		w(`<p class="ok">No critical or high severity findings.</p>`)
	}
	w(`</section></div><div class="stack">`)

	wf(`<section class="card"><span class="section-tag">Findings Summary</span><h2 style="margin-top:0.4rem">Severity distribution</h2>
<div class="badge-row" style="margin-top:0.9rem">
<span class="chip" style="border-color:#fb7185;color:#fb7185">CRITICAL: %d</span>
<span class="chip" style="border-color:#ffb86c;color:#ffb86c">HIGH: %d</span>
<span class="chip" style="border-color:#f59e0b;color:#f59e0b">MEDIUM: %d</span>
<span class="chip">LOW/INFO: %d</span>
</div>
<p class="subtle" style="margin-top:0.9rem">This summary keeps the highest-risk items front and center while preserving the underlying trust and assurance semantics.</p>
</section>`, crit, high, med, low)

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
		w(`<section class="card"><span class="section-tag">Priority Actions</span><h2 style="margin-top:0.4rem">Immediate next steps</h2>`)
		w(`<table><thead><tr><th>Category</th><th>Action</th></tr></thead><tbody>`)
		for _, s := range p1 {
			wf(`<tr><td style="width:120px">%s</td><td>%s</td></tr>`, e(s.Category), e(s.Title))
		}
		w(`</tbody></table></section>`)
	} else {
		w(`<section class="card"><span class="section-tag">Priority Actions</span><h2 style="margin-top:0.4rem">Immediate next steps</h2><p class="subtle">No priority-one remediation items were generated for this scan.</p></section>`)
	}
	w(`</div></div>`)

	// Footer
	wf(`<div class="footer">Generated by k8s-recovery-visualizer %s &nbsp;|&nbsp; Scan ID: %s</div>`,
		e(b.Metadata.ToolVersion), e(b.Scan.ScanID))
	w(`</main>`)
	w(reportDocumentEnd())
}
