package collect

import "k8s-recovery-visualizer/internal/model"

// InScope returns true when ns is within the scan's namespace scope.
// If b.ScanNamespaces is empty, all namespaces are in scope.
func InScope(ns string, b *model.Bundle) bool {
	if len(b.ScanNamespaces) == 0 {
		return true
	}
	for _, n := range b.ScanNamespaces {
		if n == ns {
			return true
		}
	}
	return false
}

// ScopeNamespaces returns the namespaces a collector should query directly.
// An empty string means "all namespaces" for cluster-scoped list operations.
func ScopeNamespaces(b *model.Bundle) []string {
	if len(b.ScanNamespaces) == 0 {
		return []string{""}
	}
	out := make([]string, 0, len(b.ScanNamespaces))
	for _, ns := range b.ScanNamespaces {
		if ns != "" {
			out = append(out, ns)
		}
	}
	if len(out) == 0 {
		return []string{""}
	}
	return out
}
