package collect

import (
	"context"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

var publicRegistries = map[string]struct{}{
	"docker.io":             {},
	"registry-1.docker.io": {},
	"ghcr.io":              {},
	"quay.io":              {},
	"gcr.io":               {},
	"registry.k8s.io":      {},
	"k8s.gcr.io":           {},
	"mcr.microsoft.com":    {},
	"public.ecr.aws":       {},
}

// Images builds a unique container image inventory from already-collected workloads.
// No additional K8s API calls are made.
func Images(_ context.Context, _ *kubernetes.Clientset, b *model.Bundle) error {
	type imgEntry struct {
		registry  string
		isPublic  bool
		workloads []string
	}
	seen := map[string]*imgEntry{}

	addImage := func(image, workloadRef string) {
		if image == "" {
			return
		}
		e, exists := seen[image]
		if !exists {
			reg := registryOf(image)
			_, pub := publicRegistries[reg]
			e = &imgEntry{registry: reg, isPublic: pub}
			seen[image] = e
		}
		e.workloads = append(e.workloads, workloadRef)
	}

	for _, d := range b.Inventory.Deployments {
		ref := d.Namespace + "/" + d.Name
		for _, img := range d.Images {
			addImage(img, ref)
		}
	}
	for _, ds := range b.Inventory.DaemonSets {
		ref := ds.Namespace + "/" + ds.Name
		for _, img := range ds.Images {
			addImage(img, ref)
		}
	}
	for _, sts := range b.Inventory.StatefulSets {
		ref := sts.Namespace + "/" + sts.Name
		// StatefulSet model doesn't carry images directly; skip for now
		_ = ref
	}

	for img, entry := range seen {
		b.Inventory.Images = append(b.Inventory.Images, model.ContainerImage{
			Image:     img,
			Registry:  entry.registry,
			IsPublic:  entry.isPublic,
			Workloads: entry.workloads,
		})
	}
	return nil
}

// registryOf extracts the registry hostname from an image reference.
// Examples:
//
//	nginx                      → docker.io
//	nginx:latest               → docker.io
//	myregistry.io/app:v1       → myregistry.io
//	ghcr.io/org/image:tag      → ghcr.io
func registryOf(image string) string {
	// Strip tag/digest
	name := image
	if i := strings.Index(name, "@"); i >= 0 {
		name = name[:i]
	}
	if i := strings.LastIndex(name, ":"); i >= 0 {
		name = name[:i]
	}
	// Split on /
	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 1 {
		return "docker.io" // bare image like "nginx"
	}
	first := parts[0]
	// Registry hostnames contain a dot or colon or are "localhost"
	if strings.ContainsAny(first, ".:") || first == "localhost" {
		return first
	}
	return "docker.io" // e.g. "library/nginx"
}
