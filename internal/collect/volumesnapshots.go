package collect

import (
	"context"
	"fmt"

	"k8s-recovery-visualizer/internal/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	gvrVolumeSnapshotClass = schema.GroupVersionResource{
		Group:    "snapshot.storage.k8s.io",
		Version:  "v1",
		Resource: "volumesnapshotclasses",
	}
	gvrVolumeSnapshot = schema.GroupVersionResource{
		Group:    "snapshot.storage.k8s.io",
		Version:  "v1",
		Resource: "volumesnapshots",
	}
)

// VolumeSnapshotClasses collects snapshot.storage.k8s.io/v1 VolumeSnapshotClasses.
// Returns nil when the CRD is not installed (404 / no matches for kind).
func VolumeSnapshotClasses(ctx context.Context, dc dynamic.Interface, b *model.Bundle) error {
	list, err := dc.Resource(gvrVolumeSnapshotClass).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, item := range list.Items {
		vsc := model.VolumeSnapshotClass{
			Name: item.GetName(),
		}
		if driver, ok := item.Object["driver"].(string); ok {
			vsc.Driver = driver
		}
		if dp, ok := item.Object["deletionPolicy"].(string); ok {
			vsc.DeletionPolicy = dp
		}
		b.Inventory.VolumeSnapshotClasses = append(b.Inventory.VolumeSnapshotClasses, vsc)
	}
	return nil
}

// VolumeSnapshots collects snapshot.storage.k8s.io/v1 VolumeSnapshots across all namespaces.
// Returns nil when the CRD is not installed.
func VolumeSnapshots(ctx context.Context, dc dynamic.Interface, b *model.Bundle) error {
	list, err := dc.Resource(gvrVolumeSnapshot).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, item := range list.Items {
		if !InScope(item.GetNamespace(), b) {
			continue
		}
		vs := model.VolumeSnapshot{
			Namespace: item.GetNamespace(),
			Name:      item.GetName(),
		}

		// spec fields
		if spec, ok := item.Object["spec"].(map[string]interface{}); ok {
			if src, ok := spec["source"].(map[string]interface{}); ok {
				if pvcName, ok := src["persistentVolumeClaimName"].(string); ok {
					vs.PVCName = pvcName
				}
			}
			if className, ok := spec["volumeSnapshotClassName"].(string); ok {
				vs.ClassName = className
			}
		}

		// status fields
		if status, ok := item.Object["status"].(map[string]interface{}); ok {
			if rtu, ok := status["readyToUse"].(bool); ok {
				vs.ReadyToUse = rtu
			}
			if ct, ok := status["creationTime"].(string); ok {
				vs.CreatedAt = ct
			}
			// restoreSize is a resource.Quantity string like "10Gi"
			if rs, ok := status["restoreSize"].(string); ok {
				vs.SizeGB = parseQuantityGB(rs)
			}
		}

		// creationTimestamp fallback
		if vs.CreatedAt == "" {
			ts := item.GetCreationTimestamp()
			if !ts.IsZero() {
				vs.CreatedAt = ts.UTC().Format("2006-01-02T15:04:05Z")
			}
		}

		b.Inventory.VolumeSnapshots = append(b.Inventory.VolumeSnapshots, vs)
	}
	return nil
}

// parseQuantityGB converts a Kubernetes resource.Quantity string (e.g. "10Gi", "500Mi")
// into a float64 number of gigabytes. Returns 0 on parse failure.
func parseQuantityGB(s string) float64 {
	if s == "" {
		return 0
	}
	// Parse suffix manually to avoid importing resource package in collect layer.
	suffixes := []struct {
		suffix string
		factor float64
	}{
		{"Ti", 1024.0},
		{"Gi", 1.0},
		{"Mi", 1.0 / 1024.0},
		{"Ki", 1.0 / (1024.0 * 1024.0)},
		{"T", 1000.0},
		{"G", 1.0},
		{"M", 1.0 / 1000.0},
		{"K", 1.0 / (1000.0 * 1000.0)},
	}
	for _, sf := range suffixes {
		if len(s) > len(sf.suffix) {
			tail := s[len(s)-len(sf.suffix):]
			if tail == sf.suffix {
				numStr := s[:len(s)-len(sf.suffix)]
				var v float64
				_, err := fmt.Sscanf(numStr, "%f", &v)
				if err == nil {
					return v * sf.factor
				}
			}
		}
	}
	// Plain number — assume bytes, convert to GB
	var v float64
	if _, err := fmt.Sscanf(s, "%f", &v); err == nil {
		return v / (1024 * 1024 * 1024)
	}
	return 0
}
