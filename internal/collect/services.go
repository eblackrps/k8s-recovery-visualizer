package collect

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func Services(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, svc := range list.Items {
		var ports []model.ServicePort
		for _, p := range svc.Spec.Ports {
			ports = append(ports, model.ServicePort{
				Name:     p.Name,
				Port:     p.Port,
				Protocol: string(p.Protocol),
				NodePort: p.NodePort,
			})
		}
		externalIP := ""
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			ing := svc.Status.LoadBalancer.Ingress[0]
			if ing.IP != "" {
				externalIP = ing.IP
			} else {
				externalIP = ing.Hostname
			}
		}
		b.Inventory.Services = append(b.Inventory.Services, model.Service{
			Namespace:  svc.Namespace,
			Name:       svc.Name,
			Type:       string(svc.Spec.Type),
			ClusterIP:  svc.Spec.ClusterIP,
			ExternalIP: externalIP,
			Ports:      ports,
			Selector:   svc.Spec.Selector,
		})
	}
	return nil
}
