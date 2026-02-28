package collect

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func Deployments(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, d := range list.Items {
		var images []string
		for _, c := range d.Spec.Template.Spec.Containers {
			images = append(images, c.Image)
		}
		for _, c := range d.Spec.Template.Spec.InitContainers {
			images = append(images, c.Image)
		}
		desired := int32(1)
		if d.Spec.Replicas != nil {
			desired = *d.Spec.Replicas
		}
		b.Inventory.Deployments = append(b.Inventory.Deployments, model.Deployment{
			Namespace: d.Namespace,
			Name:      d.Name,
			Replicas:  desired,
			Ready:     d.Status.ReadyReplicas,
			Images:    images,
		})
	}
	return nil
}
