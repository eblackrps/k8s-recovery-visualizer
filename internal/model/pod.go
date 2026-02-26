package model

type Pod struct {
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	UsesHostPath bool   `json:"usesHostPath"`
}
