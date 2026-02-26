package model

type PersistentVolume struct {
	Name          string `json:"name"`
	StorageClass  string `json:"storageClass,omitempty"`
	Capacity      string `json:"capacity,omitempty"`
	ReclaimPolicy string `json:"reclaimPolicy,omitempty"`
	Backend       string `json:"backend,omitempty"`
	ClaimRef      string `json:"claimRef,omitempty"`
}
