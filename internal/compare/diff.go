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

	// Findings delta
	prevSet := findingSet(prev.Inventory.Findings)
	currSet := findingSet(curr.Inventory.Findings)
	for k, f := range currSet {
		if _, ok := prevSet[k]; !ok {
			r.FindingsNew = append(r.FindingsNew, f)
		}
	}
	for k, f := range prevSet {
		if _, ok := currSet[k]; !ok {
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
		m[f.ID+"|"+f.ResourceID] = f
	}
	return m
}
