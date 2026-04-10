package collect

import (
	"context"

	"k8s-recovery-visualizer/internal/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func LimitRanges(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	for _, ns := range ScopeNamespaces(b) {
		list, err := cs.CoreV1().LimitRanges(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, lr := range list.Items {
			if !InScope(lr.Namespace, b) {
				continue
			}
			var items []model.LimitRangeItem
			for _, lri := range lr.Spec.Limits {
				item := model.LimitRangeItem{
					Type: string(lri.Type),
				}
				if v, ok := lri.Max[corev1.ResourceCPU]; ok {
					item.MaxCPU = v.String()
				}
				if v, ok := lri.Max[corev1.ResourceMemory]; ok {
					item.MaxMemory = v.String()
				}
				if lri.Type == corev1.LimitTypeContainer {
					if v, ok := lri.Default[corev1.ResourceCPU]; ok {
						item.DefaultCPU = v.String()
					}
					if v, ok := lri.Default[corev1.ResourceMemory]; ok {
						item.DefaultMemory = v.String()
					}
				}
				items = append(items, item)
			}
			b.Inventory.LimitRanges = append(b.Inventory.LimitRanges, model.LimitRange{
				Namespace: lr.Namespace,
				Name:      lr.Name,
				Items:     items,
			})
		}
	}
	return nil
}
