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
