package collect

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func ResourceQuotas(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.CoreV1().ResourceQuotas("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, rq := range list.Items {
		var items []model.ResourceQuotaItem
		for resource, hard := range rq.Spec.Hard {
			used := ""
			if u, ok := rq.Status.Used[resource]; ok {
				used = u.String()
			}
			items = append(items, model.ResourceQuotaItem{
				Resource: string(resource),
				Hard:     hard.String(),
				Used:     used,
			})
		}
		b.Inventory.ResourceQuotas = append(b.Inventory.ResourceQuotas, model.ResourceQuota{
			Namespace: rq.Namespace,
			Name:      rq.Name,
			Items:     items,
		})
	}
	return nil
}
