package model

type PersistentVolumeClaim struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Namespace     string   `json:"namespace"`
	StorageClass  string   `json:"storageClass,omitempty"`
	AccessModes   []string `json:"accessModes,omitempty"`
	RequestedSize string   `json:"requestedSize,omitempty"`
}
