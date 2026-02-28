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

// Deployment represents a Kubernetes Deployment.
type Deployment struct {
	Namespace   string   `json:"namespace"`
	Name        string   `json:"name"`
	Replicas    int32    `json:"replicas"`
	Ready       int32    `json:"ready"`
	Images      []string `json:"images,omitempty"`
}

// DaemonSet represents a Kubernetes DaemonSet.
type DaemonSet struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Desired   int32    `json:"desired"`
	Ready     int32    `json:"ready"`
	Images    []string `json:"images,omitempty"`
}

// Job represents a Kubernetes Job.
type Job struct {
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	Succeeded  int32  `json:"succeeded"`
	Failed     int32  `json:"failed"`
	Active     int32  `json:"active"`
	Completed  bool   `json:"completed"`
}

// CronJob represents a Kubernetes CronJob.
type CronJob struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Schedule    string `json:"schedule"`
	Suspended   bool   `json:"suspended"`
	LastRunTime string `json:"lastRunTime,omitempty"`
	ActiveJobs  int    `json:"activeJobs"`
}

// ServicePort holds port info for a Service.
type ServicePort struct {
	Name     string `json:"name,omitempty"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol,omitempty"`
	NodePort int32  `json:"nodePort,omitempty"`
}

// Service represents a Kubernetes Service.
type Service struct {
	Namespace  string            `json:"namespace"`
	Name       string            `json:"name"`
	Type       string            `json:"type"` // ClusterIP, NodePort, LoadBalancer, ExternalName
	ClusterIP  string            `json:"clusterIp,omitempty"`
	ExternalIP string            `json:"externalIp,omitempty"`
	Ports      []ServicePort     `json:"ports,omitempty"`
	Selector   map[string]string `json:"selector,omitempty"`
}

// IngressRule holds a single host rule for an Ingress.
type IngressRule struct {
	Host    string `json:"host,omitempty"`
	Backend string `json:"backend,omitempty"` // "service:port"
}

// Ingress represents a Kubernetes Ingress.
type Ingress struct {
	Namespace   string        `json:"namespace"`
	Name        string        `json:"name"`
	ClassName   string        `json:"className,omitempty"`
	TLS         bool          `json:"tls"`
	Rules       []IngressRule `json:"rules,omitempty"`
}

// ConfigMap represents metadata for a Kubernetes ConfigMap (no values).
type ConfigMap struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	KeyCount  int    `json:"keyCount"`
}

// Secret represents metadata for a Kubernetes Secret (no values or data).
type Secret struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	KeyCount  int    `json:"keyCount"`
}

// ClusterRole represents a Kubernetes ClusterRole.
type ClusterRole struct {
	Name            string   `json:"name"`
	Custom          bool     `json:"custom"` // true if not a built-in system: role
	RuleCount       int      `json:"ruleCount"`
	HasWildcardVerb bool     `json:"hasWildcardVerb,omitempty"`  // any rule grants wildcard verb
	HasSecretAccess bool     `json:"hasSecretAccess,omitempty"`  // any rule grants read on secrets
	HasEscalatePriv bool     `json:"hasEscalatePriv,omitempty"`  // escalate/bind/impersonate verbs
	DangerousRules  []string `json:"dangerousRules,omitempty"`   // human-readable risk summary
}

// ClusterRoleBinding represents a Kubernetes ClusterRoleBinding.
type ClusterRoleBinding struct {
	Name     string   `json:"name"`
	RoleName string   `json:"roleName"`
	Subjects []string `json:"subjects,omitempty"` // "kind:name" strings
}

// NetworkPolicy represents a Kubernetes NetworkPolicy.
type NetworkPolicy struct {
	Namespace     string `json:"namespace"`
	Name          string `json:"name"`
	PodSelector   string `json:"podSelector,omitempty"`
	HasIngress    bool   `json:"hasIngress"`
	HasEgress     bool   `json:"hasEgress"`
}

// HPA represents a Kubernetes HorizontalPodAutoscaler.
type HPA struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Target      string `json:"target"`
	MinReplicas int32  `json:"minReplicas"`
	MaxReplicas int32  `json:"maxReplicas"`
	CurrentReplicas int32 `json:"currentReplicas"`
}

// PodDisruptionBudget represents a Kubernetes PodDisruptionBudget.
type PodDisruptionBudget struct {
	Namespace        string `json:"namespace"`
	Name             string `json:"name"`
	MinAvailable     string `json:"minAvailable,omitempty"`
	MaxUnavailable   string `json:"maxUnavailable,omitempty"`
}

// ResourceQuotaItem holds a single resource limit.
type ResourceQuotaItem struct {
	Resource string `json:"resource"`
	Hard     string `json:"hard"`
	Used     string `json:"used,omitempty"`
}

// ResourceQuota represents a Kubernetes ResourceQuota.
type ResourceQuota struct {
	Namespace string              `json:"namespace"`
	Name      string              `json:"name"`
	Items     []ResourceQuotaItem `json:"items,omitempty"`
}

// CRD represents a Kubernetes CustomResourceDefinition.
type CRD struct {
	Name     string   `json:"name"`
	Group    string   `json:"group"`
	Versions []string `json:"versions,omitempty"`
	Scope    string   `json:"scope"` // Namespaced or Cluster
}

// HelmRelease represents an installed Helm release detected from cluster secrets.
type HelmRelease struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Chart     string `json:"chart"`
	Version   string `json:"version"`
	AppVersion string `json:"appVersion,omitempty"`
	Status    string `json:"status"` // deployed, failed, pending, etc.
}

// ContainerImage represents a unique container image found across all workloads.
type ContainerImage struct {
	Image     string   `json:"image"`
	Registry  string   `json:"registry"`
	IsPublic  bool     `json:"isPublic"` // true if known public registry
	Workloads []string `json:"workloads,omitempty"` // "ns/name" references
}

// Certificate represents a cert-manager Certificate resource.
type Certificate struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	SecretName  string `json:"secretName,omitempty"`
	Issuer      string `json:"issuer,omitempty"`
	Ready       bool   `json:"ready"`
	NotAfter    string `json:"notAfter,omitempty"`
	DaysToExpiry int   `json:"daysToExpiry"`
}

// Platform holds detected cluster platform/provider information.
type Platform struct {
	Provider   string `json:"provider"` // EKS, AKS, GKE, Rancher, k3s, vanilla
	K8sVersion string `json:"k8sVersion,omitempty"`
	ClusterUID string `json:"clusterUID,omitempty"`
}
