package collect

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func HPAs(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.AutoscalingV2().HorizontalPodAutoscalers("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, hpa := range list.Items {
		if !InScope(hpa.Namespace, b) {
			continue
		}
		minRep := int32(1)
		if hpa.Spec.MinReplicas != nil {
			minRep = *hpa.Spec.MinReplicas
		}
		target := fmt.Sprintf("%s/%s", hpa.Spec.ScaleTargetRef.Kind, hpa.Spec.ScaleTargetRef.Name)
		b.Inventory.HPAs = append(b.Inventory.HPAs, model.HPA{
			Namespace:       hpa.Namespace,
			Name:            hpa.Name,
			Target:          target,
			MinReplicas:     minRep,
			MaxReplicas:     hpa.Spec.MaxReplicas,
			CurrentReplicas: hpa.Status.CurrentReplicas,
		})
	}
	return nil
}
