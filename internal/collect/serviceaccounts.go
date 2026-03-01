package collect

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s-recovery-visualizer/internal/model"
)

// ServiceAccounts collects all ServiceAccounts across namespaces.
func ServiceAccounts(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.CoreV1().ServiceAccounts("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, sa := range list.Items {
		if !InScope(sa.Namespace, b) {
			continue
		}
		m := model.ServiceAccount{
			Namespace:                    sa.Namespace,
			Name:                         sa.Name,
			AutomountServiceAccountToken: sa.AutomountServiceAccountToken,
		}
		b.Inventory.ServiceAccounts = append(b.Inventory.ServiceAccounts, m)
	}
	return nil
}
