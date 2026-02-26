package collect

import (
	"context"

	"k8s-recovery-visualizer/internal/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PVs(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pv := range list.Items {

		capacity := ""
		if qty, ok := pv.Spec.Capacity[v1.ResourceStorage]; ok {
			capacity = qty.String()
		}

		claim := ""
		if pv.Spec.ClaimRef != nil {
			claim = pv.Spec.ClaimRef.Namespace + "/" + pv.Spec.ClaimRef.Name
		}

		b.Inventory.PVs = append(b.Inventory.PVs, model.PersistentVolume{
			Name:          pv.Name,
			StorageClass:  pv.Spec.StorageClassName,
			Capacity:      capacity,
			ReclaimPolicy: string(pv.Spec.PersistentVolumeReclaimPolicy),
			Backend:       detectBackend(&pv),
			ClaimRef:      claim,
		})
	}

	return nil
}

func detectBackend(pv *v1.PersistentVolume) string {
	switch {
	case pv.Spec.HostPath != nil:
		return "hostPath"
	case pv.Spec.CSI != nil:
		return "csi"
	case pv.Spec.NFS != nil:
		return "nfs"
	case pv.Spec.Local != nil:
		return "local"
	default:
		return "unknown"
	}
}
