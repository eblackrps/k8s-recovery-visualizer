package collect

import (
	"context"

	"k8s-recovery-visualizer/internal/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PVCs(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pvc := range list.Items {
		size := ""
		if qty, ok := pvc.Spec.Resources.Requests["storage"]; ok {
			size = qty.String()
		}

		b.Inventory.PVCs = append(b.Inventory.PVCs, model.PersistentVolumeClaim{
			ID:            "pvc:" + pvc.Namespace + ":" + pvc.Name,
			Name:          pvc.Name,
			Namespace:     pvc.Namespace,
			StorageClass:  deref(pvc.Spec.StorageClassName),
			AccessModes:   accessModesToStrings(pvc.Spec.AccessModes),
			RequestedSize: size,
		})
	}

	return nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func accessModesToStrings(modes []v1.PersistentVolumeAccessMode) []string {
	out := []string{}
	for _, m := range modes {
		out = append(out, string(m))
	}
	return out
}
