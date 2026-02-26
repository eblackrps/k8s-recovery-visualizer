package collect

import (
	"context"

	"k8s-recovery-visualizer/internal/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Pods(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {

	list, err := cs.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range list.Items {

		usesHostPath := false

		for _, vol := range pod.Spec.Volumes {
			if vol.HostPath != nil {
				usesHostPath = true
				break
			}
		}

		b.Inventory.Pods = append(b.Inventory.Pods, model.Pod{
			Namespace:    pod.Namespace,
			Name:         pod.Name,
			UsesHostPath: usesHostPath,
		})
	}

	return nil
}
