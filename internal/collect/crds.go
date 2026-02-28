package collect

import (
	"context"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

// nativeGroups are built-in K8s API groups that are not CRDs.
var nativeGroups = map[string]struct{}{
	"":                                  {},
	"apps":                              {},
	"batch":                             {},
	"networking.k8s.io":                 {},
	"storage.k8s.io":                    {},
	"rbac.authorization.k8s.io":         {},
	"autoscaling":                       {},
	"policy":                            {},
	"apiextensions.k8s.io":              {},
	"apiregistration.k8s.io":            {},
	"coordination.k8s.io":               {},
	"discovery.k8s.io":                  {},
	"events.k8s.io":                     {},
	"admissionregistration.k8s.io":      {},
	"certificates.k8s.io":               {},
	"scheduling.k8s.io":                 {},
	"flowcontrol.apiserver.k8s.io":      {},
	"resource.k8s.io":                   {},
	"node.k8s.io":                       {},
	"authentication.k8s.io":             {},
	"authorization.k8s.io":              {},
	"internal.apiserver.k8s.io":         {},
	"storagemigration.k8s.io":           {},
}

// CRDs detects installed CustomResourceDefinitions via the discovery API.
// It identifies non-native API groups as CRDs without requiring the
// apiextensions-apiserver package.
func CRDs(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	groups, err := cs.Discovery().ServerGroups()
	if err != nil {
		return err
	}

	seen := map[string]struct{}{}
	for _, g := range groups.Groups {
		if _, native := nativeGroups[g.Name]; native {
			continue
		}
		if _, already := seen[g.Name]; already {
			continue
		}
		seen[g.Name] = struct{}{}

		// Collect served versions for this group from discovery.
		var versions []string
		for _, v := range g.Versions {
			if v.Version != "" {
				versions = append(versions, v.Version)
			}
		}

		// Determine scope by querying resources in the preferred version.
		scope := "Unknown"
		pv := g.PreferredVersion.GroupVersion
		if resources, err := cs.Discovery().ServerResourcesForGroupVersion(pv); err == nil {
			for _, r := range resources.APIResources {
				if strings.Contains(r.Name, "/") {
					continue // skip subresources
				}
				if r.Namespaced {
					scope = "Namespaced"
				} else {
					scope = "Cluster"
				}
				break
			}
		}

		b.Inventory.CRDs = append(b.Inventory.CRDs, model.CRD{
			Name:     g.Name,
			Group:    g.Name,
			Versions: versions,
			Scope:    scope,
		})
	}
	return nil
}
