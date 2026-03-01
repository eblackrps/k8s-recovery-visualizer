package collect

import (
  "context"
  "strings"

  v1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/client-go/kubernetes"

  "k8s-recovery-visualizer/internal/model"
)

func Nodes(ctx context.Context, client kubernetes.Interface, b *model.Bundle) error {
  list, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
  if err != nil {
    return err
  }

  out := make([]model.Node, 0, len(list.Items))
  for _, n := range list.Items {
    out = append(out, toModelNode(&n))
  }

  b.Inventory.Nodes = out
  return nil
}

func toModelNode(n *v1.Node) model.Node {
  m := model.Node{
    Name:   n.Name,
    Labels: map[string]string{},
  }

  // Labels
  for k, v := range n.Labels {
    m.Labels[k] = v
  }

  // Zone — prefer stable GA label, fall back to deprecated beta label
  if z, ok := n.Labels["topology.kubernetes.io/zone"]; ok && z != "" {
    m.Zone = z
  } else if z, ok := n.Labels["failure-domain.beta.kubernetes.io/zone"]; ok && z != "" {
    m.Zone = z
  }

  // Roles from labels (common convention)
  // node-role.kubernetes.io/<role>=true or empty
  for k := range n.Labels {
    if strings.HasPrefix(k, "node-role.kubernetes.io/") {
      role := strings.TrimPrefix(k, "node-role.kubernetes.io/")
      if role == "" {
        role = "master"
      }
      m.Roles = append(m.Roles, role)
    }
  }

  // Ready condition
  m.Ready = false
  for _, c := range n.Status.Conditions {
    if c.Type == v1.NodeReady {
      m.Ready = (c.Status == v1.ConditionTrue)
      break
    }
  }

  // Versions / OS
  m.KubeletVersion = n.Status.NodeInfo.KubeletVersion
  m.OSImage = n.Status.NodeInfo.OSImage
  m.KernelVersion = n.Status.NodeInfo.KernelVersion
  m.ContainerRuntime = n.Status.NodeInfo.ContainerRuntimeVersion

  // IPs
  for _, a := range n.Status.Addresses {
    if a.Type == v1.NodeInternalIP && m.InternalIP == "" {
      m.InternalIP = a.Address
    }
    if a.Type == v1.NodeExternalIP && m.ExternalIP == "" {
      m.ExternalIP = a.Address
    }
  }

  // Taints (stringified)
  for _, t := range n.Spec.Taints {
    m.Taints = append(m.Taints, t.Key+"="+t.Value+":"+string(t.Effect))
  }

  return m
}
