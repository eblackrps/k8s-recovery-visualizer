package collect

import (
	"context"

	"k8s-recovery-visualizer/internal/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DaemonSets(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	for _, ns := range ScopeNamespaces(b) {
		list, err := cs.AppsV1().DaemonSets(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, ds := range list.Items {
			if !InScope(ds.Namespace, b) {
				continue
			}
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
	}
	return nil
}
