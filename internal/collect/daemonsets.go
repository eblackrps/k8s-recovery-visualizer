package collect

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func DaemonSets(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, ds := range list.Items {
		var images []string
		for _, c := range ds.Spec.Template.Spec.Containers {
			images = append(images, c.Image)
		}
		b.Inventory.DaemonSets = append(b.Inventory.DaemonSets, model.DaemonSet{
			Namespace: ds.Namespace,
			Name:      ds.Name,
			Desired:   ds.Status.DesiredNumberScheduled,
			Ready:     ds.Status.NumberReady,
			Images:    images,
		})
	}
	return nil
}
