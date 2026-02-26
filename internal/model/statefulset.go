package model

type StatefulSet struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	Replicas       int32  `json:"replicas"`
	HasVolumeClaim bool   `json:"hasVolumeClaim"`
}
