package model

type CategoryScore struct {
	Name     string  `json:"name"`
	Raw      float64 `json:"raw"`
	Weight   float64 `json:"weight"`
	Weighted float64 `json:"weighted"`
	Max      float64 `json:"max"`
	Grade    string  `json:"grade"`
}
