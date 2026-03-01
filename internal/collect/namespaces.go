package collect

import (
	"context"

	"k8s-recovery-visualizer/internal/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Namespaces(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, ns := range list.Items {
		if !InScope(ns.Name, b) {
			continue
		}
		labels := ns.Labels
		b.Inventory.Namespaces = append(b.Inventory.Namespaces, model.Namespace{
			ID:         "ns:" + ns.Name,
			Name:       ns.Name,
			PSAEnforce: labels["pod-security.kubernetes.io/enforce"],
			PSAWarn:    labels["pod-security.kubernetes.io/warn"],
			PSAAudit:   labels["pod-security.kubernetes.io/audit"],
		})
	}
	return nil
}
