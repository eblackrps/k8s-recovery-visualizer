package collect

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func NetworkPolicies(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, np := range list.Items {
		if !InScope(np.Namespace, b) {
			continue
		}
		sel := ""
		if len(np.Spec.PodSelector.MatchLabels) > 0 {
			for k, v := range np.Spec.PodSelector.MatchLabels {
				sel = fmt.Sprintf("%s=%s", k, v)
				break
			}
		}
		hasIngress := len(np.Spec.Ingress) > 0
		hasEgress := len(np.Spec.Egress) > 0
		b.Inventory.NetworkPolicies = append(b.Inventory.NetworkPolicies, model.NetworkPolicy{
			Namespace:   np.Namespace,
			Name:        np.Name,
			PodSelector: sel,
			HasIngress:  hasIngress,
			HasEgress:   hasEgress,
		})
	}
	return nil
}
