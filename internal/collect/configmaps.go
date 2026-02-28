package collect

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func ConfigMaps(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, cm := range list.Items {
		if !InScope(cm.Namespace, b) {
			continue
		}
		b.Inventory.ConfigMaps = append(b.Inventory.ConfigMaps, model.ConfigMap{
			Namespace: cm.Namespace,
			Name:      cm.Name,
			KeyCount:  len(cm.Data) + len(cm.BinaryData),
		})
	}
	return nil
}
