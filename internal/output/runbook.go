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
	activeProfile := b.Profile
	if activeProfile == "" {
		activeProfile = "standard"
	}
	writeMetaChip := func(label, value string) {
		if value == "" {
			return
		}
		wf(`<span class="meta-chip">%s <strong>%s</strong></span>`, e(label), e(value))
	}

	w(reportDocumentStart("DR Runbook", "runbook-page", runbookPageCSS()))
	w(`<div class="page-toolbar"><button type="button" class="print-btn" onclick="window.print()">Print / Save as PDF</button></div>`)
	wf(`<header class="hero">
<div class="hero-copy">
<span class="eyebrow">Operational runbook</span>
<h1>Kubernetes DR Recovery Runbook</h1>
<p>Customer-facing recovery inventory, backup posture, restore simulation, and remediation guidance for operational planning and executive review.</p>
<div class="hero-meta-grid">`)
	writeMetaChip("Customer", b.Metadata.CustomerID)
	writeMetaChip("Cluster", b.Metadata.ClusterName)
	writeMetaChip("Site", b.Metadata.Site)
	writeMetaChip("Environment", b.Metadata.Environment)
	writeMetaChip("Generated", b.Metadata.GeneratedAt)
	writeMetaChip("Profile", activeProfile)
	w(`</div></div>`)
	wf(`<div class="hero-panel">
<div class="hero-pill">Recovery target: %s</div>
<div class="hero-score">
<div>
<div class="hero-score-label">Overall DR Score</div>
<div class="hero-score-value" style="color:%s">%d</div>
</div>
<div class="badge" style="color:%s;border-color:%s">%s</div>
</div>
<div class="badge-row">
<span class="badge badge-subtle">Platform: %s</span>
<span class="badge badge-subtle">Backup: %s</span>
<span class="badge badge-subtle">Coverage: %s</span>
</div>
<div class="hero-stat-grid">
<div class="hero-stat"><span>Nodes</span><strong>%d</strong></div>
<div class="hero-stat"><span>Namespaces</span><strong>%d</strong></div>
<div class="hero-stat"><span>PVCs</span><strong>%d</strong></div>
<div class="hero-stat"><span>Findings</span><strong>%d</strong></div>
</div>
</div></header>`,
		e(b.Target), overallTone, b.Score.Overall.Final,
		matColor, matColor, e(b.Score.Maturity),
		e(platform), e(backupTool), e(backupCoverageStatusText(b.Inventory.Backup)),
		len(b.Inventory.Nodes), len(b.Inventory.Namespaces), len(b.Inventory.PVCs), len(b.Inventory.Findings))

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
		wf(`<div class="sbox"><div class="v">%d</div><div class="l">%s <span style="color:var(--accent)">%s</span></div><div class="bar"><div class="fill" style="width:%d%%"></div></div></div>`,
			d.score, e(d.label), e(d.weight), d.score)
	}
	w(`</div>`)

	w(`<div class="toc"><h3>Contents</h3><ol>
<li><a href="#s1">Cluster Inventory</a></li>
<li><a href="#s2">Backup &amp; Recovery Status</a></li>
<li><a href="#s3">Restore Simulation</a></li>
<li><a href="#s4">Findings</a></li>
<li><a href="#s5">DR Remediation Playbook</a></li>
<li><a href="#s6">Scan Metadata</a></li>
</ol></div>`)

	w(`<section class="card"><span class="section-tag">Section 1</span><h2 id="s1" style="margin-top:0.4rem">1. Cluster Inventory</h2>`)
	wf(`<table><tbody>
<tr><th style="width:180px">Provider</th><td>%s</td><th style="width:180px">K8s Version</th><td>%s</td></tr>
<tr><th>Cluster UID</th><td>%s</td><th>Platform</th><td>%s</td></tr>
<tr><th>Nodes</th><td>%d</td><th>Namespaces</th><td>%d</td></tr>
<tr><th>PVCs</th><td>%d</td><th>PVs</th><td>%d</td></tr>
<tr><th>Deployments</th><td>%d</td><th>StatefulSets</th><td>%d</td></tr>
<tr><th>Helm Releases</th><td>%d</td><th>Certificates</th><td>%d</td></tr>
<tr><th>CRDs</th><td>%d</td><th>Recovery Target</th><td>%s</td></tr>
<tr><th>Backup Coverage Status</th><td colspan="3">%s</td></tr>
</tbody></table>`,
		e(platform), e(b.Cluster.Platform.K8sVersion),
		e(b.Cluster.Platform.ClusterUID), e(platform),
		len(b.Inventory.Nodes), len(b.Inventory.Namespaces),
		len(b.Inventory.PVCs), len(b.Inventory.PVs),
		len(b.Inventory.Deployments), len(b.Inventory.StatefulSets),
		len(b.Inventory.HelmReleases), len(b.Inventory.Certificates),
		len(b.Inventory.CRDs), e(b.Target), e(backupCoverageStatusText(b.Inventory.Backup)))

	if len(b.Inventory.Nodes) > 0 {
		w(`<h3 style="margin-top:1rem">Nodes</h3><table><thead><tr><th>Name</th><th>Roles</th><th>Ready</th><th>OS</th><th>Kubelet</th></tr></thead><tbody>`)
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

	if len(b.Inventory.PVCs) > 0 {
		w(`<h3 style="margin-top:1rem">PVC Summary</h3><table><thead><tr><th>Namespace</th><th>Name</th><th>StorageClass</th><th>Size</th><th>Status</th></tr></thead><tbody>`)
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
	w(`</section>`)

	w(`<section class="card section-break"><span class="section-tag">Section 2</span><h2 id="s2" style="margin-top:0.4rem">2. Backup &amp; Recovery Status</h2>`)
	inv := b.Inventory.Backup
	backupClass := "bad"
	if inv.PrimaryTool != "none" && inv.PrimaryTool != "" {
		backupClass = "ok"
	}
	offsiteStr := `<span class="bad">No</span>`
	if !inv.CoverageVerified && inv.PrimaryTool != "none" && inv.PrimaryTool != "" {
		offsiteStr = `<span style="color:var(--warning-medium)">Unknown</span>`
	} else if inv.HasOffsite {
		offsiteStr = `<span class="ok">Yes</span>`
	}
	wf(`<table><tbody>
<tr><th style="width:200px">Primary Backup Tool</th><td class="%s">%s</td></tr>
<tr><th>Policy Coverage Verified</th><td>%s</td></tr>
<tr><th>Coverage Status</th><td>%s</td></tr>
<tr><th>Coverage Detail</th><td>%s</td></tr>
<tr><th>Backup Assurance</th><td>%s</td></tr>
<tr><th>Assurance Summary</th><td>%s</td></tr>
<tr><th>Offsite / Export Configured</th><td>%s</td></tr>
<tr><th>Offsite Detail</th><td>%s</td></tr>
<tr><th>Policies / Schedules Found</th><td>%d</td></tr>
<tr><th>Covered Namespaces</th><td>%s</td></tr>
<tr><th>Uncovered Stateful Namespaces</th><td class="%s">%s</td></tr>
</tbody></table>`,
		backupClass, e(backupTool),
		func() string {
			if inv.PrimaryTool == "none" || inv.PrimaryTool == "" {
				return "n/a"
			}
			if inv.CoverageVerified {
				return `<span class="ok">Yes</span>`
			}
			return `<span style="color:var(--warning-medium)">No</span>`
		}(),
		e(backupCoverageStatusText(inv)),
		e(backupCoverageReasonText(inv)),
		e(backupAssuranceConclusionText(inv.Assurance)),
		e(func() string {
			if inv.Assurance == nil {
				return "Backup assurance was not assessed."
			}
			return inv.Assurance.Summary
		}()),
		offsiteStr,
		e(backupOffsiteDetailText(inv)),
		len(inv.Policies),
		func() string {
			if !inv.CoverageVerified && inv.PrimaryTool != "none" && inv.PrimaryTool != "" {
				return "unknown"
			}
			if len(inv.CoveredNamespaces) == 0 {
				return "none"
			}
			return strings.Join(inv.CoveredNamespaces, ", ")
		}(),
		func() string {
			if !inv.CoverageVerified && inv.PrimaryTool != "none" && inv.PrimaryTool != "" {
				return ""
			}
			if len(inv.UncoveredStatefulNS) > 0 {
				return "bad"
			}
			return "ok"
		}(),
		func() string {
			if !inv.CoverageVerified && inv.PrimaryTool != "none" && inv.PrimaryTool != "" {
				return "unknown"
			}
			if len(inv.UncoveredStatefulNS) == 0 {
				return "none"
			}
			return strings.Join(inv.UncoveredStatefulNS, ", ")
		}())

	if inv.PrimaryTool != "none" && inv.PrimaryTool != "" && !inv.CoverageVerified {
		wf(`<p class="subtle" style="margin-top:8px">This backup product was detected, but policy coverage could not be verified (%s). %s</p>`,
			e(backupCoverageStatusText(inv)), e(backupCoverageReasonText(inv)))
	} else if len(inv.Policies) > 0 {
		w(`<h3 style="margin-top:1rem">Backup Policies</h3><table><thead><tr><th>Tool</th><th>Name</th><th>Namespaces</th><th>Schedule</th><th>RPO (h)</th><th>Last Success</th><th>Evidence</th><th>Offsite</th><th>Retention</th></tr></thead><tbody>`)
		for _, p := range inv.Policies {
			nsCell := "all"
			if len(p.IncludedNS) > 0 {
				nsCell = strings.Join(p.IncludedNS, ", ")
			}
			rpoCell := "unknown"
			if p.RPOHours >= 0 {
				rpoCell = fmt.Sprintf("%d", p.RPOHours)
			}
			lastSuccessCell := "unknown"
			if p.LastSuccessAt != "" {
				lastSuccessCell = p.LastSuccessAt
				if p.LastSuccessAgeHours > 0 {
					lastSuccessCell = fmt.Sprintf("%s (%dh ago)", p.LastSuccessAt, p.LastSuccessAgeHours)
				}
			}
			evidenceCell := string(p.Confidence)
			if evidenceCell == "" {
				evidenceCell = string(model.EvidenceConfidenceUnknown)
			}
			offsiteCell := "no"
			if p.HasOffsite {
				offsiteCell = "yes"
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(p.Tool), e(p.Name), e(nsCell), e(p.Schedule), rpoCell, e(lastSuccessCell), e(evidenceCell), offsiteCell, e(p.RetentionTTL))
		}
		w(`</tbody></table>`)
	}
	if inv.Assurance != nil && len(inv.Assurance.Signals) > 0 {
		w(`<h3 style="margin-top:1rem">Assurance Signals</h3><table><thead><tr><th>Signal</th><th>Status</th><th>Confidence</th><th>Summary</th><th>Detail</th></tr></thead><tbody>`)
		for _, signal := range inv.Assurance.Signals {
			detail := "—"
			if signal.Detail != "" {
				detail = signal.Detail
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(signal.ID), e(signal.Status), e(string(signal.Confidence)), e(signal.Summary), e(detail))
		}
		w(`</tbody></table>`)
	}
	w(`</section>`)

	w(`<section class="card"><span class="section-tag">Section 3</span><h2 id="s3" style="margin-top:0.4rem">3. Restore Simulation</h2>`)
	if sim := inv.RestoreSim; sim == nil || len(sim.Namespaces) == 0 {
		w(`<p class="subtle">No stateful namespaces found — nothing to simulate.</p>`)
	} else {
		unknownCount := 0
		for _, ns := range sim.Namespaces {
			if !ns.CoverageKnown {
				unknownCount++
			}
		}
		covPct := 0.0
		if sim.TotalPVCsGB > 0 {
			covPct = sim.CoveredPVCsGB / sim.TotalPVCsGB * 100
		}
		coverageVolumeText := fmt.Sprintf("%.0f%%", covPct)
		coverageCountLabel := "Uncovered namespaces"
		coverageCount := len(sim.UncoveredNS)
		coverageCountClass := "ok"
		if coverageCount > 0 {
			coverageCountClass = "bad"
		}
		if unknownCount > 0 {
			coverageVolumeText = "unknown"
			coverageCountLabel = "Unverified namespaces"
			coverageCount = unknownCount
			coverageCountClass = ""
		}
		wf(`<p style="margin-bottom:8px">Total PVC data: <strong>%.1f GB</strong> &nbsp; Coverage by volume: <strong>%s</strong> &nbsp; %s: <strong class="%s">%d</strong></p>`,
			sim.TotalPVCsGB, coverageVolumeText, coverageCountLabel, coverageCountClass, coverageCount)
		wf(`<p style="margin-bottom:8px">Ready namespaces: <strong class="ok">%d</strong> &nbsp; Blocked namespaces: <strong class="bad">%d</strong> &nbsp; Warning namespaces: <strong>%d</strong> &nbsp; Unknown namespaces: <strong>%d</strong> &nbsp; Estimated data at risk: <strong>%.1f GB</strong></p>`,
			sim.ReadyNamespaces, sim.BlockedNamespaces, sim.WarningNamespaces, sim.UnknownNamespaces, sim.EstimatedDataAtRiskGB)
		if len(sim.BlockingReasons) > 0 {
			wf(`<p class="subtle" style="margin-bottom:8px">Top blocking reasons: %s</p>`, e(strings.Join(sim.BlockingReasons, " • ")))
		}
		w(`<table><thead><tr><th>Namespace</th><th>Readiness</th><th>Coverage</th><th>RPO (h)</th><th>PVC Data (GB)</th><th>Blockers</th><th>Warnings</th></tr></thead><tbody>`)
		for _, ns := range sim.Namespaces {
			covCell := `<span class="bad">none</span>`
			if !ns.CoverageKnown {
				covCell = `<span style="color:var(--warning-medium)">unverified</span>`
			} else if ns.HasCoverage {
				covCell = `<span class="ok">covered</span>`
			}
			readinessCell := `<span style="color:var(--muted)">unknown</span>`
			switch ns.Readiness {
			case "ready":
				readinessCell = `<span class="ok">ready</span>`
			case "warning", "uncovered":
				readinessCell = `<span style="color:var(--warning-medium)">` + e(ns.Readiness) + `</span>`
			case "blocked":
				readinessCell = `<span class="bad">blocked</span>`
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
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%.1f</td><td>%s</td><td>%s</td></tr>`,
				e(ns.Namespace), readinessCell, covCell, rpoCell, ns.PVCSizeGB, blockersCell, warningsCell)
		}
		w(`</tbody></table>`)
	}
	if len(inv.DrillPlan) > 0 {
		w(`<h3 style="margin-top:1rem">Recommended Restore Drill Plan</h3><table><thead><tr><th>Phase</th><th>Step</th><th>Owner</th><th>Detail</th><th>Validation</th></tr></thead><tbody>`)
		for _, step := range inv.DrillPlan {
			owner := step.OwnerHint
			if owner == "" {
				owner = "—"
			}
			validation := "—"
			if len(step.Validation) > 0 {
				validation = strings.Join(step.Validation, "; ")
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(step.Phase), e(step.Title), e(owner), e(step.Detail), e(validation))
		}
		w(`</tbody></table>`)
	}
	w(`</section>`)

	w(`<section class="card section-break"><span class="section-tag">Section 4</span><h2 id="s4" style="margin-top:0.4rem">4. Findings</h2>`)
	if len(b.Inventory.Findings) == 0 {
		w(`<p class="ok">No findings recorded.</p>`)
	} else {
		sevOrder := map[string]int{"CRITICAL": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3, "INFO": 4}
		sorted := make([]model.Finding, len(b.Inventory.Findings))
		copy(sorted, b.Inventory.Findings)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].Rank > 0 && sorted[j].Rank > 0 && sorted[i].Rank != sorted[j].Rank {
				return sorted[i].Rank < sorted[j].Rank
			}
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
<span class="chip" style="border-color:var(--danger);color:var(--danger)">CRITICAL: %d</span>
<span class="chip" style="border-color:var(--warning-high);color:var(--warning-high)">HIGH: %d</span>
<span class="chip" style="border-color:var(--warning-medium);color:var(--warning-medium)">MEDIUM: %d</span>
<span class="chip">LOW/INFO: %d</span>
</p>`, crit, high, med, low)
		w(`<table><thead><tr><th>Rank</th><th>Severity</th><th>Owner</th><th>Impact</th><th>Effort</th><th>Resource</th><th>Finding</th><th>Recommendation</th></tr></thead><tbody>`)
		for _, f := range sorted {
			rank := "—"
			if f.Rank > 0 {
				rank = fmt.Sprintf("%d", f.Rank)
			}
			owner := f.OwnerHint
			if owner == "" {
				owner = "—"
			}
			impact := f.Impact
			if impact == "" {
				impact = "—"
			}
			effort := f.Effort
			if effort == "" {
				effort = "—"
			}
			wf(`<tr><td>%s</td><td class="c-%s">%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(rank), e(f.Severity), e(f.Severity), e(owner), e(impact), e(effort), e(f.ResourceID), e(f.Message), e(f.Recommendation))
		}
		w(`</tbody></table>`)
	}
	w(`</section>`)

	w(`<section class="card section-break"><span class="section-tag">Section 5</span><h2 id="s5" style="margin-top:0.4rem">5. DR Remediation Playbook</h2>`)
	if len(b.Inventory.RemediationSteps) == 0 {
		w(`<p class="subtle">No remediation steps generated. Run against a live cluster to produce findings.</p>`)
	} else {
		priLabel := map[int]string{
			1: "Priority 1 — Must Fix Before DR",
			2: "Priority 2 — Recommended",
			3: "Priority 3 — Optional",
		}
		priClass := map[int]string{1: "c-CRITICAL", 2: "c-HIGH", 3: "c-LOW"}
		chipBorder := map[int]string{1: "var(--danger)", 2: "var(--warning-high)", 3: "var(--muted)"}
		curPri := -1
		for _, step := range b.Inventory.RemediationSteps {
			if step.Priority != curPri {
				curPri = step.Priority
				wf(`<h3 class="%s">%s</h3>`, priClass[step.Priority], e(priLabel[step.Priority]))
			}
			wf(`<div class="step"><div class="step-h" style="cursor:default"><span class="chip" style="border-color:%s;color:%s">%s</span> %s</div>`,
				chipBorder[step.Priority], chipBorder[step.Priority], e(step.Category), e(step.Title))
			w(`<div class="step-b open">`)
			w(remediationBodyPrintHTML(step))
			w(`</div></div>`)
		}
	}
	w(`</section>`)

	w(`<section class="card"><span class="section-tag">Section 6</span><h2 id="s6" style="margin-top:0.4rem">6. Scan Metadata</h2>`)
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
		e(activeProfile),
		e(b.SchemaVersion))
	w(`</section>`)

	wf(`<div class="footer">Generated by k8s-recovery-visualizer %s &nbsp;|&nbsp; Scan ID: %s</div>`,
		e(b.Metadata.ToolVersion), e(b.Scan.ScanID))
	w(`</main>`)
	w(reportDocumentEnd())
}
