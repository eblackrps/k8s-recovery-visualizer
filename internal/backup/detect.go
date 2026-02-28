package backup

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

type toolSpec struct {
	Name            string
	Namespaces      []string
	CRDGroupParts   []string // substrings to match against CRD group names
	PodLabelKey     string
	PodLabelValue   string
}

var knownTools = []toolSpec{
	{
		Name:          "kasten",
		Namespaces:    []string{"kasten-io"},
		CRDGroupParts: []string{"kio.kasten.io", "config.kio.kasten.io"},
		PodLabelKey:   "app",
		PodLabelValue: "k10",
	},
	{
		Name:          "velero",
		Namespaces:    []string{"velero"},
		CRDGroupParts: []string{"velero.io"},
		PodLabelKey:   "app.kubernetes.io/name",
		PodLabelValue: "velero",
	},
	{
		Name:          "rubrik",
		Namespaces:    []string{"rubrik", "rbs"},
		CRDGroupParts: []string{"rubrik.com"},
		PodLabelKey:   "app",
		PodLabelValue: "rubrik-backup-service",
	},
	{
		Name:          "longhorn",
		Namespaces:    []string{"longhorn-system"},
		CRDGroupParts: []string{"longhorn.io"},
		PodLabelKey:   "app",
		PodLabelValue: "longhorn-manager",
	},
	{
		Name:          "trilio",
		Namespaces:    []string{"trilio-system"},
		CRDGroupParts: []string{"triliovault.trilio.io"},
		PodLabelKey:   "app",
		PodLabelValue: "trilio",
	},
	{
		Name:          "stash",
		Namespaces:    []string{"stash"},
		CRDGroupParts: []string{"stash.appscode.com"},
		PodLabelKey:   "app",
		PodLabelValue: "stash",
	},
	{
		Name:          "cloudcasa",
		Namespaces:    []string{"cloudcasa-io"},
		CRDGroupParts: []string{"cloudcasa.io"},
		PodLabelKey:   "app",
		PodLabelValue: "cloudcasa",
	},
}

// Detect scans the cluster for known backup tools and populates b.Inventory.Backup.
func Detect(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) {
	// Build quick lookup sets from already-collected data
	nsSet := map[string]struct{}{}
	for _, ns := range b.Inventory.Namespaces {
		nsSet[ns.Name] = struct{}{}
	}
	crdGroups := map[string]struct{}{}
	for _, crd := range b.Inventory.CRDs {
		crdGroups[crd.Group] = struct{}{}
	}

	inv := model.BackupInventory{
		PrimaryTool: "none",
		Tools:       []model.BackupDetectedTool{},
	}

	for _, spec := range knownTools {
		tool := model.BackupDetectedTool{
			Name:     spec.Name,
			Detected: false,
		}

		// Check namespace presence
		foundNS := ""
		for _, ns := range spec.Namespaces {
			if _, ok := nsSet[ns]; ok {
				foundNS = ns
				tool.Detected = true
				tool.Namespace = ns
				break
			}
		}

		// Check CRD presence
		for group := range crdGroups {
			for _, part := range spec.CRDGroupParts {
				if strings.Contains(group, part) {
					tool.Detected = true
					tool.CRDsFound = append(tool.CRDsFound, group)
				}
			}
		}

		// If namespace found, check for pods to confirm and get version
		if foundNS != "" && spec.PodLabelKey != "" {
			selector := spec.PodLabelKey + "=" + spec.PodLabelValue
			pods, err := cs.CoreV1().Pods(foundNS).List(ctx, metav1.ListOptions{
				LabelSelector: selector,
				Limit:         1,
			})
			if err == nil && len(pods.Items) > 0 {
				pod := pods.Items[0]
				// Try to get version from common label patterns
				if v := pod.Labels["app.kubernetes.io/version"]; v != "" {
					tool.Version = v
				} else if v := pod.Labels["helm.sh/chart"]; v != "" {
					tool.Version = v
				}
			}
		}

		inv.Tools = append(inv.Tools, tool)

		if tool.Detected && inv.PrimaryTool == "none" {
			inv.PrimaryTool = spec.Name
		}
	}

	// Determine which namespaces with StatefulSets are not covered.
	// For now, coverage detection only works for Kasten and Velero
	// (where we can inspect backup policies). For others, we mark all as covered.
	if inv.PrimaryTool != "none" {
		inv.CoveredNamespaces = coveredNamespaces(ctx, cs, inv.PrimaryTool, b)
		inv.UncoveredStatefulNS = uncoveredStatefulNamespaces(b, inv.CoveredNamespaces)
	}

	b.Inventory.Backup = inv
}

// coveredNamespaces attempts to determine which namespaces have backup policies.
// Returns a best-effort list.
func coveredNamespaces(ctx context.Context, cs *kubernetes.Clientset, tool string, b *model.Bundle) []string {
	switch tool {
	case "velero":
		return veleroScheduledNamespaces(ctx, cs)
	default:
		// For other tools, assume all namespaces are covered (conservative)
		var ns []string
		for _, n := range b.Inventory.Namespaces {
			ns = append(ns, n.Name)
		}
		return ns
	}
}

// veleroScheduledNamespaces reads Velero Schedule CRs to find which namespaces are included.
func veleroScheduledNamespaces(ctx context.Context, cs *kubernetes.Clientset) []string {
	raw, err := cs.RESTClient().
		Get().
		AbsPath("/apis/velero.io/v1/schedules").
		DoRaw(ctx)
	if err != nil {
		return nil
	}
	// Parse included namespaces from schedules
	// Simple JSON extraction — no external deps
	content := string(raw)
	if strings.Contains(content, `"includedNamespaces"`) {
		// Return marker indicating schedules exist but namespace list parsing skipped
		return []string{"*"}
	}
	return nil
}

// uncoveredStatefulNamespaces finds namespaces with StatefulSets not in the covered list.
func uncoveredStatefulNamespaces(b *model.Bundle, covered []string) []string {
	if len(covered) == 1 && covered[0] == "*" {
		return nil // wildcard coverage
	}
	coveredSet := map[string]struct{}{}
	for _, ns := range covered {
		coveredSet[ns] = struct{}{}
	}
	seen := map[string]struct{}{}
	var uncovered []string
	for _, sts := range b.Inventory.StatefulSets {
		if _, ok := coveredSet[sts.Namespace]; !ok {
			if _, already := seen[sts.Namespace]; !already {
				uncovered = append(uncovered, sts.Namespace)
				seen[sts.Namespace] = struct{}{}
			}
		}
	}
	return uncovered
}
