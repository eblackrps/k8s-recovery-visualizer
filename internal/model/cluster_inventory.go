package model

// Node represents a Kubernetes node with enough detail for DR scoring and reporting.
type Node struct {
  Name             string            `json:"name"`
  Roles            []string          `json:"roles,omitempty"`
  Ready            bool              `json:"ready"`
  KubeletVersion   string            `json:"kubeletVersion,omitempty"`
  OSImage          string            `json:"osImage,omitempty"`
  KernelVersion    string            `json:"kernelVersion,omitempty"`
  ContainerRuntime string            `json:"containerRuntime,omitempty"`
  InternalIP       string            `json:"internalIp,omitempty"`
  ExternalIP       string            `json:"externalIp,omitempty"`
  Labels           map[string]string `json:"labels,omitempty"`
  Taints           []string          `json:"taints,omitempty"`
}

// StorageClass represents a Kubernetes StorageClass for DR suitability checks.
type StorageClass struct {
  Name                 string            `json:"name"`
  Provisioner          string            `json:"provisioner,omitempty"`
  ReclaimPolicy        string            `json:"reclaimPolicy,omitempty"`
  VolumeBindingMode    string            `json:"volumeBindingMode,omitempty"`
  AllowVolumeExpansion *bool             `json:"allowVolumeExpansion,omitempty"`
  Parameters           map[string]string `json:"parameters,omitempty"`
  Annotations          map[string]string `json:"annotations,omitempty"`
}
