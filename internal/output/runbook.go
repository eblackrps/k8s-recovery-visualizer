package output

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"sort"
	"strings"

	"k8s-recovery-visualizer/internal/model"
)

// WriteRunbook writes a customer-facing, print-ready DR runbook to path.
func WriteRunbook(path string, b *model.Bundle) error {
	var buf bytes.Buffer
	buildRunbook(&buf, b)
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func buildRunbook(buf *bytes.Buffer, b *model.Bundle) {
	w := func(s string) { buf.WriteString(s) }
	wf := func(f string, a ...any) { buf.WriteString(fmt.Sprintf(f, a...)) }
	e := html.EscapeString

	matColor := map[string]string{
		"PLATINUM": "#1a56a0", "GOLD": "#b8860b",
		"SILVER": "#555", "BRONZE": "#b5521a",
	}[b.Score.Maturity]
	if matColor == "" {
		matColor = "#555"
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
<title>DR Runbook</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#fff;color:#1a1a1a;font-family:"Segoe UI",Arial,sans-serif;font-size:13px;line-height:1.6;padding:32px 48px;max-width:1100px;margin:0 auto}
h1{font-size:1.6em;font-weight:800;margin-bottom:4px}
h2{font-size:1.05em;font-weight:700;margin:28px 0 8px;padding-bottom:4px;border-bottom:2px solid #1a1a1a}
h3{font-size:.92em;font-weight:600;margin:14px 0 6px;color:#333}
.meta{color:#555;font-size:.82em;margin-bottom:24px}
.cover{border:2px solid #1a1a1a;border-radius:6px;padding:24px 28px;margin-bottom:32px;background:#fafafa}
.cover h1{margin-bottom:6px}
.score-row{display:flex;align-items:center;gap:20px;margin-top:14px}
.score-big{font-size:3em;font-weight:800;line-height:1;color:#1a1a1a}
.badge{display:inline-block;padding:5px 16px;border-radius:14px;font-weight:700;font-size:1.1em;border:2px solid}
.domains{display:grid;grid-template-columns:repeat(4,1fr);gap:8px;margin:14px 0}
.dom{border:1px solid #ccc;border-radius:4px;padding:8px;text-align:center;background:#fff}
.dom .v{font-size:1.7em;font-weight:700}
.dom .l{font-size:.74em;color:#555}
.dom .bar{background:#e0e0e0;border-radius:3px;height:4px;margin-top:5px}
.dom .fill{height:4px;border-radius:3px;background:#555}
table{width:100%;border-collapse:collapse;margin-top:4px;font-size:.84em}
th{background:#f0f0f0;text-align:left;padding:6px 9px;border:1px solid #ccc;font-weight:600}
td{padding:5px 9px;border:1px solid #ddd;vertical-align:top;word-break:break-word}
tr:nth-child(even) td{background:#fafafa}
.c-CRITICAL{color:#b91c1c;font-weight:700}
.c-HIGH{color:#c2410c;font-weight:600}
.c-MEDIUM{color:#854d0e}
.c-LOW,.c-INFO{color:#555}
.ok{color:#166534}.bad{color:#b91c1c}
.chip{display:inline-block;padding:1px 8px;border-radius:10px;font-size:.77em;border:1px solid #ccc;margin:1px}
.step{border:1px solid #ccc;border-radius:4px;margin-bottom:10px;page-break-inside:avoid}
.step-hdr{background:#f5f5f5;padding:8px 12px;font-weight:600;border-bottom:1px solid #ccc;display:flex;align-items:center;gap:8px}
.step-body{padding:10px 12px}
pre{background:#f5f5f5;border:1px solid #ddd;border-radius:4px;padding:9px;font-size:.8em;overflow-x:auto;margin-top:6px;white-space:pre-wrap}
.note{background:#fffbeb;border-left:3px solid #b8860b;padding:6px 10px;margin-top:6px;font-size:.84em;border-radius:0 3px 3px 0}
.section-break{page-break-before:always}
.footer{margin-top:36px;color:#888;font-size:.78em;border-top:1px solid #ddd;padding-top:10px}
.print-btn{display:inline-block;margin-bottom:20px;padding:7px 18px;background:#1a1a2e;color:#fff;border:none;border-radius:4px;cursor:pointer;font-size:.86em}
.toc{background:#f9f9f9;border:1px solid #ddd;border-radius:4px;padding:14px 18px;margin-bottom:28px;display:inline-block;min-width:260px}
.toc h3{margin:0 0 8px}
.toc ol{padding-left:20px;font-size:.88em;line-height:1.9}
.toc a{color:#1a56a0;text-decoration:none}
.toc a:hover{text-decoration:underline}
@media print{
  .print-btn,.toc{display:none}
  body{padding:12px 20px}
  h2{page-break-after:avoid}
  @page{margin:1.5cm}
}
</style></head><body>
`)

	w(`<button class="print-btn" onclick="window.print()">Print / Save as PDF</button>`)

	// ── Cover ────────────────────────────────────────────────────────────────
	w(`<div class="cover">`)
	wf(`<h1>Kubernetes DR Recovery Runbook</h1>`)
	parts := []string{}
	if b.Metadata.CustomerID != "" {
		parts = append(parts, "Customer: "+b.Metadata.CustomerID)
	}
	if b.Metadata.ClusterName != "" {
		parts = append(parts, "Cluster: "+b.Metadata.ClusterName)
	}
	if b.Metadata.Site != "" {
		parts = append(parts, "Site: "+b.Metadata.Site)
	}
	if b.Metadata.Environment != "" {
		parts = append(parts, "Environment: "+b.Metadata.Environment)
	}
	parts = append(parts, "Generated: "+b.Metadata.GeneratedAt)
	parts = append(parts, "Profile: "+func() string {
		if b.Profile == "" {
			return "standard"
		}
		return b.Profile
	}())
	wf(`<div class="meta">`)
	for i, p := range parts {
		if i > 0 {
			w(` &nbsp;|&nbsp; `)
		}
		w(e(p))
	}
	w(`</div>`)

	wf(`<div class="score-row">
<div class="score-big">%d<span style="font-size:.35em;color:#555">/100</span></div>
<div>
<div class="badge" style="color:%s;border-color:%s;font-size:1.2em">%s</div>
<div style="color:#555;font-size:.82em;margin-top:6px">Platform: %s &nbsp; Backup: %s &nbsp; Recovery Target: %s</div>
</div>
</div>`,
		b.Score.Overall.Final,
		matColor, matColor, e(b.Score.Maturity),
		e(platform), e(backupTool), e(b.Target))

	w(`<div class="domains" style="margin-top:14px">`)
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
	w(`</div>`) // cover

	// ── Table of contents ────────────────────────────────────────────────────
	w(`<div class="toc"><h3>Contents</h3><ol>
<li><a href="#s1">Cluster Inventory</a></li>
<li><a href="#s2">Backup &amp; Recovery Status</a></li>
<li><a href="#s3">Restore Simulation</a></li>
<li><a href="#s4">Findings</a></li>
<li><a href="#s5">DR Remediation Playbook</a></li>
<li><a href="#s6">Scan Metadata</a></li>
</ol></div>`)

	// ── Section 1: Cluster Inventory ─────────────────────────────────────────
	w(`<h2 id="s1">1. Cluster Inventory</h2>`)
	wf(`<table><tbody>
<tr><th style="width:180px">Provider</th><td>%s</td><th style="width:180px">K8s Version</th><td>%s</td></tr>
<tr><th>Cluster UID</th><td>%s</td><th>Platform</th><td>%s</td></tr>
<tr><th>Nodes</th><td>%d</td><th>Namespaces</th><td>%d</td></tr>
<tr><th>PVCs</th><td>%d</td><th>PVs</th><td>%d</td></tr>
<tr><th>Deployments</th><td>%d</td><th>StatefulSets</th><td>%d</td></tr>
<tr><th>Helm Releases</th><td>%d</td><th>Certificates</th><td>%d</td></tr>
<tr><th>CRDs</th><td>%d</td><th>Recovery Target</th><td>%s</td></tr>
</tbody></table>`,
		e(platform), e(b.Cluster.Platform.K8sVersion),
		e(b.Cluster.Platform.ClusterUID), e(platform),
		len(b.Inventory.Nodes), len(b.Inventory.Namespaces),
		len(b.Inventory.PVCs), len(b.Inventory.PVs),
		len(b.Inventory.Deployments), len(b.Inventory.StatefulSets),
		len(b.Inventory.HelmReleases), len(b.Inventory.Certificates),
		len(b.Inventory.CRDs), e(b.Target))

	// Node list (condensed)
	if len(b.Inventory.Nodes) > 0 {
		w(`<h3>Nodes</h3><table><thead><tr><th>Name</th><th>Roles</th><th>Ready</th><th>OS</th><th>Kubelet</th></tr></thead><tbody>`)
		for _, n := range b.Inventory.Nodes {
			rdStr := `<span class="bad">✗</span>`
			if n.Ready {
				rdStr = `<span class="ok">✓</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(n.Name), e(strings.Join(n.Roles, ",")), rdStr, e(n.OSImage), e(n.KubeletVersion))
		}
		w(`</tbody></table>`)
	}

	// Storage summary
	if len(b.Inventory.PVCs) > 0 {
		w(`<h3>PVC Summary</h3><table><thead><tr><th>Namespace</th><th>Name</th><th>StorageClass</th><th>Size</th><th>Status</th></tr></thead><tbody>`)
		pvMap := map[string]model.PersistentVolume{}
		for _, pv := range b.Inventory.PVs {
			pvMap[pv.ClaimRef] = pv
		}
		for _, pvc := range b.Inventory.PVCs {
			status := `<span class="ok">bound</span>`
			if _, ok := pvMap[pvc.Namespace+"/"+pvc.Name]; !ok {
				status = `<span class="bad">unbound</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(pvc.Namespace), e(pvc.Name), e(pvc.StorageClass), e(pvc.RequestedSize), status)
		}
		w(`</tbody></table>`)
	}

	// ── Section 2: Backup & Recovery Status ──────────────────────────────────
	w(`<h2 id="s2" class="section-break">2. Backup &amp; Recovery Status</h2>`)
	inv := b.Inventory.Backup
	backupClass := "bad"
	if inv.PrimaryTool != "none" && inv.PrimaryTool != "" {
		backupClass = "ok"
	}
	offsiteStr := `<span class="bad">No</span>`
	if inv.HasOffsite {
		offsiteStr = `<span class="ok">Yes</span>`
	}
	wf(`<table><tbody>
<tr><th style="width:200px">Primary Backup Tool</th><td class="%s">%s</td></tr>
<tr><th>Offsite / Export Configured</th><td>%s</td></tr>
<tr><th>Policies / Schedules Found</th><td>%d</td></tr>
<tr><th>Covered Namespaces</th><td>%s</td></tr>
<tr><th>Uncovered Stateful Namespaces</th><td class="%s">%s</td></tr>
</tbody></table>`,
		backupClass, e(backupTool),
		offsiteStr,
		len(inv.Policies),
		func() string {
			if len(inv.CoveredNamespaces) == 0 {
				return "none"
			}
			return strings.Join(inv.CoveredNamespaces, ", ")
		}(),
		func() string {
			if len(inv.UncoveredStatefulNS) > 0 {
				return "bad"
			}
			return "ok"
		}(),
		func() string {
			if len(inv.UncoveredStatefulNS) == 0 {
				return "none"
			}
			return strings.Join(inv.UncoveredStatefulNS, ", ")
		}())

	if len(inv.Policies) > 0 {
		w(`<h3>Backup Policies</h3><table><thead><tr><th>Tool</th><th>Name</th><th>Namespaces</th><th>Schedule</th><th>RPO (h)</th><th>Offsite</th><th>Retention</th></tr></thead><tbody>`)
		for _, p := range inv.Policies {
			nsCell := "all"
			if len(p.IncludedNS) > 0 {
				nsCell = strings.Join(p.IncludedNS, ", ")
			}
			rpoCell := "unknown"
			if p.RPOHours >= 0 {
				rpoCell = fmt.Sprintf("%d", p.RPOHours)
			}
			offsiteCell := "no"
			if p.HasOffsite {
				offsiteCell = "yes"
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(p.Tool), e(p.Name), e(nsCell), e(p.Schedule), rpoCell, offsiteCell, e(p.RetentionTTL))
		}
		w(`</tbody></table>`)
	}

	// ── Section 3: Restore Simulation ────────────────────────────────────────
	w(`<h2 id="s3">3. Restore Simulation</h2>`)
	if sim := inv.RestoreSim; sim == nil || len(sim.Namespaces) == 0 {
		w(`<p style="color:#555">No stateful namespaces found — nothing to simulate.</p>`)
	} else {
		covPct := 0.0
		if sim.TotalPVCsGB > 0 {
			covPct = sim.CoveredPVCsGB / sim.TotalPVCsGB * 100
		}
		wf(`<p style="margin-bottom:8px">Total PVC data: <strong>%.1f GB</strong> &nbsp; Coverage by volume: <strong>%.0f%%</strong> &nbsp; Uncovered namespaces: <strong class="%s">%d</strong></p>`,
			sim.TotalPVCsGB, covPct,
			func() string {
				if len(sim.UncoveredNS) > 0 {
					return "bad"
				}
				return "ok"
			}(), len(sim.UncoveredNS))
		w(`<table><thead><tr><th>Namespace</th><th>Coverage</th><th>RPO (h)</th><th>PVC Data (GB)</th><th>Blockers</th><th>Warnings</th></tr></thead><tbody>`)
		for _, ns := range sim.Namespaces {
			covCell := `<span class="bad">none</span>`
			if ns.HasCoverage {
				covCell = `<span class="ok">covered</span>`
			}
			rpoCell := "unknown"
			if ns.RPOHours >= 0 {
				rpoCell = fmt.Sprintf("%d", ns.RPOHours)
			}
			blockersCell := "—"
			if len(ns.Blockers) > 0 {
				blockersCell = `<span class="c-CRITICAL">` + e(strings.Join(ns.Blockers, "; ")) + `</span>`
			}
			warningsCell := "—"
			if len(ns.Warnings) > 0 {
				warningsCell = `<span class="c-MEDIUM">` + e(strings.Join(ns.Warnings, "; ")) + `</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%.1f</td><td>%s</td><td>%s</td></tr>`,
				e(ns.Namespace), covCell, rpoCell, ns.PVCSizeGB, blockersCell, warningsCell)
		}
		w(`</tbody></table>`)
	}

	// ── Section 4: Findings ───────────────────────────────────────────────────
	w(`<h2 id="s4" class="section-break">4. Findings</h2>`)
	if len(b.Inventory.Findings) == 0 {
		w(`<p class="ok">No findings recorded.</p>`)
	} else {
		// Sort: CRITICAL → HIGH → MEDIUM → LOW → INFO
		sevOrder := map[string]int{"CRITICAL": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3, "INFO": 4}
		sorted := make([]model.Finding, len(b.Inventory.Findings))
		copy(sorted, b.Inventory.Findings)
		sort.Slice(sorted, func(i, j int) bool {
			oi := sevOrder[sorted[i].Severity]
			oj := sevOrder[sorted[j].Severity]
			if oi != oj {
				return oi < oj
			}
			return sorted[i].ResourceID < sorted[j].ResourceID
		})
		crit, high, med, low := 0, 0, 0, 0
		for _, f := range sorted {
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
		wf(`<p style="margin-bottom:8px">
<span class="chip" style="border-color:#b91c1c;color:#b91c1c">CRITICAL: %d</span>
<span class="chip" style="border-color:#c2410c;color:#c2410c">HIGH: %d</span>
<span class="chip" style="border-color:#854d0e;color:#854d0e">MEDIUM: %d</span>
<span class="chip">LOW/INFO: %d</span>
</p>`, crit, high, med, low)
		w(`<table><thead><tr><th>Severity</th><th>Resource</th><th>Finding</th><th>Recommendation</th></tr></thead><tbody>`)
		for _, f := range sorted {
			wf(`<tr><td class="c-%s">%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(f.Severity), e(f.Severity), e(f.ResourceID), e(f.Message), e(f.Recommendation))
		}
		w(`</tbody></table>`)
	}

	// ── Section 5: Remediation Playbook ──────────────────────────────────────
	w(`<h2 id="s5" class="section-break">5. DR Remediation Playbook</h2>`)
	if len(b.Inventory.RemediationSteps) == 0 {
		w(`<p style="color:#555">No remediation steps generated. Run against a live cluster to produce findings.</p>`)
	} else {
		priLabel := map[int]string{
			1: "Priority 1 — Must Fix Before DR",
			2: "Priority 2 — Recommended",
			3: "Priority 3 — Optional",
		}
		priClass := map[int]string{1: "c-CRITICAL", 2: "c-HIGH", 3: "c-LOW"}
		chipBorder := map[int]string{1: "#b91c1c", 2: "#c2410c", 3: "#888"}
		curPri := -1
		for _, step := range b.Inventory.RemediationSteps {
			if step.Priority != curPri {
				curPri = step.Priority
				wf(`<h3 class="%s">%s</h3>`, priClass[step.Priority], e(priLabel[step.Priority]))
			}
			wf(`<div class="step"><div class="step-hdr"><span class="chip" style="border-color:%s;color:%s">%s</span> %s</div>`,
				chipBorder[step.Priority], chipBorder[step.Priority], e(step.Category), e(step.Title))
			wf(`<div class="step-body"><p>%s</p>`, e(step.Detail))
			if step.TargetNotes != "" {
				wf(`<div class="note">%s</div>`, e(step.TargetNotes))
			}
			if len(step.Commands) > 0 {
				wf(`<pre>%s</pre>`, e(strings.Join(step.Commands, "\n")))
			}
			w(`</div></div>`)
		}
	}

	// ── Section 6: Scan Metadata ─────────────────────────────────────────────
	w(`<h2 id="s6">6. Scan Metadata</h2>`)
	wf(`<table><tbody>
<tr><th style="width:200px">Scan ID</th><td>%s</td></tr>
<tr><th>Tool Version</th><td>%s</td></tr>
<tr><th>Scan Started</th><td>%s</td></tr>
<tr><th>Scan Duration</th><td>%d seconds</td></tr>
<tr><th>Scoring Profile</th><td>%s</td></tr>
<tr><th>Schema Version</th><td>%s</td></tr>
</tbody></table>`,
		e(b.Scan.ScanID), e(b.Metadata.ToolVersion),
		e(b.Scan.StartedAt.Format("2006-01-02 15:04:05 UTC")),
		b.Scan.DurationSeconds,
		e(func() string {
			if b.Profile == "" {
				return "standard"
			}
			return b.Profile
		}()),
		e(b.SchemaVersion))

	// Footer
	wf(`<div class="footer">Generated by k8s-recovery-visualizer %s &nbsp;|&nbsp; Scan ID: %s</div>`,
		e(b.Metadata.ToolVersion), e(b.Scan.ScanID))

	w(`</body></html>`)
}
