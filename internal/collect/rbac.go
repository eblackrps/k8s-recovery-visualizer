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

		var hasWildcard, hasSecrets, hasEscalate bool
		for _, rule := range cr.Rules {
			// wildcard verb check
			for _, v := range rule.Verbs {
				if v == "*" {
					hasWildcard = true
				}
			}
			// escalate / bind / impersonate verb check
			for _, v := range rule.Verbs {
				if v == "escalate" || v == "bind" || v == "impersonate" {
					hasEscalate = true
				}
			}
			// secrets read access: verbs allow read AND resource is secrets or *
			verbsAllowRead := false
			for _, v := range rule.Verbs {
				if v == "*" || v == "get" || v == "list" || v == "watch" {
					verbsAllowRead = true
					break
				}
			}
			if verbsAllowRead {
				for _, res := range rule.Resources {
					if res == "*" || res == "secrets" {
						hasSecrets = true
					}
				}
			}
		}

		var dangerous []string
		if hasWildcard {
			dangerous = append(dangerous, "wildcard verbs")
		}
		if hasSecrets {
			dangerous = append(dangerous, "secrets read access")
		}
		if hasEscalate {
			dangerous = append(dangerous, "escalate/bind/impersonate")
		}

		b.Inventory.ClusterRoles = append(b.Inventory.ClusterRoles, model.ClusterRole{
			Name:            cr.Name,
			Custom:          custom,
			RuleCount:       len(cr.Rules),
			HasWildcardVerb: hasWildcard,
			HasSecretAccess: hasSecrets,
			HasEscalatePriv: hasEscalate,
			DangerousRules:  dangerous,
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
