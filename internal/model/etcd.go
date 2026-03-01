package model

// EtcdBackupEvidence records how (and whether) etcd backup was detected.
type EtcdBackupEvidence struct {
	// Detected is true when at least one credible etcd-backup mechanism was found.
	Detected bool `json:"detected"`
	// Source describes where the evidence was found.
	// Possible values: "cronjob", "velero-cluster-backup", "provider-managed", "none"
	Source string `json:"source"`
	// Detail is a human-readable description of the evidence.
	Detail string `json:"detail,omitempty"`
}
