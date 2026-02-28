package collect

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func ClusterRoles(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, cr := range list.Items {
		custom := !strings.HasPrefix(cr.Name, "system:")
		b.Inventory.ClusterRoles = append(b.Inventory.ClusterRoles, model.ClusterRole{
			Name:      cr.Name,
			Custom:    custom,
			RuleCount: len(cr.Rules),
		})
	}
	return nil
}

func ClusterRoleBindings(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, crb := range list.Items {
		var subjects []string
		for _, s := range crb.Subjects {
			subjects = append(subjects, fmt.Sprintf("%s:%s", s.Kind, s.Name))
		}
		b.Inventory.ClusterRoleBindings = append(b.Inventory.ClusterRoleBindings, model.ClusterRoleBinding{
			Name:     crb.Name,
			RoleName: crb.RoleRef.Name,
			Subjects: subjects,
		})
	}
	return nil
}
