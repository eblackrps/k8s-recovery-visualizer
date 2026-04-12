package output

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"strings"

	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/profile"
)

// WriteReport writes the full tabbed dark-mode HTML report to path.
func WriteReport(path string, b *model.Bundle) error {
	var buf bytes.Buffer
	buildReport(&buf, b)
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func buildReport(buf *bytes.Buffer, b *model.Bundle) {
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
	generatedAt := b.Metadata.GeneratedAt
	if generatedAt == "" {
		generatedAt = "unknown"
	}
	scopeLabel := "all namespaces"
	if len(b.ScanNamespaces) > 0 {
		scopeLabel = strings.Join(b.ScanNamespaces, ", ")
	}
	activeProfile := b.Profile
	if activeProfile == "" {
		activeProfile = "standard"
	}
	coverageStatus := backupCoverageStatusText(b.Inventory.Backup)
	assuranceText := backupAssuranceConclusionText(b.Inventory.Backup.Assurance)
	criticalOrHigh := 0
	for _, f := range b.Inventory.Findings {
		if f.Severity == "CRITICAL" || f.Severity == "HIGH" {
			criticalOrHigh++
		}
	}
	priorityOne := 0
	for _, step := range b.Inventory.RemediationSteps {
		if step.Priority == 1 {
			priorityOne++
		}
	}
	actionHeadline := "No urgent actions"
	actionDetail := "No priority-one remediation steps were generated for this assessment."
	if priorityOne > 0 {
		actionHeadline = fmt.Sprintf("%d priority-one action(s)", priorityOne)
		actionDetail = "Immediate remediation focus for the current report scope."
	} else if criticalOrHigh > 0 {
		actionHeadline = fmt.Sprintf("%d critical/high finding(s)", criticalOrHigh)
		actionDetail = "Review the Findings and Remediation tabs for the highest-risk gaps."
	}
	w(reportDocumentStart("K8s DR Recovery Report", "report-page", reportPageCSS()))

	// Header
	wf(`<header class="hero">
<div class="hero-copy">
<span class="eyebrow">Recovery assessment</span>
<h1>K8s DR Recovery Report</h1>
<p>Offline-ready disaster recovery readiness, backup trust, and remediation guidance for the scanned Kubernetes environment.</p>
<div class="hero-meta-grid">
<span class="meta-chip">Cluster <strong>%s</strong></span>
<span class="meta-chip">Platform <strong>%s</strong></span>
<span class="meta-chip">Scope <strong>%s</strong></span>
<span class="meta-chip">Profile <strong>%s</strong></span>
<span class="meta-chip">Target <strong>%s</strong></span>
</div>
<div class="hero-brief-grid">
<div class="hero-brief">
<span class="hero-brief-label">Assessment posture</span>
<strong>%s maturity at score %d</strong>
<p>Scored against the <strong>%s</strong> profile for the current cluster scope.</p>
</div>
<div class="hero-brief">
<span class="hero-brief-label">Backup trust</span>
<strong>%s</strong>
<p>Backup assurance is currently <strong>%s</strong> for this environment.</p>
</div>
<div class="hero-brief">
<span class="hero-brief-label">Action focus</span>
<strong>%s</strong>
<p>%s</p>
</div>
</div>
</div>
<div class="hero-panel">
<div class="hero-pill">Generated %s</div>
<div class="hero-score">
<div>
<div class="hero-score-label">Overall DR Score</div>
<div class="hero-score-value" style="color:%s">%d</div>
</div>
<div class="badge" style="color:%s;border-color:%s">%s</div>
</div>
<div class="badge-row">
<span class="badge badge-subtle">Backup: %s</span>
<span class="badge badge-subtle">Coverage: %s</span>
<span class="badge badge-subtle">Assurance: %s</span>
</div>
<div class="hero-stat-grid">
<div class="hero-stat"><span>Nodes</span><strong>%d</strong></div>
<div class="hero-stat"><span>Namespaces</span><strong>%d</strong></div>
<div class="hero-stat"><span>Findings</span><strong>%d</strong></div>
<div class="hero-stat"><span>Policies</span><strong>%d</strong></div>
</div>
</div></header>`,
		e(clusterName), e(platform), e(scopeLabel), e(activeProfile), e(b.Target),
		e(b.Score.Maturity), b.Score.Overall.Final, e(activeProfile),
		e(coverageStatus), e(assuranceText),
		e(actionHeadline), e(actionDetail),
		e(generatedAt), overallTone, b.Score.Overall.Final,
		matColor, matColor, e(b.Score.Maturity),
		e(backupTool), e(coverageStatus), e(assuranceText),
		len(b.Inventory.Nodes), len(b.Inventory.Namespaces), len(b.Inventory.Findings), len(b.Inventory.Backup.Policies))

	// Tab bar — add Compare tab only when comparison data is present
	tabNames := []string{"Summary", "Nodes", "Workloads", "Storage", "Networking", "Config", "Images", "Backup", "DR Score", "Findings", "Remediation"}
	if b.Comparison != nil {
		tabNames = append(tabNames, "Compare")
	}
	w(`<nav class="tabs" role="tablist" aria-label="Report sections">`)
	for i, t := range tabNames {
		cls := "tab"
		selected := "false"
		tabIndex := -1
		if i == 0 {
			cls += " active"
			selected = "true"
			tabIndex = 0
		}
		wf(`<button type="button" id="tab-%d" class="%s" role="tab" aria-selected="%s" aria-controls="p%d" tabindex="%d" onclick="show(%d)">%s</button>`,
			i, cls, selected, i, tabIndex, i, e(t))
	}
	w(`</nav>`)

	// ── Tab 0: Summary ───────────────────────────────────────────────────────
	w(`<div class="pane active" id="p0" role="tabpanel" aria-labelledby="tab-0">`)
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
		wf(`<div class="sbox"><div class="v">%d</div><div class="l">%s <span style="color:var(--accent-soft)">%s</span></div><div class="bar"><div class="fill" style="width:%d%%"></div></div></div>`,
			d.score, e(d.label), e(d.weight), d.score)
	}
	w(`</div>`)

	btClass := "ok"
	if backupTool == "none" {
		btClass = "bad"
	}
	wf(`<div class="card"><h2>Environment</h2><table><tbody>
<tr><td>Provider</td><td>%s</td></tr>
<tr><td>K8s Version</td><td>%s</td></tr>
<tr><td>Cluster UID</td><td>%s</td></tr>
<tr><td>Backup Tool</td><td class="%s">%s</td></tr>
<tr><td>Backup Coverage</td><td>%s</td></tr>
<tr><td>Nodes</td><td>%d</td></tr>
<tr><td>Namespaces</td><td>%d</td></tr>
<tr><td>Helm Releases</td><td>%d</td></tr>
<tr><td>Certificates</td><td>%d</td></tr>
<tr><td>Recovery Target</td><td>%s</td></tr>
<tr><td>Namespace Scope</td><td>%s</td></tr>
</tbody></table></div>`,
		e(platform), e(b.Cluster.Platform.K8sVersion), e(b.Cluster.Platform.ClusterUID),
		btClass, e(backupTool), e(backupCoverageStatusText(b.Inventory.Backup)),
		len(b.Inventory.Nodes), len(b.Inventory.Namespaces),
		len(b.Inventory.HelmReleases), len(b.Inventory.Certificates),
		e(b.Target), e(scopeLabel))

	crit, high, med, low := 0, 0, 0, 0
	for _, f := range b.Inventory.Findings {
		switch f.Severity {
		case "CRITICAL":
			crit++
		case "HIGH":
			high++
		case "MEDIUM":
			med++
		case "LOW", "INFO":
			low++
		}
	}
	// Severity bar chart — proportional bars, max bar = 200px
	sevMax := crit
	if high > sevMax {
		sevMax = high
	}
	if med > sevMax {
		sevMax = med
	}
	if low > sevMax {
		sevMax = low
	}
	sevBar := func(count, maxCount int, color string) string {
		if maxCount == 0 || count == 0 {
			return fmt.Sprintf(`<div style="width:0;height:10px;border-radius:3px;background:%s;display:inline-block"></div>`, color)
		}
		w := count * 200 / maxCount
		if w < 4 {
			w = 4
		}
		return fmt.Sprintf(`<div style="width:%dpx;height:10px;border-radius:3px;background:%s;display:inline-block"></div>`, w, color)
	}
	w(`<div class="card"><h2>Findings Summary</h2>`)
	w(`<table style="border-collapse:collapse;margin-top:6px;font-size:.86em">`)
	for _, row := range []struct {
		label, color string
		count        int
	}{
		{"CRITICAL", "var(--danger)", crit},
		{"HIGH", "var(--warning-high)", high},
		{"MEDIUM", "var(--warning-medium)", med},
		{"LOW / INFO", "var(--muted)", low},
	} {
		wf(`<tr><td style="color:%s;width:80px;padding:3px 0">%s</td><td style="padding:3px 8px">%s</td><td style="padding:3px 0;color:%s">%d</td></tr>`,
			row.color, row.label, sevBar(row.count, sevMax, row.color), row.color, row.count)
	}
	w(`</table>`)
	w(`<p class="subtle" style="margin-top:10px">Full details -> <button type="button" class="inline-link" onclick="showTab('Findings')">Findings</button> tab. Action steps -> <button type="button" class="inline-link" onclick="showTab('Remediation')">Remediation</button> tab.</p>`)
	w(`</div>`)

	// Scan coverage / skipped collectors callout
	totalCollectors := 25 // total number of optional collectors attempted
	skipped := len(b.CollectorSkips)
	rbacSkips := 0
	for _, sk := range b.CollectorSkips {
		if sk.RBAC {
			rbacSkips++
		}
	}
	if skipped > 0 {
		w(`<div class="card" style="border-color:var(--warning-medium)">`)
		wf(`<h2 style="color:var(--warning-medium)">Scan Coverage — %d/%d collectors skipped</h2>`, skipped, totalCollectors)
		if rbacSkips > 0 {
			wf(`<p style="color:var(--muted);font-size:.86em;margin-bottom:8px">%d skip(s) appear to be RBAC / permissions errors. Grant the service account read access to the listed resources to improve coverage.</p>`, rbacSkips)
		}
		w(`<table style="margin-top:4px"><thead><tr>`)
		for _, h := range []string{"Collector", "Reason", "RBAC?"} {
			wf(`<th>%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, sk := range b.CollectorSkips {
			rbacCell := `<span class="bad">✗ No</span>`
			if sk.RBAC {
				rbacCell = `<span style="color:var(--warning-medium)">⚠ Yes</span>`
			}
			// Truncate long reasons for display
			reason := sk.Reason
			if len(reason) > 120 {
				reason = reason[:117] + "..."
			}
			wf(`<tr><td>%s</td><td style="color:var(--muted);font-size:.84em">%s</td><td>%s</td></tr>`,
				e(sk.Name), e(reason), rbacCell)
		}
		w(`</tbody></table></div>`)
	} else {
		wf(`<div class="card" style="border-color:var(--line-soft)"><h2>Scan Coverage</h2>
<p style="color:var(--success);font-size:.86em">All %d collectors completed successfully — full inventory captured.</p></div>`, totalCollectors)
	}

	// ── Round 10c: Score Trend sparkline ──────────────────────────────────
	if len(b.TrendHistory) > 1 {
		n := len(b.TrendHistory)
		const svgW, svgH = 500, 90
		last := b.TrendHistory[n-1]
		prev := b.TrendHistory[n-2]
		trendColor := "var(--accent)"
		trendLabel := "STABLE"
		if last.Overall > prev.Overall {
			trendColor = "var(--success)"
			trendLabel = "IMPROVING"
		} else if last.Overall < prev.Overall {
			trendColor = "var(--danger)"
			trendLabel = "DECLINING"
		}

		var ptsSlice []string
		type dot struct{ cx, cy float64 }
		var dots []dot
		for i, tp := range b.TrendHistory {
			x := float64(10) + float64(i)*float64(svgW-20)/float64(n-1)
			y := float64(svgH-18) - float64(tp.Overall)*float64(svgH-28)/100.0 + 4
			ptsSlice = append(ptsSlice, fmt.Sprintf("%.1f,%.1f", x, y))
			dots = append(dots, dot{x, y})
		}

		w(`<div class="card"><h2>Score Trend`)
		wf(` &nbsp;<span style="color:%s;font-size:.82em;font-weight:400">%s</span></h2>`, trendColor, trendLabel)
		wf(`<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg" style="width:100%%;max-width:500px;height:%dpx;display:block;margin:4px 0">`,
			svgW, svgH, svgH)
		for _, ref := range []struct {
			v   int
			lbl string
		}{{100, "100"}, {75, "75"}, {50, "50"}, {0, "0"}} {
			ry := float64(svgH-18) - float64(ref.v)*float64(svgH-28)/100.0 + 4
			dash := "4,4"
			if ref.v == 75 {
				dash = "2,6"
			}
			wf(`<line x1="30" y1="%.1f" x2="%d" y2="%.1f" stroke="var(--line)" stroke-dasharray="%s" stroke-width="1"/>`, ry, svgW-5, ry, dash)
			wf(`<text x="0" y="%.1f" fill="var(--muted)" font-size="9" dominant-baseline="middle">%s</text>`, ry, ref.lbl)
		}
		wf(`<defs><linearGradient id="sg" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="%s" stop-opacity="0.25"/><stop offset="1" stop-color="%s" stop-opacity="0"/></linearGradient></defs>`,
			trendColor, trendColor)
		firstDot := dots[0]
		lastDot := dots[n-1]
		wf(`<polygon points="%s %.1f,%.1f %.1f,%.1f" fill="url(#sg)"/>`,
			strings.Join(ptsSlice, " "),
			lastDot.cx, float64(svgH-14),
			firstDot.cx, float64(svgH-14))
		wf(`<polyline points="%s" fill="none" stroke="%s" stroke-width="2" stroke-linejoin="round"/>`,
			strings.Join(ptsSlice, " "), trendColor)
		for i, d := range dots {
			r := "2.5"
			if i == n-1 {
				r = "4"
			}
			wf(`<circle cx="%.1f" cy="%.1f" r="%s" fill="%s"/>`, d.cx, d.cy, r, trendColor)
		}
		w(`</svg>`)
		wf(`<div style="color:var(--muted);font-size:.83em;margin-top:4px">Last <strong style="color:var(--text)">%d</strong> scans &mdash; Current: <strong style="color:%s">%d</strong> (%s)</div>`,
			n, trendColor, last.Overall, e(last.Maturity))
		w(`</div>`)
	}

	w(`</div>`) // p0

	// ── Tab 1: Nodes ─────────────────────────────────────────────────────────
	w(`<div class="pane" id="p1" role="tabpanel" aria-labelledby="tab-1"><h2>Nodes</h2>`)
	if len(b.Inventory.Nodes) == 0 {
		w(`<div class="empty">No node data collected.</div>`)
	} else {
		w(`<table id="t-nodes"><thead><tr>`)
		for _, h := range []string{"Name", "Roles", "Ready", "Zone", "OS", "Kernel", "Runtime", "Kubelet", "Internal IP", "Taints"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, n := range b.Inventory.Nodes {
			rdStr := `<span class="bad">✗</span>`
			if n.Ready {
				rdStr = `<span class="ok">✓</span>`
			}
			zone := n.Zone
			if zone == "" {
				zone = "—"
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(n.Name), e(strings.Join(n.Roles, ",")), rdStr, e(zone),
				e(n.OSImage), e(n.KernelVersion), e(n.ContainerRuntime),
				e(n.KubeletVersion), e(n.InternalIP), e(strings.Join(n.Taints, " ")))
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // p1

	// ── Tab 2: Workloads ─────────────────────────────────────────────────────
	w(`<div class="pane" id="p2" role="tabpanel" aria-labelledby="tab-2"><h2>Workloads</h2>`)

	// ── Round 11: Resource Governance + Round 12 pod security summary ──────
	{
		totalPods := len(b.Inventory.Pods)
		noReq, noLim, priv, hostNS := 0, 0, 0, 0
		for _, pod := range b.Inventory.Pods {
			if pod.Namespace == "kube-system" {
				continue
			}
			if !pod.HasRequests {
				noReq++
			}
			if !pod.HasLimits {
				noLim++
			}
			if pod.Privileged {
				priv++
			}
			if pod.HostNetwork || pod.HostPID {
				hostNS++
			}
		}
		reqColor := "var(--success)"
		if noReq > 0 {
			reqColor = "var(--warning-high)"
		}
		limColor := "var(--success)"
		if noLim > 0 {
			limColor = "var(--warning-medium)"
		}
		privColor := "var(--success)"
		if priv > 0 {
			privColor = "var(--danger)"
		}
		hostColor := "var(--success)"
		if hostNS > 0 {
			hostColor = "var(--warning-high)"
		}
		wf(`<div class="card"><h2>Resource Governance &amp; Pod Security</h2>
<p style="color:var(--muted);font-size:.84em;margin-bottom:10px">kube-system pods are excluded from governance checks.</p>
<div class="grid">
<div class="sbox"><div class="v">%d</div><div class="l">Total Pods</div></div>
<div class="sbox"><div class="v" style="color:%s">%d</div><div class="l">Missing Requests</div></div>
<div class="sbox"><div class="v" style="color:%s">%d</div><div class="l">Missing Limits</div></div>
<div class="sbox"><div class="v" style="color:%s">%d</div><div class="l">Privileged</div></div>
<div class="sbox"><div class="v" style="color:%s">%d</div><div class="l">Host Net/PID</div></div>
</div>`,
			totalPods, reqColor, noReq, limColor, noLim, privColor, priv, hostColor, hostNS)

		if totalPods > 0 {
			w(`<h3 style="margin-top:12px">Pod Inventory</h3>`)
			w(`<table id="t-pods"><thead><tr>`)
			for _, h := range []string{"Namespace", "Name", "Containers", "Requests", "Limits", "Privileged", "HostNet/PID"} {
				wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
			}
			w(`</tr></thead><tbody>`)
			for _, pod := range b.Inventory.Pods {
				reqCell := `<span class="ok">✓</span>`
				if !pod.HasRequests {
					reqCell = `<span class="bad">✗</span>`
				}
				limCell := `<span class="ok">✓</span>`
				if !pod.HasLimits {
					limCell = `<span class="c-MEDIUM">✗</span>`
				}
				privCell := `<span style="color:var(--muted)">—</span>`
				if pod.Privileged {
					privCell = `<span class="bad">yes</span>`
				}
				hnCell := `<span style="color:var(--muted)">—</span>`
				if pod.HostNetwork || pod.HostPID {
					parts := []string{}
					if pod.HostNetwork {
						parts = append(parts, "net")
					}
					if pod.HostPID {
						parts = append(parts, "pid")
					}
					hnCell = fmt.Sprintf(`<span class="c-HIGH">%s</span>`, e(strings.Join(parts, "+")))
				}
				wf(`<tr><td>%s</td><td>%s</td><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
					e(pod.Namespace), e(pod.Name), pod.ContainerCount, reqCell, limCell, privCell, hnCell)
			}
			w(`</tbody></table>`)
		}
		w(`</div>`) // governance card
	}

	w(`<table id="t-workloads"><thead><tr>`)
	for _, h := range []string{"Type", "Namespace", "Name", "Replicas", "Ready/Status", "Images"} {
		wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
	}
	w(`</tr></thead><tbody>`)
	for _, d := range b.Inventory.Deployments {
		wf(`<tr><td>Deployment</td><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%s</td></tr>`,
			e(d.Namespace), e(d.Name), d.Replicas, d.Ready, e(strings.Join(d.Images, ", ")))
	}
	for _, ds := range b.Inventory.DaemonSets {
		wf(`<tr><td>DaemonSet</td><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%s</td></tr>`,
			e(ds.Namespace), e(ds.Name), ds.Desired, ds.Ready, e(strings.Join(ds.Images, ", ")))
	}
	for _, sts := range b.Inventory.StatefulSets {
		pvcBadge := `<span class="chip f">no PVC</span>`
		if sts.HasVolumeClaim {
			pvcBadge = `<span class="chip p">has PVC</span>`
		}
		wf(`<tr><td>StatefulSet</td><td>%s</td><td>%s</td><td>%d</td><td>%s</td><td></td></tr>`,
			e(sts.Namespace), e(sts.Name), sts.Replicas, pvcBadge)
	}
	for _, j := range b.Inventory.Jobs {
		done := `<span style="color:var(--muted)">active</span>`
		if j.Completed {
			done = `<span class="ok">done</span>`
		}
		wf(`<tr><td>Job</td><td>%s</td><td>%s</td><td>–</td><td>%s</td><td></td></tr>`,
			e(j.Namespace), e(j.Name), done)
	}
	for _, cj := range b.Inventory.CronJobs {
		wf(`<tr><td>CronJob</td><td>%s</td><td>%s</td><td>–</td><td>%s</td><td></td></tr>`,
			e(cj.Namespace), e(cj.Name), e(cj.Schedule))
	}
	w(`</tbody></table></div>`) // p2

	// ── Tab 3: Storage ───────────────────────────────────────────────────────
	w(`<div class="pane" id="p3" role="tabpanel" aria-labelledby="tab-3"><h2>Storage</h2>`)
	pvMap := map[string]model.PersistentVolume{}
	for _, pv := range b.Inventory.PVs {
		pvMap[pv.ClaimRef] = pv
	}
	w(`<h3>PersistentVolumeClaims</h3><table id="t-pvcs"><thead><tr>`)
	for _, h := range []string{"Namespace", "Name", "StorageClass", "Access", "Size", "DR Risk"} {
		wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
	}
	w(`</tr></thead><tbody>`)
	for _, pvc := range b.Inventory.PVCs {
		key := pvc.Namespace + "/" + pvc.Name
		risk := `<span class="ok">Low</span>`
		pv, bound := pvMap[key]
		if !bound {
			risk = `<span class="bad">Unbound</span>`
		} else if pv.Backend == "hostPath" {
			risk = `<span class="bad">hostPath</span>`
		} else if pv.ReclaimPolicy == "Delete" {
			risk = `<span class="c-HIGH">Delete policy</span>`
		}
		wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			e(pvc.Namespace), e(pvc.Name), e(pvc.StorageClass),
			e(strings.Join(pvc.AccessModes, ",")), e(pvc.RequestedSize), risk)
	}
	w(`</tbody></table>`)
	w(`<h3>PersistentVolumes</h3><table id="t-pvs"><thead><tr>`)
	for _, h := range []string{"Name", "StorageClass", "Capacity", "Backend", "Reclaim", "Bound To"} {
		wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
	}
	w(`</tr></thead><tbody>`)
	for _, pv := range b.Inventory.PVs {
		wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			e(pv.Name), e(pv.StorageClass), e(pv.Capacity), e(pv.Backend), e(pv.ReclaimPolicy), e(pv.ClaimRef))
	}
	w(`</tbody></table>`)
	w(`<h3>StorageClasses</h3><table id="t-sc"><thead><tr>`)
	for _, h := range []string{"Name", "Provisioner", "Reclaim", "Binding Mode", "Expandable"} {
		wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
	}
	w(`</tr></thead><tbody>`)
	for _, sc := range b.Inventory.StorageClasses {
		exp := "–"
		if sc.AllowVolumeExpansion != nil {
			if *sc.AllowVolumeExpansion {
				exp = "yes"
			} else {
				exp = "no"
			}
		}
		wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			e(sc.Name), e(sc.Provisioner), e(sc.ReclaimPolicy), e(sc.VolumeBindingMode), exp)
	}
	w(`</tbody></table>`)

	// ── Round 13: VolumeSnapshot coverage ─────────────────────────────────
	{
		snappedPVCs := map[string]bool{}
		for _, vs := range b.Inventory.VolumeSnapshots {
			if vs.PVCName != "" {
				snappedPVCs[vs.Namespace+"/"+vs.PVCName] = true
			}
		}

		w(`<h3>VolumeSnapshot Coverage</h3>`)
		if len(b.Inventory.VolumeSnapshotClasses) == 0 {
			w(`<div class="empty" style="color:var(--warning-high)">No VolumeSnapshotClasses found — CSI snapshot infrastructure not configured.</div>`)
		} else {
			// Snapshot class table
			w(`<table id="t-vsc" style="margin-bottom:12px"><thead><tr>`)
			for _, h := range []string{"VolumeSnapshotClass", "Driver", "Deletion Policy"} {
				wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
			}
			w(`</tr></thead><tbody>`)
			for _, vsc := range b.Inventory.VolumeSnapshotClasses {
				dpColor := "var(--success)"
				if vsc.DeletionPolicy == "Delete" {
					dpColor = "var(--warning-high)"
				}
				wf(`<tr><td>%s</td><td>%s</td><td style="color:%s">%s</td></tr>`,
					e(vsc.Name), e(vsc.Driver), dpColor, e(vsc.DeletionPolicy))
			}
			w(`</tbody></table>`)
		}

		if len(b.Inventory.PVCs) > 0 {
			w(`<table id="t-vsnap"><thead><tr>`)
			for _, h := range []string{"Namespace", "PVC", "Snapshot", "Ready", "Size (GB)", "Created"} {
				wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
			}
			w(`</tr></thead><tbody>`)
			// Show PVCs first, annotated with snapshot status
			for _, pvc := range b.Inventory.PVCs {
				key := pvc.Namespace + "/" + pvc.Name
				if snappedPVCs[key] {
					continue // will be shown via snapshots below
				}
				snapCell := `<span class="bad">none</span>`
				wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>—</td><td>—</td><td>—</td></tr>`,
					e(pvc.Namespace), e(pvc.Name), snapCell)
			}
			for _, vs := range b.Inventory.VolumeSnapshots {
				rdyCell := `<span class="bad">✗</span>`
				if vs.ReadyToUse {
					rdyCell = `<span class="ok">✓</span>`
				}
				sizeCell := "–"
				if vs.SizeGB > 0 {
					sizeCell = fmt.Sprintf("%.1f", vs.SizeGB)
				}
				wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
					e(vs.Namespace), e(vs.PVCName), e(vs.Name), rdyCell, sizeCell, e(vs.CreatedAt))
			}
			w(`</tbody></table>`)
		}
	}

	w(`</div>`) // p3

	// ── Tab 4: Networking ────────────────────────────────────────────────────
	w(`<div class="pane" id="p4" role="tabpanel" aria-labelledby="tab-4"><h2>Networking</h2>`)

	// ── Round 10b: Network Topology card ──────────────────────────────────
	{
		// Count service types
		lbCount, npCount, nodePortCount := 0, 0, 0
		npNamespaces := map[string]bool{}
		for _, svc := range b.Inventory.Services {
			switch svc.Type {
			case "LoadBalancer":
				lbCount++
			case "NodePort":
				nodePortCount++
			}
		}
		for _, np := range b.Inventory.NetworkPolicies {
			npNamespaces[np.Namespace] = true
			npCount++
		}
		totalNS := len(b.Inventory.Namespaces)
		coveredNS := len(npNamespaces)

		npColor := "var(--success)"
		if coveredNS < totalNS/2 {
			npColor = "var(--danger)"
		} else if coveredNS < totalNS {
			npColor = "var(--warning-high)"
		}

		// Count total ingress rules
		ingressRules := 0
		for _, ing := range b.Inventory.Ingresses {
			ingressRules += len(ing.Rules)
		}

		w(`<div class="card"><h2>Network Topology</h2>`)
		wf(`<div class="grid" style="margin-bottom:14px">
<div class="sbox"><div class="v">%d</div><div class="l">LoadBalancer</div></div>
<div class="sbox"><div class="v">%d</div><div class="l">NodePort</div></div>
<div class="sbox"><div class="v">%d</div><div class="l">Ingress Rules</div></div>
<div class="sbox"><div class="v" style="color:%s">%d / %d</div><div class="l">NS w/ NetworkPolicy</div></div>
</div>`, lbCount, nodePortCount, ingressRules, npColor, coveredNS, totalNS)

		// Ingress → Service adjacency map (build from ingress rules)
		if len(b.Inventory.Ingresses) > 0 {
			w(`<h3 style="margin-top:10px">Ingress → Service Connectivity</h3>`)
			w(`<table style="margin-top:6px"><thead><tr>`)
			for _, h := range []string{"Ingress", "Namespace", "Host", "Backend Service", "TLS", "NP Covered"} {
				wf(`<th>%s</th>`, e(h))
			}
			w(`</tr></thead><tbody>`)
			for _, ing := range b.Inventory.Ingresses {
				tlsStr := "–"
				if ing.TLS {
					tlsStr = `<span class="ok">✓ TLS</span>`
				}
				npCovered := npNamespaces[ing.Namespace]
				npCell := `<span class="bad">✗</span>`
				if npCovered {
					npCell = `<span class="ok">✓</span>`
				}
				if len(ing.Rules) == 0 {
					wf(`<tr><td>%s</td><td>%s</td><td colspan="2"><span style="color:var(--muted)">no rules</span></td><td>%s</td><td>%s</td></tr>`,
						e(ing.Name), e(ing.Namespace), tlsStr, npCell)
					continue
				}
				for i, r := range ing.Rules {
					if i == 0 {
						wf(`<tr><td rowspan="%d">%s</td><td rowspan="%d">%s</td><td>%s</td><td>%s</td><td rowspan="%d">%s</td><td rowspan="%d">%s</td></tr>`,
							len(ing.Rules), e(ing.Name),
							len(ing.Rules), e(ing.Namespace),
							e(r.Host), e(r.Backend),
							len(ing.Rules), tlsStr,
							len(ing.Rules), npCell)
					} else {
						wf(`<tr><td>%s</td><td>%s</td></tr>`, e(r.Host), e(r.Backend))
					}
				}
			}
			w(`</tbody></table>`)
		}

		// Namespace protection summary
		if totalNS > 0 {
			w(`<h3 style="margin-top:14px">Namespace Exposure Summary</h3>`)
			w(`<table style="margin-top:6px"><thead><tr>`)
			for _, h := range []string{"Namespace", "Ingress", "LB Service", "NetworkPolicy"} {
				wf(`<th>%s</th>`, e(h))
			}
			w(`</tr></thead><tbody>`)

			// Build lookup maps
			nsHasIngress := map[string]bool{}
			nsHasLB := map[string]bool{}
			for _, ing := range b.Inventory.Ingresses {
				nsHasIngress[ing.Namespace] = true
			}
			for _, svc := range b.Inventory.Services {
				if svc.Type == "LoadBalancer" {
					nsHasLB[svc.Namespace] = true
				}
			}

			for _, ns := range b.Inventory.Namespaces {
				ingCell := `<span style="color:var(--muted)">—</span>`
				if nsHasIngress[ns.Name] {
					ingCell = `<span class="c-HIGH">exposed</span>`
				}
				lbCell := `<span style="color:var(--muted)">—</span>`
				if nsHasLB[ns.Name] {
					lbCell = `<span class="c-HIGH">exposed</span>`
				}
				npCell := `<span class="bad">none</span>`
				if npNamespaces[ns.Name] {
					npCell = `<span class="ok">covered</span>`
				}
				wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
					e(ns.Name), ingCell, lbCell, npCell)
			}
			w(`</tbody></table>`)
		}
		w(`</div>`) // topology card
	}

	w(`<h3>Services</h3><table id="t-svc"><thead><tr>`)
	for _, h := range []string{"Namespace", "Name", "Type", "Cluster IP", "External IP"} {
		wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
	}
	w(`</tr></thead><tbody>`)
	for _, svc := range b.Inventory.Services {
		wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			e(svc.Namespace), e(svc.Name), e(svc.Type), e(svc.ClusterIP), e(svc.ExternalIP))
	}
	w(`</tbody></table>`)
	w(`<h3>Ingresses</h3><table id="t-ing"><thead><tr>`)
	for _, h := range []string{"Namespace", "Name", "Class", "TLS", "Rules"} {
		wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
	}
	w(`</tr></thead><tbody>`)
	for _, ing := range b.Inventory.Ingresses {
		tls := "–"
		if ing.TLS {
			tls = "✓"
		}
		var rules []string
		for _, r := range ing.Rules {
			rules = append(rules, r.Host+" → "+r.Backend)
		}
		wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			e(ing.Namespace), e(ing.Name), e(ing.ClassName), e(tls), e(strings.Join(rules, "; ")))
	}
	w(`</tbody></table>`)
	w(`<h3>NetworkPolicies</h3><table id="t-np"><thead><tr>`)
	for _, h := range []string{"Namespace", "Name", "Pod Selector", "Ingress", "Egress"} {
		wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
	}
	w(`</tr></thead><tbody>`)
	for _, np := range b.Inventory.NetworkPolicies {
		hasI, hasE := "–", "–"
		if np.HasIngress {
			hasI = "✓"
		}
		if np.HasEgress {
			hasE = "✓"
		}
		wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			e(np.Namespace), e(np.Name), e(np.PodSelector), hasI, hasE)
	}
	w(`</tbody></table></div>`) // p4

	// ── Tab 5: Config ────────────────────────────────────────────────────────
	w(`<div class="pane" id="p5" role="tabpanel" aria-labelledby="tab-5"><h2>Config</h2>`)
	w(`<h3>Helm Releases</h3>`)
	if len(b.Inventory.HelmReleases) == 0 {
		w(`<div class="empty">No Helm releases detected.</div>`)
	} else {
		w(`<table id="t-helm"><thead><tr>`)
		for _, h := range []string{"Namespace", "Release", "Chart", "Version", "Status"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, hr := range b.Inventory.HelmReleases {
			sc := "var(--muted)"
			if hr.Status == "deployed" {
				sc = "var(--success)"
			} else if hr.Status == "failed" {
				sc = "var(--danger)"
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td style="color:%s">%s</td></tr>`,
				e(hr.Namespace), e(hr.Name), e(hr.Chart), e(hr.Version), sc, e(hr.Status))
		}
		w(`</tbody></table>`)
	}
	w(`<h3>Certificates (cert-manager)</h3>`)
	if len(b.Inventory.Certificates) == 0 {
		w(`<div class="empty">No cert-manager certificates detected.</div>`)
	} else {
		w(`<table id="t-certs"><thead><tr>`)
		for _, h := range []string{"Namespace", "Name", "Issuer", "Ready", "Expires", "Days Left"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, c := range b.Inventory.Certificates {
			rdStr := `<span class="ok">✓</span>`
			if !c.Ready {
				rdStr = `<span class="bad">✗</span>`
			}
			dc := "var(--success)"
			if c.DaysToExpiry < 30 {
				dc = "var(--danger)"
			} else if c.DaysToExpiry < 60 {
				dc = "var(--warning-high)"
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td style="color:%s">%d</td></tr>`,
				e(c.Namespace), e(c.Name), e(c.Issuer), rdStr, e(c.NotAfter), dc, c.DaysToExpiry)
		}
		w(`</tbody></table>`)
	}
	w(`<h3>Custom Resource Definitions</h3>`)
	if len(b.Inventory.CRDs) == 0 {
		w(`<div class="empty">No custom API groups detected.</div>`)
	} else {
		w(`<table id="t-crds"><thead><tr>`)
		for _, h := range []string{"Group", "Versions", "Scope"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, crd := range b.Inventory.CRDs {
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(crd.Group), e(strings.Join(crd.Versions, ",")), e(crd.Scope))
		}
		w(`</tbody></table>`)
	}
	type nsCS struct{ cm, sec int }
	nsCounts := map[string]*nsCS{}
	for _, cm := range b.Inventory.ConfigMaps {
		if nsCounts[cm.Namespace] == nil {
			nsCounts[cm.Namespace] = &nsCS{}
		}
		nsCounts[cm.Namespace].cm++
	}
	for _, s := range b.Inventory.Secrets {
		if nsCounts[s.Namespace] == nil {
			nsCounts[s.Namespace] = &nsCS{}
		}
		nsCounts[s.Namespace].sec++
	}
	if len(nsCounts) > 0 {
		w(`<h3>ConfigMaps &amp; Secrets by Namespace</h3><table id="t-cms"><thead><tr>`)
		for _, h := range []string{"Namespace", "ConfigMaps", "Secrets"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for ns, c := range nsCounts {
			wf(`<tr><td>%s</td><td>%d</td><td>%d</td></tr>`, e(ns), c.cm, c.sec)
		}
		w(`</tbody></table>`)
	}
	if len(b.Inventory.ResourceQuotas) > 0 {
		w(`<h3>Resource Quotas</h3><table id="t-rq"><thead><tr>`)
		for _, h := range []string{"Namespace", "Name", "Resource", "Hard", "Used"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, rq := range b.Inventory.ResourceQuotas {
			for _, item := range rq.Items {
				wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
					e(rq.Namespace), e(rq.Name), e(item.Resource), e(item.Hard), e(item.Used))
			}
		}
		w(`</tbody></table>`)
	}

	// ── Round 14: LimitRange enforcement ─────────────────────────────────
	w(`<h3>LimitRange Enforcement</h3>`)
	if len(b.Inventory.LimitRanges) == 0 {
		w(`<div class="empty" style="color:var(--warning-high)">No LimitRanges found — namespaces have no default resource constraints.</div>`)
	} else {
		w(`<table id="t-lr"><thead><tr>`)
		for _, h := range []string{"Namespace", "Name", "Type", "Max CPU", "Max Memory", "Default CPU", "Default Memory"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, lr := range b.Inventory.LimitRanges {
			if len(lr.Items) == 0 {
				wf(`<tr><td>%s</td><td>%s</td><td colspan="5"><span style="color:var(--muted)">no items</span></td></tr>`,
					e(lr.Namespace), e(lr.Name))
				continue
			}
			for _, item := range lr.Items {
				dash := `<span style="color:var(--muted)">—</span>`
				maxCPU, maxMem, defCPU, defMem := dash, dash, dash, dash
				if item.MaxCPU != "" {
					maxCPU = e(item.MaxCPU)
				}
				if item.MaxMemory != "" {
					maxMem = e(item.MaxMemory)
				}
				if item.DefaultCPU != "" {
					defCPU = e(item.DefaultCPU)
				}
				if item.DefaultMemory != "" {
					defMem = e(item.DefaultMemory)
				}
				wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
					e(lr.Namespace), e(lr.Name), e(item.Type), maxCPU, maxMem, defCPU, defMem)
			}
		}
		w(`</tbody></table>`)
	}

	// ── Round 14: PSA label coverage ─────────────────────────────────────
	w(`<h3>Pod Security Admission (PSA) Coverage</h3>`)
	w(`<p style="color:var(--muted);font-size:.84em;margin-bottom:8px">Namespaces should carry <code>pod-security.kubernetes.io/enforce</code> labels to activate PSA admission control. System namespaces are excluded.</p>`)
	if len(b.Inventory.Namespaces) == 0 {
		w(`<div class="empty">No namespace data collected.</div>`)
	} else {
		w(`<table id="t-psa"><thead><tr>`)
		for _, h := range []string{"Namespace", "Enforce", "Warn", "Audit", "Status"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, ns := range b.Inventory.Namespaces {
			if ns.Name == "kube-system" || ns.Name == "kube-public" || ns.Name == "kube-node-lease" {
				continue
			}
			dash := `<span style="color:var(--muted)">—</span>`
			enf := dash
			if ns.PSAEnforce != "" {
				enf = fmt.Sprintf(`<span class="chip p">%s</span>`, e(ns.PSAEnforce))
			}
			wrn := dash
			if ns.PSAWarn != "" {
				wrn = e(ns.PSAWarn)
			}
			aud := dash
			if ns.PSAAudit != "" {
				aud = e(ns.PSAAudit)
			}
			statusCell := `<span class="bad">missing</span>`
			if ns.PSAEnforce != "" {
				statusCell = `<span class="ok">✓</span>`
			} else if ns.PSAWarn != "" || ns.PSAAudit != "" {
				statusCell = `<span style="color:var(--warning-high)">warn/audit only</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(ns.Name), enf, wrn, aud, statusCell)
		}
		w(`</tbody></table>`)
	}

	// RBAC privilege audit — custom ClusterRoles only
	w(`<h3>RBAC Security — Custom ClusterRoles</h3>`)
	var customRoles []model.ClusterRole
	for _, cr := range b.Inventory.ClusterRoles {
		if cr.Custom {
			customRoles = append(customRoles, cr)
		}
	}
	if len(customRoles) == 0 {
		w(`<div class="empty">No custom ClusterRoles detected (only built-in system: roles present).</div>`)
	} else {
		w(`<table id="t-rbac"><thead><tr>`)
		for _, h := range []string{"Name", "Rules", "Wildcard Verb", "Secret Access", "Escalate/Bind", "Risk"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, cr := range customRoles {
			wcCell := `<span style="color:var(--muted)">—</span>`
			if cr.HasWildcardVerb {
				wcCell = `<span class="bad">yes</span>`
			}
			saCell := `<span style="color:var(--muted)">—</span>`
			if cr.HasSecretAccess {
				saCell = `<span class="c-HIGH">yes</span>`
			}
			esCell := `<span style="color:var(--muted)">—</span>`
			if cr.HasEscalatePriv {
				esCell = `<span class="c-HIGH">yes</span>`
			}
			riskLabel, riskColor := "clean", "var(--success)"
			if cr.HasWildcardVerb {
				riskLabel, riskColor = "CRITICAL", "var(--danger)"
			} else if cr.HasSecretAccess || cr.HasEscalatePriv {
				riskLabel, riskColor = "HIGH", "var(--warning-high)"
			}
			riskCell := fmt.Sprintf(`<span style="color:%s;font-weight:700">%s</span>`, riskColor, riskLabel)
			wf(`<tr><td>%s</td><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(cr.Name), cr.RuleCount, wcCell, saCell, esCell, riskCell)
		}
		w(`</tbody></table>`)
	}

	// ── Round 10a: ClusterRoleBinding subject audit ───────────────────────
	if len(b.Inventory.ClusterRoleBindings) > 0 {
		// Build role-risk index from collected ClusterRoles.
		roleRisk := map[string]string{}
		for _, cr := range b.Inventory.ClusterRoles {
			switch {
			case cr.Name == "cluster-admin":
				roleRisk[cr.Name] = "CRITICAL"
			case cr.HasWildcardVerb || cr.HasEscalatePriv:
				if roleRisk[cr.Name] == "" {
					roleRisk[cr.Name] = "HIGH"
				}
			case cr.HasSecretAccess:
				if roleRisk[cr.Name] == "" {
					roleRisk[cr.Name] = "MEDIUM"
				}
			}
		}

		w(`<h3>RBAC: ClusterRoleBinding Subject Audit</h3>`)
		w(`<p style="color:var(--muted);font-size:.84em;margin-bottom:8px">Shows who holds cluster-level permissions. System-to-system bindings (both role and all subjects prefixed with <code>system:</code>) are hidden for clarity.</p>`)
		w(`<table id="t-crbs"><thead><tr>`)
		for _, h := range []string{"Binding", "Role", "Subjects", "Risk"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, crb := range b.Inventory.ClusterRoleBindings {
			// Skip pure system-to-system noise: role AND every subject start with system:
			allSysSubjects := true
			for _, s := range crb.Subjects {
				if !strings.HasPrefix(s, "ServiceAccount:system:") &&
					!strings.HasPrefix(s, "User:system:") &&
					!strings.HasPrefix(s, "Group:system:") {
					allSysSubjects = false
					break
				}
			}
			if strings.HasPrefix(crb.RoleName, "system:") && allSysSubjects {
				continue
			}

			risk := roleRisk[crb.RoleName]
			if crb.RoleName == "cluster-admin" && risk == "" {
				risk = "CRITICAL"
			}
			if risk == "" {
				risk = "LOW"
			}

			riskColor := map[string]string{
				"CRITICAL": "var(--danger)", "HIGH": "var(--warning-high)", "MEDIUM": "var(--warning-medium)", "LOW": "var(--muted)",
			}[risk]
			riskCell := fmt.Sprintf(`<span style="color:%s;font-weight:700">%s</span>`, riskColor, risk)
			subjCell := e(strings.Join(crb.Subjects, ", "))
			if len(crb.Subjects) == 0 {
				subjCell = `<span style="color:var(--muted)">—</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td style="font-size:.82em">%s</td><td>%s</td></tr>`,
				e(crb.Name), e(crb.RoleName), subjCell, riskCell)
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // p5

	// ── Tab 6: Images ────────────────────────────────────────────────────────
	w(`<div class="pane" id="p6" role="tabpanel" aria-labelledby="tab-6"><h2>Container Images</h2>`)
	if len(b.Inventory.Images) == 0 {
		w(`<div class="empty">No image data collected (run against a live cluster with workloads).</div>`)
	} else {
		w(`<table id="t-images"><thead><tr>`)
		for _, h := range []string{"Image", "Registry", "Type", "Used By"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, img := range b.Inventory.Images {
			cls, lbl := "prv", "private"
			if img.IsPublic {
				cls, lbl = "pub", "PUBLIC"
			}
			wf(`<tr><td>%s</td><td>%s</td><td><span class="chip %s">%s</span></td><td>%s</td></tr>`,
				e(img.Image), e(img.Registry), cls, lbl, e(strings.Join(img.Workloads, ", ")))
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // p6

	// ── Tab 7: Backup ────────────────────────────────────────────────────────
	w(`<div class="pane" id="p7" role="tabpanel" aria-labelledby="tab-7">`)
	backupInv := b.Inventory.Backup

	// Detected tools card
	w(`<div class="card"><h2>Detected Backup Tools</h2>`)
	if len(backupInv.Tools) == 0 {
		w(`<div class="empty">No backup tools scanned.</div>`)
	} else {
		w(`<table id="t-bktools"><thead><tr>`)
		for _, h := range []string{"Tool", "Detected", "Namespace", "Version", "Policy Inspection", "Detail", "CRDs Found"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, t := range backupInv.Tools {
			detectedCell := `<span class="chip n">no</span>`
			if t.Detected {
				detectedCell = `<span class="chip p">yes</span>`
			}
			inspectionStatus := `<span class="chip n">not inspected</span>`
			inspectionDetail := `<span style="color:var(--muted)">—</span>`
			if t.Detected {
				switch t.PolicyInspectionStatus {
				case model.BackupCoverageStatusVerified:
					inspectionStatus = `<span class="chip p">verified</span>`
				case model.BackupCoverageStatusUnsupported:
					inspectionStatus = `<span class="chip w">unsupported</span>`
				case model.BackupCoverageStatusPermissionDenied:
					inspectionStatus = `<span class="chip w">permission denied</span>`
				case model.BackupCoverageStatusParseError, model.BackupCoverageStatusAPIError:
					inspectionStatus = `<span class="chip f">inspection failed</span>`
				}
				if t.PolicyInspectionDetail != "" {
					inspectionDetail = e(t.PolicyInspectionDetail)
				}
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(t.Name), detectedCell, e(t.Namespace), e(t.Version), inspectionStatus, inspectionDetail,
				e(strings.Join(t.CRDsFound, ", ")))
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // detected tools card

	// Backup policies card
	w(`<div class="card"><h2>Backup Policies / Schedules</h2>`)
	if backupInv.PrimaryTool == "none" || backupInv.PrimaryTool == "" {
		w(`<div class="empty">No backup tool detected — no policies to display.</div>`)
	} else if !backupInv.CoverageVerified {
		wf(`<div class="empty" style="color:var(--warning-medium)">%s detected, but policy coverage could not be verified (%s). %s</div>`,
			e(backupInv.PrimaryTool), e(backupCoverageStatusText(backupInv)), e(backupCoverageReasonText(backupInv)))
	} else if len(backupInv.Policies) == 0 {
		wf(`<div class="empty" style="color:var(--warning-high)">%s detected but no policies or schedules found. Create backup schedules to establish coverage.</div>`,
			e(backupInv.PrimaryTool))
	} else {
		offsiteCount := 0
		for _, p := range backupInv.Policies {
			if p.HasOffsite {
				offsiteCount++
			}
		}
		wf(`<p style="color:var(--muted);font-size:.84em;margin-bottom:8px">%d policies found &mdash; %d with offsite/export &mdash; coverage sources: %s</p>`,
			len(backupInv.Policies), offsiteCount, e(strings.Join(backupInv.CoverageSourceTools, ", ")))
		if len(backupInv.OffsiteMissingNS) > 0 {
			wf(`<p style="color:var(--warning-high);font-size:.84em;margin-bottom:8px">Offsite evidence is missing for covered namespaces: %s</p>`,
				e(strings.Join(backupInv.OffsiteMissingNS, ", ")))
		}
		w(`<table id="t-policies"><thead><tr>`)
		for _, h := range []string{"Tool", "Name", "Namespaces", "Schedule", "RPO (h)", "Last Success", "Evidence", "Offsite", "Retention"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, p := range backupInv.Policies {
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
			offsiteCell := `<span class="chip n">no</span>`
			if p.HasOffsite {
				offsiteCell = `<span class="chip p">yes</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(p.Tool), e(p.Name), e(nsCell), e(p.Schedule), rpoCell, e(lastSuccessCell), e(evidenceCell), offsiteCell, e(p.RetentionTTL))
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // policies card

	// Backup assurance card
	w(`<div class="card"><h2>Backup Assurance</h2>`)
	if backupInv.Assurance == nil {
		w(`<div class="empty">Backup assurance was not calculated.</div>`)
	} else {
		color := backupAssuranceColor(backupInv.Assurance)
		wf(`<p style="margin-bottom:8px"><strong style="color:%s">%s</strong> &nbsp; <span style="color:var(--muted)">confidence: %s</span></p>`,
			color, e(backupAssuranceConclusionText(backupInv.Assurance)), e(string(backupInv.Assurance.Confidence)))
		wf(`<p style="color:var(--muted);font-size:.85em;margin-bottom:10px">%s</p>`, e(backupInv.Assurance.Summary))
		if len(backupInv.Assurance.Signals) > 0 {
			w(`<table id="t-assurance"><thead><tr>`)
			for _, h := range []string{"Signal", "Status", "Confidence", "Summary", "Detail"} {
				wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
			}
			w(`</tr></thead><tbody>`)
			for _, signal := range backupInv.Assurance.Signals {
				statusColor := "var(--muted)"
				switch signal.Status {
				case "confirmed":
					statusColor = "var(--success)"
				case "warning", "unverified":
					statusColor = "var(--warning-medium)"
				case "missing":
					statusColor = "var(--danger)"
				}
				detail := "—"
				if signal.Detail != "" {
					detail = signal.Detail
				}
				wf(`<tr><td>%s</td><td style="color:%s">%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
					e(signal.ID), statusColor, e(signal.Status), e(string(signal.Confidence)), e(signal.Summary), e(detail))
			}
			w(`</tbody></table>`)
		}
	}
	w(`</div>`) // assurance card

	// Restore simulation card
	w(`<div class="card"><h2>Restore Simulation</h2>`)
	if sim := backupInv.RestoreSim; sim == nil {
		w(`<div class="empty">Restore simulation not available (dry-run mode or no cluster data).</div>`)
	} else if len(sim.Namespaces) == 0 {
		w(`<div class="empty" style="color:var(--success)">No stateful namespaces found — nothing to simulate.</div>`)
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
		coverageCountLabel := "Uncovered"
		coverageCount := len(sim.UncoveredNS)
		coverageCountColor := "var(--success)"
		if coverageCount > 0 {
			coverageCountColor = "var(--danger)"
		}
		if unknownCount > 0 {
			coverageVolumeText = "unknown"
			coverageCountLabel = "Unverified"
			coverageCount = unknownCount
			coverageCountColor = "var(--warning-medium)"
		}
		wf(`<div class="grid" style="margin-bottom:12px">
<div class="sbox"><div class="v">%d</div><div class="l">Namespaces</div></div>
<div class="sbox"><div class="v" style="color:%s">%d</div><div class="l">%s</div></div>
<div class="sbox"><div class="v">%.1f GB</div><div class="l">Total PVC Data</div></div>
<div class="sbox"><div class="v">%s</div><div class="l">Coverage by Volume</div></div>
</div>`,
			len(sim.Namespaces),
			coverageCountColor,
			coverageCount,
			coverageCountLabel,
			sim.TotalPVCsGB,
			coverageVolumeText)
		w(`<table id="t-sim"><thead><tr>`)
		for _, h := range []string{"Namespace", "Coverage", "RPO (h)", "PVC Data (GB)", "Blockers", "Warnings"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, ns := range sim.Namespaces {
			covCell := `<span class="chip f">none</span>`
			if !ns.CoverageKnown {
				covCell = `<span class="chip w">unverified</span>`
			} else if ns.HasCoverage {
				covCell = `<span class="chip p">covered</span>`
			}
			rpoCell := `<span style="color:var(--muted)">unknown</span>`
			if ns.RPOHours >= 0 {
				color := "var(--success)"
				if ns.RPOHours > 24 {
					color = "var(--warning-high)"
				}
				rpoCell = fmt.Sprintf(`<span style="color:%s">%d</span>`, color, ns.RPOHours)
			}
			sizeCell := fmt.Sprintf("%.1f", ns.PVCSizeGB)
			blockersCell := `<span style="color:var(--muted)">—</span>`
			if len(ns.Blockers) > 0 {
				blockersCell = fmt.Sprintf(`<span class="c-CRITICAL">%s</span>`, e(strings.Join(ns.Blockers, "; ")))
			}
			warningsCell := `<span style="color:var(--muted)">—</span>`
			if len(ns.Warnings) > 0 {
				warningsCell = fmt.Sprintf(`<span class="c-MEDIUM">%s</span>`, e(strings.Join(ns.Warnings, "; ")))
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(ns.Namespace), covCell, rpoCell, sizeCell, blockersCell, warningsCell)
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // restore sim card

	// ── Round 14: etcd backup status ─────────────────────────────────────
	w(`<div class="card"><h2>etcd Backup Status</h2>`)
	if eb := b.Inventory.EtcdBackup; eb == nil {
		w(`<div class="empty">etcd backup detection not run (dry-run mode or collector skipped).</div>`)
	} else if eb.Detected {
		sourceLabel := map[string]string{
			"provider-managed": "Provider-managed",
			"cronjob":          "CronJob",
			"configmap":        "ConfigMap",
			"velero-cluster":   "Velero cluster-scoped backup",
		}[eb.Source]
		if sourceLabel == "" {
			sourceLabel = eb.Source
		}
		wf(`<p><span class="ok">✓</span> <strong>etcd backup detected</strong> — source: <em>%s</em></p>`, e(sourceLabel))
		if eb.Detail != "" {
			wf(`<p style="color:var(--muted);font-size:.84em">%s</p>`, e(eb.Detail))
		}
	} else {
		w(`<p><span class="bad">✗</span> <strong style="color:var(--danger)">No etcd backup evidence found</strong></p>`)
		w(`<p style="color:var(--muted);font-size:.84em">etcd holds all cluster state. Without a backup, the cluster cannot be recovered after catastrophic failure. Configure periodic <code>etcdctl snapshot save</code> via a CronJob, or migrate to a managed K8s service.</p>`)
	}
	w(`</div>`) // etcd backup card

	w(`</div>`) // p7

	// ── Tab 8: DR Score ──────────────────────────────────────────────────────
	w(`<div class="pane" id="p8" role="tabpanel" aria-labelledby="tab-8"><h2>DR Score Breakdown</h2>`)
	w(`<table id="t-score"><thead><tr>`)
	for _, h := range []string{"Domain", "Score", "Max", "Weight"} {
		wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
	}
	w(`</tr></thead><tbody>`)
	for _, d := range []struct {
		n string
		s int
		w string
	}{
		{"Storage", b.Score.Storage.Final, domainWeightLabel("storage")},
		{"Workload", b.Score.Workload.Final, domainWeightLabel("workload")},
		{"Config", b.Score.Config.Final, domainWeightLabel("config")},
		{"Backup / Recovery", b.Score.Backup.Final, domainWeightLabel("backup")},
		{"Overall", b.Score.Overall.Final, "100%"},
	} {
		c := "var(--success)"
		if d.s < 50 {
			c = "var(--danger)"
		} else if d.s < 75 {
			c = "var(--warning-high)"
		}
		wf(`<tr><td>%s</td><td style="color:%s;font-weight:700">%d</td><td>100</td><td>%s</td></tr>`,
			e(d.n), c, d.s, e(d.w))
	}
	w(`</tbody></table>`)

	// Profile weights card
	p := profile.Normalize(b.Profile)
	pWeights := profile.Weights(p)
	type wRow struct{ label, key string }
	wRows := []wRow{
		{"Restore Testing", "restoreTesting"},
		{"Immutability", "immutability"},
		{"Replication / Offsite", "replication"},
		{"Security", "security"},
		{"Airgap Restrictions", "airgap"},
	}
	w(`<div class="card" style="margin-top:16px"><h2>Active Scoring Profile: `)
	wf(`<span style="color:var(--accent)">%s</span></h2>`, e(activeProfile))
	hasCustom := len(pWeights) > 0
	if !hasCustom {
		w(`<p style="color:var(--muted);font-size:.86em">Standard profile — all domain weights at baseline (1.0×). Use <code>--profile enterprise|dev|airgap</code> to adjust penalty emphasis.</p>`)
	} else {
		w(`<p style="color:var(--muted);font-size:.86em;margin-bottom:8px">Penalty multipliers applied to relevant scoring rules:</p>`)
		w(`<table style="width:auto"><thead><tr><th>Category</th><th>Multiplier</th><th>Effect</th></tr></thead><tbody>`)
		for _, r := range wRows {
			if mul, ok := pWeights[r.key]; ok {
				effect := "increased penalty"
				effectColor := "var(--warning-high)"
				if mul < 1.0 {
					effect = "reduced penalty"
					effectColor = "var(--success)"
				}
				wf(`<tr><td>%s</td><td style="color:var(--text);font-weight:700">%.2f×</td><td style="color:%s">%s</td></tr>`,
					e(r.label), mul, effectColor, effect)
			}
		}
		w(`</tbody></table>`)
	}
	w(`</div>`)

	w(`<h2 style="margin-top:20px">Findings</h2>`)
	if len(b.Inventory.Findings) == 0 {
		w(`<div class="empty">No findings.</div>`)
	} else {
		// Severity filter bar
		w(`<div class="filter-bar">
<span>Filter:</span>
<button class="fbtn active" data-sev="ALL" onclick="filterSev(this)">All</button>
<button class="fbtn fc" data-sev="CRITICAL" onclick="filterSev(this)">Critical</button>
<button class="fbtn fh" data-sev="HIGH" onclick="filterSev(this)">High</button>
<button class="fbtn fm" data-sev="MEDIUM" onclick="filterSev(this)">Medium</button>
<button class="fbtn" data-sev="LOW" onclick="filterSev(this)">Low</button>
<button class="fbtn" data-sev="INFO" onclick="filterSev(this)">Info</button>
</div>`)
		w(`<table id="t-findings"><thead><tr>`)
		for _, h := range []string{"Severity", "Resource", "Finding", "Recommendation"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody id="findings-tbody">`)
		for _, f := range b.Inventory.Findings {
			wf(`<tr data-sev="%s"><td class="sev-%s">%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(f.Severity), e(f.Severity), e(f.Severity), e(f.ResourceID), e(f.Message), e(f.Recommendation))
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // p8

	// ── Tab 9: Findings ───────────────────────────────────────────────────────
	w(`<div class="pane" id="p9" role="tabpanel" aria-labelledby="tab-9"><h2>Findings</h2>`)
	if len(b.Inventory.Findings) == 0 {
		w(`<div class="empty">No findings — cluster passed all checks.</div>`)
	} else {
		// Build remediation index: findingID → step index
		remIdx := map[string]int{}
		for i, step := range b.Inventory.RemediationSteps {
			if step.FindingID != "" {
				remIdx[step.FindingID] = i
			}
		}

		// Severity breakdown bar chart
		fCrit, fHigh, fMed, fLow := 0, 0, 0, 0
		for _, f := range b.Inventory.Findings {
			switch f.Severity {
			case "CRITICAL":
				fCrit++
			case "HIGH":
				fHigh++
			case "MEDIUM":
				fMed++
			default:
				fLow++
			}
		}
		fMax := fCrit
		if fHigh > fMax {
			fMax = fHigh
		}
		if fMed > fMax {
			fMax = fMed
		}
		if fLow > fMax {
			fMax = fLow
		}
		barW := func(n, total int) int {
			if total == 0 || n == 0 {
				return 0
			}
			v := n * 180 / total
			if v < 4 {
				v = 4
			}
			return v
		}
		w(`<div class="card" style="margin-bottom:14px">`)
		w(`<h2 style="margin-bottom:10px">Severity Breakdown</h2>`)
		w(`<svg viewBox="0 0 300 80" xmlns="http://www.w3.org/2000/svg" style="width:100%;max-width:340px;height:80px;display:block">`)
		for row, sev := range []struct {
			label, color string
			count        int
		}{
			{"CRITICAL", "var(--danger)", fCrit},
			{"HIGH", "var(--warning-high)", fHigh},
			{"MEDIUM", "var(--warning-medium)", fMed},
			{"LOW/INFO", "var(--muted)", fLow},
		} {
			y := float64(row)*18 + 8
			bw := barW(sev.count, fMax)
			wf(`<text x="0" y="%.1f" fill="%s" font-size="9" dominant-baseline="middle">%s</text>`, y+4, sev.color, sev.label)
			wf(`<rect x="70" y="%.1f" width="%d" height="10" rx="3" fill="%s" opacity="0.85"/>`, y-1, bw, sev.color)
			wf(`<text x="%d" y="%.1f" fill="%s" font-size="9" dominant-baseline="middle">%d</text>`, bw+74, y+4, sev.color, sev.count)
		}
		w(`</svg></div>`)

		// Filter bar
		w(`<div class="filter-bar">
<span>Filter:</span>
<button class="fbtn active" data-sev="ALL" onclick="filterSev2(this)">All</button>
<button class="fbtn fc" data-sev="CRITICAL" onclick="filterSev2(this)">Critical</button>
<button class="fbtn fh" data-sev="HIGH" onclick="filterSev2(this)">High</button>
<button class="fbtn fm" data-sev="MEDIUM" onclick="filterSev2(this)">Medium</button>
<button class="fbtn" data-sev="LOW" onclick="filterSev2(this)">Low/Info</button>
</div>`)

		w(`<table id="t-findings2"><thead><tr>`)
		for _, h := range []string{"Severity", "ID", "Resource", "Finding", "Recommendation", "Action"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody id="findings2-tbody">`)
		for _, f := range b.Inventory.Findings {
			actionCell := `<span style="color:var(--muted);font-size:.82em">—</span>`
			if remI, ok := remIdx[f.ID]; ok {
				actionCell = fmt.Sprintf(`<a href="#" onclick="showRemStep(%d);return false;" style="color:var(--accent-soft);font-size:.82em;white-space:nowrap">-> Remediation #%d</a>`, remI, remI+1)
			}
			wf(`<tr data-sev="%s"><td class="sev-%s">%s</td><td style="color:var(--muted);font-size:.82em">%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(f.Severity), e(f.Severity), e(f.Severity), e(f.ID), e(f.ResourceID), e(f.Message), e(f.Recommendation), actionCell)
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // p9

	// ── Tab 10: Remediation ───────────────────────────────────────────────────
	w(`<div class="pane" id="p10" role="tabpanel" aria-labelledby="tab-10"><h2>Remediation Plan</h2>`)
	if len(b.Inventory.RemediationSteps) == 0 {
		w(`<div class="empty">No remediation steps generated. Run with a live cluster to produce findings.</div>`)
	} else {
		w(`<div class="rem-controls">
<button class="btn-sm" onclick="remAll(true)">Expand All</button>
<button class="btn-sm" onclick="remAll(false)">Collapse All</button>
</div>`)
		curPri := -1
		priLabel := map[int]string{1: "Priority 1 — Must Fix Before DR", 2: "Priority 2 — Recommended", 3: "Priority 3 — Optional"}
		priClass := map[int]string{1: "c-CRITICAL", 2: "c-HIGH", 3: "c-LOW"}
		chipClass := map[int]string{1: "f", 2: "w", 3: "n"}
		for i, step := range b.Inventory.RemediationSteps {
			if step.Priority != curPri {
				curPri = step.Priority
				wf(`<h3 class="%s">%s</h3>`, priClass[step.Priority], e(priLabel[step.Priority]))
			}
			wf(`<div class="step"><button type="button" class="step-h" onclick="tog(%d)" aria-controls="sb%d" aria-expanded="false"><span class="chip %s">%s</span><span>%s</span></button>
<div class="step-b" id="sb%d">`,
				i, i, chipClass[step.Priority], e(step.Category), e(step.Title),
				i)
			w(remediationBodyHTML(step))
			w(`</div></div>`)
		}
	}
	w(`</div>`) // p10

	// ── Tab 11: Compare (only rendered when --compare was used) ──────────────
	if c := b.Comparison; c != nil {
		w(`<div class="pane" id="p11" role="tabpanel" aria-labelledby="tab-11">`)
		wf(`<h2>Comparison vs scan from %s</h2>`, e(c.PreviousScannedAt))

		// Score delta card
		deltaSign := ""
		deltaColor := "var(--muted)"
		if c.ScoreDelta > 0 {
			deltaSign = "+"
			deltaColor = "var(--success)"
		} else if c.ScoreDelta < 0 {
			deltaColor = "var(--danger)"
		}
		wf(`<div class="card">
<div class="grid">
<div class="sbox"><div class="v" style="color:%s">%s%d</div><div class="l">Score Delta</div></div>
<div class="sbox"><div class="v">%d</div><div class="l">Previous Score</div></div>
<div class="sbox"><div class="v">%d</div><div class="l">Current Score</div></div>
<div class="sbox"><div class="v" style="color:var(--muted);font-size:.8em">%s → %s</div><div class="l">Maturity Change</div></div>
</div></div>`,
			deltaColor, deltaSign, c.ScoreDelta,
			c.PreviousScore, b.Score.Overall.Final,
			e(c.PreviousMaturity), e(b.Score.Maturity))

		// Backup tool change
		if c.BackupToolChanged {
			wf(`<div class="card" style="border-color:var(--warning-medium)"><h2 style="color:var(--warning-medium)">Backup Tool Changed</h2>
<p style="color:var(--muted);font-size:.86em;margin-top:4px">%s → <strong style="color:var(--text)">%s</strong></p></div>`,
				e(c.BackupToolPrevious), e(c.BackupToolCurrent))
		}

		// Resource delta table
		type rowDef struct {
			label   string
			added   []string
			removed []string
		}
		rows := []rowDef{
			{"Namespaces", c.NamespacesAdded, c.NamespacesRemoved},
			{"Workloads", c.WorkloadsAdded, c.WorkloadsRemoved},
			{"PVCs", c.PVCsAdded, c.PVCsRemoved},
			{"Images", c.ImagesAdded, c.ImagesRemoved},
		}
		w(`<div class="card"><h2>Resource Changes</h2><table><thead><tr>`)
		for _, h := range []string{"Category", "Added", "Removed"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, row := range rows {
			addedCell := fmt.Sprintf(`<span class="ok">+%d</span>`, len(row.added))
			removedCell := fmt.Sprintf(`<span class="bad">-%d</span>`, len(row.removed))
			if len(row.added) == 0 {
				addedCell = `<span style="color:var(--muted)">—</span>`
			}
			if len(row.removed) == 0 {
				removedCell = `<span style="color:var(--muted)">—</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td></tr>`, e(row.label), addedCell, removedCell)
		}
		w(`</tbody></table></div>`)

		// New findings (regressions)
		if len(c.FindingsNew) > 0 {
			w(`<div class="card" style="border-color:var(--danger)"><h2 style="color:var(--danger)">New Findings (regressions)</h2>`)
			w(`<table><thead><tr>`)
			for _, h := range []string{"Severity", "Resource", "Message"} {
				wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
			}
			w(`</tr></thead><tbody>`)
			for _, f := range c.FindingsNew {
				wf(`<tr><td class="sev-%s">%s</td><td>%s</td><td>%s</td></tr>`,
					e(f.Severity), e(f.Severity), e(f.ResourceID), e(f.Message))
			}
			w(`</tbody></table></div>`)
		}

		// Resolved findings (improvements)
		if len(c.FindingsResolved) > 0 {
			w(`<div class="card" style="border-color:var(--success)"><h2 style="color:var(--success)">Resolved Findings (improvements)</h2>`)
			w(`<table><thead><tr>`)
			for _, h := range []string{"Severity", "Resource", "Message"} {
				wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
			}
			w(`</tr></thead><tbody>`)
			for _, f := range c.FindingsResolved {
				wf(`<tr><td class="sev-%s">%s</td><td>%s</td><td>%s</td></tr>`,
					e(f.Severity), e(f.Severity), e(f.ResourceID), e(f.Message))
			}
			w(`</tbody></table></div>`)
		}

		if len(c.FindingsNew) == 0 && len(c.FindingsResolved) == 0 {
			w(`<div class="card"><p class="ok">No finding changes between scans.</p></div>`)
		}

		w(`</div>`) // p10
	}

	// JS
	w(`</main><script>
function show(n){
  document.querySelectorAll('.tab').forEach(function(t,i){
    var active=i===n;
    t.classList.toggle('active',active);
    t.setAttribute('aria-selected',active?'true':'false');
    t.tabIndex=active?0:-1;
  });
  document.querySelectorAll('.pane').forEach(function(p,i){
    var active=i===n;
    p.classList.toggle('active',active);
    p.hidden=!active;
  });
}
function setStepState(n,open){
  var b=document.getElementById('sb'+n);
  if(!b)return;
  b.classList.toggle('open',open);
  var trigger=document.querySelector('.step-h[aria-controls="sb'+n+'"]');
  if(trigger)trigger.setAttribute('aria-expanded',open?'true':'false');
}
function tog(n){
  var b=document.getElementById('sb'+n);
  if(b)setStepState(n,!b.classList.contains('open'));
}
function remAll(open){
  document.querySelectorAll('.step-b').forEach(function(b){
    setStepState(b.id.replace('sb',''),open);
  });
}
function sortTbl(th){
  var tbl=th.closest('table'),tbody=tbl.querySelector('tbody');
  if(!tbody)return;
  var rows=Array.from(tbody.querySelectorAll('tr'));
  var idx=Array.from(th.parentNode.children).indexOf(th);
  var asc=th.dataset.asc!=='1';
  th.dataset.asc=asc?'1':'0';
  tbl.querySelectorAll('th').forEach(function(h){h.classList.remove('asc','desc');delete h.dataset.asc;});
  th.dataset.asc=asc?'1':'0';
  th.classList.add(asc?'asc':'desc');
  rows.sort(function(a,b){
    var av=a.cells[idx]?a.cells[idx].textContent.trim():'';
    var bv=b.cells[idx]?b.cells[idx].textContent.trim():'';
    var an=parseFloat(av),bn=parseFloat(bv);
    if(!isNaN(an)&&!isNaN(bn))return asc?an-bn:bn-an;
    return asc?av.localeCompare(bv):bv.localeCompare(av);
  });
  rows.forEach(function(r){tbody.appendChild(r);});
}
function filterSev(btn){
  var sev=btn.dataset.sev;
  document.querySelectorAll('.filter-bar .fbtn').forEach(function(b){b.classList.remove('active');});
  btn.classList.add('active');
  document.querySelectorAll('#findings-tbody tr').forEach(function(r){
    r.style.display=(sev==='ALL'||r.dataset.sev===sev)?'':'none';
  });
}
function filterSev2(btn){
  var sev=btn.dataset.sev;
  btn.closest('.filter-bar').querySelectorAll('.fbtn').forEach(function(b){b.classList.remove('active');});
  btn.classList.add('active');
  document.querySelectorAll('#findings2-tbody tr').forEach(function(r){
    var ms=sev==='LOW'?['LOW','INFO']:null;
    r.style.display=(sev==='ALL'||(ms?ms.indexOf(r.dataset.sev)>=0:r.dataset.sev===sev))?'':'none';
  });
}
function showTab(name){
  document.querySelectorAll('.tab').forEach(function(t,i){if(t.textContent.trim()===name)show(i);});
}
function showRemStep(idx){
  showTab('Remediation');
  var el=document.getElementById('sb'+idx);
  var behavior=window.matchMedia&&window.matchMedia('(prefers-reduced-motion: reduce)').matches?'auto':'smooth';
  if(el){
    setStepState(idx,true);
    el.scrollIntoView({behavior:behavior,block:'center'});
  }
}
show(0);
</script>`)
	w(reportDocumentEnd())
}
