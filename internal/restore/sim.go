// Package restore provides restore simulation logic that assesses per-namespace
// recovery feasibility based on backup policy coverage, PVC sizing, and known
// blockers (hostPath volumes, unbound PVCs, missing StorageClasses).
package restore

import (
	"strconv"
	"strings"

	"k8s-recovery-visualizer/internal/model"
)

// Simulate builds a per-namespace restore feasibility assessment and returns
// the aggregated result. It is called after backup.Detect() so that policy
// data is already present on the bundle.
func Simulate(b *model.Bundle) model.RestoreSimResult {
	inv := b.Inventory.Backup

	// Build set of StorageClasses present in the cluster for blocker detection.
	scSet := map[string]struct{}{}
	for _, sc := range b.Inventory.StorageClasses {
		scSet[sc.Name] = struct{}{}
	}

	// Build per-namespace PVC size totals and per-PVC metadata for blocker checks.
	type pvcMeta struct {
		storageClass string
		sizeGB       float64
		bound        bool   // true when a matching PV exists
		backend      string // "hostPath" triggers a blocker
	}
	nsPVCs := map[string][]pvcMeta{}
	pvMap := map[string]model.PersistentVolume{}
	for _, pv := range b.Inventory.PVs {
		pvMap[pv.ClaimRef] = pv
	}
	for _, pvc := range b.Inventory.PVCs {
		pv, bound := pvMap[pvc.Namespace+"/"+pvc.Name]
		backend := ""
		if bound {
			backend = pv.Backend
		}
		nsPVCs[pvc.Namespace] = append(nsPVCs[pvc.Namespace], pvcMeta{
			storageClass: pvc.StorageClass,
			sizeGB:       parseGiB(pvc.RequestedSize),
			bound:        bound,
			backend:      backend,
		})
	}

	// Collect the namespaces that are relevant for restore simulation:
	// any namespace that has StatefulSets or PVCs.
	relevantNS := map[string]struct{}{}
	for _, sts := range b.Inventory.StatefulSets {
		relevantNS[sts.Namespace] = struct{}{}
	}
	for ns := range nsPVCs {
		relevantNS[ns] = struct{}{}
	}

	var result model.RestoreSimResult
	var uncoveredNS []string

	for ns := range relevantNS {
		sim := model.RestoreSimNamespace{
			Namespace:   ns,
			HasCoverage: policyCoversNamespace(inv, ns),
			RPOHours:    bestRPOForNamespace(inv, ns),
		}

		// Sum PVC sizes and collect blockers/warnings.
		var nsSizeGB float64
		for _, pvc := range nsPVCs[ns] {
			nsSizeGB += pvc.sizeGB
			if !pvc.bound {
				sim.Blockers = append(sim.Blockers, "unbound PVC")
			}
			if pvc.backend == "hostPath" {
				sim.Blockers = append(sim.Blockers, "hostPath volume — not portable across nodes")
			}
			if pvc.storageClass != "" {
				if _, ok := scSet[pvc.storageClass]; !ok {
					sim.Warnings = append(sim.Warnings,
						"StorageClass '"+pvc.storageClass+"' not present in cluster")
				}
			}
		}
		sim.PVCSizeGB = nsSizeGB

		if !sim.HasCoverage {
			uncoveredNS = append(uncoveredNS, ns)
		}

		result.Namespaces = append(result.Namespaces, sim)
		result.TotalPVCsGB += nsSizeGB
		if sim.HasCoverage {
			result.CoveredPVCsGB += nsSizeGB
		}
	}

	result.UncoveredNS = uncoveredNS
	return result
}

// policyCoversNamespace returns true when at least one policy covers the given
// namespace, or when the tool has wildcard coverage (CoveredNamespaces == ["*"]).
func policyCoversNamespace(inv model.BackupInventory, ns string) bool {
	if inv.PrimaryTool == "none" || inv.PrimaryTool == "" {
		return false
	}
	// Policy-level check (Velero, Kasten, Longhorn).
	if len(inv.Policies) > 0 {
		for _, p := range inv.Policies {
			if matchesPolicy(p, ns) {
				return true
			}
		}
		return false
	}
	// Fallback: use the legacy CoveredNamespaces list.
	if len(inv.CoveredNamespaces) == 1 && inv.CoveredNamespaces[0] == "*" {
		return true
	}
	for _, covered := range inv.CoveredNamespaces {
		if covered == ns {
			return true
		}
	}
	return false
}

// matchesPolicy returns true when a BackupPolicy covers the given namespace.
func matchesPolicy(p model.BackupPolicy, ns string) bool {
	// Excluded always wins.
	for _, ex := range p.ExcludedNS {
		if ex == ns {
			return false
		}
	}
	// Empty IncludedNS = all namespaces.
	if len(p.IncludedNS) == 0 {
		return true
	}
	for _, incl := range p.IncludedNS {
		if incl == ns || incl == "*" {
			return true
		}
	}
	return false
}

// bestRPOForNamespace returns the lowest (best) RPO hours across all policies
// that cover the namespace. Returns -1 when unknown or no coverage.
func bestRPOForNamespace(inv model.BackupInventory, ns string) int {
	best := -1
	for _, p := range inv.Policies {
		if !matchesPolicy(p, ns) {
			continue
		}
		if p.RPOHours < 0 {
			continue
		}
		if best < 0 || p.RPOHours < best {
			best = p.RPOHours
		}
	}
	return best
}

// parseGiB converts a Kubernetes quantity string (e.g. "10Gi", "500Mi", "2Ti")
// to a float64 number of GiB. Returns 0 for unparseable values.
func parseGiB(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	suffixes := []struct {
		suffix string
		factor float64
	}{
		{"Ti", 1024},
		{"Gi", 1},
		{"Mi", 1.0 / 1024},
		{"Ki", 1.0 / (1024 * 1024)},
		{"T", 1000},
		{"G", 1000.0 / 1024},
		{"M", 1000.0 / (1024 * 1024)},
	}
	for _, sf := range suffixes {
		if strings.HasSuffix(s, sf.suffix) {
			numStr := strings.TrimSuffix(s, sf.suffix)
			var n float64
			if err := parseFloat(numStr, &n); err == nil {
				return n * sf.factor
			}
		}
	}
	// Plain bytes
	var n float64
	if err := parseFloat(s, &n); err == nil {
		return n / (1024 * 1024 * 1024)
	}
	return 0
}

func parseFloat(s string, out *float64) error {
	n, err := strconv.ParseFloat(s, 64)
	*out = n
	return err
}
