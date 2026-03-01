package model

// LimitRangeItem holds a single limit/request constraint from a LimitRange.
type LimitRangeItem struct {
	Type           string `json:"type"`                     // Container, Pod, PersistentVolumeClaim
	MaxCPU         string `json:"maxCpu,omitempty"`
	MaxMemory      string `json:"maxMemory,omitempty"`
	DefaultCPU     string `json:"defaultCpu,omitempty"`
	DefaultMemory  string `json:"defaultMemory,omitempty"`
}

// LimitRange represents a Kubernetes LimitRange object.
type LimitRange struct {
	Namespace string           `json:"namespace"`
	Name      string           `json:"name"`
	Items     []LimitRangeItem `json:"items,omitempty"`
}
