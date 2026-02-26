package collect

import (
  "context"

  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/client-go/kubernetes"

  "k8s-recovery-visualizer/internal/model"
)

func StorageClasses(ctx context.Context, client kubernetes.Interface, b *model.Bundle) error {
  list, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
  if err != nil {
    return err
  }

  out := make([]model.StorageClass, 0, len(list.Items))
  for _, sc := range list.Items {
    m := model.StorageClass{
      Name:              sc.Name,
      Provisioner:       sc.Provisioner,
      Parameters:        map[string]string{},
      Annotations:       map[string]string{},
      ReclaimPolicy:     "",
      VolumeBindingMode: "",
    }

    if sc.ReclaimPolicy != nil {
      m.ReclaimPolicy = string(*sc.ReclaimPolicy)
    }
    if sc.VolumeBindingMode != nil {
      m.VolumeBindingMode = string(*sc.VolumeBindingMode)
    }
    if sc.AllowVolumeExpansion != nil {
      v := *sc.AllowVolumeExpansion
      m.AllowVolumeExpansion = &v
    }

    for k, v := range sc.Parameters {
      m.Parameters[k] = v
    }
    for k, v := range sc.Annotations {
      m.Annotations[k] = v
    }

    out = append(out, m)
  }

  b.Inventory.StorageClasses = out
  return nil
}
