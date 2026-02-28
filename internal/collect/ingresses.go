package collect

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func Ingresses(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, ing := range list.Items {
		if !InScope(ing.Namespace, b) {
			continue
		}
		hasTLS := len(ing.Spec.TLS) > 0
		className := ""
		if ing.Spec.IngressClassName != nil {
			className = *ing.Spec.IngressClassName
		}
		var rules []model.IngressRule
		for _, r := range ing.Spec.Rules {
			backend := ""
			if r.HTTP != nil && len(r.HTTP.Paths) > 0 {
				p := r.HTTP.Paths[0]
				if p.Backend.Service != nil {
					backend = fmt.Sprintf("%s:%d", p.Backend.Service.Name, p.Backend.Service.Port.Number)
				}
			}
			rules = append(rules, model.IngressRule{
				Host:    r.Host,
				Backend: backend,
			})
		}
		b.Inventory.Ingresses = append(b.Inventory.Ingresses, model.Ingress{
			Namespace: ing.Namespace,
			Name:      ing.Name,
			ClassName: className,
			TLS:       hasTLS,
			Rules:     rules,
		})
	}
	return nil
}
