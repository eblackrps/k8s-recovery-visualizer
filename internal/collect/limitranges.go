package collect

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func LimitRanges(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.CoreV1().LimitRanges("").List(ctx, metav1.ListOptions{})
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
	return nil
}
