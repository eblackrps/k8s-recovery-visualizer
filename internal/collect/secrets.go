package collect

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func Secrets(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.CoreV1().Secrets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, s := range list.Items {
		b.Inventory.Secrets = append(b.Inventory.Secrets, model.Secret{
			Namespace: s.Namespace,
			Name:      s.Name,
			Type:      string(s.Type),
			KeyCount:  len(s.Data),
		})
	}
	return nil
}
