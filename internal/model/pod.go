package model

type Pod struct {
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	UsesHostPath bool   `json:"usesHostPath"`

	// Round 11 — resource governance
	ContainerCount int  `json:"containerCount"`
	HasRequests    bool `json:"hasRequests"`   // every container defines CPU + memory requests
	HasLimits      bool `json:"hasLimits"`     // every container defines CPU + memory limits

	// Round 12 — pod security
	Privileged  bool `json:"privileged,omitempty"`  // any container runs privileged
	HostNetwork bool `json:"hostNetwork,omitempty"` // pod uses host network namespace
	HostPID     bool `json:"hostPid,omitempty"`     // pod shares host PID namespace

	// Round 18 — ServiceAccount token audit
	AutomountSAToken bool `json:"automountSaToken,omitempty"` // pod explicitly has automountServiceAccountToken=true
}
