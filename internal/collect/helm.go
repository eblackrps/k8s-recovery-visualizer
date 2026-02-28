package collect

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func HelmReleases(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	// Helm v3 stores releases as secrets with this label
	list, err := cs.CoreV1().Secrets("").List(ctx, metav1.ListOptions{
		LabelSelector: "owner=helm",
	})
	if err != nil {
		return err
	}

	// Track latest version per release (name+namespace)
	type releaseKey struct{ ns, name string }
	latest := map[releaseKey]model.HelmRelease{}

	for _, s := range list.Items {
		if s.Type != "helm.sh/release.v1" {
			continue
		}
		labels := s.Labels
		name := labels["name"]
		status := labels["status"]
		if name == "" {
			// fallback: parse from secret name "sh.helm.release.v1.NAME.vN"
			parts := strings.Split(s.Name, ".")
			if len(parts) >= 5 {
				name = parts[4]
			}
		}
		// Extract chart name from the secret name pattern
		chart := name // chart name often equals release name in simple cases
		// Use the "helm.sh/chart" annotation if present on the secret
		if c, ok := s.Annotations["meta.helm.sh/release-name"]; ok && c != "" {
			chart = c
		}

		key := releaseKey{ns: s.Namespace, name: name}
		existing, ok := latest[key]
		// keep the entry with status=deployed over others, or just keep the latest
		if !ok || status == "deployed" || existing.Status != "deployed" {
			latest[key] = model.HelmRelease{
				Namespace:  s.Namespace,
				Name:       name,
				Chart:      chart,
				Version:    labels["version"],
				AppVersion: "",
				Status:     status,
			}
		}
	}

	for _, r := range latest {
		b.Inventory.HelmReleases = append(b.Inventory.HelmReleases, r)
	}
	return nil
}
