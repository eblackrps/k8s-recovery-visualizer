package collect

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func PodDisruptionBudgets(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.PolicyV1().PodDisruptionBudgets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pdb := range list.Items {
		minAvail := ""
		if pdb.Spec.MinAvailable != nil {
			minAvail = pdb.Spec.MinAvailable.String()
		}
		maxUnavail := ""
		if pdb.Spec.MaxUnavailable != nil {
			maxUnavail = pdb.Spec.MaxUnavailable.String()
		}
		b.Inventory.PodDisruptionBudgets = append(b.Inventory.PodDisruptionBudgets, model.PodDisruptionBudget{
			Namespace:      pdb.Namespace,
			Name:           pdb.Name,
			MinAvailable:   minAvail,
			MaxUnavailable: maxUnavail,
		})
	}
	return nil
}
