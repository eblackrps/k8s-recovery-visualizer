package compare

import (
	"fmt"

	"k8s-recovery-visualizer/internal/model"
)

// Diff compares prev against curr and returns a ComparisonSummary for the bundle.
func Diff(prev, curr *model.Bundle) model.ComparisonSummary {
	r := model.ComparisonSummary{
		PreviousScanID:    prev.Scan.ScanID,
		PreviousScannedAt: prev.Metadata.GeneratedAt,
		PreviousScore:     prev.Score.Overall.Final,
		PreviousMaturity:  prev.Score.Maturity,
		CurrentScore:      curr.Score.Overall.Final,
		CurrentMaturity:   curr.Score.Maturity,
		ScoreDelta:        curr.Score.Overall.Final - prev.Score.Overall.Final,
	}

	// Namespaces
	ns := setDelta(
		keys(prev.Inventory.Namespaces, func(n model.Namespace) string { return n.Name }),
		keys(curr.Inventory.Namespaces, func(n model.Namespace) string { return n.Name }),
	)
	r.NamespacesAdded, r.NamespacesRemoved = ns.added, ns.removed

	// Workloads
	wl := setDelta(workloadKeys(prev), workloadKeys(curr))
	r.WorkloadsAdded, r.WorkloadsRemoved = wl.added, wl.removed

	// PVCs
	pv := setDelta(
		keys(prev.Inventory.PVCs, func(p model.PersistentVolumeClaim) string {
			return fmt.Sprintf("%s/%s", p.Namespace, p.Name)
		}),
		keys(curr.Inventory.PVCs, func(p model.PersistentVolumeClaim) string {
			return fmt.Sprintf("%s/%s", p.Namespace, p.Name)
		}),
	)
	r.PVCsAdded, r.PVCsRemoved = pv.added, pv.removed

	// Images
	img := setDelta(
		keys(prev.Inventory.Images, func(i model.ContainerImage) string { return i.Image }),
		keys(curr.Inventory.Images, func(i model.ContainerImage) string { return i.Image }),
	)
	r.ImagesAdded, r.ImagesRemoved = img.added, img.removed

	// Backup tool
	prevTool := prev.Inventory.Backup.PrimaryTool
	currTool := curr.Inventory.Backup.PrimaryTool
	if prevTool == "" {
		prevTool = "none"
	}
	if currTool == "" {
		currTool = "none"
	}
	r.BackupToolPrevious = prevTool
	r.BackupToolCurrent = currTool
	r.BackupToolChanged = prevTool != currTool
	r.DomainDeltas = []model.ScoreDeltaSummary{
		scoreDelta("overall", prev.Score.Overall.Final, curr.Score.Overall.Final),
		scoreDelta("storage", prev.Score.Storage.Final, curr.Score.Storage.Final),
		scoreDelta("workload", prev.Score.Workload.Final, curr.Score.Workload.Final),
		scoreDelta("config", prev.Score.Config.Final, curr.Score.Config.Final),
		scoreDelta("backup", prev.Score.Backup.Final, curr.Score.Backup.Final),
	}
	r.SeverityDeltas = buildSeverityDeltas(prev.Inventory.Findings, curr.Inventory.Findings)
	r.InventoryDeltas = []model.InventoryDeltaSummary{
		{Name: "namespaces", Added: len(r.NamespacesAdded), Removed: len(r.NamespacesRemoved)},
		{Name: "workloads", Added: len(r.WorkloadsAdded), Removed: len(r.WorkloadsRemoved)},
		{Name: "pvcs", Added: len(r.PVCsAdded), Removed: len(r.PVCsRemoved)},
		{Name: "images", Added: len(r.ImagesAdded), Removed: len(r.ImagesRemoved)},
	}

	// Findings delta
	prevSet := findingSet(prev.Inventory.Findings)
	currSet := findingSet(curr.Inventory.Findings)
	for _, f := range curr.Inventory.Findings {
		prevFinding, ok := prevSet[findingKey(f)]
		if !ok {
			r.FindingsNew = append(r.FindingsNew, f)
			continue
		}
		r.PersistentFinding++
		if severityRank(f.Severity) > severityRank(prevFinding.Severity) {
			r.FindingsRegressed = append(r.FindingsRegressed, buildFindingChange(prevFinding, f, "severity-up"))
		} else if severityRank(f.Severity) < severityRank(prevFinding.Severity) {
			r.FindingsImproved = append(r.FindingsImproved, buildFindingChange(prevFinding, f, "severity-down"))
		}
	}
	for _, f := range prev.Inventory.Findings {
		if _, ok := currSet[findingKey(f)]; !ok {
			r.FindingsResolved = append(r.FindingsResolved, f)
		}
	}

	return r
}

// ── helpers ──────────────────────────────────────────────────────────────────

type delta struct {
	added   []string
	removed []string
}

func keys[T any](items []T, key func(T) string) map[string]struct{} {
	m := make(map[string]struct{}, len(items))
	for _, item := range items {
		m[key(item)] = struct{}{}
	}
	return m
}

func setDelta(prev, curr map[string]struct{}) delta {
	var d delta
	for k := range curr {
		if _, ok := prev[k]; !ok {
			d.added = append(d.added, k)
		}
	}
	for k := range prev {
		if _, ok := curr[k]; !ok {
			d.removed = append(d.removed, k)
		}
	}
	return d
}

func workloadKeys(b *model.Bundle) map[string]struct{} {
	m := make(map[string]struct{})
	for _, w := range b.Inventory.Deployments {
		m[fmt.Sprintf("%s/%s (Deployment)", w.Namespace, w.Name)] = struct{}{}
	}
	for _, w := range b.Inventory.StatefulSets {
		m[fmt.Sprintf("%s/%s (StatefulSet)", w.Namespace, w.Name)] = struct{}{}
	}
	for _, w := range b.Inventory.DaemonSets {
		m[fmt.Sprintf("%s/%s (DaemonSet)", w.Namespace, w.Name)] = struct{}{}
	}
	for _, w := range b.Inventory.Jobs {
		m[fmt.Sprintf("%s/%s (Job)", w.Namespace, w.Name)] = struct{}{}
	}
	for _, w := range b.Inventory.CronJobs {
		m[fmt.Sprintf("%s/%s (CronJob)", w.Namespace, w.Name)] = struct{}{}
	}
	return m
}

func findingSet(findings []model.Finding) map[string]model.Finding {
	m := make(map[string]model.Finding, len(findings))
	for _, f := range findings {
		m[findingKey(f)] = f
	}
	return m
}

func findingKey(f model.Finding) string {
	return f.ID + "|" + f.ResourceID
}

func scoreDelta(name string, prev, curr int) model.ScoreDeltaSummary {
	return model.ScoreDeltaSummary{
		Name:     name,
		Previous: prev,
		Current:  curr,
		Delta:    curr - prev,
	}
}

func severityRank(severity string) int {
	switch severity {
	case "CRITICAL":
		return 4
	case "HIGH":
		return 3
	case "MEDIUM":
		return 2
	case "LOW":
		return 1
	case "INFO":
		return 0
	default:
		return -1
	}
}

func buildSeverityDeltas(prev, curr []model.Finding) []model.SeverityDeltaSummary {
	severities := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "INFO"}
	prevCounts := severityCounts(prev)
	currCounts := severityCounts(curr)
	out := make([]model.SeverityDeltaSummary, 0, len(severities))
	for _, severity := range severities {
		out = append(out, model.SeverityDeltaSummary{
			Severity: severity,
			Previous: prevCounts[severity],
			Current:  currCounts[severity],
			Delta:    currCounts[severity] - prevCounts[severity],
		})
	}
	return out
}

func severityCounts(findings []model.Finding) map[string]int {
	out := map[string]int{}
	for _, finding := range findings {
		out[finding.Severity]++
	}
	return out
}

func buildFindingChange(prev, curr model.Finding, change string) model.FindingChange {
	return model.FindingChange{
		ID:               curr.ID,
		Title:            curr.Title,
		ResourceID:       curr.ResourceID,
		Message:          curr.Message,
		Recommendation:   curr.Recommendation,
		PreviousSeverity: prev.Severity,
		CurrentSeverity:  curr.Severity,
		Change:           change,
		OwnerHint:        curr.OwnerHint,
		Impact:           curr.Impact,
		Effort:           curr.Effort,
	}
}
