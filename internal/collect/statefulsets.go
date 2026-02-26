package collect

import (
	"context"

	"k8s-recovery-visualizer/internal/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func StatefulSets(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {

	list, err := cs.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, sts := range list.Items {

		hasPVC := len(sts.Spec.VolumeClaimTemplates) > 0

		replicas := int32(0)
		if sts.Spec.Replicas != nil {
			replicas = *sts.Spec.Replicas
		}

		b.Inventory.StatefulSets = append(b.Inventory.StatefulSets, model.StatefulSet{
			Namespace:      sts.Namespace,
			Name:           sts.Name,
			Replicas:       replicas,
			HasVolumeClaim: hasPVC,
		})
	}

	return nil
}
