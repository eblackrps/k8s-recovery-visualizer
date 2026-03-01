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
	scopeLabel := "all namespaces"
	if len(b.ScanNamespaces) > 0 {
		scopeLabel = strings.Join(b.ScanNamespaces, ", ")
	}
	activeProfile := b.Profile
	if activeProfile == "" {
		activeProfile = "standard"
	}

	w(`<!DOCTYPE html><html lang="en"><head>
<meta charset="utf-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>K8s DR Recovery Report</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#0d1117;color:#c9d1d9;font-family:system-ui,"Segoe UI",Arial,sans-serif;font-size:14px;line-height:1.5}
h2{color:#f0f6fc;font-size:1.05em;margin:16px 0 8px}
h3{color:#c9d1d9;font-size:.92em;margin:14px 0 6px}
.hdr{background:#161b22;border-bottom:1px solid #30363d;padding:14px 22px;display:flex;align-items:center;gap:16px}
.hdr h1{color:#f0f6fc;font-size:1.3em}
.hdr-meta{color:#8b949e;font-size:.82em;margin-top:3px}
.badge{display:inline-block;padding:3px 10px;border-radius:12px;font-weight:700;font-size:.85em;border:1px solid}
.tabs{display:flex;background:#161b22;border-bottom:1px solid #30363d;overflow-x:auto;padding:0 16px}
.tab{padding:9px 15px;cursor:pointer;color:#8b949e;border-bottom:2px solid transparent;font-size:.88em;user-select:none;white-space:nowrap}
.tab:hover{color:#c9d1d9}.tab.active{color:#58a6ff;border-bottom-color:#58a6ff}
.pane{display:none;padding:20px}.pane.active{display:block}
table{width:100%;border-collapse:collapse;margin-top:6px;font-size:.86em}
th{background:#161b22;color:#8b949e;text-align:left;padding:7px 9px;border-bottom:1px solid #30363d;white-space:nowrap;cursor:pointer;user-select:none}
th:hover{color:#c9d1d9}
th.asc::after{content:" \2191";color:#58a6ff}
th.desc::after{content:" \2193";color:#58a6ff}
td{padding:6px 9px;border-bottom:1px solid #21262d;vertical-align:top;word-break:break-word}
tr:hover td{background:#161b22}
.card{background:#161b22;border:1px solid #30363d;border-radius:6px;padding:14px;margin-bottom:14px}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:10px;margin:10px 0}
.sbox{background:#0d1117;border:1px solid #30363d;border-radius:6px;padding:12px;text-align:center}
.sbox .v{font-size:2em;font-weight:700;color:#f0f6fc}
.sbox .l{font-size:.78em;color:#8b949e;margin-top:2px}
.sbox .bar{background:#21262d;border-radius:3px;height:5px;margin-top:7px;overflow:hidden}
.sbox .fill{height:5px;border-radius:3px;background:#58a6ff}
.c-CRITICAL,.sev-CRITICAL{color:#f85149}
.c-HIGH,.sev-HIGH{color:#ffa657}
.c-MEDIUM,.sev-MEDIUM{color:#f2cc60}
.c-LOW,.c-INFO,.sev-LOW,.sev-INFO{color:#8b949e}
.ok{color:#7ee787}.bad{color:#f85149}
.chip{display:inline-block;padding:1px 7px;border-radius:10px;font-size:.78em;margin:1px}
.chip.p{background:#1f2d1f;color:#7ee787}
.chip.f{background:#3d1f1f;color:#f85149}
.chip.w{background:#3d2400;color:#f2cc60}
.chip.n{background:#21262d;color:#8b949e}
.pub{background:#3d1f1f;color:#f85149}.prv{background:#1f2d1f;color:#7ee787}
.step{border:1px solid #30363d;border-radius:5px;margin-bottom:10px;overflow:hidden}
.step-h{background:#161b22;padding:9px 13px;cursor:pointer;display:flex;align-items:center;gap:9px}
.step-h:hover{background:#21262d}
.step-b{padding:11px 13px;display:none;border-top:1px solid #21262d}
.step-b.open{display:block}
pre{background:#0d1117;border:1px solid #21262d;border-radius:4px;padding:9px;overflow-x:auto;font-size:.8em;color:#7ee787;margin-top:7px;white-space:pre-wrap}
.note{background:#1f2d1f;border-left:3px solid #7ee787;padding:7px 10px;margin-top:7px;font-size:.84em;border-radius:0 4px 4px 0}
.empty{color:#8b949e;font-style:italic;padding:10px 0}
.filter-bar{display:flex;gap:6px;margin-bottom:10px;flex-wrap:wrap;align-items:center}
.filter-bar span{color:#8b949e;font-size:.82em;margin-right:4px}
.fbtn{padding:3px 10px;border-radius:10px;font-size:.78em;cursor:pointer;border:1px solid #30363d;background:#161b22;color:#8b949e}
.fbtn:hover{border-color:#58a6ff;color:#58a6ff}
.fbtn.active{background:#1f3a5f;border-color:#58a6ff;color:#58a6ff}
.fbtn.fc{border-color:#f85149;color:#f85149}.fbtn.fc.active{background:#3d1f1f}
.fbtn.fh{border-color:#ffa657;color:#ffa657}.fbtn.fh.active{background:#3d2400}
.fbtn.fm{border-color:#f2cc60;color:#f2cc60}.fbtn.fm.active{background:#3d3000}
.rem-controls{display:flex;gap:8px;margin-bottom:12px}
.btn-sm{padding:4px 12px;border-radius:4px;font-size:.82em;cursor:pointer;border:1px solid #30363d;background:#161b22;color:#8b949e}
.btn-sm:hover{border-color:#58a6ff;color:#58a6ff}
</style></head><body>
`)

	// Header
	wf(`<div class="hdr"><div><h1>K8s DR Recovery Report</h1>
<div class="hdr-meta">Cluster: %s &nbsp;|&nbsp; Platform: %s &nbsp;|&nbsp; Scope: %s &nbsp;|&nbsp; Profile: <strong style="color:#c9d1d9">%s</strong> &nbsp;|&nbsp; %s</div></div>
<div style="margin-left:auto;text-align:right">
<div class="badge" style="color:%s;border-color:%s;font-size:1.1em">%s</div>
<div style="color:#8b949e;font-size:.83em;margin-top:3px">Score: <strong style="color:#f0f6fc">%d / 100</strong></div>
</div></div>`,
		e(b.Metadata.ClusterName), e(platform), e(scopeLabel), e(activeProfile), e(b.Metadata.GeneratedAt),
		matColor, matColor, e(b.Score.Maturity), b.Score.Overall.Final)

	// Tab bar — add Compare tab only when comparison data is present
	tabNames := []string{"Summary", "Nodes", "Workloads", "Storage", "Networking", "Config", "Images", "Backup", "DR Score", "Remediation"}
	if b.Comparison != nil {
		tabNames = append(tabNames, "Compare")
	}
	w(`<div class="tabs">`)
	for i, t := range tabNames {
		cls := "tab"
		if i == 0 {
			cls += " active"
		}
		wf(`<div class="%s" onclick="show(%d)">%s</div>`, cls, i, e(t))
	}
	w(`</div>`)

	// ── Tab 0: Summary ───────────────────────────────────────────────────────
	w(`<div class="pane active" id="p0">`)
	w(`<div class="grid">`)
	for _, d := range []struct {
		label, weight string
		score         int
	}{
		{"Storage", "35%", b.Score.Storage.Final},
		{"Workload", "20%", b.Score.Workload.Final},
		{"Config", "15%", b.Score.Config.Final},
		{"Backup / Recovery", "30%", b.Score.Backup.Final},
	} {
		wf(`<div class="sbox"><div class="v">%d</div><div class="l">%s <span style="color:#58a6ff">%s</span></div><div class="bar"><div class="fill" style="width:%d%%"></div></div></div>`,
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
<tr><td>Nodes</td><td>%d</td></tr>
<tr><td>Namespaces</td><td>%d</td></tr>
<tr><td>Helm Releases</td><td>%d</td></tr>
<tr><td>Certificates</td><td>%d</td></tr>
<tr><td>Recovery Target</td><td>%s</td></tr>
<tr><td>Namespace Scope</td><td>%s</td></tr>
</tbody></table></div>`,
		e(platform), e(b.Cluster.Platform.K8sVersion), e(b.Cluster.Platform.ClusterUID),
		btClass, e(backupTool),
		len(b.Inventory.Nodes), len(b.Inventory.Namespaces),
		len(b.Inventory.HelmReleases), len(b.Inventory.Certificates),
		e(b.Target), e(scopeLabel))

	crit, high, med := 0, 0, 0
	for _, f := range b.Inventory.Findings {
		switch f.Severity {
		case "CRITICAL":
			crit++
		case "HIGH":
			high++
		case "MEDIUM":
			med++
		}
	}
	wf(`<div class="card"><h2>Findings Summary</h2>
<span class="chip f">CRITICAL: %d</span>
<span class="chip w">HIGH: %d</span>
<span class="chip n">MEDIUM: %d</span>
<p style="margin-top:10px;color:#8b949e;font-size:.86em">Full findings → <strong>DR Score</strong> tab. Action steps → <strong>Remediation</strong> tab.</p>
</div>`, crit, high, med)

	// Scan coverage / skipped collectors callout
	totalCollectors := 24 // total number of optional collectors attempted
	skipped := len(b.CollectorSkips)
	rbacSkips := 0
	for _, sk := range b.CollectorSkips {
		if sk.RBAC {
			rbacSkips++
		}
	}
	if skipped > 0 {
		w(`<div class="card" style="border-color:#f2cc60">`)
		wf(`<h2 style="color:#f2cc60">Scan Coverage — %d/%d collectors skipped</h2>`, skipped, totalCollectors)
		if rbacSkips > 0 {
			wf(`<p style="color:#8b949e;font-size:.86em;margin-bottom:8px">%d skip(s) appear to be RBAC / permissions errors. Grant the service account read access to the listed resources to improve coverage.</p>`, rbacSkips)
		}
		w(`<table style="margin-top:4px"><thead><tr>`)
		for _, h := range []string{"Collector", "Reason", "RBAC?"} {
			wf(`<th>%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, sk := range b.CollectorSkips {
			rbacCell := `<span class="bad">✗ No</span>`
			if sk.RBAC {
				rbacCell = `<span style="color:#f2cc60">⚠ Yes</span>`
			}
			// Truncate long reasons for display
			reason := sk.Reason
			if len(reason) > 120 {
				reason = reason[:117] + "..."
			}
			wf(`<tr><td>%s</td><td style="color:#8b949e;font-size:.84em">%s</td><td>%s</td></tr>`,
				e(sk.Name), e(reason), rbacCell)
		}
		w(`</tbody></table></div>`)
	} else {
		wf(`<div class="card" style="border-color:#30363d"><h2>Scan Coverage</h2>
<p style="color:#7ee787;font-size:.86em">All %d collectors completed successfully — full inventory captured.</p></div>`, totalCollectors)
	}

	// ── Round 10c: Score Trend sparkline ──────────────────────────────────
	if len(b.TrendHistory) > 1 {
		n := len(b.TrendHistory)
		const svgW, svgH = 500, 90
		last := b.TrendHistory[n-1]
		prev := b.TrendHistory[n-2]
		trendColor := "#58a6ff"
		trendLabel := "STABLE"
		if last.Overall > prev.Overall {
			trendColor = "#7ee787"
			trendLabel = "IMPROVING"
		} else if last.Overall < prev.Overall {
			trendColor = "#f85149"
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
			wf(`<line x1="30" y1="%.1f" x2="%d" y2="%.1f" stroke="#21262d" stroke-dasharray="%s" stroke-width="1"/>`, ry, svgW-5, ry, dash)
			wf(`<text x="0" y="%.1f" fill="#8b949e" font-size="9" dominant-baseline="middle">%s</text>`, ry, ref.lbl)
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
		wf(`<div style="color:#8b949e;font-size:.83em;margin-top:4px">Last <strong style="color:#f0f6fc">%d</strong> scans &mdash; Current: <strong style="color:%s">%d</strong> (%s)</div>`,
			n, trendColor, last.Overall, e(last.Maturity))
		w(`</div>`)
	}

	w(`</div>`) // p0

	// ── Tab 1: Nodes ─────────────────────────────────────────────────────────
	w(`<div class="pane" id="p1"><h2>Nodes</h2>`)
	if len(b.Inventory.Nodes) == 0 {
		w(`<div class="empty">No node data collected.</div>`)
	} else {
		w(`<table id="t-nodes"><thead><tr>`)
		for _, h := range []string{"Name", "Roles", "Ready", "OS", "Kernel", "Runtime", "Kubelet", "Internal IP", "Taints"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, n := range b.Inventory.Nodes {
			rdStr := `<span class="bad">✗</span>`
			if n.Ready {
				rdStr = `<span class="ok">✓</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(n.Name), e(strings.Join(n.Roles, ",")), rdStr,
				e(n.OSImage), e(n.KernelVersion), e(n.ContainerRuntime),
				e(n.KubeletVersion), e(n.InternalIP), e(strings.Join(n.Taints, " ")))
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // p1

	// ── Tab 2: Workloads ─────────────────────────────────────────────────────
	w(`<div class="pane" id="p2"><h2>Workloads</h2>`)

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
		reqColor := "#7ee787"
		if noReq > 0 {
			reqColor = "#ffa657"
		}
		limColor := "#7ee787"
		if noLim > 0 {
			limColor = "#f2cc60"
		}
		privColor := "#7ee787"
		if priv > 0 {
			privColor = "#f85149"
		}
		hostColor := "#7ee787"
		if hostNS > 0 {
			hostColor = "#ffa657"
		}
		wf(`<div class="card"><h2>Resource Governance &amp; Pod Security</h2>
<p style="color:#8b949e;font-size:.84em;margin-bottom:10px">kube-system pods are excluded from governance checks.</p>
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
				privCell := `<span style="color:#8b949e">—</span>`
				if pod.Privileged {
					privCell = `<span class="bad">yes</span>`
				}
				hnCell := `<span style="color:#8b949e">—</span>`
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
		done := `<span style="color:#8b949e">active</span>`
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
	w(`<div class="pane" id="p3"><h2>Storage</h2>`)
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
			w(`<div class="empty" style="color:#ffa657">No VolumeSnapshotClasses found — CSI snapshot infrastructure not configured.</div>`)
		} else {
			// Snapshot class table
			w(`<table id="t-vsc" style="margin-bottom:12px"><thead><tr>`)
			for _, h := range []string{"VolumeSnapshotClass", "Driver", "Deletion Policy"} {
				wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
			}
			w(`</tr></thead><tbody>`)
			for _, vsc := range b.Inventory.VolumeSnapshotClasses {
				dpColor := "#7ee787"
				if vsc.DeletionPolicy == "Delete" {
					dpColor = "#ffa657"
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
	w(`<div class="pane" id="p4"><h2>Networking</h2>`)

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

		npColor := "#7ee787"
		if coveredNS < totalNS/2 {
			npColor = "#f85149"
		} else if coveredNS < totalNS {
			npColor = "#ffa657"
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
					wf(`<tr><td>%s</td><td>%s</td><td colspan="2"><span style="color:#8b949e">no rules</span></td><td>%s</td><td>%s</td></tr>`,
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
				ingCell := `<span style="color:#8b949e">—</span>`
				if nsHasIngress[ns.Name] {
					ingCell = `<span class="c-HIGH">exposed</span>`
				}
				lbCell := `<span style="color:#8b949e">—</span>`
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
	w(`<div class="pane" id="p5"><h2>Config</h2>`)
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
			sc := "#8b949e"
			if hr.Status == "deployed" {
				sc = "#7ee787"
			} else if hr.Status == "failed" {
				sc = "#f85149"
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
			dc := "#7ee787"
			if c.DaysToExpiry < 30 {
				dc = "#f85149"
			} else if c.DaysToExpiry < 60 {
				dc = "#ffa657"
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
			wcCell := `<span style="color:#8b949e">—</span>`
			if cr.HasWildcardVerb {
				wcCell = `<span class="bad">yes</span>`
			}
			saCell := `<span style="color:#8b949e">—</span>`
			if cr.HasSecretAccess {
				saCell = `<span class="c-HIGH">yes</span>`
			}
			esCell := `<span style="color:#8b949e">—</span>`
			if cr.HasEscalatePriv {
				esCell = `<span class="c-HIGH">yes</span>`
			}
			riskLabel, riskColor := "clean", "#7ee787"
			if cr.HasWildcardVerb {
				riskLabel, riskColor = "CRITICAL", "#f85149"
			} else if cr.HasSecretAccess || cr.HasEscalatePriv {
				riskLabel, riskColor = "HIGH", "#ffa657"
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
		w(`<p style="color:#8b949e;font-size:.84em;margin-bottom:8px">Shows who holds cluster-level permissions. System-to-system bindings (both role and all subjects prefixed with <code>system:</code>) are hidden for clarity.</p>`)
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
				"CRITICAL": "#f85149", "HIGH": "#ffa657", "MEDIUM": "#f2cc60", "LOW": "#8b949e",
			}[risk]
			riskCell := fmt.Sprintf(`<span style="color:%s;font-weight:700">%s</span>`, riskColor, risk)
			subjCell := e(strings.Join(crb.Subjects, ", "))
			if len(crb.Subjects) == 0 {
				subjCell = `<span style="color:#8b949e">—</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td style="font-size:.82em">%s</td><td>%s</td></tr>`,
				e(crb.Name), e(crb.RoleName), subjCell, riskCell)
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // p5

	// ── Tab 6: Images ────────────────────────────────────────────────────────
	w(`<div class="pane" id="p6"><h2>Container Images</h2>`)
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
	w(`<div class="pane" id="p7">`)
	backupInv := b.Inventory.Backup

	// Detected tools card
	w(`<div class="card"><h2>Detected Backup Tools</h2>`)
	if len(backupInv.Tools) == 0 {
		w(`<div class="empty">No backup tools scanned.</div>`)
	} else {
		w(`<table id="t-bktools"><thead><tr>`)
		for _, h := range []string{"Tool", "Detected", "Namespace", "Version", "CRDs Found"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, t := range backupInv.Tools {
			detectedCell := `<span class="chip n">no</span>`
			if t.Detected {
				detectedCell = `<span class="chip p">yes</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(t.Name), detectedCell, e(t.Namespace), e(t.Version),
				e(strings.Join(t.CRDsFound, ", ")))
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // detected tools card

	// Backup policies card
	w(`<div class="card"><h2>Backup Policies / Schedules</h2>`)
	if backupInv.PrimaryTool == "none" || backupInv.PrimaryTool == "" {
		w(`<div class="empty">No backup tool detected — no policies to display.</div>`)
	} else if len(backupInv.Policies) == 0 {
		wf(`<div class="empty" style="color:#ffa657">%s detected but no policies or schedules found. Create backup schedules to establish coverage.</div>`,
			e(backupInv.PrimaryTool))
	} else {
		offsiteCount := 0
		for _, p := range backupInv.Policies {
			if p.HasOffsite {
				offsiteCount++
			}
		}
		wf(`<p style="color:#8b949e;font-size:.84em;margin-bottom:8px">%d policies found &mdash; %d with offsite/export</p>`,
			len(backupInv.Policies), offsiteCount)
		w(`<table id="t-policies"><thead><tr>`)
		for _, h := range []string{"Tool", "Name", "Namespaces", "Schedule", "RPO (h)", "Offsite", "Retention"} {
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
			offsiteCell := `<span class="chip n">no</span>`
			if p.HasOffsite {
				offsiteCell = `<span class="chip p">yes</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(p.Tool), e(p.Name), e(nsCell), e(p.Schedule), rpoCell, offsiteCell, e(p.RetentionTTL))
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // policies card

	// Restore simulation card
	w(`<div class="card"><h2>Restore Simulation</h2>`)
	if sim := backupInv.RestoreSim; sim == nil {
		w(`<div class="empty">Restore simulation not available (dry-run mode or no cluster data).</div>`)
	} else if len(sim.Namespaces) == 0 {
		w(`<div class="empty" style="color:#7ee787">No stateful namespaces found — nothing to simulate.</div>`)
	} else {
		covPct := 0.0
		if sim.TotalPVCsGB > 0 {
			covPct = sim.CoveredPVCsGB / sim.TotalPVCsGB * 100
		}
		wf(`<div class="grid" style="margin-bottom:12px">
<div class="sbox"><div class="v">%d</div><div class="l">Namespaces</div></div>
<div class="sbox"><div class="v" style="color:%s">%d</div><div class="l">Uncovered</div></div>
<div class="sbox"><div class="v">%.1f GB</div><div class="l">Total PVC Data</div></div>
<div class="sbox"><div class="v">%.0f%%</div><div class="l">Coverage by Volume</div></div>
</div>`,
			len(sim.Namespaces),
			func() string {
				if len(sim.UncoveredNS) > 0 {
					return "#f85149"
				}
				return "#7ee787"
			}(),
			len(sim.UncoveredNS),
			sim.TotalPVCsGB,
			covPct)
		w(`<table id="t-sim"><thead><tr>`)
		for _, h := range []string{"Namespace", "Coverage", "RPO (h)", "PVC Data (GB)", "Blockers", "Warnings"} {
			wf(`<th onclick="sortTbl(this)">%s</th>`, e(h))
		}
		w(`</tr></thead><tbody>`)
		for _, ns := range sim.Namespaces {
			covCell := `<span class="chip f">none</span>`
			if ns.HasCoverage {
				covCell = `<span class="chip p">covered</span>`
			}
			rpoCell := `<span style="color:#8b949e">unknown</span>`
			if ns.RPOHours >= 0 {
				color := "#7ee787"
				if ns.RPOHours > 24 {
					color = "#ffa657"
				}
				rpoCell = fmt.Sprintf(`<span style="color:%s">%d</span>`, color, ns.RPOHours)
			}
			sizeCell := fmt.Sprintf("%.1f", ns.PVCSizeGB)
			blockersCell := `<span style="color:#8b949e">—</span>`
			if len(ns.Blockers) > 0 {
				blockersCell = fmt.Sprintf(`<span class="c-CRITICAL">%s</span>`, e(strings.Join(ns.Blockers, "; ")))
			}
			warningsCell := `<span style="color:#8b949e">—</span>`
			if len(ns.Warnings) > 0 {
				warningsCell = fmt.Sprintf(`<span class="c-MEDIUM">%s</span>`, e(strings.Join(ns.Warnings, "; ")))
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				e(ns.Namespace), covCell, rpoCell, sizeCell, blockersCell, warningsCell)
		}
		w(`</tbody></table>`)
	}
	w(`</div>`) // restore sim card
	w(`</div>`) // p7

	// ── Tab 8: DR Score ──────────────────────────────────────────────────────
	w(`<div class="pane" id="p8"><h2>DR Score Breakdown</h2>`)
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
		{"Storage", b.Score.Storage.Final, "35%"},
		{"Workload", b.Score.Workload.Final, "20%"},
		{"Config", b.Score.Config.Final, "15%"},
		{"Backup / Recovery", b.Score.Backup.Final, "30%"},
		{"Overall", b.Score.Overall.Final, "100%"},
	} {
		c := "#7ee787"
		if d.s < 50 {
			c = "#f85149"
		} else if d.s < 75 {
			c = "#ffa657"
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
	wf(`<span style="color:#58a6ff">%s</span></h2>`, e(activeProfile))
	hasCustom := len(pWeights) > 0
	if !hasCustom {
		w(`<p style="color:#8b949e;font-size:.86em">Standard profile — all domain weights at baseline (1.0×). Use <code>--profile enterprise|dev|airgap</code> to adjust penalty emphasis.</p>`)
	} else {
		w(`<p style="color:#8b949e;font-size:.86em;margin-bottom:8px">Penalty multipliers applied to relevant scoring rules:</p>`)
		w(`<table style="width:auto"><thead><tr><th>Category</th><th>Multiplier</th><th>Effect</th></tr></thead><tbody>`)
		for _, r := range wRows {
			if mul, ok := pWeights[r.key]; ok {
				effect := "increased penalty"
				effectColor := "#ffa657"
				if mul < 1.0 {
					effect = "reduced penalty"
					effectColor = "#7ee787"
				}
				wf(`<tr><td>%s</td><td style="color:#f0f6fc;font-weight:700">%.2f×</td><td style="color:%s">%s</td></tr>`,
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

	// ── Tab 9: Remediation ───────────────────────────────────────────────────
	w(`<div class="pane" id="p9"><h2>Remediation Plan</h2>`)
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
			wf(`<div class="step"><div class="step-h" onclick="tog(%d)"><span class="chip %s">%s</span><span>%s</span></div>
<div class="step-b" id="sb%d"><p>%s</p>`,
				i, chipClass[step.Priority], e(step.Category), e(step.Title),
				i, e(step.Detail))
			if step.TargetNotes != "" {
				wf(`<div class="note">%s</div>`, e(step.TargetNotes))
			}
			if len(step.Commands) > 0 {
				wf(`<pre>%s</pre>`, e(strings.Join(step.Commands, "\n")))
			}
			w(`</div></div>`)
		}
	}
	w(`</div>`) // p9

	// ── Tab 10: Compare (only rendered when --compare was used) ──────────────
	if c := b.Comparison; c != nil {
		w(`<div class="pane" id="p10">`)
		wf(`<h2>Comparison vs scan from %s</h2>`, e(c.PreviousScannedAt))

		// Score delta card
		deltaSign := ""
		deltaColor := "#8b949e"
		if c.ScoreDelta > 0 {
			deltaSign = "+"
			deltaColor = "#7ee787"
		} else if c.ScoreDelta < 0 {
			deltaColor = "#f85149"
		}
		wf(`<div class="card">
<div class="grid">
<div class="sbox"><div class="v" style="color:%s">%s%d</div><div class="l">Score Delta</div></div>
<div class="sbox"><div class="v">%d</div><div class="l">Previous Score</div></div>
<div class="sbox"><div class="v">%d</div><div class="l">Current Score</div></div>
<div class="sbox"><div class="v" style="color:#8b949e;font-size:.8em">%s → %s</div><div class="l">Maturity Change</div></div>
</div></div>`,
			deltaColor, deltaSign, c.ScoreDelta,
			c.PreviousScore, b.Score.Overall.Final,
			e(c.PreviousMaturity), e(b.Score.Maturity))

		// Backup tool change
		if c.BackupToolChanged {
			wf(`<div class="card" style="border-color:#f2cc60"><h2 style="color:#f2cc60">Backup Tool Changed</h2>
<p style="color:#8b949e;font-size:.86em;margin-top:4px">%s → <strong style="color:#f0f6fc">%s</strong></p></div>`,
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
				addedCell = `<span style="color:#8b949e">—</span>`
			}
			if len(row.removed) == 0 {
				removedCell = `<span style="color:#8b949e">—</span>`
			}
			wf(`<tr><td>%s</td><td>%s</td><td>%s</td></tr>`, e(row.label), addedCell, removedCell)
		}
		w(`</tbody></table></div>`)

		// New findings (regressions)
		if len(c.FindingsNew) > 0 {
			w(`<div class="card" style="border-color:#f85149"><h2 style="color:#f85149">New Findings (regressions)</h2>`)
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
			w(`<div class="card" style="border-color:#7ee787"><h2 style="color:#7ee787">Resolved Findings (improvements)</h2>`)
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
	w(`<script>
function show(n){
  document.querySelectorAll('.tab').forEach(function(t,i){t.classList.toggle('active',i===n)});
  document.querySelectorAll('.pane').forEach(function(p,i){p.classList.toggle('active',i===n)});
}
function tog(n){var b=document.getElementById('sb'+n);if(b)b.classList.toggle('open');}
function remAll(open){document.querySelectorAll('.step-b').forEach(function(b){b.classList.toggle('open',open)});}
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
</script></body></html>`)
}
